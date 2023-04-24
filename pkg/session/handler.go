package session

import "context"

// Handler is an interface for mProxy hooks
type Handler interface {
	// Authorization on client `CONNECT`
	// Each of the params are passed by reference, so that it can be changed
	AuthConnect(ctx context.Context, client *Client) error

	// Authorization on client `PUBLISH`
	// Topic is passed by reference, so that it can be modified
	AuthPublish(ctx context.Context, client *Client, topic *string, payload *[]byte) error

	// Authorization on client `SUBSCRIBE`
	// Topics are passed by reference, so that they can be modified
	AuthSubscribe(ctx context.Context, client *Client, topics *[]string) error

	// After client successfully connected
	Connect(ctx context.Context, client *Client)

	// After client successfully published
	Publish(ctx context.Context, client *Client, topic *string, payload *[]byte)

	// After client successfully subscribed
	Subscribe(ctx context.Context, client *Client, topics *[]string)

	// After client unsubscribed
	Unsubscribe(ctx context.Context, client *Client, topics *[]string)

	// Disconnect on connection with client lost
	Disconnect(ctx context.Context, client *Client)
}
