package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"golang.org/x/sync/errgroup"
)

type TranscodeJob struct {
	ffmpegPath string
	input      string
	output     string
	packager   string
}

func (t TranscodeJob) Run() error {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	// Define the 3 output configs
	profiles := []struct {
		name    string
		height  string
		profile string
		level   string
		bitrate string
		minrate string
		maxrate string
		bufsize string
	}{
		{"480p", "480", "main", "3.1", "1000k", "1000k", "1000k", "1000k"},
		{"720p", "720", "main", "4.0", "3000k", "3000k", "3000k", "3000k"},
		{"1080p", "1080", "high", "4.2", "6000k", "6000k", "6000k", "6000k"},
	}

	var paths = make([]string, len(profiles))

	for i, profile := range profiles {
		// capture loop variables
		i, profile := i, profile
		g.Go(func() error {
			outputFile := fmt.Sprintf("h264_%s_%s_%s.mp4", profile.profile, profile.name, profile.bitrate)

			cmd := exec.CommandContext(ctx, t.ffmpegPath,
				"-hwaccel", "cuvid",
				"-hwaccel_output_format", "cuda",
				"-sn",
				"-i", t.input,
				"-c:a", "copy",
				"-vf", fmt.Sprintf("scale_npp=-2:%s,hwdownload,format=nv12", profile.height),
				"-c:v", "h264_nvenc",
				"-profile:v", profile.profile,
				"-level:v", profile.level,
				"-x264-params", "scenecut=0:open_gop=0:min-keyint=72:keyint=72",
				"-minrate", profile.minrate,
				"-maxrate", profile.maxrate,
				"-bufsize", profile.bufsize,
				"-b:v", profile.bitrate,
				"-y", outputFile,
			)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Printf("Starting FFmpeg for %s\n", profile.name)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("ffmpeg failed for %s: %w", profile.name, err)
			}

			paths[i] = outputFile
			return nil
		})
	}

	// Wait for all jobs
	if err := g.Wait(); err != nil {
		return err
	}

	// Now build Shaka Packager command using actual generated files
	// Your actual files will be:
	// - h264_main_480p_1000k.mp4
	// - h264_main_720p_3000k.mp4
	// - h264_high_1080p_6000k.mp4

	var shakaInputs []string

	// Audio from highest quality file (1080p)
	var audioSource string
	for i, profile := range profiles {
		if profile.name == "1080p" {
			audioSource = paths[i]
			break
		}
	}

	// Add audio input
	shakaInputs = append(shakaInputs, fmt.Sprintf(
		"'in=%s,stream=audio,init_segment=audio/init.mp4,segment_template=audio/$Number$.m4s,playlist_name=audio.m3u8,hls_group_id=audio,hls_name=ENGLISH'",
		audioSource))

	// Add video inputs for each profile
	for i, profile := range profiles {
		outputFile := paths[i]

		// Basic video stream
		var videoInput string
		if profile.name == "720p" || profile.name == "1080p" {
			// Add iframe playlist for higher qualities
			videoInput = fmt.Sprintf(
				"'in=%s,stream=video,init_segment=h264_%s/init.mp4,segment_template=h264_%s/$Number$.m4s,playlist_name=h264_%s.m3u8,iframe_playlist_name=h264_%s_iframe.m3u8'",
				outputFile, profile.name, profile.name, profile.name, profile.name)
		} else {
			videoInput = fmt.Sprintf(
				"'in=%s,stream=video,init_segment=h264_%s/init.mp4,segment_template=h264_%s/$Number$.m4s,playlist_name=h264_%s.m3u8'",
				outputFile, profile.name, profile.name, profile.name)
		}

		shakaInputs = append(shakaInputs, videoInput)

		// Add trick-play tracks for 720p and 1080p
		if profile.name == "720p" || profile.name == "1080p" {
			trickPlayInput := fmt.Sprintf(
				"'in=%s,stream=video,init_segment=h264_%s_trick/init.mp4,segment_template=h264_%s_trick/$Number$.m4s,trick_play_factor=1'",
				outputFile, profile.name, profile.name)
			shakaInputs = append(shakaInputs, trickPlayInput)
		}
	}

	// Build final command
	cmdstr := fmt.Sprintf("%s %s --generate_static_live_mpd --mpd_output h264.mpd --hls_master_playlist_output h264_master.m3u8",
		t.packager,
		strings.Join(shakaInputs, " \\\n  "))

	fmt.Printf("Shaka Packager command:\n%s\n", cmdstr)
	fmt.Println("All transcodes complete.")
	fmt.Printf("Generated MP4 files: %v\n", paths)
	return nil
}
