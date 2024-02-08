package requestdecoder

import (
	"context"
	"github.com/dnbsd/services/convertor"
	"github.com/dnbsd/services/jsondecoder"
	"github.com/dnbsd/services/jsonrpc/messages"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	InputCh  <-chan InputMessage
	OutputCh chan<- OutputMessage
}

type InputMessage struct {
	Data []byte
}

type OutputMessage struct {
	Data  messages.JsonRpcRequest
	Error error
}

type Service struct{}

func New() *Service {
	return &Service{}
}

func (s *Service) Start(ctx context.Context, args Arguments) error {
	var (
		jsonDecoderInCh  = make(chan jsondecoder.InputMessage, 1024)
		jsonDecoderOutCh = make(chan jsondecoder.OutputMessage[messages.JsonRpcRequest], 1024)
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return convertor.New[
			InputMessage,
			jsondecoder.InputMessage,
		](
			args.InputCh,
			jsonDecoderInCh,
			func(msg InputMessage) jsondecoder.InputMessage {
				return jsondecoder.InputMessage{
					Data: msg.Data,
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		return jsondecoder.New[messages.JsonRpcRequest](
			jsonDecoderInCh,
			jsonDecoderOutCh,
		).Start(
			groupCtx,
			jsondecoder.Arguments{},
		)
	})

	group.Go(func() error {
		return convertor.New[
			jsondecoder.OutputMessage[messages.JsonRpcRequest],
			OutputMessage,
		](
			jsonDecoderOutCh,
			args.OutputCh,
			func(msg jsondecoder.OutputMessage[messages.JsonRpcRequest]) OutputMessage {
				err := msg.Error
				if err == nil {
					err = msg.Data.Validate()
				}

				return OutputMessage{
					Data: messages.JsonRpcRequest{
						Version: msg.Data.Version,
						ID:      msg.Data.ID,
						Method:  msg.Data.Method,
						Params:  msg.Data.Params,
					},
					Error: err,
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
