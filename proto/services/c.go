package services

import (
	"context"
	"github.com/dnbsd/services/proto"
	"github.com/dnbsd/services/proto/adapters"
	"golang.org/x/sync/errgroup"
)

var _ proto.Service = &ServiceA{}
var _ proto.EventProducer = &ServiceA{}

type ServiceC struct {
	a  *ServiceA
	ca *adapters.SourceAdapter[proto.Event, proto.Request]
	cb *adapters.DestinationAdapter[proto.Response, proto.Event]
	b  *ServiceB
}

func NewServiceC() *ServiceC {
	a := &ServiceA{
		outputCh: make(chan proto.Event),
	}
	b := &ServiceB{
		inputCh: make(chan proto.Event),
	}
	cb := adapters.NewDestinationAdapter[proto.Response, proto.Event](b, func(response proto.Response) proto.Event {
		return proto.Event{
			Params: response.Result,
		}
	})
	ca := adapters.NewSourceAdapter[proto.Event, proto.Request](a, func(event proto.Event) proto.Request {
		return proto.Request{
			Params:    event.Params,
			RespondTo: cb,
		}
	})
	return &ServiceC{
		a:  a,
		ca: ca,
		cb: cb,
		b:  b,
	}
}

func (s *ServiceC) Start(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return s.a.Start(groupCtx)
	})
	group.Go(func() error {
		return s.ca.Start(groupCtx)
	})
	group.Go(func() error {
		return s.cb.Start(groupCtx)
	})
	group.Go(func() error {
		return s.b.Start(groupCtx)
	})

	return group.Wait()
}

func (s *ServiceC) Output() <-chan proto.Request {
	return s.ca.Output()
}
