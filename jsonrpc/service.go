package jsonrpc

import (
	"context"
	"github.com/dnbsd/services/convertor"
	"github.com/dnbsd/services/jsonrpc/internal/services/requestdecoder"
	"github.com/dnbsd/services/jsonrpc/internal/services/responseencoder"
	"github.com/dnbsd/services/jsonrpc/messages"
	"golang.org/x/sync/errgroup"
)

type Arguments struct {
	InputCh  <-chan InputMessage
	OutputCh chan<- OutputMessage
}

type InputMessage struct {
	Data messages.JsonRpcRequest
}

type OutputMessage struct {
	Data messages.JsonRpcResponse
}

type Service struct {
	//logger *slog.Logger
}

func New() *Service {
	return &Service{}
}

func (s *Service) Start(ctx context.Context, args Arguments) error {
	var (
		requestDecoderInCh   = make(chan requestdecoder.InputMessage, 1024)
		requestDecoderOutCh  = make(chan requestdecoder.OutputMessage, 1024)
		responseCh           = make(chan responseMessage, 1024)
		responseEncoderInCh  = make(chan responseencoder.InputMessage, 1024)
		responseEncoderOutCh = make(chan responseencoder.OutputMessage, 1024)
	)

	group, groupCtx := errgroup.WithContext(ctx)
	group.Go(func() error {
		return convertor.New[
			InputMessage,
			requestdecoder.InputMessage,
		](
			args.InputCh,
			requestDecoderInCh,
			func(msg InputMessage) requestdecoder.InputMessage {
				return requestdecoder.InputMessage{
					Data: msg.Data,
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		return requestdecoder.New().Start(
			groupCtx,
			requestdecoder.Arguments{
				InputCh:  requestDecoderInCh,
				OutputCh: requestDecoderOutCh,
			},
		)
	})

	group.Go(func() error {
		return convertor.New[
			requestdecoder.OutputMessage,
			OutputMessage,
		](
			requestDecoderOutCh,
			args.OutputCh,
			func(msg requestdecoder.OutputMessage) OutputMessage {
				return OutputMessage{
					Data:  msg.Data,
					Error: msg.Error,
					ResponseFn: func(ctx context.Context, result any, err error) error {
						if msg.Data.IsNotification() {
							return messages.ErrIsNotification
						}

						select {
						case responseCh <- responseMessage{
							ID:     msg.Data.ID,
							Result: result,
							Error:  err,
						}:
						case <-ctx.Done():
							return ctx.Err()
						}
						return nil
					},
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		return convertor.New[
			responseMessage,
			responseencoder.InputMessage,
		](
			responseCh,
			responseEncoderInCh,
			func(msg responseMessage) responseencoder.InputMessage {
				var data messages.JsonRpcResponse
				if msg.Error != nil {
					data = messages.NewErrorResponse(msg.ID, msg.Error)
				} else {
					data = messages.NewResponse(msg.ID, msg.Result)
				}

				return responseencoder.InputMessage{
					Data: data,
				}
			},
		).Start(
			groupCtx,
			convertor.Arguments{},
		)
	})

	group.Go(func() error {
		return responseencoder.New().Start(
			groupCtx,
			responseencoder.Arguments{
				InputCh:  responseEncoderInCh,
				OutputCh: responseEncoderOutCh,
			},
		)
	})

	err := group.Wait()
	if err != nil {
		return err
	}

	return nil
}
