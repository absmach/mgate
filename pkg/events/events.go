package events

// Event is an interface for mProxy hooks
type Event interface {
	// Athorization on client `CONNECT`
	// Each of the params are passed by reference, so that it can be changed
	AuthRegister(username, clientID *string, password *[]byte) error

	// Athorization on client `PUBLISH`
	// Topic is passed by reference, so that it can be modified
	AuthPublish(username, clientID string, topic *string, payload *[]byte) error

	// Athorization on client `SUBSCRIBE`
	// Topics are passed by reference, so that they can be modified
	AuthSubscribe(username, clientID string, topics *[]string) error

	// After client sucesfully connected
	Register(clientID string)

	// After client sucesfully published
	Publish(clientID, topic string, payload []byte)

	// After client sucesfully subscribed
	Subscribe(clientID string, topics []string)

	// After client unsubscribed
	Unubscribe(clientID string, topics []string)
}
