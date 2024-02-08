package jsonencoder

import (
	"context"
	"encoding/json"
)

type Arguments struct{}

type InputMessage[I any] struct {
	Data I
}

type OutputMessage struct {
	Data  []byte
	Error error
}

type Service[I any] struct {
	InputCh  <-chan InputMessage[I]
	OutputCh chan<- OutputMessage
}

func New[I any](inputCh <-chan InputMessage[I], outputCh chan<- OutputMessage) *Service[I] {
	return &Service[I]{
		InputCh:  inputCh,
		OutputCh: outputCh,
	}
}

func (s *Service[I]) Start(ctx context.Context, args Arguments) error {
	for {
		select {
		case inMsg := <-s.InputCh:
			b, err := json.Marshal(inMsg.Data)
			outMsg := OutputMessage{
				Data:  b,
				Error: err,
			}

			select {
			case s.OutputCh <- outMsg:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
