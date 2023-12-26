package natsrequester

import (
	"context"
	"fmt"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	const subject = "test"
	nc := newNatsServerAndConnection(t)
	outputCh := make(chan OutputMessage)
	s := New(nc)
	go func() {
		err := s.Start(context.Background(), Arguments{
			Subject:  subject,
			OutputCh: outputCh,
		})
		assert.NoError(t, err)
	}()

	go func() {
		msg := <-outputCh
		assert.NotEmpty(t, msg.Subject)
		assert.Equal(t, "Test request", string(msg.Data))
		err := msg.ReplyFn(context.Background(), []byte("Test reply"))
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	msg, err := nc.RequestMsgWithContext(context.Background(), &nats.Msg{
		Subject: subject,
		Reply:   nats.NewInbox(),
		Data:    []byte("Test request"),
	})
	assert.NoError(t, err)
	assert.Equal(t, "Test reply", string(msg.Data))
}

func newNatsServerAndConnection(t *testing.T) *nats.Conn {
	opts := natsserver.DefaultTestOptions
	opts.NoLog = false
	opts.Port = 14444
	s := natsserver.RunServer(&opts)
	uri := fmt.Sprintf("nats://%s:%d", opts.Host, opts.Port)
	nc, err := nats.Connect(uri)
	assert.NoError(t, err)
	t.Cleanup(func() {
		nc.Close()
		s.Shutdown()
	})
	return nc
}
