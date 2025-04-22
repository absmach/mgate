// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/absmach/mgate"
	"github.com/absmach/mgate/pkg/common"
	"github.com/absmach/mgate/pkg/session"
	mptls "github.com/absmach/mgate/pkg/tls"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

const contentType = "application/json"

// ErrMissingAuthentication returned when no basic or Authorization header is set.
var ErrMissingAuthentication = errors.New("missing authorization")

func isWebSocketRequest(r *http.Request) bool {
	return strings.EqualFold(r.Header.Get("Connection"), "Upgrade") &&
		strings.EqualFold(r.Header.Get("Upgrade"), "websocket")
}

func (p Proxy) getUserPass(r *http.Request) (string, string, error) {
	username, password, ok := r.BasicAuth()
	var err error
	switch {
	case ok:
		return username, password, nil
	case len(r.URL.Query()["authorization"]) != 0:
		password = r.URL.Query()["authorization"][0]
		return username, password, nil
	case r.Header.Get("Authorization") != "":
		password = r.Header.Get("Authorization")
		return username, password, nil
	default:
		err = ErrMissingAuthentication
	}
	return username, password, err
}

func (p Proxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !strings.HasPrefix(r.URL.Path, p.config.PathPrefix) {
		http.NotFound(w, r)
		return
	}

	if p.bypassMatcher.ShouldBypass(r) {
		p.target.ServeHTTP(w, r)
		return
	}

	username, password, err := p.getUserPass(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	s := &session.Session{
		Password: []byte(password),
		Username: username,
	}

	if isWebSocketRequest(r) {
		p.handleWebSocket(w, r, s)
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

func (p *Proxy) handleWebSocket(w http.ResponseWriter, r *http.Request, s *session.Session) {
	headers := http.Header{}

	targetUrl := p.targetUrl
	targetUrl.Scheme = "ws"
	target := fmt.Sprintf("%s%s", targetUrl.String(), r.RequestURI)

	targetConn, _, err := websocket.DefaultDialer.Dial(target, headers)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	topic := r.URL.Path
	ctx := session.NewContext(context.Background(), s)
	if err := p.session.AuthConnect(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err := p.session.AuthSubscribe(ctx, &[]string{topic}); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}
	if err := p.session.Subscribe(ctx, &[]string{topic}); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	inConn, err := p.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Warn("WS Proxy failed to upgrade connection", slog.Any("error", err))
		return
	}
	defer inConn.Close()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		err := p.stream(ctx, topic, inConn, targetConn, true)
		_ = targetConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closed"))
		_ = targetConn.Close()
		return err
	})
	g.Go(func() error {
		err := p.stream(ctx, topic, targetConn, inConn, false)
		_ = inConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closed"))
		_ = inConn.Close()
		return err
	})

	gErr := g.Wait()
	if err := p.session.Unsubscribe(ctx, &[]string{topic}); err != nil {
		p.logger.Error("Unsubscribe failed", slog.String("topic", topic), slog.Any("error", err))
	}
	if gErr != nil {
		p.logger.Error("WS Proxy session terminated", slog.Any("error", gErr))
		return
	}
	p.logger.Info("WS Proxy session terminated", slog.Any("error", gErr))
}

func (p *Proxy) stream(ctx context.Context, topic string, src, dest *websocket.Conn, upstream bool) error {
	for {
		messageType, payload, err := src.ReadMessage()
		if err != nil {
			return handleStreamErr(err, upstream)
		}
		if upstream {
			if err := p.session.AuthPublish(ctx, &topic, &payload); err != nil {
				return err
			}
			if err := p.session.Publish(ctx, &topic, &payload); err != nil {
				return err
			}
		}
		if err := dest.WriteMessage(messageType, payload); err != nil {
			return err
		}
	}
}

func handleStreamErr(err error, upstream bool) error {
	if err == nil {
		return nil
	}

	if upstream && websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseNoStatusReceived) {
		return nil
	}
	if errors.Is(err, net.ErrClosed) {
		return nil
	}

	prefix := "downstream"
	if upstream {
		prefix = "upstream"
	}
	return fmt.Errorf("%s error: %w", prefix, err)
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
	config        mgate.Config
	targetUrl     *url.URL
	target        *httputil.ReverseProxy
	session       session.Handler
	logger        *slog.Logger
	wsUpgrader    websocket.Upgrader
	bypassMatcher *common.BypassMatcher
}

func NewProxy(config mgate.Config, handler session.Handler, logger *slog.Logger, allowedOrigins []string, bypassPaths []string) (Proxy, error) {
	targetUrl, err := url.Parse(config.Target)
	if err != nil {
		return Proxy{}, err
	}

	oc := common.NewOriginChecker(logger, allowedOrigins)
	wsUpgrader := websocket.Upgrader{CheckOrigin: oc.CheckOrigin}
	bypassMatcher, err := common.NewBypassMatcher(bypassPaths)
	if err != nil {
		return Proxy{}, err
	}

	return Proxy{
		config:        config,
		targetUrl:     targetUrl,
		target:        httputil.NewSingleHostReverseProxy(targetUrl),
		session:       handler,
		logger:        logger,
		wsUpgrader:    wsUpgrader,
		bypassMatcher: bypassMatcher,
	}, nil
}

func (p Proxy) Listen(ctx context.Context) error {
	l, err := net.Listen("tcp", p.config.Address)
	if err != nil {
		return err
	}

	if p.config.TLSConfig != nil {
		l = tls.NewListener(l, p.config.TLSConfig)
	}
	status := mptls.SecurityStatus(p.config.TLSConfig)

	p.logger.Info(fmt.Sprintf("HTTP proxy server started at %s%s with %s", p.config.Address, p.config.PathPrefix, status))

	var server http.Server
	g, ctx := errgroup.WithContext(ctx)

	mux := http.NewServeMux()
	mux.Handle(p.config.PathPrefix, p)
	server.Handler = mux

	g.Go(func() error {
		return server.Serve(l)
	})

	g.Go(func() error {
		<-ctx.Done()
		return server.Close()
	})
	if err := g.Wait(); err != nil {
		p.logger.Info(fmt.Sprintf("HTTP proxy server at %s%s with %s exiting with errors", p.config.Address, p.config.PathPrefix, status), slog.String("error", err.Error()))
	} else {
		p.logger.Info(fmt.Sprintf("HTTP proxy server at %s%s with %s exiting...", p.config.Address, p.config.PathPrefix, status))
	}
	return nil
}
