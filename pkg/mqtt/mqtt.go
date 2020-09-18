package mqtt

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"io/ioutil"
	"net"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mainflux/pkg/errors"
	"github.com/mainflux/mproxy/pkg/session"
)

var (
	errCreateListener = errors.New("failed creating TLS listener")
	errParseRoot      = errors.New("failed to parse root certificate")
)

// Proxy is main MQTT proxy struct
type Proxy struct {
	address string
	target  string
	handler session.Handler
	logger  logger.Logger
	dialer  net.Dialer
	ca      string
	crt     string
	key     string
}

// New returns a new mqtt Proxy instance.
func New(address, target string, handler session.Handler, logger logger.Logger, ca, crt, key string) *Proxy {
	return &Proxy{
		address: address,
		target:  target,
		handler: handler,
		logger:  logger,
		ca:      ca,
		crt:     crt,
		key:     key,
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
	outbound, err := p.dialer.Dial("tcp", p.target)
	if err != nil {
		p.logger.Error("Cannot connect to remote broker " + p.target + " due to: " + err.Error())
		return
	}
	defer p.close(outbound)

	clientCert, err := session.ClientCert(inbound)
	if err != nil {
		p.logger.Error("Failed to get client certificate, reason: " + err.Error())
		return
	}

	s := session.New(inbound, outbound, p.handler, p.logger, clientCert)

	if err = s.Stream(); !errors.Contains(err, io.EOF) {
		p.logger.Warn("Broken connection for client: " + s.Client.ID + " with error: " + err.Error())
	}
}

// Listen of the server, this will block.
func (p Proxy) Listen() error {
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

func (p Proxy) certConfig() (tls.Config, error) {
	caCertPEM, err := ioutil.ReadFile(p.ca)
	if err != nil {
		return tls.Config{}, err
	}

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM(caCertPEM)
	if !ok {
		return tls.Config{}, errParseRoot
	}

	cert, err := tls.LoadX509KeyPair(p.crt, p.key)
	if err != nil {
		return tls.Config{}, err
	}
	return tls.Config{
		Certificates: []tls.Certificate{cert},
		ClientAuth:   tls.RequireAndVerifyClientCert,
		ClientCAs:    roots,
	}, nil
}

// ListenTLS - version of Listen with TLS encryption
func (p Proxy) ListenTLS() error {
	config, err := p.certConfig()

	if err != nil {
		return err
	}

	l, err := tls.Listen("tcp", p.address, &config)
	if err != nil {
		return errors.Wrap(errCreateListener, err)
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
