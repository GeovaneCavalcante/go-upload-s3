package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

var (
	s3Client *s3.S3
	s3Bucket string
	wg       sync.WaitGroup
)

func init() {

	cfg := aws.Config{
		Region:   aws.String("us-east-1"),
		Endpoint: aws.String("http://172.26.2.44:4566"),
	}
	cfg.WithCredentials(credentials.AnonymousCredentials)
	sess, err := session.NewSession(&cfg)

	if err != nil {
		panic(err)
	}

	s3Client = s3.New(sess)
	s3Bucket = "go-expert"
}

func main() {
	dir, err := os.Open("./tmp")
	if err != nil {
		panic(err)
	}
	defer dir.Close()
	uploadControl := make(chan struct{}, 100)
	errorFileUpload := make(chan string)

	go func() {
		for file := range errorFileUpload {
			uploadControl <- struct{}{}
			wg.Add(1)
			go uploadFile(file, uploadControl, errorFileUpload)
		}
	}()

	for {
		files, err := dir.ReadDir(1)
		if err != nil {
			if err == io.EOF {
				break
			}
			continue
		}
		wg.Add(1)
		uploadControl <- struct{}{}
		go uploadFile(files[0].Name(), uploadControl, errorFileUpload)
	}
	wg.Wait()
}

func uploadFile(filename string, uploadControl chan struct{}, errorFileUpload chan string) {
	defer wg.Done()
	completeFileName := "./tmp/" + filename
	f, err := os.Open(completeFileName)
	if err != nil {
		fmt.Println(err)
		<-uploadControl
		return
	}
	defer f.Close()
	_, err = s3Client.PutObject(&s3.PutObjectInput{
		Bucket: aws.String(s3Bucket),
		Key:    aws.String(filename),
		Body:   f,
	})
	if err != nil {
		fmt.Println(err)
		<-uploadControl
		return
	}
	fmt.Printf("Successfully uploaded %q to %q\n", filename, s3Bucket)
	<-uploadControl
}
