// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package coap

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"

	"github.com/absmach/mproxy"
	"github.com/absmach/mproxy/pkg/session"
	"github.com/plgd-dev/go-coap/v3/dtls"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/net"
	"github.com/plgd-dev/go-coap/v3/options"
	"github.com/plgd-dev/go-coap/v3/udp"
)

type Proxy struct {
	config  mproxy.Config
	session session.Handler
	logger  *slog.Logger
}

func NewProxy(config mproxy.Config, handler session.Handler, logger *slog.Logger) *Proxy {
	return &Proxy{
		config:  config,
		session: handler,
		logger:  logger,
	}
}

func sendErrorMessage(cc mux.Conn, token []byte, err error, code codes.Code) error {
	m := cc.AcquireMessage(cc.Context())
	defer cc.ReleaseMessage(m)
	m.SetCode(code)
	m.SetBody(bytes.NewReader(([]byte)(err.Error())))
	m.SetToken(token)
	m.SetContentFormat(message.TextPlain)
	return cc.WriteMessage(m)
}

func (p *Proxy) postUpstream(cc mux.Conn, req *mux.Message, token []byte) error {
	format, err := req.ContentFormat()
	if err != nil {
		return err
	}
	path, err := req.Options().Path()
	if err != nil {
		return err
	}

	targetConn, err := udp.Dial(p.config.Target)
	if err != nil {
		return err
	}
	defer targetConn.Close()
	pm, err := targetConn.Post(cc.Context(), path, format, req.Body(), req.Options()...)
	if err != nil {
		return err
	}
	pm.SetToken(token)
	return cc.WriteMessage(pm)
}

func (p *Proxy) getUpstream(cc mux.Conn, req *mux.Message, token []byte) error {
	path, err := req.Options().Path()
	if err != nil {
		return err
	}

	targetConn, err := udp.Dial(p.config.Target)
	if err != nil {
		return err
	}
	defer targetConn.Close()
	pm, err := targetConn.Get(cc.Context(), path, req.Options()...)
	if err != nil {
		return err
	}
	pm.SetToken(token)
	return cc.WriteMessage(pm)
}

func (p *Proxy) observeUpstream(ctx context.Context, cc mux.Conn, opts []message.Option, token []byte, path string) {
	targetConn, err := udp.Dial(p.config.Target)
	if err != nil {
		if err := sendErrorMessage(cc, token, err, codes.BadGateway); err != nil {
			p.logger.Error("cannot send error response: %v", err)
		}
	}
	defer targetConn.Close()
	doneObserving := make(chan struct{})

	obs, err := targetConn.Observe(ctx, path, func(req *pool.Message) {
		req.SetToken(token)
		if err := cc.WriteMessage(req); err != nil {
			if err := sendErrorMessage(cc, token, err, codes.BadGateway); err != nil {
				p.logger.Error(err.Error())
			}
			p.logger.Error(err.Error())
		}
		if req.Code() == codes.NotFound {
			close(doneObserving)
		}
	}, opts...)
	if err != nil {
		if err := sendErrorMessage(cc, token, err, codes.BadGateway); err != nil {
			p.logger.Error("cannot send error response: %v", err)
		}
	}

	select {
	case <-doneObserving:
		if err := obs.Cancel(ctx); err != nil {
			p.logger.Error("failed to cancel observation: %v", err)
		}
	case <-ctx.Done():
		return
	}
}

func (p *Proxy) handler(w mux.ResponseWriter, r *mux.Message) {
	tok, err := r.Options().GetBytes(message.URIQuery)
	if err != nil {
		if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.Unauthorized); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
	ctx := session.NewContext(r.Context(), &session.Session{Password: tok})
	if err := p.session.AuthConnect(ctx); err != nil {
		if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.Unauthorized); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
	path, err := r.Options().Path()
	if err != nil {
		if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadOption); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
	switch r.Code() {
	case codes.GET:
		obs, err := r.Options().Observe()
		if err != nil {
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadRequest); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
		p.handleGet(ctx, path, w.Conn(), r.Token(), obs, r)

	case codes.POST:
		body, err := r.ReadBody()
		if err != nil {
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadRequest); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
		p.handlePost(ctx, w.Conn(), body, r.Token(), path, r)
	}
}

func (p *Proxy) handleGet(ctx context.Context, path string, con mux.Conn, token []byte, obs uint32, r *mux.Message) {
	if err := p.session.AuthSubscribe(ctx, &[]string{path}); err != nil {
		if err := sendErrorMessage(con, token, err, codes.Unauthorized); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
	if err := p.session.Subscribe(ctx, &[]string{path}); err != nil {
		if err := sendErrorMessage(con, token, err, codes.Unauthorized); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
	switch {
	// obs == 0, start observe
	case obs == 0:
		go p.observeUpstream(ctx, con, r.Options(), token, path)

	default:
		if err := p.getUpstream(con, r, token); err != nil {
			p.logger.Error(fmt.Sprintf("error performing get: %v\n", err))
			if err := sendErrorMessage(con, token, err, codes.BadGateway); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
	}
}

func (p *Proxy) handlePost(ctx context.Context, con mux.Conn, body, token []byte, path string, r *mux.Message) {
	if err := p.session.AuthPublish(ctx, &path, &body); err != nil {
		if err := sendErrorMessage(con, token, err, codes.Unauthorized); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
	if err := p.session.Publish(ctx, &path, &body); err != nil {
		if err := sendErrorMessage(con, token, err, codes.BadRequest); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
	if err := p.postUpstream(con, r, token); err != nil {
		p.logger.Debug(fmt.Sprintf("error performing post: %v\n", err))
		if err := sendErrorMessage(con, token, err, codes.BadGateway); err != nil {
			p.logger.Error(err.Error())
		}
		return
	}
}

func (p *Proxy) Listen(ctx context.Context) error {
	switch {
	case p.config.DTLSConfig == nil:
		l, err := net.NewListenUDP("udp", p.config.Address)
		if err != nil {
			return err
		}
		defer l.Close()

		p.logger.Info(fmt.Sprintf("CoAP proxy server started at %s without DTLS", p.config.Address))
		s := udp.NewServer(options.WithMux(mux.HandlerFunc(p.handler)))

		errCh := make(chan error)
		go func() {
			errCh <- s.Serve(l)
		}()

		select {
		case <-ctx.Done():
			p.logger.Info(fmt.Sprintf("CoAP proxy server at %s without DTLS exiting ...", p.config.Address))
			l.Close()
		case err := <-errCh:
			p.logger.Error(fmt.Sprintf("CoAP proxy server at %s without DTLS exiting with errors: %s", p.config.Address, err.Error()))
			return err
		}
		return nil
	case p.config.DTLSConfig != nil:
		l, err := net.NewDTLSListener("udp", p.config.Address, p.config.DTLSConfig)
		if err != nil {
			return err
		}
		defer l.Close()

		p.logger.Info(fmt.Sprintf("CoAP proxy server started at %s with DTLS", p.config.Address))
		s := dtls.NewServer(options.WithMux(mux.HandlerFunc(p.handler)))

		errCh := make(chan error)
		go func() {
			errCh <- s.Serve(l)
		}()

		select {
		case <-ctx.Done():
			p.logger.Info(fmt.Sprintf("CoAP proxy server at %s with DTLS exiting ...", p.config.Address))
			l.Close()
		case err := <-errCh:
			p.logger.Error(fmt.Sprintf("CoAP proxy server at %s with DTLS exiting with errors: %s", p.config.Address, err.Error()))
			return err
		}
		return nil
	default:
		return errors.New("unsupported CoAP configuration")
	}
}
