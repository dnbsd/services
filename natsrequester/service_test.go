package natsrequester

import (
	"context"
	"fmt"
	natsserver "github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"log/slog"
	"os"
	"testing"
	"time"
)

func TestService(t *testing.T) {
	const subject = "test"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	nc := newNatsServerAndConnection(t)
	s := New(Arguments{
		Logger: slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			Level: slog.LevelDebug,
		})),
		Conn:    nc,
		Subject: subject,
		Group:   "",
	})
	go func() {
		err := s.Start(ctx)
		assert.NoError(t, err)
	}()

	// Request replier
	go func() {
		msg := <-s.Output()
		assert.NotEmpty(t, msg.Subject)
		assert.Equal(t, "Test request", string(msg.Data))
		err := msg.Reply(ctx, []byte("Test reply"))
		assert.NoError(t, err)
	}()

	time.Sleep(1 * time.Second)

	// Requester
	msg, err := nc.RequestMsgWithContext(ctx, &nats.Msg{
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
