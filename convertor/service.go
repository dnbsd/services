package convertor

import "context"

type Arguments struct{}

type Service[I, O any] struct {
	inputCh   <-chan I
	outputCh  chan<- O
	convertFn func(I) O
}

func New[I, O any](inputCh <-chan I, outputCh chan<- O, convertFn func(I) O) *Service[I, O] {
	return &Service[I, O]{
		inputCh:   inputCh,
		outputCh:  outputCh,
		convertFn: convertFn,
	}
}

func (s *Service[I, O]) Start(ctx context.Context, _ Arguments) error {
	for {
		select {
		case msgIn := <-s.inputCh:
			msgOut := s.convertFn(msgIn)

			select {
			case s.outputCh <- msgOut:
			case <-ctx.Done():
				return nil
			}

		case <-ctx.Done():
			return nil
		}
	}
}
