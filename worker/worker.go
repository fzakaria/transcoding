package worker

import (
	log "github.com/Sirupsen/logrus"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	sqs "github.com/aws/aws-sdk-go/service/sqs"
)

//Much of the base for this file was originally attributed to
//https://github.com/nabeken/golang-sqs-worker-example/blob/master/worker/worker.go

type Handler interface {
	HandleMessage(msg * string) error
}

type HandlerFunc func(msg * string) error

func (f HandlerFunc) HandleMessage(msg * string) error {
	return f(msg)
}

//An SQS worker for simplicity is tied to a specific queue
type Worker struct {
	QueueUrl string

	client *sqs.SQS

	Handler
}

//Return a default client for a given region and handler.
//The default SQS will lookup the necessary credentials in a variety of means
//such as ec2 instance metadata, environment variables etc..
func NewDefaultWorker(queueUrl, region string, f Handler) *Worker {
	return &Worker{
		queueUrl,
		sqs.New(session.New(&aws.Config{Region: aws.String(region)})),
		f}
}

func (w *Worker) Start() {
	params := &sqs.ReceiveMessageInput{
		QueueUrl: aws.String(w.QueueUrl), // Required
		AttributeNames: []*string{
			aws.String("All"), //include all diagnostic attributes
		},
		MaxNumberOfMessages: aws.Int64(10), //10 is the maximum
		MessageAttributeNames: []*string{
			aws.String("All"), // Required
		},
		VisibilityTimeout: aws.Int64(5), //The duration (in seconds) that the received messages are hidden from subsequent retrieve requests
		WaitTimeSeconds:   aws.Int64(1), //The duration (in seconds) for which the call will wait for a message to arrive in the queue before returning
	}

	resp, err := w.client.ReceiveMessage(params)
	if err != nil {
		awsErr := err.(awserr.Error)
		log.WithFields(log.Fields{
			"error":   err,
			"code":    awsErr.Code(),
			"message": awsErr.Message(),
		}).Error("Error occured receiving message.")
		return
	}

	log.WithFields(log.Fields{
			"resp":   resp,
	}).Debug("Polled response from sqs.")

	for _, message := range resp.Messages {
		//spawn a goroutine to handle each message concurrently
		go func(m *sqs.Message) {
			if err := w.HandleMessage(m.Body); err == nil {
				//only delete the message if everything was okay
				w.deleteMessage(m)
			}
		}(message)
	}
}

func (w *Worker) deleteMessage(m *sqs.Message) {
	params := &sqs.DeleteMessageInput{
		QueueUrl:      aws.String(w.QueueUrl), // Required
		ReceiptHandle: m.ReceiptHandle,        // Required
	}
	_, err := w.client.DeleteMessage(params)
	if err != nil {
		awsErr := err.(awserr.Error)
		log.WithFields(log.Fields{
			"error":   err,
			"code":    awsErr.Code(),
			"message": awsErr.Message(),
		}).Error("Error occured deleting message.")
	}
}
