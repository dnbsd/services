package responseencoder

import (
	"context"
	"github.com/dnbsd/services/convertor"
	"github.com/dnbsd/services/jsonencoder"
	"github.com/dnbsd/services/jsonrpc/messages"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	InputCh  <-chan InputMessage
	OutputCh chan<- OutputMessage
}

type InputMessage struct {
	Data messages.JsonRpcResponse
}

type OutputMessage struct {
	Data  []byte
	Error error
}

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Start(ctx context.Context, args Arguments) error {
	var (
		jsonEncoderInCh  = make(chan jsonencoder.InputMessage[messages.JsonRpcResponse], 1024)
		jsonEncoderOutCh = make(chan jsonencoder.OutputMessage, 1024)
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return convertor.New[
			InputMessage,
			jsonencoder.InputMessage[messages.JsonRpcResponse],
		](
			args.InputCh,
			jsonEncoderInCh,
			func(msg InputMessage) jsonencoder.InputMessage[messages.JsonRpcResponse] {
				return jsonencoder.InputMessage[messages.JsonRpcResponse]{
					Data: msg.Data,
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		return jsonencoder.New[messages.JsonRpcResponse](
			jsonEncoderInCh,
			jsonEncoderOutCh,
		).Start(
			groupCtx,
			jsonencoder.Arguments{},
		)
	})

	group.Go(func() error {
		return convertor.New[
			jsonencoder.OutputMessage,
			OutputMessage,
		](
			jsonEncoderOutCh,
			args.OutputCh,
			func(msg jsonencoder.OutputMessage) OutputMessage {
				return OutputMessage{
					Data:  msg.Data,
					Error: msg.Error,
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}
