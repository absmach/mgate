package mqtt

import (
	"errors"
	"fmt"
	"io"
	"net"
	"time"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/pkg/session"
)

var errNotTCPConn = errors.New("not a TCP connection")

// Proxy is main MQTT proxy struct
type Proxy struct {
	address   string
	target    string
	handler   session.Handler
	logger    logger.Logger
	keepAlive time.Duration
}

// New returns a new mqtt Proxy instance.
func New(address, target string, keepAlive time.Duration, handler session.Handler, logger logger.Logger) *Proxy {
	return &Proxy{
		address:   address,
		target:    target,
		handler:   handler,
		logger:    logger,
		keepAlive: keepAlive,
	}
}

func (p Proxy) accept(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			p.logger.Warn("Accept error " + err.Error())
			continue
		}

		p.logger.Info("Accepted new client")
		go p.handle(conn)
	}
}

func (p Proxy) handle(inbound net.Conn) {
	defer p.close(inbound)
	outbound, err := net.Dial("tcp", p.target)
	if err != nil {
		p.logger.Error("Cannot connect to remote broker " + p.target + " due to: " + err.Error())
		return
	}
	defer p.close(outbound)

	if err := setKeepAlive(inbound, p.keepAlive); err != nil {
		p.logger.Error("Client conenction error: " + err.Error())
		return
	}
	if err := setKeepAlive(outbound, p.keepAlive); err != nil {
		p.logger.Error("Broker connection error: " + err.Error())
		return
	}

	s := session.New(inbound, outbound, p.handler, p.logger)

	if err = s.Stream(); err != io.EOF {
		p.logger.Warn("Broken connection for client: " + s.Client.ID + " with error: " + err.Error())
	}
}

// Proxy of the server, this will block.
func (p Proxy) Proxy() error {
	l, err := net.Listen("tcp", p.address)
	if err != nil {
		return err
	}
	defer l.Close()

	// Acceptor loop
	p.accept(l)

	p.logger.Info("Server Exiting...")
	return nil
}

func (p Proxy) close(conn net.Conn) {
	if err := conn.Close(); err != nil {
		p.logger.Warn(fmt.Sprintf("Error closing connection %s", err.Error()))
	}
}

func setKeepAlive(conn net.Conn, period time.Duration) error {
	tcpConn, ok := conn.(*net.TCPConn)
	if !ok {
		return errNotTCPConn
	}
	if err := tcpConn.SetKeepAlive(true); err != nil {
		return err
	}
	return tcpConn.SetKeepAlivePeriod(period)
}
