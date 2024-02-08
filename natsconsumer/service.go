package natsconsumer

import (
	"context"
	"github.com/dnbsd/services/proto"
	"github.com/nats-io/nats.go"
	"log/slog"
)

var _ proto.Service = &Service{}
var _ proto.EventProducer = &Service{}

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
	outputCh chan proto.Event
}

func New(args Arguments) *Service {
	return &Service{
		args:     args,
		outputCh: make(chan proto.Event),
	}
}

func (s *Service) Start(ctx context.Context) error {
	inputCh := make(chan *nats.Msg, 1024)
	subscription, err := s.args.Conn.ChanQueueSubscribe(s.args.Subject, s.args.Group, inputCh)
	if err != nil {
		return err
	}
	defer subscription.Drain()

	for {
		select {
		case inMsg := <-inputCh:
			outMsg := proto.Event{
				Body: OutputMessage{
					Subject:      inMsg.Subject,
					ReplySubject: inMsg.Reply,
					Data:         inMsg.Data,
				},
			}

			select {
			case s.outputCh <- outMsg:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service) Output() <-chan proto.Event {
	return s.outputCh
}
