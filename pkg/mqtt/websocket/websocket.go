// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package websocket

import (
	"context"
	"crypto/tls"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/absmach/mproxy"
	"github.com/absmach/mproxy/pkg/session"
	mptls "github.com/absmach/mproxy/pkg/tls"
	"github.com/gorilla/websocket"
)

// Proxy represents WS Proxy.
type Proxy struct {
	config      mproxy.Config
	path        string
	scheme      string
	handler     session.Handler
	interceptor session.Interceptor
	logger      *slog.Logger
}

// New - creates new WS proxy
func New(config mproxy.Config, path, scheme string, handler session.Handler, interceptor session.Interceptor, logger *slog.Logger) *Proxy {
	return &Proxy{
		config:      config,
		path:        path,
		scheme:      scheme,
		handler:     handler,
		interceptor: interceptor,
		logger:      logger,
	}
}

var upgrader = websocket.Upgrader{
	// Timeout for WS upgrade request handshake
	HandshakeTimeout: 10 * time.Second,
	// Paho JS client expecting header Sec-WebSocket-Protocol:mqtt in Upgrade response during handshake.
	Subprotocols: []string{"mqttv3.1", "mqtt"},
	// Allow CORS
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handler - proxies WS traffic.
func (p Proxy) Handler() http.Handler {
	return p.handle()
}

func (p Proxy) handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cconn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			p.logger.Error("Error upgrading connection", slog.Any("error", err))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		go p.pass(r.Context(), cconn)
	})
}

func (p Proxy) pass(ctx context.Context, in *websocket.Conn) {
	defer in.Close()

	websocketURL := url.URL{
		Scheme: p.scheme,
		Host:   p.config.Target,
		Path:   p.path,
	}

	dialer := &websocket.Dialer{
		Subprotocols: []string{"mqtt"},
	}
	srv, _, err := dialer.Dial(websocketURL.String(), nil)
	if err != nil {
		p.logger.Error("Unable to connect to broker", slog.Any("error", err))
		return
	}

	errc := make(chan error, 1)
	inboundConn := newConn(in)
	outboundConn := newConn(srv)

	defer inboundConn.Close()
	defer outboundConn.Close()

	clientCert, err := mptls.ClientCert(in.UnderlyingConn())
	if err != nil {
		p.logger.Error("Failed to get client certificate", slog.Any("error", err))
		return
	}

	err = session.Stream(ctx, inboundConn, outboundConn, p.handler, p.interceptor, clientCert)
	errc <- err
	p.logger.Warn("Broken connection for client", slog.Any("error", err))
}

func (p Proxy) Listen() error {
	tlsCfg, secure, err := mptls.LoadTLSCfg(p.config.ServerCAFile, p.config.ClientCAFile, p.config.CertFile, p.config.KeyFile)
	if err != nil {
		return err
	}

	l, err := net.Listen("tcp", p.config.Address)
	if err != nil {
		return err
	}
	defer l.Close()

	if secure > mptls.WithoutTLS {
		l = tls.NewListener(l, tlsCfg)
	}

	p.logger.Info(fmt.Sprintf("http proxy server started %s", secure.String()))

	var server http.Server

	return server.Serve(l)
}
