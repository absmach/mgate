package session

import (
	"context"
	"crypto/x509"
	"errors"
)

// The ctxKey type is unexported to prevent collisions with context keys defined in
// other packages.
type ctxKey int

// clientKey is the context key for the session client.  Its value of one is
// arbitrary.  If this package defined other context keys, they would have
// different integer values.
const clientKey ctxKey = 1

// ErrClientNotInContext failed to retrieve client from context
var ErrClientNotInContext error = errors.New("client not set in context")

// Client stores MQTT client data.
type Client struct {
	ID       string
	Username string
	Password []byte
	Cert     x509.Certificate
}

// ToContext store Client in context.Context values
func (c Client) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, clientKey, c)
}

// FromContext retrieve client from context.Context
func FromContext(ctx context.Context) (Client, error) {
	if client, ok := ctx.Value(clientKey).(Client); ok {
		return client, nil
	}
	return Client{}, ErrClientNotInContext
}
