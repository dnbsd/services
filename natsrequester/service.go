package natsrequester

import (
	"context"
	"github.com/dnbsd/services/natsconsumer"
	"github.com/dnbsd/services/natspublisher"
	"github.com/dnbsd/services/natsrequester/internal/services/replyhandler"
	"github.com/dnbsd/services/natsrequester/internal/services/requesthandler"
	"github.com/dnbsd/services/proto"
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
	Subject      string
	Data         []byte
	replySubject string
	replyTo      proto.Consumer[replyMessage]
}

func (m OutputMessage) Reply(ctx context.Context, data []byte) error {
	message := replyMessage{
		Subject: m.replySubject,
		Data:    data,
	}
	select {
	case m.replyTo.Input() <- message:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

type replyMessage struct {
	Subject string
	Data    []byte
}

type Service struct {
	args           Arguments
	consumer       *natsconsumer.Service
	requestHandler *requesthandler.Service[natsconsumer.OutputMessage, OutputMessage]
	replyHandler   *replyhandler.Service[replyMessage]
	publisher      *natspublisher.Service[replyMessage]
}

func New(args Arguments) *Service {
	logger := args.Logger
	return &Service{
		args: args,
		consumer: natsconsumer.New(natsconsumer.Arguments{
			Logger:  logger.With("component", "natsconsumer"),
			Conn:    args.Conn,
			Subject: args.Subject,
			Group:   args.Group,
		}),
		requestHandler: requesthandler.New[natsconsumer.OutputMessage, OutputMessage](requesthandler.Arguments{
			Logger: logger.With("component", "requesthandler"),
		}),
		replyHandler: replyhandler.New[replyMessage](replyhandler.Arguments{
			Logger: logger.With("component", "replyhandler"),
		}),
		publisher: natspublisher.New[replyMessage](natspublisher.Arguments{
			Logger: logger.With("component", "natspublisher"),
			Conn:   args.Conn,
		}),
	}
}

func (s *Service) Start(ctx context.Context) error {
	logger := s.args.Logger
	logger.Info("started")
	defer logger.Info("stopped")

	defer func() {
		err := recover()
		if err != nil {
			logger.Error("recovered from panic", "error", err)
			return
		}
	}()

	s.requestHandler.Connect(s.consumer, s.consumerConvertor)
	s.publisher.Connect(s.replyHandler, s.replyConvertor)

	group, groupCtx := errgroup.WithContext(ctx)

	group.Go(func() error {
		return s.consumer.Start(groupCtx)
	})

	group.Go(func() error {
		return s.requestHandler.Start(groupCtx)
	})

	group.Go(func() error {
		return s.replyHandler.Start(groupCtx)
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

func (s *Service) Output() <-chan OutputMessage {
	return s.requestHandler.Output()
}

func (s *Service) consumerConvertor(message natsconsumer.OutputMessage) OutputMessage {
	return OutputMessage{
		Subject:      message.Subject,
		Data:         message.Data,
		replySubject: message.ReplySubject,
		replyTo:      s.replyHandler,
	}
}

func (s *Service) replyConvertor(message replyMessage) natspublisher.InputMessage {
	return natspublisher.InputMessage{
		Subject: message.Subject,
		Data:    message.Data,
	}
}
