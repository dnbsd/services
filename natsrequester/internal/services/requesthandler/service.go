package requesthandler

import (
	"context"
	"github.com/dnbsd/services/proto"
	"log/slog"
)

var (
// TODO: add interface tests
)

type Arguments struct {
	Logger *slog.Logger
}

type Service[IT, OT any] struct {
	args      Arguments
	source    proto.Producer[IT]
	outputCh  chan OT
	convertor func(IT) OT
}

func New[IT, OT any](args Arguments) *Service[IT, OT] {
	return &Service[IT, OT]{
		args:     args,
		outputCh: make(chan OT),
	}
}

func (s *Service[IT, OT]) Start(ctx context.Context) error {
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

			select {
			case s.outputCh <- s.convertor(inMsg):
				logger.Debug("pushed a message to output channel")
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service[IT, OT]) Connect(source proto.Producer[IT], convertor func(IT) OT) {
	s.source = source
	s.convertor = convertor
}

func (s *Service[IT, OT]) Output() <-chan OT {
	return s.outputCh
}
