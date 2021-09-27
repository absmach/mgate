package session

import "github.com/eclipse/paho.mqtt.golang/packets"

// Interceptor is an interface for mProxy intercept hook
type Interceptor interface {
	// Intercept is called on every packet flowing through the proxy
	// packets can be modified before being sent to the broker or the client
	// If the interceptor returns a non-nil packet, the modified packet is sent
	// If the interceptor returns nil or error, the original packet is send
	Intercept(pkt packets.ControlPacket, c *Client, d Direction) (packets.ControlPacket, error)
}
