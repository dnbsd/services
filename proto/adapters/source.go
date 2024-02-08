package adapters

import (
	"context"
	"github.com/dnbsd/services/proto"
)

type Source[ST, OT proto.Request | proto.Event] struct {
	source    proto.Producer[ST]
	outputCh  chan OT
	convertor func(ST) OT
}

func NewSource[ST, OT proto.Request | proto.Event](source proto.Producer[ST], convertor func(ST) OT) *Source[ST, OT] {
	return &Source[ST, OT]{
		source:    source,
		outputCh:  make(chan OT),
		convertor: convertor,
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

func (s *Source[ST, OT]) Output() <-chan OT {
	return s.outputCh
}
