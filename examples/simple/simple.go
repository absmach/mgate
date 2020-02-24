package simple

import (
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/pkg/events"
)

var _ events.Event = (*Event)(nil)

// Event implements events.Event interface
type Event struct {
	logger logger.Logger
}

// New creates new Event entity
func New(logger logger.Logger) *Event {
	return &Event{
		logger: logger,
	}
}

// AuthRegister is called on device connection,
// prior forwarding to the MQTT broker
func (e *Event) AuthRegister(username, clientID *string, password *[]byte) error {
	e.logger.Info(fmt.Sprintf("AuthRegister() - clientID: %s, username: %s, password: %s", *clientID, *username, string(*password)))
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker
func (e *Event) AuthPublish(username, clientID string, topic *string, payload *[]byte) error {
	e.logger.Info(fmt.Sprintf("AuthPublish() - clientID: %s, topic: %s, payload: %s", clientID, *topic, string(*payload)))
	return nil
}

// AuthSubscribe is called on device publish,
// prior forwarding to the MQTT broker
func (e *Event) AuthSubscribe(username, clientID string, topics *[]string) error {
	e.logger.Info(fmt.Sprintf("AuthSubscribe() - clientID: %s, topics: %s", clientID, strings.Join(*topics, ",")))
	return nil
}

// Register - after client successfully connected
func (e *Event) Register(clientID string) {
	e.logger.Info(fmt.Sprintf("Register() - clientID: %s", clientID))
}

// Publish - after client successfully published
func (e *Event) Publish(clientID, topic string, payload []byte) {
	e.logger.Info(fmt.Sprintf("Publish() - clientID: %s, topic: %s, payload: %s", clientID, topic, string(payload)))
}

// Subscribe - after client successfully subscribed
func (e *Event) Subscribe(clientID string, topics []string) {
	e.logger.Info(fmt.Sprintf("Subscribe() - clientID: %s, topics: %s", clientID, strings.Join(topics, ",")))
}

// Unsubscribe - after client unsubscribed
func (e *Event) Unsubscribe(clientID string, topics []string) {
	e.logger.Info(fmt.Sprintf("Unsubscribe() - clientID: %s, topics: %s", clientID, strings.Join(topics, ",")))
}
