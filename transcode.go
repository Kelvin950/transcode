package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sync"

	"golang.org/x/sync/errgroup"
)

type TranscodeJob struct {
	ffmpegPath string
	input      string
	output     string
}

func (t TranscodeJob) Run() error {
	ctx := context.Background()
	g, ctx := errgroup.WithContext(ctx)

	// Define the 4 output configs
	profiles := []struct {
		name      string
		res       string
		bitrate   string
		framerate string
		audioBr   string
	}{
		{"1080p", "1920x1080", "6500k", "60", "128k"},
		{"720p60", "1280x720", "4000k", "60", "128k"},
		{"720p30", "1280x720", "2500k", "30", "64k"},
		{"480p", "854x480", "1500k", "30", "64k"},
	}

	var mu sync.Mutex
	var paths []string

	for _, p := range profiles {
		profile := p // capture loop variable
		g.Go(func() error {
			outputDir := filepath.Join(".", "stream_"+profile.name)
			if err := os.MkdirAll(outputDir, 0755); err != nil {
				return err
			}

			cmd := exec.CommandContext(ctx, t.ffmpegPath,
				"-i", t.input,
				"-c:v", "libx264",
				"-s", profile.res,
				"-r", profile.framerate,
				"-b:v", profile.bitrate,
				"-preset", "ultrafast",
				"-c:a", "aac",
				"-b:a", profile.audioBr,
				"-ac", "2",
				"-f", "hls",
				"-hls_time", "5",
				"-hls_playlist_type", "vod",
				"-hls_flags", "independent_segments",
				"-hls_segment_type", "mpegts",
				"-hls_segment_filename", filepath.Join(outputDir, "data%03d.ts"),
				filepath.Join(outputDir, "playlist.m3u8"),
			)

			cmd.Stdout = os.Stdout
			cmd.Stderr = os.Stderr

			fmt.Printf("Starting FFmpeg for %s\n", profile.name)
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("ffmpeg failed for %s: %w", profile.name, err)
			}

			mu.Lock()
			paths = append(paths, profile.name)
			mu.Unlock()

			return nil
		})
	}

	// Wait for all jobs
	if err := g.Wait(); err != nil {
		return err
	}

	// Generate master.m3u8
	if err := generateMasterPlaylist("."); err != nil {
		return fmt.Errorf("failed to generate master playlist: %w", err)
	}

	fmt.Println("All transcodes complete.")
	return nil
}

func generateMasterPlaylist(outputDir string) error {
	content := `#EXTM3U
#EXT-X-VERSION:3

#EXT-X-STREAM-INF:BANDWIDTH=6800000,RESOLUTION=1920x1080,FRAME-RATE=60
stream_1080p/playlist.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=4300000,RESOLUTION=1280x720,FRAME-RATE=60
stream_720p60/playlist.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=2700000,RESOLUTION=1280x720,FRAME-RATE=30
stream_720p30/playlist.m3u8

#EXT-X-STREAM-INF:BANDWIDTH=1700000,RESOLUTION=854x480,FRAME-RATE=30
stream_480p/playlist.m3u8
`
	masterPath := filepath.Join(outputDir, "master.m3u8")
	return os.WriteFile(masterPath, []byte(content), 0644)
}
