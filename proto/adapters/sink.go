package adapters

import (
	"context"
	"github.com/dnbsd/services/proto"
)

type Sink[IT, DT proto.Request | proto.Response | proto.Event] struct {
	inputCh     chan IT
	destination proto.Consumer[DT]
	convertor   func(IT) DT
}

func NewSink[IT, DT proto.Request | proto.Response | proto.Event](destination proto.Consumer[DT], convertor func(IT) DT) *Sink[IT, DT] {
	return &Sink[IT, DT]{
		inputCh:     make(chan IT),
		destination: destination,
		convertor:   convertor,
	}
}

func (s *Sink[IT, DT]) Start(ctx context.Context) error {
	for {
		select {
		case inMsg := <-s.inputCh:
			outMsg := s.convertor(inMsg)

			select {
			case s.destination.Input() <- outMsg:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Sink[IT, DT]) Input() chan<- IT {
	return s.inputCh
}
