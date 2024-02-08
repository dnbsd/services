package proto

import "context"

type Service interface {
	Start(context.Context) error
}

type Producer[MT Request | Event] interface {
	Output() <-chan MT
}

type EventProducer interface {
	Producer[Event]
}

type RequestProducer interface {
	Producer[Request]
}

// FIXME: using MT any because of recursive type error!
type Consumer[MT any] interface {
	Input() chan<- MT
}

type EventConsumer interface {
	Consumer[Event]
}

type RequestConsumer interface {
	Consumer[Request]
}

type ResponseConsumer interface {
	Consumer[Response]
}
