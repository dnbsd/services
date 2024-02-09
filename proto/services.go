package proto

import "context"

type Service interface {
	Start(context.Context) error
}

type Producer[MT any] interface {
	Output() <-chan MT
}

type Consumer[MT any] interface {
	Input() chan<- MT
}

type Adapter[ST, IT any] interface {
	Connect(Producer[ST], func(ST) IT)
}
