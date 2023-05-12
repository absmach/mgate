package session

import (
	"context"
	"crypto/x509"
)

// The ctxKey type is unexported to prevent collisions with context keys defined in
// other packages.
type ctxKey int

// SessionKey is the context key for the session. Its value (1) is
// arbitrary. If this package defined other context keys, they would have
// different integer values.
const sessionKey ctxKey = 1

// Session stores MQTT session data.
type Session struct {
	ID       string
	Username string
	Password []byte
	Cert     x509.Certificate
}

// NewContext stores Session in context.Context values.
// It uses pointer to the session so it can be modified by handler.
func NewContext(ctx context.Context, s *Session) context.Context {
	return context.WithValue(ctx, sessionKey, s)
}

// FromContext retrieves Session from context.Context.
// Second value indicates if session is present in the context
// and if it's safe to use it (it's not nil).
func FromContext(ctx context.Context) (*Session, bool) {
	if s, ok := ctx.Value(sessionKey).(*Session); ok && s != nil {
		return s, true
	}
	return nil, false
}
