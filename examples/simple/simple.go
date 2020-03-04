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

// AuthConnect is called on device connection,
// prior forwarding to the MQTT broker
func (e *Event) AuthConnect(username, clientID *string, password *[]byte) error {
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

// Connect - after client successfully connected
func (e *Event) Connect(username, clientID string) {
	e.logger.Info(fmt.Sprintf("Register() - username: %s, clientID: %s", username, clientID))
}

// Publish - after client successfully published
func (e *Event) Publish(username, clientID, topic string, payload []byte) {
	e.logger.Info(fmt.Sprintf("Publish() - username: %s, clientID: %s, topic: %s, payload: %s", username, clientID, topic, string(payload)))
}

// Subscribe - after client successfully subscribed
func (e *Event) Subscribe(username, clientID string, topics []string) {
	e.logger.Info(fmt.Sprintf("Subscribe() - username: %s, clientID: %s, topics: %s", username, clientID, strings.Join(topics, ",")))
}

// Unsubscribe - after client unsubscribed
func (e *Event) Unsubscribe(username, clientID string, topics []string) {
	e.logger.Info(fmt.Sprintf("Unsubscribe() - username: %s, clientID: %s, topics: %s", username, clientID, strings.Join(topics, ",")))
}

// Disconnect on conection lost
func (e *Event) Disconnect(username, clientID string) {
	e.logger.Info(fmt.Sprintf("Disconnect() - client with username: %s and ID: %s disconenectd", username, clientID))
}
