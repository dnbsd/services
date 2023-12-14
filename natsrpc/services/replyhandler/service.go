package replyhandler

import (
	"context"
	"github.com/nats-io/nats.go"
)

type Arguments struct {
	InputCh  <-chan InputMessage
	OutputCh chan<- OutputMessage
}

type InputMessage struct {
	Subject string
	Data    []byte
}

type OutputMessage struct {
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
			outMsg := OutputMessage{
				Subject: inMsg.Subject,
				Data:    inMsg.Data,
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
