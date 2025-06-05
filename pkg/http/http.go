// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/absmach/mgate"
	"github.com/absmach/mgate/pkg/session"
	mptls "github.com/absmach/mgate/pkg/tls"
	"github.com/absmach/mgate/pkg/transport"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

const (
	contentType      = "application/json"
	authzQueryKey    = "authorization"
	authzHeaderKey   = "Authorization"
	connHeaderKey    = "Connection"
	connHeaderVal    = "upgrade"
	upgradeHeaderKey = "Upgrade"
	upgradeHeaderVal = "websocket"
)

type Checker interface {
	Check(r *http.Request) error
}

func isWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get(connHeaderKey), connHeaderVal) &&
		strings.EqualFold(r.Header.Get(upgradeHeaderKey), upgradeHeaderVal)
}

func (p Proxy) getUserPass(r *http.Request) (string, string) {
	username, password, ok := r.BasicAuth()
	switch {
	case ok:
		return username, password
	case r.URL.Query().Get(authzQueryKey) != "":
		password = r.URL.Query().Get(authzQueryKey)
		return username, password
	case r.Header.Get(authzHeaderKey) != "":
		password = r.Header.Get(authzHeaderKey)
		return username, password
	}
	return username, password
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, transport.AddSuffixSlash(p.config.PathPrefix+p.config.TargetPath)) {
		http.NotFound(w, r)
		return
	}

	r.URL.Path = strings.TrimPrefix(r.URL.Path, p.config.PathPrefix)

	if err := p.bypass.Check(r); err == nil {
		p.target.ServeHTTP(w, r)
		return
	}

	username, password := p.getUserPass(r)
	s := &session.Session{
		Password: []byte(password),
		Username: username,
	}

	if isWebSocketRequest(r) {
		p.handleWebSocket(w, r, s) //nolint:contextcheck // handleWebSocket does not need context
		return
	}

	ctx := session.NewContext(r.Context(), s)
	payload, err := io.ReadAll(r.Body)
	if err != nil {
		encodeError(w, http.StatusBadRequest, err)
		p.logger.Error("Failed to read body", slog.Any("error", err))
		return
	}
	if err := r.Body.Close(); err != nil {
		encodeError(w, http.StatusInternalServerError, err)
		p.logger.Error("Failed to close body", slog.Any("error", err))
		return
	}

	// r.Body is reset to ensure it can be safely copied by httputil.ReverseProxy.
	// no close method is required since NopClose Close() always returns nill.
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	if err := p.session.AuthConnect(ctx); err != nil {
		encodeError(w, http.StatusUnauthorized, err)
		p.logger.Error("Failed to authorize connect", slog.Any("error", err))
		return
	}
	if err := p.session.Publish(ctx, &r.RequestURI, &payload); err != nil {
		encodeError(w, http.StatusBadRequest, err)
		p.logger.Error("Failed to publish", slog.Any("error", err))
		return
	}

	p.target.ServeHTTP(w, r)
}

func checkOrigin(allowedOrigins []string) func(r *http.Request) bool {
	oc := NewOriginChecker(allowedOrigins)
	return func(r *http.Request) bool {
		return oc.Check(r) == nil
	}
}

func encodeError(w http.ResponseWriter, defStatusCode int, err error) {
	hpe, ok := err.(HTTPProxyError)
	if !ok {
		hpe = NewHTTPProxyError(defStatusCode, err)
	}
	w.WriteHeader(hpe.StatusCode())
	w.Header().Set("Content-Type", contentType)
	if err := json.NewEncoder(w).Encode(err); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Proxy represents HTTP Proxy.
type Proxy struct {
	config     mgate.Config
	target     *httputil.ReverseProxy
	session    session.Handler
	logger     *slog.Logger
	wsUpgrader websocket.Upgrader
	bypass     Checker
}

func NewProxy(config mgate.Config, handler session.Handler, logger *slog.Logger, allowedOrigins []string, bypassPaths []string) (Proxy, error) {
	targetUrl := &url.URL{
		Scheme: config.TargetProtocol,
		Host:   net.JoinHostPort(config.TargetHost, config.TargetPort),
	}

	bpc, err := NewBypassChecker(bypassPaths)
	if err != nil {
		return Proxy{}, err
	}

	wsUpgrader := websocket.Upgrader{CheckOrigin: checkOrigin(allowedOrigins)}

	return Proxy{
		config:     config,
		target:     httputil.NewSingleHostReverseProxy(targetUrl),
		session:    handler,
		logger:     logger,
		wsUpgrader: wsUpgrader,
		bypass:     bpc,
	}, nil
}

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

	p.logger.Info(fmt.Sprintf("HTTP proxy server started at %s%s with %s", listenAddress, p.config.PathPrefix, status))

	var server http.Server
	g, ctx := errgroup.WithContext(ctx)

	mux := http.NewServeMux()

	mux.Handle(transport.AddSuffixSlash(p.config.PathPrefix), p)
	server.Handler = mux

	g.Go(func() error {
		return server.Serve(l)
	})

	g.Go(func() error {
		<-ctx.Done()
		return server.Close()
	})
	if err := g.Wait(); err != nil {
		p.logger.Info(fmt.Sprintf("HTTP proxy server at %s%s with %s exiting with errors", listenAddress, p.config.PathPrefix, status), slog.String("error", err.Error()))
	} else {
		p.logger.Info(fmt.Sprintf("HTTP proxy server at %s%s with %s exiting...", listenAddress, p.config.PathPrefix, status))
	}
	return nil
}
