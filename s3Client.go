package main

import (
	"context"
	"fmt"
	

	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/smithy-go/logging"
	"golang.org/x/sync/errgroup"
)

type S3Client struct {
	S3         *s3.Client
	s3Download *manager.Downloader
	s3Upload   *manager.Uploader
}

func NewS3Client(cfg aws.Config) *S3Client {

	s3client := s3.NewFromConfig(cfg)
	downloadManager := manager.NewDownloader(s3client)
	uploadManager := manager.NewUploader(s3client)

	return &S3Client{
		S3:         s3client,
		s3Download: downloadManager,
		s3Upload:   uploadManager,
	}

}

func (s S3Client) DownloadContents(bucket, key string) error {

	fs, err := os.Create(key)

	if err != nil {
		return err
	}

	_, err = s.s3Download.Download(context.TODO(), fs, &s3.GetObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
	}, func(d *manager.Downloader) {
		d.Logger = logging.NewStandardLogger(os.Stderr)
		d.PartSize = 20 * 1024 * 1024 // 5MB
	})
	return err

}

func (s S3Client) UploadContents(bucket, key string)(string , string ,error) {

	queue := []string{
		"./encoded_output",
	}

	s3files := []string{}

	visited := make(map[string]bool)

	for len(queue) > 0 {

		dirc := queue[0]
		queue = queue[1:]

		visited[dirc] = true

		dir, err := os.ReadDir(dirc)

		if err != nil {
			return "" ,"",err
		}
		for _, dirOrFile := range dir {

			if !visited[filepath.Join(dirc, dirOrFile.Name())] {
				if dirOrFile.IsDir() {
					queue = append(queue, filepath.Join(dirc, dirOrFile.Name()))
					continue
				}

		
					s3files = append(s3files, filepath.Join(dirc, dirOrFile.Name()))

				
			}

		}

	}

	errG, ctx := errgroup.WithContext(context.Background())

	for _, file := range s3files {

		func(f string) {

			errG.Go(func() error {
				fmt.Println("Uploading file:", f)
				fs, err := os.Open(f)

				if err != nil {
					fmt.Println("Uploading fildde:", f)
					return err
				}
				defer fs.Close()
				output, err := s.s3Upload.Upload(ctx, &s3.PutObjectInput{
					Bucket: aws.String(bucket),
					Key:    aws.String(key + "/" + f),
					Body:   fs,
				}, func(u *manager.Uploader) {
					u.PartSize = 20 * 1024 * 1024 // 5MB
				})

				fmt.Println(output.Location)
				if err != nil {
					fmt.Println("Uploading fdildde:", f)
					return err
				}
				return nil
			})
		}(file)
	}

	return fmt.Sprintf("%s/%s/h264.mpd" ,key ,outputDir  ) ,  fmt.Sprintf("%s/%s/h264_master.m3u8" ,key ,outputDir  ), errG.Wait()
}