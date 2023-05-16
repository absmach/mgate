package session

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

type direction int

const (
	up direction = iota
	down
)

const unknownID = "unknown"

var (
	errBroker = "failed to proxy from MQTT client with id %s to MQTT broker with error: %s"
	errClient = "failed to proxy from MQTT broker to client with id %s with error: %s"
)

// Stream starts proxy between client and broker.
func Stream(ctx context.Context, inbound, outbound net.Conn, handler Handler, cert x509.Certificate) error {
	// Authorize CONNECT.
	pkt, err := packets.ReadPacket(inbound)
	if err != nil {
		return err
	}
	ctx, err = authorize(ctx, pkt, handler, cert)
	if err != nil {
		return err
	}
	// Send CONNECT to broker.
	if err = pkt.Write(outbound); err != nil {
		return wrap(ctx, err, up)
	}
	// In parallel read from client, send to broker
	// and read from broker, send to client.
	errs := make(chan error, 2)

	go stream(ctx, up, inbound, outbound, handler, cert, errs)
	go stream(ctx, down, outbound, inbound, handler, cert, errs)

	// Handle whichever error happens first.
	// The other routine won't be blocked when writing
	// to the errors channel because it is buffered.
	err = <-errs

	handler.Disconnect(ctx)
	return err
}

func stream(ctx context.Context, dir direction, r, w net.Conn, h Handler, cert x509.Certificate, errs chan error) {
	for {
		// Read from one connection.
		pkt, err := packets.ReadPacket(r)
		if err != nil {
			errs <- wrap(ctx, err, dir)
			return
		}

		if dir == up {
			ctx, err = authorize(ctx, pkt, h, cert)
			if err != nil {
				errs <- wrap(ctx, err, dir)
				return
			}
		}

		// Send to another.
		if err := pkt.Write(w); err != nil {
			errs <- wrap(ctx, err, dir)
			return
		}

		if dir == up {
			notify(ctx, pkt, h)
		}
	}
}

func authorize(ctx context.Context, pkt packets.ControlPacket, h Handler, cert x509.Certificate) (context.Context, error) {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		s := Session{
			ID:       p.ClientIdentifier,
			Username: p.Username,
			Password: p.Password,
			Cert:     cert,
		}

		ctx = NewContext(ctx, &s)
		if err := h.AuthConnect(ctx); err != nil {
			return ctx, err
		}
		// Copy back to the packet in case values are changed by Event handler.
		// This is specific to CONN, as only that package type has credentials.
		p.ClientIdentifier = s.ID
		p.Username = s.Username
		p.Password = s.Password
		return ctx, nil
	case *packets.PublishPacket:
		return ctx, h.AuthPublish(ctx, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		return ctx, h.AuthSubscribe(ctx, &p.Topics)
	default:
		return ctx, nil
	}
}

func notify(ctx context.Context, pkt packets.ControlPacket, h Handler) {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		h.Connect(ctx)
	case *packets.PublishPacket:
		h.Publish(ctx, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		h.Subscribe(ctx, &p.Topics)
	case *packets.UnsubscribePacket:
		h.Unsubscribe(ctx, &p.Topics)
	default:
		return
	}
}

func wrap(ctx context.Context, err error, dir direction) error {
	if err == io.EOF {
		return err
	}
	cid := unknownID
	if s, ok := FromContext(ctx); ok {
		cid = s.ID
	}
	switch dir {
	case up:
		return fmt.Errorf(errClient, cid, err.Error())
	case down:
		return fmt.Errorf(errBroker, cid, err.Error())
	default:
		return err
	}
}
