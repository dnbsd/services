package proto

import (
	"context"
	"errors"
)

var ErrNoResponder = errors.New("request has no responder")

type Event struct {
	Body any
}

type Request struct {
	Body      any
	RespondTo ResponseConsumer
}

func (m Request) reply(ctx context.Context, response Response) error {
	if m.RespondTo == nil {
		return ErrNoResponder
	}

	select {
	case m.RespondTo.Input() <- response:
	case <-ctx.Done():
		return ctx.Err()
	}

	return nil
}

func (m Request) Reply(ctx context.Context, result any) error {
	return m.reply(ctx, Response{
		Result: result,
	})
}

func (m Request) ReplyError(ctx context.Context, err error) error {
	return m.reply(ctx, Response{
		Error: err,
	})
}

type Response struct {
	Result any
	Error  error
}
