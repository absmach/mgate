package mproxy

import (
	"context"
	"net"
	"net/http"
)

// Handler is an interface for mProxy hooks.
type Handler interface {
	// Authorization on client `CONNECT`
	// Each of the params are passed by reference, so that it can be changed
	AuthConnect(ctx context.Context) error

	// Authorization on client `PUBLISH`
	// Topic is passed by reference, so that it can be modified
	AuthPublish(ctx context.Context, topic *string, payload *[]byte) error

	// Authorization on client `SUBSCRIBE`
	// Topics are passed by reference, so that they can be modified
	AuthSubscribe(ctx context.Context, topics *[]string) error

	// After client successfully connected
	Connect(ctx context.Context) error

	// After client successfully published
	Publish(ctx context.Context, topic *string, payload *[]byte) error

	// After client successfully subscribed
	Subscribe(ctx context.Context, topics *[]string) error

	// After client unsubscribed
	Unsubscribe(ctx context.Context, topics *[]string) error

	// Disconnect on connection with client lost
	Disconnect(ctx context.Context) error
}

// Interceptor is an interface for mProxy intercept hook.
type Interceptor interface {
	// Intercept is called on every packet flowing through the Proxy.
	// Packets can be modified before being sent to the broker or the client.
	// If the interceptor returns a non-nil packet, the modified packet is sent.
	// The error indicates unsuccessful interception and mProxy is cancelling the packet.
	Intercept(ctx context.Context, pkt interface{}) (interface{}, error)
}

type Streamer interface {
	Stream(ctx context.Context, r, w net.Conn, h Handler, ic Interceptor, errs chan error)
}

type Passer interface {
	Pass(rw http.ResponseWriter, r *http.Request)
}
