package natspublisher

import (
	"context"
	"github.com/nats-io/nats.go"
)

type Arguments struct {
	InputCh <-chan InputMessage
}

type InputMessage struct {
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
	for {
		select {
		case inMsg := <-args.InputCh:
			outMsg := &nats.Msg{
				Subject: inMsg.Subject,
				Data:    inMsg.Data,
			}

			err := s.nc.PublishMsg(outMsg)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return nil
		}
	}
}
