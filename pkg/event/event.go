package event

// MessageHandler is an abstraction of message handler. Will be excecute when the message is received.
type MessageHandler interface {
	Handle(interface{}) error
}

// Receiver is message receiver (could be a consumer or subscriber).
type Receiver interface {
	Receive()
}

// Sender is a message sender (could be a produce or publisher).
type Sender interface {
	Send(topic, key string, headers map[string]string, body []byte)
}
