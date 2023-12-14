package natsrpc

import (
	"context"
	"github.com/dnbsd/services/convertor"
	"github.com/dnbsd/services/natsconsumer"
	"github.com/dnbsd/services/natspublisher"
	"github.com/dnbsd/services/natsrpc/services/replyhandler"
	"github.com/dnbsd/services/natsrpc/services/requesthandler"
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

type ReplyMessage struct {
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
		natsConsumerOutCh     = make(chan natsconsumer.OutputMessage, 1024)
		requestHandlerInCh    = make(chan requesthandler.InputMessage, 1024)
		requestHandlerOutCh   = make(chan requesthandler.OutputMessage, 1024)
		requestHandlerReplyCh = make(chan requesthandler.ReplyMessage, 1024)
		replyHandlerInCh      = make(chan replyhandler.InputMessage, 1024)
		replyHandlerOutCh     = make(chan replyhandler.OutputMessage, 1024)
		natsPublisherInCh     = make(chan natspublisher.InputMessage, 1024)
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
			requesthandler.InputMessage,
		](
			natsConsumerOutCh,
			requestHandlerInCh,
			func(msg natsconsumer.OutputMessage) requesthandler.InputMessage {
				return requesthandler.InputMessage{
					Subject:      msg.Subject,
					ReplySubject: msg.ReplySubject,
					Data:         msg.Data,
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		return requesthandler.New(
			s.nc,
		).Start(
			groupCtx,
			requesthandler.Arguments{
				InputCh:  requestHandlerInCh,
				OutputCh: requestHandlerOutCh,
				ReplyCh:  requestHandlerReplyCh,
			},
		)
	})

	group.Go(func() error {
		return convertor.New[
			requesthandler.OutputMessage,
			OutputMessage,
		](
			requestHandlerOutCh,
			args.OutputCh,
			func(msg requesthandler.OutputMessage) OutputMessage {
				return OutputMessage{
					Subject: msg.Subject,
					Data:    msg.Data,
					ReplyFn: func(ctx context.Context, data []byte) error {
						return msg.ReplyFn(ctx, data)
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
			requesthandler.ReplyMessage,
			replyhandler.InputMessage,
		](
			requestHandlerReplyCh,
			replyHandlerInCh,
			func(msg requesthandler.ReplyMessage) replyhandler.InputMessage {
				return replyhandler.InputMessage{
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
		args := replyhandler.Arguments{
			InputCh:  replyHandlerInCh,
			OutputCh: replyHandlerOutCh,
		}
		return replyhandler.New(
			s.nc,
		).Start(
			groupCtx,
			args,
		)
	})

	group.Go(func() error {
		return convertor.New[
			replyhandler.OutputMessage,
			natspublisher.InputMessage,
		](
			replyHandlerOutCh,
			natsPublisherInCh,
			func(msg replyhandler.OutputMessage) natspublisher.InputMessage {
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
