package natsconsumer

import (
	"context"
	"github.com/dnbsd/services/proto"
	"github.com/nats-io/nats.go"
	"log/slog"
)

var (
	_ proto.Service                 = &Service{}
	_ proto.Producer[OutputMessage] = &Service{}
)

type Arguments struct {
	Logger  *slog.Logger
	Conn    *nats.Conn
	Subject string
	Group   string
}

type OutputMessage struct {
	Subject      string
	ReplySubject string
	Data         []byte
}

type Service struct {
	args     Arguments
	outputCh chan OutputMessage
}

func New(args Arguments) *Service {
	return &Service{
		args:     args,
		outputCh: make(chan OutputMessage),
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

	inputCh := make(chan *nats.Msg, 1024)
	subscription, err := s.args.Conn.ChanQueueSubscribe(s.args.Subject, s.args.Group, inputCh)
	if err != nil {
		return err
	}
	defer subscription.Drain()

	for {
		select {
		case inMsg := <-inputCh:
			logger.Debug("received a message")

			outMsg := OutputMessage{
				Subject:      inMsg.Subject,
				ReplySubject: inMsg.Reply,
				Data:         inMsg.Data,
			}
			select {
			case s.outputCh <- outMsg:
				logger.Debug("pushed a message to output channel")
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) Output() <-chan OutputMessage {
	return s.outputCh
}
