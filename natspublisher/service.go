package natspublisher

import (
	"context"
	"github.com/dnbsd/services/proto"
	"github.com/nats-io/nats.go"
	"log/slog"
)

var _ proto.Service = &Service[proto.Request]{}
var _ proto.Service = &Service[proto.Event]{}

type Arguments struct {
	Logger *slog.Logger
	Conn   *nats.Conn
}

type InputMessage struct {
	Subject string
	Data    []byte
}

type Service[MT proto.Request | proto.Event] struct {
	args      Arguments
	source    proto.Producer[MT]
	convertor func(MT) InputMessage
}

func New[MT proto.Request | proto.Event](args Arguments) *Service[MT] {
	return &Service[MT]{
		args: args,
	}
}

func (s *Service[MT]) Start(ctx context.Context) error {
	for {
		select {
		case inMsg := <-s.source.Output():
			msg := s.convertor(inMsg)
			err := s.args.Conn.PublishMsg(&nats.Msg{
				Subject: msg.Subject,
				Data:    msg.Data,
			})
			if err != nil {
				return err
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service[MT]) Connect(source proto.Producer[MT], convertor func(MT) InputMessage) {
	s.source = source
	s.convertor = convertor
}
