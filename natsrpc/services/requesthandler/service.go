package requesthandler

import (
	"context"
	"github.com/nats-io/nats.go"
)

type Arguments struct {
	InputCh  <-chan InputMessage
	OutputCh chan<- OutputMessage
	ReplyCh  chan<- ReplyMessage
}

type InputMessage struct {
	Subject      string
	ReplySubject string
	Data         []byte
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
	nc *nats.Conn
}

func New(nc *nats.Conn) *Service {
	return &Service{
		nc: nc,
	}
}

func (s *Service) Start(ctx context.Context, args Arguments) error {
	for {
		select {
		case inMsg := <-args.InputCh:
			if inMsg.ReplySubject == "" {
				continue
			}

			outMsg := OutputMessage{
				Subject: inMsg.Subject,
				Data:    inMsg.Data,
				ReplyFn: func(ctx context.Context, data []byte) error {
					select {
					case args.ReplyCh <- ReplyMessage{
						Subject: inMsg.ReplySubject,
						Data:    data,
					}:
					case <-ctx.Done():
						return ctx.Err()
					}
					return nil
				},
			}
			select {
			case args.OutputCh <- outMsg:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
