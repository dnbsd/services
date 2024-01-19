package adapters

import (
	"context"
	"github.com/dnbsd/services/proto"
)

type SourceAdapter[ST, OT proto.Request | proto.Event] struct {
	source    proto.Producer[ST]
	outputCh  chan OT
	convertor func(ST) OT
}

func NewSourceAdapter[ST, OT proto.Request | proto.Event](source proto.Producer[ST], convertor func(ST) OT) *SourceAdapter[ST, OT] {
	return &SourceAdapter[ST, OT]{
		source:    source,
		outputCh:  make(chan OT),
		convertor: convertor,
	}
}

func (s *SourceAdapter[ST, OT]) Start(ctx context.Context) error {
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

func (s *SourceAdapter[ST, OT]) Output() <-chan OT {
	return s.outputCh
}

type DestinationAdapter[IT, DT proto.Request | proto.Response | proto.Event] struct {
	inputCh     chan IT
	destination proto.Consumer[DT]
	convertor   func(IT) DT
}

func NewDestinationAdapter[IT, DT proto.Request | proto.Response | proto.Event](destination proto.Consumer[DT], convertor func(IT) DT) *DestinationAdapter[IT, DT] {
	return &DestinationAdapter[IT, DT]{
		inputCh:     make(chan IT),
		destination: destination,
		convertor:   convertor,
	}
}

func (s *DestinationAdapter[IT, DT]) Start(ctx context.Context) error {
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

func (s *DestinationAdapter[IT, DT]) Input() chan<- IT {
	return s.inputCh
}
