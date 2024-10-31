// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package session

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"golang.org/x/sync/errgroup"
)

type Direction int

const (
	Up Direction = iota
	Down
)

const unknownID = "unknown"

var (
	errBroker = "failed to proxy from MQTT client with id %s to MQTT broker with error: %s"
	errClient = "failed to proxy from MQTT broker to client with id %s with error: %s"
)

// Stream starts proxy between client and broker.
func Stream(ctx context.Context, in, out net.Conn, h Handler, ic Interceptor, cert x509.Certificate) error {
	s := Session{
		Cert: cert,
	}
	ctx = NewContext(ctx, &s)

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		return stream(ctx, Up, in, out, h, ic)
	})

	g.Go(func() error {
		return stream(ctx, Down, out, in, h, ic)
	})

	err := g.Wait()

	disconnectErr := h.Disconnect(ctx)

	return errors.Join(err, disconnectErr)
}

func stream(ctx context.Context, dir Direction, r, w net.Conn, h Handler, ic Interceptor) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// Read from one connection.
		pkt, err := packets.ReadPacket(r)
		if err != nil {
			return wrap(ctx, err, dir)
		}

		switch dir {
		case Up:
			if err = authorize(ctx, pkt, h); err != nil {
				return wrap(ctx, err, dir)
			}
		default:
			if p, ok := pkt.(*packets.PublishPacket); ok {
				topics := []string{p.TopicName}
				if err = h.AuthSubscribe(ctx, &topics); err != nil {
					pkt = packets.NewControlPacket(packets.Disconnect).(*packets.DisconnectPacket)
					if wErr := pkt.Write(w); wErr != nil {
						err = errors.Join(err, wErr)
					}
					return wrap(ctx, err, dir)
				}
			}
		}

		if ic != nil {
			pkt, err = ic.Intercept(ctx, pkt, dir)
			if err != nil {
				return wrap(ctx, err, dir)
			}
		}

		// Send to another.
		if err := pkt.Write(w); err != nil {
			return wrap(ctx, err, dir)
		}

		// Notify only for packets sent from client to broker (incoming packets).
		if dir == Up {
			if err := notify(ctx, pkt, h); err != nil {
				return wrap(ctx, err, dir)
			}
		}
	}
}

func authorize(ctx context.Context, pkt packets.ControlPacket, h Handler) error {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		s, ok := FromContext(ctx)
		if ok {
			s.ID = p.ClientIdentifier
			s.Username = p.Username
			s.Password = p.Password
		}

		ctx = NewContext(ctx, s)
		if err := h.AuthConnect(ctx); err != nil {
			return err
		}
		// Copy back to the packet in case values are changed by Event handler.
		// This is specific to CONN, as only that package type has credentials.
		p.ClientIdentifier = s.ID
		p.Username = s.Username
		p.Password = s.Password
		return nil
	case *packets.PublishPacket:
		return h.AuthPublish(ctx, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		return h.AuthSubscribe(ctx, &p.Topics)
	default:
		return nil
	}
}

func notify(ctx context.Context, pkt packets.ControlPacket, h Handler) error {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		return h.Connect(ctx)
	case *packets.PublishPacket:
		return h.Publish(ctx, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		return h.Subscribe(ctx, &p.Topics)
	case *packets.UnsubscribePacket:
		return h.Unsubscribe(ctx, &p.Topics)
	default:
		return nil
	}
}

func wrap(ctx context.Context, err error, dir Direction) error {
	if err == io.EOF {
		return err
	}
	cid := unknownID
	if s, ok := FromContext(ctx); ok {
		cid = s.ID
	}
	switch dir {
	case Up:
		return fmt.Errorf(errClient, cid, err.Error())
	case Down:
		return fmt.Errorf(errBroker, cid, err.Error())
	default:
		return err
	}
}
