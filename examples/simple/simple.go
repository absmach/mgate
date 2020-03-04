package simple

import (
	"fmt"
	"strings"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/pkg/mqtt"
)

var _ mqtt.Event = (*Event)(nil)

// Event implements mqtt.Event interface
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
func (e *Event) AuthConnect(c mqtt.Client) error {
	e.logger.Info(fmt.Sprintf("AuthRegister() - clientID: %s, username: %s, password: %s", c.ID, c.Username, string(c.Password)))
	return nil
}

// AuthPublish is called on device publish,
// prior forwarding to the MQTT broker
func (e *Event) AuthPublish(c mqtt.Client, topic string, payload []byte) error {
	e.logger.Info(fmt.Sprintf("AuthPublish() - clientID: %s, topic: %s, payload: %s", c.ID, topic, string(payload)))
	return nil
}

// AuthSubscribe is called on device publish,
// prior forwarding to the MQTT broker
func (e *Event) AuthSubscribe(c mqtt.Client, topics []string) error {
	e.logger.Info(fmt.Sprintf("AuthSubscribe() - clientID: %s, topics: %s", c.ID, strings.Join(topics, ",")))
	return nil
}

// Connect - after client successfully connected
func (e *Event) Connect(c mqtt.Client) {
	e.logger.Info(fmt.Sprintf("Register() - username: %s, clientID: %s", c.Username, c.ID))
}

// Publish - after client successfully published
func (e *Event) Publish(c mqtt.Client, topic string, payload []byte) {
	e.logger.Info(fmt.Sprintf("Publish() - username: %s, clientID: %s, topic: %s, payload: %s", c.Username, c.ID, topic, string(payload)))
}

// Subscribe - after client successfully subscribed
func (e *Event) Subscribe(c mqtt.Client, topics []string) {
	e.logger.Info(fmt.Sprintf("Subscribe() - username: %s, clientID: %s, topics: %s", c.Username, c.ID, strings.Join(topics, ",")))
}

// Unsubscribe - after client unsubscribed
func (e *Event) Unsubscribe(c mqtt.Client, topics []string) {
	e.logger.Info(fmt.Sprintf("Unsubscribe() - username: %s, clientID: %s, topics: %s", c.Username, c.ID, strings.Join(topics, ",")))
}

// Disconnect on conection lost
func (e *Event) Disconnect(c mqtt.Client) {
	e.logger.Info(fmt.Sprintf("Disconnect() - client with username: %s and ID: %s disconenectd", c.Username, c.ID))
}
