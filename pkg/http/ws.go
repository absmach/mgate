// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"

	"github.com/absmach/mgate/pkg/session"
	"github.com/gorilla/websocket"
	"golang.org/x/sync/errgroup"
)

const (
	upstreamDesc   = "from mGate Proxy to websocket server"
	downStreamDesc = "from websocket server to mGate Proxy"
)

func (p *Proxy) handleWebSocket(w http.ResponseWriter, r *http.Request, s *session.Session) {
	topic := r.URL.Path
	ctx := session.NewContext(context.Background(), s)
	if err := p.session.AuthConnect(ctx); err != nil {
		encodeError(w, http.StatusUnauthorized, err)
		return
	}
	if err := p.session.AuthSubscribe(ctx, &[]string{topic}); err != nil {
		encodeError(w, http.StatusUnauthorized, err)
		return
	}
	if err := p.session.Subscribe(ctx, &[]string{topic}); err != nil {
		encodeError(w, http.StatusBadRequest, err)
		return
	}

	header := http.Header{}

	if auth := r.Header.Get(authzHeaderKey); auth != "" {
		header.Set(authzHeaderKey, auth)
	}

	target := fmt.Sprintf("%s://%s:%s%s", wsScheme(p.config.TargetProtocol), p.config.TargetHost, p.config.TargetPort, r.URL.RequestURI())

	targetConn, _, err := websocket.DefaultDialer.Dial(target, header)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	defer targetConn.Close()

	inConn, err := p.wsUpgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Warn("WS Proxy failed to upgrade connection", slog.Any("error", err))
		return
	}
	defer inConn.Close()

	g, ctx := errgroup.WithContext(ctx)

	g.Go(func() error {
		upstream := true
		err := p.stream(ctx, topic, inConn, targetConn, upstream)
		if err := targetConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closed")); err != nil {
			p.logger.Warn(fmt.Sprintf("failed to send close connection %s", getPrefix(upstream)), slog.Any("error", err))
		}
		if err := targetConn.Close(); err != nil {
			p.logger.Warn("failed to send close connection to websocket server", slog.Any("error", err))
		}
		return err
	})
	g.Go(func() error {
		upstream := false
		err := p.stream(ctx, topic, targetConn, inConn, upstream)
		if err := inConn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, "client closed")); err != nil {
			p.logger.Warn(fmt.Sprintf("failed to send close connection %s", getPrefix(upstream)), slog.Any("error", err))
		}
		if err := inConn.Close(); err != nil {
			p.logger.Warn("failed to send close connection to client", slog.Any("error", err))
		}
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
	p.logger.Info("WS Proxy session terminated")
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
	return fmt.Errorf("%s error: %w", getPrefix(upstream), err)
}

func getPrefix(upstream bool) string {
	prefix := downStreamDesc
	if upstream {
		prefix = upstreamDesc
	}
	return prefix
}

func wsScheme(scheme string) string {
	switch scheme {
	case "http":
		return "ws"
	case "https":
		return "wss"
	default:
		return scheme
	}
}
