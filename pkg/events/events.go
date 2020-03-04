package events

// Event is an interface for mProxy hooks
type Event interface {
	// Authorization on client `CONNECT`
	// Each of the params are passed by reference, so that it can be changed
	AuthConnect(username, clientID *string, password *[]byte) error

	// Authorization on client `PUBLISH`
	// Topic is passed by reference, so that it can be modified
	AuthPublish(username, clientID string, topic *string, payload *[]byte) error

	// Authorization on client `SUBSCRIBE`
	// Topics are passed by reference, so that they can be modified
	AuthSubscribe(username, clientID string, topics *[]string) error

	// After client successfully connected
	Connect(username, clientID string)

	// After client successfully published
	Publish(username, clientID, topic string, payload []byte)

	// After client successfully subscribed
	Subscribe(username, clientID string, topics []string)

	// After client unsubscribed
	Unsubscribe(username, clientID string, topics []string)

	// Disconnect on connection with client lost
	Disconnect(username, clientID string)
}
