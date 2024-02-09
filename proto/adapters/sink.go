package adapters

import (
	"context"
	"github.com/dnbsd/services/proto"
)

var (
// TODO: add interface tests
)

type Sink[IT, OT any] struct {
	source    proto.Producer[IT]
	inputCh   chan IT
	outputCh  chan OT
	convertor func(IT) OT
}

func NewSink[IT, OT any]() *Sink[IT, OT] {
	return &Sink[IT, OT]{
		inputCh:  make(chan IT),
		outputCh: make(chan OT),
	}
}

func (s *Sink[IT, OT]) Start(ctx context.Context) error {
	for {
		select {
		case inMsg := <-s.source.Output():
			select {
			case s.outputCh <- s.convertor(inMsg):
			case <-ctx.Done():
				return nil
			}

		case inMsg := <-s.inputCh:
			select {
			case s.outputCh <- s.convertor(inMsg):
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Sink[IT, OT]) Connect(source proto.Producer[IT], convertor func(IT) OT) {
	s.source = source
	s.convertor = convertor
}

func (s *Sink[IT, OT]) Input() chan<- IT {
	return s.inputCh
}

func (s *Sink[IT, OT]) Output() <-chan OT {
	return s.outputCh
}
