package services

import (
	"context"
	"github.com/dnbsd/services/proto"
)

var _ proto.Service = &ServiceA{}
var _ proto.EventProducer = &ServiceA{}

type Arguments struct{}

type ServiceA struct {
	outputCh chan proto.Event
}

func (s *ServiceA) Start(ctx context.Context) error {
	for {
		s.outputCh <- proto.Event{
			Body: `{"request":"ok"}`,
		}
	}
}

func (s *ServiceA) Output() <-chan proto.Event {
	return s.outputCh
}
