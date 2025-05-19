package main

import (
	"fmt"
	"os"
	"os/exec"
)

type TranscodeJob struct {
	ffmpegPath string
	input      string
	output     string
	outputDir  string
}

func (t TranscodeJob) Run() error {
	cmdString := fmt.Sprintf(`%s -i %s \
-filter_complex "[0:v]split=4[v1][v2][v3][v4]; \
[v1]scale=w=1920:h=1080[v1out]; \
[v2]scale=w=1280:h=720[v2out]; \
[v3]scale=w=1280:h=720[v3out_30fps]; \
[v3out_30fps]fps=fps=30[v3scaled_30fps]; \
[v4]scale=w=854:h=480[v4out]; \
[0:a]asplit=4[a1][a2][a3][a4]" \
-map "[v1out]" -c:v:0 libx264 -s:v:0 1920x1080 -r:v:0 60 -b:v:0 6500k -preset veryfast \
-map "[v2out]" -c:v:1 libx264 -s:v:1 1280x720 -r:v:1 60 -b:v:1 4000k -preset veryfast \
-map "[v3scaled_30fps]" -c:v:2 libx264 -s:v:2 1280x720 -r:v:2 30 -b:v:2 2500k -preset veryfast \
-map "[v4out]" -c:v:3 libx264 -s:v:3 854x480 -r:v:3 30 -b:v:3 1500k -preset veryfast \
-map "[a1]" -c:a:0 aac -b:a:0 128k -ac:a:0 2 \
-map "[a2]" -c:a:1 aac -b:a:1 128k -ac:a:1 2 \
-map "[a3]" -c:a:2 aac -b:a:2 64k  -ac:a:2 2 \
-map "[a4]" -c:a:3 aac -b:a:3 64k  -ac:a:3 2 \
-f hls \
-hls_time 5 \
-hls_playlist_type vod \
-hls_flags independent_segments \
-hls_segment_type mpegts \
-hls_segment_filename "stream_%%v/data%%03d.ts" \
-master_pl_name master.m3u8 \
-var_stream_map "v:0,a:0 v:1,a:1 v:2,a:2 v:3,a:3" \
"stream_%%v/playlist.m3u8"`,
		t.ffmpegPath, t.input)

	fmt.Println(cmdString)
	// Split the command string into arguments.

	// Create the command.
	cmd := exec.Command("sh", "-c", cmdString)

	// Optional: Pipe ffmpeg output to stdout/stderr
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	// Run the command
	if err := cmd.Run(); err != nil {

		return err
	}

	return nil
}
