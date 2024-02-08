package natspublisher

import (
	"context"
	"github.com/dnbsd/services/proto"
	"github.com/nats-io/nats.go"
	"log/slog"
)

var _ proto.Service = &Service{}
var _ proto.EventConsumer = &Service{}

type Arguments struct {
	Logger *slog.Logger
	Conn   *nats.Conn
}

type InputMessage struct {
	Subject string
	Data    []byte
}

type Service struct {
	args    Arguments
	inputCh chan proto.Event
}

func New(args Arguments) *Service {
	return &Service{
		args:    args,
		inputCh: make(chan proto.Event),
	}
}

func (s *Service) Start(ctx context.Context) error {
	for {
		select {
		case inMsg := <-s.inputCh:
			body := inMsg.Body.(InputMessage)
			outMsg := &nats.Msg{
				Subject: body.Subject,
				Data:    body.Data,
			}

			err := s.args.Conn.PublishMsg(outMsg)
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) Input() chan<- proto.Event {
	return s.inputCh
}
