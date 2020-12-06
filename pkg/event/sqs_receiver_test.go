package event_test

import (
	"fmt"
	"go-sqs-receiver-demo/pkg/event"
	"go-sqs-receiver-demo/pkg/event/mocks"
	"os"
	"testing"

	"github.com/aws/aws-sdk-go/service/sqs"

	"github.com/stretchr/testify/mock"

	"github.com/sirupsen/logrus"
)

type signalMock struct{}

func (signalMock) String() string { return "signal" }
func (signalMock) Signal()        { return }

func getEmptySignalChan() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	return signalChan
}

func getFilledSignalChan() chan os.Signal {
	signalChan := make(chan os.Signal, 1)
	signalChan <- signalMock{}
	return signalChan
}

var logger = logrus.New()

func TestSQSReceiverAdapterReceive_ErroWhenGettingURL(t *testing.T) {
	signalChan := getEmptySignalChan()

	sqsClient := &mocks.SQSClient{}
	sqsClient.On("GetQueueUrl", mock.AnythingOfType("*sqs.GetQueueUrlInput")).Return(nil, fmt.Errorf("URL Error"))

	receiver := event.NewSQSReceiverAdapter(
		signalChan, logger, sqsClient, "test-queue", 20, 30, nil,
	)

	receiver.Receive()

	sqsClient.AssertExpectations(t)
}

func TestSQSReceiverAdapterReceive_StopByInterruption(t *testing.T) {
	signalChan := getFilledSignalChan()
	queueURLMock := "http://test.co.id"
	sqsClient := &mocks.SQSClient{}
	sqsClient.On("GetQueueUrl", mock.AnythingOfType("*sqs.GetQueueUrlInput")).Return(&sqs.GetQueueUrlOutput{
		QueueUrl: &queueURLMock,
	}, nil)

	receiver := event.NewSQSReceiverAdapter(
		signalChan, logger, sqsClient, "test-queue", 20, 30, nil,
	)

	receiver.Receive()

	sqsClient.AssertExpectations(t)
}
