package simple

import (
	"context"
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/pkg/session"
)

var _ session.Handler = (*Handler)(nil)

// Handler implements mqtt.Handler interface
type Handler struct {
	logger logger.Logger
}

// New creates new Event entity
func New(logger logger.Logger) *Handler {
	return &Handler{
		logger: logger,
	}
}

// AuthConnect is called on device connection,
// prior forwarding to the MQTT broker
func (h *Handler) AuthConnect(ctx context.Context) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return nil
	}
	h.logger.Info(fmt.Sprintf("AuthConnect() - sessionID: %s, username: %s, password: %s, client_CN: %s", s.ID, s.Username, string(s.Password), s.Cert.Subject.CommonName))

	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker
func (h *Handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return nil
	}
	h.logger.Info(fmt.Sprintf("AuthPublish() - sessionID: %s, topic: %s, payload: %s", s.ID, *topic, string(*payload)))

	return nil
}

// AuthSubscribe is called on device publish,
// prior forwarding to the MQTT broker
func (h *Handler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return nil
	}
	h.logger.Info(fmt.Sprintf("AuthSubscribe() - sessionID: %s, topics: %s", s.ID, strings.Join(*topics, ",")))

	return nil
}

// Connect - after client successfully connected
func (h *Handler) Connect(ctx context.Context) {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return
	}
	h.logger.Info(fmt.Sprintf("Connect() - username: %s, sessionID: %s", s.Username, s.ID))
}

// Publish - after client successfully published
func (h *Handler) Publish(ctx context.Context, topic *string, payload *[]byte) {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return
	}
	h.logger.Info(fmt.Sprintf("Publish() - username: %s, sessionID: %s, topic: %s, payload: %s", s.Username, s.ID, *topic, string(*payload)))
}

// Subscribe - after client successfully subscribed
func (h *Handler) Subscribe(ctx context.Context, topics *[]string) {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return
	}
	h.logger.Info(fmt.Sprintf("Subscribe() - username: %s, sessionID: %s, topics: %s", s.Username, s.ID, strings.Join(*topics, ",")))
}

// Unsubscribe - after client unsubscribed
func (h *Handler) Unsubscribe(ctx context.Context, topics *[]string) {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return
	}
	h.logger.Info(fmt.Sprintf("Unsubscribe() - username: %s, sessionID: %s, topics: %s", s.Username, s.ID, strings.Join(*topics, ",")))
}

// Disconnect on connection lost
func (h *Handler) Disconnect(ctx context.Context) {
	s, ok := session.FromContext(ctx)
	if !ok {
		h.logger.Error("Session is missing!")
		return
	}
	h.logger.Info(fmt.Sprintf("Disconnect() - client with username: %s and ID: %s disconnected", s.Username, s.ID))
}
