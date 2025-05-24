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

	key := os.Getenv("key")
	bucket := os.Getenv("bucket")
	path := os.Getenv("path")
	taskid := os.Getenv("taskid")
	startTime := os.Getenv("timestarted")

	tableName := "Task_State"

	cfg, err := config.LoadDefaultConfig(context.TODO())

	dynamoCl := NewDynamoClient(cfg, tableName)
	if err != nil {

		log.Fatal(err)
	}

	fmt.Println(key, bucket)
	s3c := NewS3Client(cfg)

	if err != nil {
		log.Fatal(err)

	}

	err = s3c.DownloadContents(bucket, key)

	fmt.Println(
		"Downloaded file from S3:", key,
	)
	if err != nil {

		dynamoCl.PutITem(Ec2TaskState{
			TaskID: taskid,

			State:      "failed",
			StartedAt:  startTime,
			FinishedAt: time.Now(),
			ErrMsg:     err.Error(),
		})
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

		dynamoCl.PutITem(Ec2TaskState{
			TaskID: taskid,

			State:      "failed",
			StartedAt:  startTime,
			FinishedAt: time.Now(),
			ErrMsg:     err.Error(),
		})

		fmt.Println("Error running ffmpeg:", err)
		log.Fatal(err, 3)
	}

	err = s3c.UploadContents("streamtestke", time.Now().Format(time.RFC1123))

	if err != nil {

		dynamoCl.PutITem(Ec2TaskState{
			TaskID: taskid,

			State:      "failed",
			StartedAt:  startTime,
			FinishedAt: time.Now(),
			ErrMsg:     err.Error(),
		})
		log.Fatal(err)
	}
	_, err = dynamoCl.PutITem(Ec2TaskState{
		TaskID: taskid,

		StartedAt:  startTime,
		State:      "finished",
		FinishedAt: time.Now(),
	})

	if err != nil {
		log.Fatal(err)
	}

}
