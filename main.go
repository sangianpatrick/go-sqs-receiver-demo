package main

import (
	"fmt"
	"go-sqs-receiver-demo/pkg/event"
	"os"
	"os/signal"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sqs"
	"github.com/sirupsen/logrus"
)

type myEventHandler struct {
	logger *logrus.Logger
}

func (h myEventHandler) Handle(data interface{}) error {
	message, ok := data.(*sqs.Message)
	if !ok {
		err := fmt.Errorf("Not an expected message")
		h.logger.Error(err)
		return err
	}

	fmt.Printf("Incoming Message: %s", *message.Body)
	return nil
}

func main() {
	queueName := "first-queue"
	region := "ap-southeast-1"
	sess := session.Must(session.NewSession(&aws.Config{
		Credentials: credentials.NewEnvCredentials(),
		Region:      &region,
	}))

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt, os.Kill)

	sqsClient := sqs.New(sess)
	receiver := event.NewSQSReceiverAdapter(
		signalChan,
		sqsClient,
		queueName,
		20,
		30,
		new(myEventHandler),
	)

	receiver.Receive()
}
