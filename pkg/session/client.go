package session

import (
	"context"
	"crypto/x509"
	"errors"
)

type ctxkey string

const MPROXY_CLIENT_KEY ctxkey = "mproxy-client"

var ErrClientNotInContext error = errors.New("client not set in context")

// Client stores MQTT client data.
type Client struct {
	ID       string
	Username string
	Password []byte
	Cert     x509.Certificate
}

func (c Client) ToContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, MPROXY_CLIENT_KEY, c)
}

func (c *Client) FromContext(ctx context.Context) error {
	if client, ok := ctx.Value(MPROXY_CLIENT_KEY).(Client); ok {
		c.ID = client.ID
		c.Password = client.Password
		c.Username = client.Username
		c.Cert = client.Cert
		return nil
	}
	return ErrClientNotInContext
}
