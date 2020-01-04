package mqtt

import (
	"net"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/mainflux/mainflux/logger"
)

const (
	up direction = iota
	down
)

type direction int

type session struct {
	id       string
	logger   logger.Logger
	inbound  net.Conn
	outbound net.Conn
}

func newSession(uuid string, inbound, outbound net.Conn, logger logger.Logger) *session {
	return &session{
		id:       uuid,
		logger:   logger,
		inbound:  inbound,
		outbound: outbound,
	}
}

func (s *session) stream() error {
	// In parallel read from client, send to broker
	// and read from broker, send to client
	errs := make(chan error, 2)

	go s.streamUnidir(up, s.inbound, s.outbound, errs)
	go s.streamUnidir(down, s.outbound, s.inbound, errs)

	return <-errs
}

func (s *session) streamUnidir(dir direction, r, w net.Conn, errs chan error) {
	for {
		// Read from one connection
		pkt, err := packets.ReadPacket(r)
		if err != nil {
			errs <- err
			return
		}

		// Send to another
		if err := pkt.Write(w); err != nil {
			errs <- err
			return
		}
	}
}
