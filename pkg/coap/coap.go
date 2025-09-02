// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package coap

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"sync"
	"time"

	"github.com/absmach/mgate"
	"github.com/absmach/mgate/pkg/session"
	mptls "github.com/absmach/mgate/pkg/tls"
	gocoap "github.com/dustin/go-coap"
	"github.com/pion/dtls/v2"
	"golang.org/x/sync/errgroup"
)

const (
	bufferSize   uint64 = 1280
	startObserve uint32 = 0
)

var (
	ConnMap = make(map[string]*Conn)
	mutex   sync.Mutex
)

type Conn struct {
	clientAddr *net.UDPAddr
	serverConn *net.UDPConn
}

type Proxy struct {
	config  mgate.Config
	session session.Handler
	logger  *slog.Logger
}

func NewProxy(config mgate.Config, handler session.Handler, logger *slog.Logger) *Proxy {
	return &Proxy{
		config:  config,
		session: handler,
		logger:  logger,
	}
}

func (p *Proxy) proxyUDP(ctx context.Context, l *net.UDPConn) {
	buffer := make([]byte, bufferSize)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			n, clientAddr, err := l.ReadFromUDP(buffer)
			if err != nil {
				return
			}
			mutex.Lock()
			conn, ok := ConnMap[clientAddr.String()]
			if !ok {
				conn, err = p.newConn(clientAddr)
				if err != nil {
					p.logger.Error("Failed to create new connection", slog.Any("error", err))
					mutex.Unlock()
					return
				}
				ConnMap[clientAddr.String()] = conn
				go p.downUDP(l, conn)
			}
			mutex.Unlock()
			//nolint:contextcheck // upUDP does not need context
			p.upUDP(conn, buffer[:n])
		}
	}
}

func (p *Proxy) Listen(ctx context.Context) error {
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(p.config.Host, p.config.Port))
	if err != nil {
		p.logger.Error("Failed to resolve UDP address", slog.Any("error", err))
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	switch {
	case p.config.DTLSConfig != nil:
		l, err := dtls.Listen("udp", addr, p.config.DTLSConfig)
		if err != nil {
			return err
		}
		defer l.Close()

		g.Go(func() error {
			p.proxyDTLS(ctx, l)
			return nil
		})

		g.Go(func() error {
			<-ctx.Done()
			return l.Close()
		})

	default:
		l, err := net.ListenUDP("udp", addr)
		if err != nil {
			return err
		}
		defer l.Close()

		g.Go(func() error {
			p.proxyUDP(ctx, l)
			return nil
		})

		g.Go(func() error {
			<-ctx.Done()
			return l.Close()
		})
	}

	status := mptls.SecurityStatus(p.config.DTLSConfig)
	p.logger.Info(fmt.Sprintf("COAP proxy server started at %s  with %s", net.JoinHostPort(p.config.Host, p.config.Port), status))

	if err := g.Wait(); err != nil {
		p.logger.Info(fmt.Sprintf("COAP proxy server at %s exiting with errors", net.JoinHostPort(p.config.Host, p.config.Port)), slog.String("error", err.Error()))
	} else {
		p.logger.Info(fmt.Sprintf("COAP proxy server at %s exiting...", net.JoinHostPort(p.config.Host, p.config.Port)))
	}
	return nil
}

func (p *Proxy) newConn(clientAddr *net.UDPAddr) (*Conn, error) {
	conn := new(Conn)
	conn.clientAddr = clientAddr
	addr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(p.config.TargetHost, p.config.TargetPort))
	if err != nil {
		return nil, err
	}
	t, err := net.DialUDP("udp", nil, addr)
	if err != nil {
		return nil, err
	}
	conn.serverConn = t
	return conn, nil
}

func (p *Proxy) upUDP(conn *Conn, buffer []byte) {
	p.handleCoAPMessage(buffer)
	_, err := conn.serverConn.Write(buffer)
	if err != nil {
		return
	}
}

func (p *Proxy) downUDP(l *net.UDPConn, conn *Conn) {
	buffer := make([]byte, bufferSize)
	for {
		err := conn.serverConn.SetReadDeadline(time.Now().Add(10 * time.Second))
		if err != nil {
			return
		}
		n, err := conn.serverConn.Read(buffer)
		if err != nil {
			p.close(conn)
			return
		}
		_, err = l.WriteToUDP(buffer[:n], conn.clientAddr)
		if err != nil {
			return
		}
	}
}

func (p *Proxy) close(conn *Conn) {
	mutex.Lock()
	defer mutex.Unlock()
	delete(ConnMap, conn.clientAddr.String())
	conn.serverConn.Close()
}

func (p *Proxy) proxyDTLS(ctx context.Context, l net.Listener) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			conn, err := l.Accept()
			if err != nil {
				p.logger.Warn("Accept error " + err.Error())
				continue
			}
			p.logger.Info("Accepted new client")
			//nolint:contextcheck // p.handleDTLS does not need context
			go p.handleDTLS(conn)
		}
	}
}

func (p *Proxy) handleDTLS(inbound net.Conn) {
	outboundAddr, err := net.ResolveUDPAddr("udp", net.JoinHostPort(p.config.Host, p.config.Port))
	if err != nil {
		return
	}

	outbound, err := net.DialUDP("udp", nil, outboundAddr)
	if err != nil {
		p.logger.Error("Cannot connect to remote broker " + net.JoinHostPort(p.config.Host, p.config.Port) + " due to: " + err.Error())
		return
	}

	go p.dtlsUp(outbound, inbound)
	go p.dtlsDown(inbound, outbound)
}

func (p *Proxy) dtlsUp(outbound *net.UDPConn, inbound net.Conn) {
	buffer := make([]byte, bufferSize)
	for {
		n, err := inbound.Read(buffer)
		if err != nil {
			return
		}
		p.handleCoAPMessage(buffer[:n])

		_, err = outbound.Write(buffer[:n])
		if err != nil {
			slog.Error("Failed to write to server", slog.Any("err", err))
		}
	}
}

func (p *Proxy) dtlsDown(inbound net.Conn, outbound *net.UDPConn) {
	buffer := make([]byte, bufferSize)
	for {
		err := outbound.SetReadDeadline(time.Now().Add(1 * time.Minute))
		if err != nil {
			return
		}
		n, err := outbound.Read(buffer)
		defer outbound.Close()
		if err != nil {
			return
		}

		_, err = inbound.Write(buffer[:n])
		defer inbound.Close()
		if err != nil {
			return
		}
	}
}

func (p *Proxy) handleCoAPMessage(buffer []byte) {
	msg, err := gocoap.ParseMessage(buffer)
	if err != nil {
		p.logger.Error("Failed to parse message", slog.Any("error", err))
		return
	}

	token := msg.Token
	path := msg.Path()
	ctx := session.NewContext(context.Background(), &session.Session{Password: token})

	switch msg.Code {
	case gocoap.POST:
		if err := p.session.AuthConnect(ctx); err != nil {
			return
		}
		if err := p.session.AuthPublish(ctx, &path[0], &msg.Payload); err != nil {
			return
		}
		if err := p.session.Publish(ctx, &path[0], &msg.Payload); err != nil {
			return
		}
	case gocoap.GET:
		if err := p.session.AuthConnect(ctx); err != nil {
			return
		}
		if msg.Option(gocoap.Observe) == startObserve {
			if err := p.session.AuthSubscribe(ctx, &path); err != nil {
				return
			}
			if err := p.session.Subscribe(ctx, &path); err != nil {
				return
			}
		}
	}
}
