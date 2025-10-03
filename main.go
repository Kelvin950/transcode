package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	"os"

	"time"

	"github.com/ThreeDotsLabs/watermill"
	"github.com/ThreeDotsLabs/watermill/message"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sqs"
	Sqssubpub "github.com/kelvin950/sqspubsub"
)

var logger = watermill.NewStdLogger(false, false)

func main() {

	// outputDir := "outpush_dash"
	output := "output.m3u8"

	key := os.Getenv("key")
	bucket := os.Getenv("bucket")
	path := os.Getenv("path")
	taskid := os.Getenv("taskid")
	startTime := os.Getenv("timestarted")
	packagerpath := os.Getenv("path1")

	// tableName := "Task_State"

	cfg, err := config.LoadDefaultConfig(context.TODO())

	if err != nil {

		log.Fatal(err)
	}

	pub, err := Sqssubpub.NewSqsSub(&cfg, logger, func(o *sqs.Options) {

		
	})

	if err != nil {
		log.Fatal(err)
	}

	err = pub.CreatePub()

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(key, bucket)
	s3c := NewS3Client(cfg)

	err = s3c.DownloadContents(bucket, key)

	fmt.Println(
		"Downloaded file from S3:", key,
	)
	if err != nil {

		// dynamoCl.PutITem(Ec2TaskState{
		// 	TaskID: taskid,

		// 	State:      "failed",
		// 	StartedAt:  startTime,
		// 	FinishedAt: time.Now(),
		// 	ErrMsg:     err.Error(),

		x := map[string]interface{}{
			"taskid": taskid,

			"state":      "failed",
			"startedAt":  startTime,
			"finishedAt": time.Now(),
			"contentid":  79,
			"errmsg":     err.Error(),
		}

		p, _ := json.Marshal(&x)

		pub.Publisher.Publish("Transcode_job", message.NewMessage(watermill.NewUUID(), p))
		// })
		fmt.Println("Error running ffmpeg:", err)
		log.Fatal(err)
	}

	job := TranscodeJob{

		ffmpegPath: path,
		input:      key,
		output:     output,
		packager:   packagerpath,
	}

	fmt.Println("Transcoding job created:", job)
	err = job.Run()

	if err != nil {

		// dynamoCl.PutITem(Ec2TaskState{
		// 	TaskID: taskid,

		// 	State:      "failed",
		// 	StartedAt:  startTime,
		// 	FinishedAt: time.Now(),
		// 	ErrMsg:     err.Error(),
		// })

		x := map[string]interface{}{
			"taskid": taskid,

			"state":      "failed",
			"startedAt":  startTime,
			"finishedAt": time.Now(),
			"contentid":  79,
			"errmsg":     err.Error(),
		}

		p, _ := json.Marshal(&x)

		pub.Publisher.Publish("Transcode_job", message.NewMessage(watermill.NewUUID(), p))

		fmt.Println("Error running ffmpeg:", err)
		log.Fatal(err, 3)
	}

	if err != nil {
		log.Fatal(err)
	}

	dash, hls, err := s3c.UploadContents("streamtestke", time.Now().Format(time.RFC1123))

	if err != nil {

		// dynamoCl.PutITem(Ec2TaskState{
		// 	TaskID: taskid,

		// 	State:      "failed",
		// 	StartedAt:  startTime,
		// 	FinishedAt: time.Now(),
		// 	ErrMsg:     err.Error(),
		// })

		x := map[string]interface{}{
			"taskid": taskid,

			"state":      "failed",
			"startedAt":  startTime,
			"finishedAt": time.Now(),
			"contentid":  79,
			"errmsg":     err.Error(),
		}

		p, _ := json.Marshal(&x)

		pub.Publisher.Publish("Transcode_job", message.NewMessage(watermill.NewUUID(), p))
		log.Fatal(err)
	}

	x := map[string]interface{}{
		"taskid": taskid,

		"state":        "failed",
		"startedAt":    startTime,
		"finishedAt":   time.Now(),
		"contentid":    79,
		"manifest_url": fmt.Sprintf("%s:%s", dash, hls),
	}

	p, _ := json.Marshal(&x)

	err = pub.Publisher.Publish("Transcode_job", message.NewMessage(watermill.NewUUID(), p))

	if err != nil {
		log.Println(err.Error())
	}

}
