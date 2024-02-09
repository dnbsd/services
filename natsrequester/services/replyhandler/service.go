package replyhandler

import (
	"context"
	"log/slog"
)

var (
// TODO: add interface tests
)

type Arguments struct {
	Logger *slog.Logger
}

type Service[IT any] struct {
	args     Arguments
	inputCh  chan IT
	outputCh chan IT
}

func New[IT any](args Arguments) *Service[IT] {
	return &Service[IT]{
		args:     args,
		inputCh:  make(chan IT),
		outputCh: make(chan IT),
	}
}

func (s *Service[IT]) Start(ctx context.Context) error {
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
		case inMsg := <-s.inputCh:
			logger.Debug("received a message")

			select {
			case s.outputCh <- inMsg:
				logger.Debug("pushed a message to output channel")
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}

func (s *Service[IT]) Input() chan<- IT {
	return s.inputCh
}

func (s *Service[IT]) Output() <-chan IT {
	return s.outputCh
}
