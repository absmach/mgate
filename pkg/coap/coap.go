package coap

import (
	"bytes"
	"crypto/tls"
	"fmt"

	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/pion/dtls/v2"
	coap "github.com/plgd-dev/go-coap/v3"
	"github.com/plgd-dev/go-coap/v3/message"
	"github.com/plgd-dev/go-coap/v3/message/codes"
	"github.com/plgd-dev/go-coap/v3/message/pool"
	"github.com/plgd-dev/go-coap/v3/mux"
	"github.com/plgd-dev/go-coap/v3/udp"
	"github.com/plgd-dev/go-coap/v3/udp/client"
)

type Proxy struct {
	logger     logger.Logger
	targetConn *client.Conn
	address    string
	event      session.Handler
}

func sendPoolMessage(cc mux.Conn, pm *pool.Message, token []byte) error {
	m := cc.AcquireMessage(pm.Context())
	defer cc.ReleaseMessage(m)
	m.SetCode(pm.Code())
	m.SetBody(pm.Body())
	m.SetToken(token)
	formt, err := pm.ContentFormat()
	switch err {
	case nil:
		m.SetContentFormat(formt)
	case message.ErrOptionNotFound:
		m.SetContentFormat(message.TextPlain)
	default:
		return err
	}
	obs, err := pm.Observe()
	switch err {
	case nil:
		m.SetObserve(uint32(obs))
	case message.ErrOptionNotFound:
		break
	default:
		return err
	}

	return cc.WriteMessage(m)
}

func sendErrorMessage(cc mux.Conn, token []byte, err error, code codes.Code) error {
	m := cc.AcquireMessage(cc.Context())
	defer cc.ReleaseMessage(m)
	m.SetCode(code)
	m.SetBody(bytes.NewReader([]byte(err.Error())))
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

	pm, err := p.targetConn.Post(cc.Context(), path, format, req.Body(), req.Options()...)
	if err != nil {
		return err
	}
	return sendPoolMessage(cc, pm, token)
}

func (p *Proxy) getUpstream(cc mux.Conn, req *mux.Message, token []byte) error {
	path, err := req.Options().Path()
	if err != nil {
		return err
	}

	pm, err := p.targetConn.Get(cc.Context(), path, req.Options()...)
	if err != nil {
		return err
	}
	return sendPoolMessage(cc, pm, token)
}

func (p *Proxy) observeUpstream(cc mux.Conn, req *mux.Message, token []byte) error {
	path, err := req.Options().Path()
	if err != nil {
		return err
	}
	if _, err = p.targetConn.Client.Observe(req.Context(), path, func(req *pool.Message) {
		if err := sendPoolMessage(cc, req, token); err != nil {
			p.logger.Error(err.Error())
		}
	}, req.Options()...); err != nil {
		return err
	}
	return nil
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
	if err := p.event.AuthConnect(ctx); err != nil {
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
	p.logger.Debug(fmt.Sprintf("Got message path=%v: %+v from %v", path, r, w.Conn().RemoteAddr()))
	switch r.Code() {
	case codes.GET:
		if err := p.event.AuthSubscribe(ctx, &[]string{path}); err != nil {
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.Unauthorized); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
		if err := p.event.Subscribe(ctx, &[]string{path}); err != nil {
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.Unauthorized); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
		obs, err := r.Options().Observe()
		switch {
		// obs == 0, start observe
		case obs == 0 && err == nil:
			if err := p.observeUpstream(w.Conn(), r, r.Token()); err != nil {
				p.logger.Debug(fmt.Sprintf("error performing get: %v\n", err))
				if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadGateway); err != nil {
					p.logger.Error(err.Error())
				}
				return
			}
		default:
			if err := p.getUpstream(w.Conn(), r, r.Token()); err != nil {
				p.logger.Debug(fmt.Sprintf("error performing get: %v\n", err))
				if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadGateway); err != nil {
					p.logger.Error(err.Error())
				}
				return
			}
		}

	case codes.POST:
		body, err := r.ReadBody()
		if err != nil {
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadRequest); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
		if err := p.event.AuthPublish(ctx, &path, &body); err != nil {
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.Unauthorized); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
		if err := p.event.Publish(ctx, &path, &body); err != nil {
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadRequest); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
		if err := p.postUpstream(w.Conn(), r, r.Token()); err != nil {
			p.logger.Debug(fmt.Sprintf("error performing post: %v\n", err))
			if err := sendErrorMessage(w.Conn(), r.Token(), err, codes.BadGateway); err != nil {
				p.logger.Error(err.Error())
			}
			return
		}
	}
}

func NewProxy(address, target string, logger logger.Logger, handler session.Handler) (*Proxy, error) {
	targetConn, err := udp.Dial(address)
	if err != nil {
		return nil, err
	}

	return &Proxy{
		targetConn: targetConn,
		logger:     logger,
		address:    address,
		event:      handler,
	}, nil
}

func (p *Proxy) Listen() error {
	return coap.ListenAndServe("udp", p.address, mux.HandlerFunc(p.handler))
}

func (p Proxy) ListenTLS(tlsCfg *tls.Config) error {
	return coap.ListenAndServeTCPTLS("udp", p.address, tlsCfg, mux.HandlerFunc(p.handler))
}

func (p Proxy) ListenDLS(dtlsCfg *dtls.Config) error {
	return coap.ListenAndServeDTLS("udp", p.address, dtlsCfg, mux.HandlerFunc(p.handler))
}
