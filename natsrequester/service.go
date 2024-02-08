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

type replyMessage struct {
	Subject string
	Data    []byte
}

type Service struct {
	args            Arguments
	consumer        *natsconsumer.Service
	consumerAdapter *adapters.Source[proto.Event, proto.Request]
	responseAdapter *adapters.Sink[proto.Response, proto.Event]
	publisher       *natspublisher.Service
}

func New(args Arguments) *Service {
	consumer := natsconsumer.New(natsconsumer.Arguments{
		// TODO: logger
		Conn:    args.Conn,
		Subject: args.Subject,
		Group:   args.Group,
	})
	consumerAdapter := adapters.NewSource(consumer, func(event proto.Event) proto.Request {
		body := event.Body.(natsconsumer.OutputMessage)
		return proto.Request{
			Body: OutputMessage{
				Subject: body.Subject,
				Data:    body.Data,
			},
			RespondTo: nil, // TODO
		}
	})
	publisher := natspublisher.New(natspublisher.Arguments{
		// TODO: logger
		Logger: nil,
		Conn:   args.Conn,
	})
	responseAdapter := adapters.NewSink(publisher, func(response proto.Response) proto.Event {
		if response.Error != nil {
			return proto.Event{
				Body: response.Error,
			}
		}
		return proto.Event{
			Body: response.Result
		}
	})

	return &Service{
		args:     args,
		consumer: consumer,
		outputCh: make(chan proto.Request),
	}
}

func (s *Service) Start(ctx context.Context) error {
	group, groupCtx := errgroup.WithContext(ctx)

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}

func (s *Service) consumerConvertor(event proto.Event) proto.Request {

}
