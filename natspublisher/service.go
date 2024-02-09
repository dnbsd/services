package natspublisher

import (
	"context"
	"github.com/dnbsd/services/proto"
	"github.com/nats-io/nats.go"
	"log/slog"
)

var (
	_ proto.Service                         = &Service[struct{}]{}
	_ proto.Adapter[struct{}, InputMessage] = &Service[struct{}]{}
)

type Arguments struct {
	Logger *slog.Logger
	Conn   *nats.Conn
}

type InputMessage struct {
	Subject string
	Data    []byte
}

type Service[MT any] struct {
	args      Arguments
	source    proto.Producer[MT]
	convertor func(MT) InputMessage
}

func New[MT any](args Arguments) *Service[MT] {
	return &Service[MT]{
		args: args,
	}
}

func (s *Service[MT]) Start(ctx context.Context) error {
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

	for {
		select {
		case inMsg := <-s.source.Output():
			logger.Debug("received a message")

			msg := s.convertor(inMsg)
			err := s.args.Conn.PublishMsg(&nats.Msg{
				Subject: msg.Subject,
				Data:    msg.Data,
			})
			if err != nil {
				logger.Warn("cannot publish a NATS message", "error", err)
				continue
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
