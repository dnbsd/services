package natsconsumer

import (
	"context"
	"github.com/nats-io/nats.go"
)

type Arguments struct {
	Subject  string
	Group    string
	OutputCh chan<- OutputMessage
}

type OutputMessage struct {
	Subject      string
	ReplySubject string
	Data         []byte
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
	inputCh := make(chan *nats.Msg, 1024)
	subscription, err := s.nc.ChanQueueSubscribe(args.Subject, args.Group, inputCh)
	if err != nil {
		return err
	}
	defer subscription.Drain()

	select {
	case inMsg := <-inputCh:
		outMsg := OutputMessage{
			Subject:      inMsg.Subject,
			ReplySubject: inMsg.Reply,
			Data:         inMsg.Data,
		}

		select {
		case args.OutputCh <- outMsg:
		case <-ctx.Done():
			return nil
		}

	case <-ctx.Done():
		return nil
	}

	return nil
}
