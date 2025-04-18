// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package mqtt

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"log/slog"
	"net"

	"github.com/absmach/mgate"
	"github.com/absmach/mgate/pkg/session"
	mptls "github.com/absmach/mgate/pkg/tls"
	"golang.org/x/sync/errgroup"
)

// Proxy is main MQTT proxy struct.
type Proxy struct {
	config        mgate.Config
	handler       session.Handler
	beforeHandler session.Interceptor
	afterHandler  session.Interceptor
	logger        *slog.Logger
	dialer        net.Dialer
}

// New returns a new MQTT Proxy instance.
func New(config mgate.Config, handler session.Handler, beforeHandler, afterHandler session.Interceptor, logger *slog.Logger) *Proxy {
	return &Proxy{
		config:        config,
		handler:       handler,
		logger:        logger,
		beforeHandler: beforeHandler,
		afterHandler:  afterHandler,
	}
}

func (p Proxy) accept(ctx context.Context, l net.Listener) {
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
			go p.handle(ctx, conn)
		}
	}
}

func (p Proxy) handle(ctx context.Context, inbound net.Conn) {
	defer p.close(inbound)
	targetAddress := net.JoinHostPort(p.config.TargetHost, p.config.TargetPort)
	outbound, err := p.dialer.Dial("tcp", targetAddress)
	if err != nil {
		p.logger.Error("Cannot connect to remote broker " + targetAddress + " due to: " + err.Error())
		return
	}
	defer p.close(outbound)

	clientCert, err := mptls.ClientCert(inbound)
	if err != nil {
		p.logger.Error("Failed to get client certificate: " + err.Error())
		return
	}

	if err = session.Stream(ctx, inbound, outbound, p.handler, p.beforeHandler, p.afterHandler, clientCert); err != io.EOF {
		p.logger.Warn(err.Error())
	}
}

// Listen of the server, this will block.
func (p Proxy) Listen(ctx context.Context) error {
	listenAddress := net.JoinHostPort(p.config.Host, p.config.Port)
	l, err := net.Listen("tcp", listenAddress)
	if err != nil {
		return err
	}

	if p.config.TLSConfig != nil {
		l = tls.NewListener(l, p.config.TLSConfig)
	}
	status := mptls.SecurityStatus(p.config.TLSConfig)
	p.logger.Info(fmt.Sprintf("MQTT proxy server started at %s  with %s", listenAddress, status))
	g, ctx := errgroup.WithContext(ctx)

	// Acceptor loop
	g.Go(func() error {
		p.accept(ctx, l)
		return nil
	})

	g.Go(func() error {
		<-ctx.Done()
		return l.Close()
	})
	if err := g.Wait(); err != nil {
		p.logger.Info(fmt.Sprintf("MQTT proxy server at %s with %s exiting with errors", listenAddress, status), slog.String("error", err.Error()))
	} else {
		p.logger.Info(fmt.Sprintf("MQTT proxy server at %s with %s exiting...", listenAddress, status))
	}
	return nil
}

func (p Proxy) close(conn net.Conn) {
	if err := conn.Close(); err != nil {
		p.logger.Warn(fmt.Sprintf("Error closing connection %s", err.Error()))
	}
}
