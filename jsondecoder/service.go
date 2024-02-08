package jsondecoder

import (
	"context"
	"encoding/json"
)

type Arguments struct{}

type InputMessage struct {
	Data []byte
}

type OutputMessage[O any] struct {
	Data  O
	Error error
}

type Service[O any] struct {
	InputCh  <-chan InputMessage
	OutputCh chan<- OutputMessage[O]
}

func New[O any](inputCh <-chan InputMessage, outputCh chan<- OutputMessage[O]) *Service[O] {
	return &Service[O]{
		InputCh:  inputCh,
		OutputCh: outputCh,
	}
}

func (s *Service[O]) Start(ctx context.Context, args Arguments) error {
	for {
		select {
		case inMsg := <-s.InputCh:
			var o O
			err := json.Unmarshal(inMsg.Data, &o)
			outMsg := OutputMessage[O]{
				Data:  o,
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
