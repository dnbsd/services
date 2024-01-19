package services

import (
	"context"
	"fmt"
	"github.com/dnbsd/services/proto"
)

var _ proto.Service = &ServiceB{}
var _ proto.EventConsumer = &ServiceB{}

type ServiceB struct {
	inputCh chan proto.Event
}

func (s *ServiceB) Start(ctx context.Context) error {
	for {
		inMsg := <-s.inputCh
		fmt.Printf("received a response %v\n", inMsg)
	}
}

func (s *ServiceB) Input() chan<- proto.Event {
	return s.inputCh
}
