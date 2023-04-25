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
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("AuthConnect() - clientID: %s, username: %s, password: %s, client_CN: %s", c.ID, c.Username, string(c.Password), c.Cert.Subject.CommonName))
	} else {
		h.logger.Error("client is missing")
	}
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker
func (h *Handler) AuthPublish(ctx context.Context, topic *string, payload *[]byte) error {
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("AuthPublish() - clientID: %s, topic: %s, payload: %s", c.ID, *topic, string(*payload)))
	} else {
		h.logger.Error("client is missing")
	}
	return nil
}

// AuthSubscribe is called on device publish,
// prior forwarding to the MQTT broker
func (h *Handler) AuthSubscribe(ctx context.Context, topics *[]string) error {
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("AuthSubscribe() - clientID: %s, topics: %s", c.ID, strings.Join(*topics, ",")))
	} else {
		h.logger.Error("client is missing")
	}
	return nil
}

// Connect - after client successfully connected
func (h *Handler) Connect(ctx context.Context) {
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("Connect() - username: %s, clientID: %s", c.Username, c.ID))
	} else {
		h.logger.Error("client is missing")
	}
}

// Publish - after client successfully published
func (h *Handler) Publish(ctx context.Context, topic *string, payload *[]byte) {
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("Publish() - username: %s, clientID: %s, topic: %s, payload: %s", c.Username, c.ID, *topic, string(*payload)))
	} else {
		h.logger.Error("client is missing")
	}
}

// Subscribe - after client successfully subscribed
func (h *Handler) Subscribe(ctx context.Context, topics *[]string) {
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("Subscribe() - username: %s, clientID: %s, topics: %s", c.Username, c.ID, strings.Join(*topics, ",")))
	} else {
		h.logger.Error("client is missing")
	}
}

// Unsubscribe - after client unsubscribed
func (h *Handler) Unsubscribe(ctx context.Context, topics *[]string) {
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("Unsubscribe() - username: %s, clientID: %s, topics: %s", c.Username, c.ID, strings.Join(*topics, ",")))
	} else {
		h.logger.Error("client is missing")
	}
}

// Disconnect on conection lost
func (h *Handler) Disconnect(ctx context.Context) {
	var c session.Client
	if err := c.FromContext(ctx); err != nil {
		h.logger.Info(fmt.Sprintf("Disconnect() - client with username: %s and ID: %s disconenected", c.Username, c.ID))
	} else {
		h.logger.Error("client is missing")
	}
}
