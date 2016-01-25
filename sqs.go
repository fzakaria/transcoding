package main

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

//Much of the base for this file was originally attributed to
//https://github.com/nabeken/golang-sqs-worker-example/blob/master/worker/worker.go

type Handler interface {
	HandleMessage(msg *sqs.Message) error
}

type HandlerFunc func(msg *sqs.Message) error

func (f HandlerFunc) HandleMessage(msg *sqs.Message) error {
	return f(msg)
}

func NewDefaultClient(region string) *sqs.SQS {
	return sqs.New(session.New(&aws.Config{Region: aws.String(region)}))
}
