package services

import (
	"context"
	"fmt"
	"testing"
	"time"
)

func Test(t *testing.T) {
	c := NewServiceC()
	go c.Start(context.Background())
	time.Sleep(1 * time.Second)

	for {
		select {
		case req := <-c.Output():
			fmt.Printf("received a request! %v\n", req)
			req.Reply(context.Background(), `{"response":"ok"}`)
		}
	}
}
