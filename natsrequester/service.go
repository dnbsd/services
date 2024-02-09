package natsrequester

import (
	"context"
	"github.com/dnbsd/services/natsconsumer"
	"github.com/dnbsd/services/natspublisher"
	"github.com/dnbsd/services/proto"
	"github.com/dnbsd/services/proto/adapters"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
	"log/slog"
)

type Arguments struct {
	Logger  *slog.Logger
	Conn    *nats.Conn
	Subject string
	Group   string
}

type OutputMessage struct {
	Subject string
	Data    []byte
}

type Service struct {
	args            Arguments
	consumer        *natsconsumer.Service
	consumerAdapter *adapters.Sink[proto.Event, proto.Request]
	responseAdapter *adapters.Sink[proto.Response, proto.Event]
	publisher       *natspublisher.Service[proto.Event]
}

func New(args Arguments) *Service {
	return &Service{
		args: args,
		consumer: natsconsumer.New(natsconsumer.Arguments{
			// TODO: logger
			Logger:  nil,
			Conn:    args.Conn,
			Subject: args.Subject,
			Group:   args.Group,
		}),
		consumerAdapter: adapters.NewSink[proto.Event, proto.Request](),
		responseAdapter: adapters.NewSink[proto.Response, proto.Event](),
		publisher: natspublisher.New[proto.Event](natspublisher.Arguments{
			// TODO: logger
			Logger: nil,
			Conn:   args.Conn,
		}),
	}
}

func (s *Service) Start(ctx context.Context) error {
	s.consumerAdapter.Connect(s.consumer, s.consumerAdapterConvertor)

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return s.consumer.Start(groupCtx)
	})

	group.Go(func() error {
		return s.consumerAdapter.Start(groupCtx)
	})

	group.Go(func() error {
		return s.responseAdapter.Start(groupCtx)
	})

	group.Go(func() error {
		return s.publisher.Start(groupCtx)
	})

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) Output() <-chan proto.Request {
	return s.consumerAdapter.Output()
}

func (s *Service) consumerAdapterConvertor(event proto.Event) proto.Request {
	// TODO: convert an event!
	return proto.Request{
		Body:      nil,
		RespondTo: s.responseAdapter,
	}
}
