package adapters

import (
	"context"
	"github.com/dnbsd/services/proto"
)

type Bridge[ST proto.Request | proto.Event, DT proto.Request | proto.Response | proto.Event] struct {
	source      proto.Producer[ST]
	destination proto.Consumer[DT]
	convertor   func(ST) DT
}

// TODO: NewBridge

func (s *Bridge[ST, DT]) Start(ctx context.Context) error {
	for {
		select {
		case inMsg := <-s.source.Output():
			outMsg := s.convertor(inMsg)
			select {
			case s.destination.Input() <- outMsg:
				// TODO: ctx!
			}
			// TODO: ctx!
		}
	}
}
