package mqtt

import (
	"fmt"
	"net"

	"github.com/google/uuid"
	"github.com/mainflux/mainflux/logger"
)

// Proxy is main MQTT proxy struct
type Proxy struct {
	host       string
	port       string
	targetHost string
	targetPort string
	sessions   map[string]*session
	logger     logger.Logger
}

// New will setup a new Proxy struct after parsing the options
func New(host, port, targetHost, targetPort string, logger logger.Logger) *Proxy {
	return &Proxy{
		host:       host,
		port:       port,
		targetHost: targetHost,
		targetPort: targetPort,
		logger:     logger,
		sessions:   make(map[string]*session),
	}
}

func (p *Proxy) accept(l net.Listener) {
	for {
		conn, err := l.Accept()
		if err != nil {
			p.logger.Warn(fmt.Sprintf("Accept error %s", err))
			//continue
		}

		p.logger.Info("Accepted new client")
		go p.handleConnection(conn)
	}
}

func (p *Proxy) handleConnection(inbound net.Conn) {
	addr := fmt.Sprintf("%s:%s", p.targetHost, p.targetPort)

	println("CONNECTING TO", addr)
	outbound, err := net.Dial("tcp", addr)
	if err != nil {
		p.logger.Warn(fmt.Sprintf("Cannot connect to remote broker %s", addr))
		return
	}

	uuid, err := uuid.NewRandom()
	if err != nil {
		return
	}

	println(uuid.String())

	s := newSession(uuid.String(), inbound, outbound, p.logger)
	p.sessions[s.id] = s
	s.stream()
}

// Proxy of the server, this will block.
func (p *Proxy) Proxy() error {
	addr := fmt.Sprintf("%s:%s", p.host, p.port)
	l, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	defer l.Close()

	// Acceptor loop
	p.accept(l)

	p.logger.Info("Server Exiting...")
	return nil
}
