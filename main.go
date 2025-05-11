package main

import (
	"context"
	"log"
	"os"

	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/config"
)

func main() {

	outputDir := "outpush_dash"
	output := filepath.Join(outputDir ,  "output.m3u8")
	 

	 cfg , err:= config.LoadDefaultConfig(context.TODO())

	 if err!=nil{
		log.Fatal(err) 
 	 }
 
	 key:= os.Getenv("key")
	 bucket := os.Getenv("bucket")
	 s3c :=  NewS3Client(cfg) 

	 err = s3c.DownloadContents(bucket , key)
	 
	 if err!=nil{
		log.Fatal(err) 
 	 }


	job := TranscodeJob{

		ffmpegPath: "/snap/bin/ffmpeg",
		input:  key,
		output: output , 
	}

	err = job.Run()

	if err!=nil{
		log.Fatal(err)
	}
}