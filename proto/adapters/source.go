package adapters

import (
	"context"
	"github.com/dnbsd/services/proto"
)

var (
	_ proto.Service         = &Source[proto.Request, proto.Event]{}
	_ proto.RequestProducer = &Source[proto.Request, proto.Request]{}
	_ proto.RequestProducer = &Source[proto.Event, proto.Request]{}
	_ proto.EventProducer   = &Source[proto.Request, proto.Event]{}
	_ proto.EventProducer   = &Source[proto.Event, proto.Event]{}
)

type Source[ST, OT proto.Request | proto.Event] struct {
	source    proto.Producer[ST]
	outputCh  chan OT
	convertor func(ST) OT
}

func NewSource[ST, OT proto.Request | proto.Event]() *Source[ST, OT] {
	return &Source[ST, OT]{
		outputCh: make(chan OT),
	}
}

func (s *Source[ST, OT]) Start(ctx context.Context) error {
	for {
		select {
		case inMsg := <-s.source.Output():
			outMsg := s.convertor(inMsg)

			select {
			case s.outputCh <- outMsg:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Source[ST, OT]) Connect(source proto.Producer[ST], convertor func(ST) OT) {
	s.source = source
	s.convertor = convertor
}

func (s *Source[ST, OT]) Output() <-chan OT {
	return s.outputCh
}
