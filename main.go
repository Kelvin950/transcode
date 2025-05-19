package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {

	outputDir := "outpush_dash"
	output := "output.m3u8"

	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {
		log.Fatal(err)
	}

	key := os.Getenv("key")
	bucket := os.Getenv("bucket")
	path := os.Getenv("path")

	fmt.Println(key, bucket)
	s3c := NewS3Client(cfg)

	err = s3c.DownloadContents(bucket, key)

	fmt.Println(
		"Downloaded file from S3:", key,
	)
	if err != nil {
		fmt.Println("Error running ffmpeg:", err)
		log.Fatal(err)
	}

	job := TranscodeJob{

		ffmpegPath: path,
		input:      key,
		output:     output,
		outputDir:  outputDir,
	}

	fmt.Println("Transcoding job created:", job)
	err = job.Run()

	if err != nil {
		fmt.Println("Error running ffmpeg:", err)
		log.Fatal(err, 3)
	}

	err = s3c.UploadContents("streamtestke", fmt.Sprintf("%d", time.Now().UnixMicro()))

	if err != nil {
		log.Fatal(err)
	}
}
