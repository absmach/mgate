package mqtt

import (
	"fmt"
	"net"
	"sync"

	"github.com/eclipse/paho.mqtt.golang/packets"
	"github.com/mainflux/mainflux/logger"
)

const (
	up Direction = iota
	down
)

type Direction int

type session struct {
	id       string
	wg       sync.WaitGroup
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

func (s *session) stream() {

	// In parallel reas from client, send to broker
	// and read from broker, send to client
	s.wg.Add(2)
	go s.streamUnidir(up, s.inbound, s.outbound)
	go s.streamUnidir(down, s.outbound, s.inbound)
	s.wg.Wait()

	s.logger.Info(fmt.Sprintf("Session %s closed: %s", s.id, s.outbound.LocalAddr().String()))
}

func (s *session) streamUnidir(dir Direction, r, w net.Conn) error {

	// Read from one connection
	pkt, err := packets.ReadPacket(r)
	if err != nil {
		return err
	}
	
	// Send to another
	if err := pkt.Write(w); err != nil {
		return err
	}

	return nil
}

