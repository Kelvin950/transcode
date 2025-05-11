package main

import (
	
	"context"
	"os"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/feature/s3/manager"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type S3Client struct{

 S3  *s3.Client
 s3Download *manager.Downloader 
 s3Upload  *manager.Uploader

}


func NewS3Client(cfg aws.Config)*S3Client{

 
	s3client := s3.NewFromConfig(cfg); 
    downloadManager:= manager.NewDownloader(s3client) 
	uploadManager:= manager.NewUploader(s3client) 

	return  &S3Client{
		S3:s3client,
		s3Download :downloadManager ,
		s3Upload : uploadManager,
	}

}


func(s S3Client) DownloadContents(bucket , key string)(error){
     

	 fs , err:= os.Create(key)

	 if err!=nil{
		return err
	 }
  
	_ ,err = s.s3Download.Download(context.TODO() , fs , &s3.GetObjectInput{
		Bucket:aws.String(bucket),
		Key: aws.String(key),
	})
	return err

}


func(s S3Client)UploadContents(bucket , key string){
  

	


}