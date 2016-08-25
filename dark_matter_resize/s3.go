package main

import (
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func GetOriginFile(fileName string) error {
	downloadFile := fmt.Sprintf("%s/%s", downloadPath, fileName)
	fs, err := os.Create(downloadFile)
	if err != nil {
		logLn(fmt.Sprintf("origin create err / %s", err.Error()))
		return err
	}
	defer fs.Close()

	creds := credentials.NewStaticCredentials("aws_access_key", "aws_access_key_secret", "")

	downloader := s3manager.NewDownloader(session.New(&aws.Config{Credentials: creds, Region: aws.String("region")}))
	key := fmt.Sprintf("files/%s", fileName)
	_, err = downloader.Download(fs,
		&s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    aws.String(key),
		})
	if err != nil {
		logLn(fmt.Sprintf("origin download err / %s", err.Error()))
		return err
	}

	return nil
}

func UploadThumbFile(fileName string) error {
	thumbFile := fmt.Sprintf("%s/%s%s", downloadPath, fileName, sufix)
	file, err := os.Open(thumbFile)
	if err != nil {
		return err
	}
	defer file.Close()
	creds := credentials.NewStaticCredentials("aws_access_key", "aws_access_key_secret", "")
	newSession := session.New(&aws.Config{Credentials: creds, Region: aws.String("region")})
	s3Client := s3.New(newSession)
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fmt.Sprintf("files/%s%s", fileName, sufix)),
		Body:   file,
	})

	if err != nil {
		return err
	}

	return nil
}
