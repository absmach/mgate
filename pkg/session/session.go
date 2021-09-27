package session

import (
	"crypto/x509"
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
)

const (
	Upstream Direction = iota
	Downstream
)

var (
	errBroker = errors.New("failed proxying from MQTT client to MQTT broker")
	errClient = errors.New("failed proxying from MQTT broker to MQTT client")
)

type Direction int

// Session represents MQTT Proxy session between client and broker.
type Session struct {
	logger      logger.Logger
	inbound     net.Conn
	outbound    net.Conn
	handler     Handler
	interceptor Interceptor
	Client      Client
}

// New creates a new Session.
func New(inbound, outbound net.Conn, handler Handler, interceptor Interceptor, logger logger.Logger, cert x509.Certificate) *Session {
	return &Session{
		logger:      logger,
		inbound:     inbound,
		outbound:    outbound,
		handler:     handler,
		interceptor: interceptor,
		Client: Client{
			Cert: cert,
		},
	}
}

// Stream starts proxying traffic between client and broker.
func (s *Session) Stream() error {
	// In parallel read from client, send to broker
	// and read from broker, send to client.
	errs := make(chan error, 2)

	go s.stream(Upstream, s.inbound, s.outbound, errs)
	go s.stream(Downstream, s.outbound, s.inbound, errs)

	// Handle whichever error happens first.
	// The other routine won't be blocked when writing
	// to the errors channel because it is buffered.
	err := <-errs

	s.handler.Disconnect(&s.Client)
	return err
}

func (s *Session) stream(dir Direction, r, w net.Conn, errs chan error) {
	for {
		// Read from one connection
		pkt, err := packets.ReadPacket(r)
		if err != nil {
			errs <- wrap(err, dir)
			return
		}

		if dir == Upstream {
			if err := s.authorize(pkt); err != nil {
				errs <- wrap(err, dir)
				return
			}
		}

		pkt = s.intercept(pkt, dir)

		// Send to another
		if err := pkt.Write(w); err != nil {
			errs <- wrap(err, dir)
			return
		}

		if dir == Upstream {
			s.notify(pkt)
		}
	}
}

func (s *Session) authorize(pkt packets.ControlPacket) error {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		s.Client.ID = p.ClientIdentifier
		s.Client.Username = p.Username
		s.Client.Password = p.Password
		if err := s.handler.AuthConnect(&s.Client); err != nil {
			return err
		}
		// Copy back to the packet in case values are changed by Event handler.
		// This is specific to CONN, as only that package type has credentials.
		p.ClientIdentifier = s.Client.ID
		p.Username = s.Client.Username
		p.Password = s.Client.Password
		return nil
	case *packets.PublishPacket:
		return s.handler.AuthPublish(&s.Client, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		return s.handler.AuthSubscribe(&s.Client, &p.Topics)
	default:
		return nil
	}
}

func (s *Session) intercept(pkt packets.ControlPacket, dir Direction) packets.ControlPacket {
	if s.interceptor == nil {
		return pkt
	}
	npkt, err := s.interceptor.Intercept(pkt, &s.Client, dir)
	if err != nil {
		return pkt
	}
	if npkt == nil {
		return pkt
	}
	return npkt
}

func (s *Session) notify(pkt packets.ControlPacket) {
	switch p := pkt.(type) {
	case *packets.ConnectPacket:
		s.handler.Connect(&s.Client)
	case *packets.PublishPacket:
		s.handler.Publish(&s.Client, &p.TopicName, &p.Payload)
	case *packets.SubscribePacket:
		s.handler.Subscribe(&s.Client, &p.Topics)
	case *packets.UnsubscribePacket:
		s.handler.Unsubscribe(&s.Client, &p.Topics)
	default:
		return
	}
}

func wrap(err error, dir Direction) error {
	switch dir {
	case Upstream:
		return errors.Wrap(errClient, err)
	case Downstream:
		return errors.Wrap(errBroker, err)
	default:
		return err
	}
}
