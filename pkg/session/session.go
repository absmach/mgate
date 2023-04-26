package session

import (
	"context"
	"crypto/x509"
	"errors"
	"io"
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/mainflux/mproxy/pkg/logger"
)

const (
	up direction = iota
	down
)

var (
	errBroker = "failed proxying from MQTT client to MQTT broker"
	errClient = "failed proxying from MQTT broker to MQTT client"
)

type direction int

// Session represents MQTT Proxy session between client and broker.
type Session struct {
	logger   logger.Logger
	inbound  net.Conn
	outbound net.Conn
	handler  Handler
	cert     x509.Certificate
	context.Context
}

// New creates a new Session.
func New(ctx context.Context, inbound, outbound net.Conn, handler Handler, logger logger.Logger, cert x509.Certificate) *Session {
	return &Session{
		logger:   logger,
		inbound:  inbound,
		outbound: outbound,
		handler:  handler,
		cert:     cert,
		Context:  ctx,
	}
}

// Stream starts proxying traffic between client and broker.
func (s *Session) Stream() error {
	// In parallel read from client, send to broker
	// and read from broker, send to client.
	errs := make(chan error, 2)

	go s.stream(up, s.inbound, s.outbound, errs)
	go s.stream(down, s.outbound, s.inbound, errs)

	// Handle whichever error happens first.
	// The other routine won't be blocked when writing
	// to the errors channel because it is buffered.
	err := <-errs

	s.handler.Disconnect(s.Context)
	return err
}

func (s *Session) stream(dir direction, r, w net.Conn, errs chan error) {
	for {
		// Read from one connection
		pkt, err := packets.ReadPacket(r)
		if err != nil {
			errs <- wrap(err, dir)
			return
		}

		if dir == up {
			if err := s.authorize(s.Context, pkt); err != nil {
				errs <- wrap(err, dir)
				return
			}
		}

		// Send to another
		if err := pkt.Write(w); err != nil {
			errs <- wrap(err, dir)
			return
		}

		if dir == up {
			s.notify(s.Context, pkt)
		}
	}
}

func (s *Session) authorize(ctx context.Context, pkt packets.ControlPacket) error {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		var c Client
		c.ID = p.ClientIdentifier
		c.Username = p.Username
		c.Password = p.Password
		c.Cert = s.cert
		s.Context = c.ToContext(ctx)
		ctx = s.Context

		if err := s.handler.AuthConnect(ctx); err != nil {
			return err
		}
		// Copy back to the packet in case values are changed by Event handler.
		// This is specific to CONN, as only that package type has credentials.
		if err := c.FromContext(s.Context); err != nil {
			p.ClientIdentifier = c.ID
			p.Username = c.Username
			p.Password = c.Password
		}
		return nil
	case *packets.PublishPacket:
		return s.handler.AuthPublish(ctx, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		return s.handler.AuthSubscribe(ctx, &p.Topics)
	default:
		return nil
	}
}

func (s *Session) notify(ctx context.Context, pkt packets.ControlPacket) {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		s.handler.Connect(ctx)
	case *packets.PublishPacket:
		s.handler.Publish(ctx, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		s.handler.Subscribe(ctx, &p.Topics)
	case *packets.UnsubscribePacket:
		s.handler.Unsubscribe(ctx, &p.Topics)
	default:
		return
	}
}

func wrap(err error, dir direction) error {
	if err == io.EOF {
		return err
	}
	switch dir {
	case up:
		return errors.New(errClient + err.Error())
	case down:
		return errors.New(errBroker + err.Error())
	default:
		return err
	}
}
