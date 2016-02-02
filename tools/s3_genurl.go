package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"fmt"
	"time"
)

func main() {
	url, _ := GenSignedURL("slinger-test", "output.mp4")
	fmt.Println(url)
}


func GenSignedURL(bucket, key string) (string, error) {
	svc := s3.New(session.New(&aws.Config{Region: aws.String("us-east-1")}))
    req, _ := svc.PutObjectRequest(&s3.PutObjectInput{
        Bucket: aws.String(bucket),
        Key:    aws.String(key),
    })

    URL, err := req.Presign(300 * time.Second)
    if err != nil {
        return "", err
    }
    return URL, nil
}