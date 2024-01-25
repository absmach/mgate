package session

import (
	"context"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

// Interceptor is an interface for mProxy intercept hook.

//go:generate mockery --name Interceptor --filename interceptor.go --quiet --note "Copyright (c) Abstract Machines"
type Interceptor interface {
	// Intercept is called on every packet flowing through the Proxy.
	// Packets can be modified before being sent to the broker or the client.
	// If the interceptor returns a non-nil packet, the modified packet is sent.
	// The error indicates unsuccessful interception and mProxy is cancelling the packet.
	Intercept(ctx context.Context, pkt packets.ControlPacket, dir Direction) (packets.ControlPacket, error)
}
