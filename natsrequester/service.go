package natsrequester

import (
	"context"
	"github.com/dnbsd/services/convertor"
	"github.com/dnbsd/services/natsconsumer"
	"github.com/dnbsd/services/natspublisher"
	"github.com/nats-io/nats.go"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	Subject  string
	Group    string
	OutputCh chan<- OutputMessage
}

type OutputMessage struct {
	Subject string
	Data    []byte
	ReplyFn func(context.Context, []byte) error
}

type replyMessage struct {
	Subject string
	Data    []byte
}

type Service struct {
	//logger *slog.Logger
	nc *nats.Conn
}

func New(nc *nats.Conn) *Service {
	return &Service{
		nc: nc,
	}
}

func (s *Service) Start(ctx context.Context, args Arguments) error {
	var (
		natsConsumerOutCh = make(chan natsconsumer.OutputMessage, 1024)
		replyCh           = make(chan replyMessage, 1024)
		natsPublisherInCh = make(chan natspublisher.InputMessage, 1024)
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return natsconsumer.New(
			s.nc,
		).Start(
			groupCtx,
			natsconsumer.Arguments{
				Subject:  args.Subject,
				Group:    args.Group,
				OutputCh: natsConsumerOutCh,
			},
		)
	})

	group.Go(func() error {
		return convertor.New[
			natsconsumer.OutputMessage,
			OutputMessage,
		](
			natsConsumerOutCh,
			args.OutputCh,
			func(msg natsconsumer.OutputMessage) OutputMessage {
				return OutputMessage{
					Subject: msg.Subject,
					Data:    msg.Data,
					ReplyFn: func(ctx context.Context, b []byte) error {
						select {
						case replyCh <- replyMessage{
							Subject: msg.ReplySubject,
							Data:    b,
						}:
						case <-ctx.Done():
							return ctx.Err()
						}
						return nil
					},
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		return convertor.New[
			replyMessage,
			natspublisher.InputMessage,
		](
			replyCh,
			natsPublisherInCh,
			func(msg replyMessage) natspublisher.InputMessage {
				return natspublisher.InputMessage{
					Subject: msg.Subject,
					Data:    msg.Data,
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		args := natspublisher.Arguments{
			InputCh: natsPublisherInCh,
		}
		return natspublisher.New(
			s.nc,
		).Start(
			groupCtx,
			args,
		)
	})

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}
