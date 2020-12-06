package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
)

func main() {
	queueName := "demo-queue"
	region := "ap-southeast-1"
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      &region,
	}))

	svc := sqs.New(sess)

	queueURLResult, err := svc.GetQueueUrl(&sqs.GetQueueUrlInput{
		QueueName: &queueName,
	})

	if err != nil {
		log.Fatal(err)
	}

	queueURL := queueURLResult.QueueUrl

	go func() {
		for {
			msgResult, err := svc.ReceiveMessage(&sqs.ReceiveMessageInput{
				AttributeNames: []*string{
					aws.String(sqs.MessageSystemAttributeNameSentTimestamp),
				},
				MessageAttributeNames: []*string{
					aws.String(sqs.QueueAttributeNameAll),
				},
				QueueUrl:            queueURL,
				MaxNumberOfMessages: aws.Int64(1),
				VisibilityTimeout:   aws.Int64(30),
				WaitTimeSeconds:     aws.Int64(20),
			})

			if err != nil {
				log.Fatal(err)
			}

			if len(msgResult.Messages) < 1 {
				continue
			}

			for _, message := range msgResult.Messages {
				fmt.Println(message.String())
				svc.DeleteMessage(&sqs.DeleteMessageInput{
					QueueUrl:      queueURL,
					ReceiptHandle: message.ReceiptHandle,
				})
			}
		}
	}()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	<-signalChan

	fmt.Println("Shuting down ...")
	time.Sleep(time.Second * 3)
}
