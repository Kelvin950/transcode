package main

 import (
	"os"
	"os/exec"
)

type TranscodeJob struct {
	ffmpegPath string
	input      string
	output     string
}

func (t TranscodeJob) Run() error {
	width := "1920"
height := "1080"
maxBitrate := "5000k"
hrdBufferSize := "10000k"
gopSize := "48"
numBFrames := "3"
numRefFrames := "4"
framerate := "30/1"
audioBitrate := "128k"
sampleRate := "48000"
segmentLength := "180" // <-- 3 minutes per segment


	cmd := exec.Command(
		t.ffmpegPath,
		"-sn",
		"-i", t.input,
		"-vf", "scale="+width+":"+height,
		"-c:v", "libx264",
		"-profile:v", "high",
		"-level", "4.0",
		"-preset", "medium",
		"-b:v", maxBitrate,
		"-maxrate", maxBitrate,
		"-bufsize", hrdBufferSize,
		"-g", gopSize,
		"-sc_threshold", "0",
		"-bf", numBFrames,
		"-refs", numRefFrames,
		"-pix_fmt", "yuv420p",
		"-x264opts", "b-pyramid=1:weightb=1:scenecut=40:open_gop=0:cabac=1",
		"-r", framerate,
		"-c:a", "aac",
		"-b:a", audioBitrate,
		"-ac", "2",
		"-ar", sampleRate,
		"-strict", "-2",
		"-hls_time", segmentLength,
		"-hls_segment_filename", t.output+"/output_%03d.ts",
		"-hls_playlist_type", "vod",
		"-hls_flags", "independent_segments",
		"-f", "hls",
		t.output,
	)

	// Optional: Pipe ffmpeg output to stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {

		return err
	}

	return nil
}
