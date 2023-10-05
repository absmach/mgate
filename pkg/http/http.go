package http

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
)

// Handler default handler reads authorization header and
// performs authorization before proxying the request.
func (p *Proxy) Handler(w http.ResponseWriter, r *http.Request) {
	username, password, ok := r.BasicAuth()
	switch {
	case ok:
		break
	case r.Header.Get("Authorization") != "":
		password = r.Header.Get("Authorization")
	default:
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	s := &session.Session{
		Password: []byte(password),
		Username: username,
	}
	ctx := session.NewContext(r.Context(), s)
	if err := p.event.AuthConnect(ctx); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		p.logger.Error(err.Error())
		return
	}
	payload, err := io.ReadAll(r.Body)
	r.Body.Close() //  must close
	r.Body = io.NopCloser(bytes.NewBuffer(payload))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		p.logger.Error(err.Error())
		return
	}
	if err := p.event.AuthPublish(ctx, &r.RequestURI, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		p.logger.Error(err.Error())
		return
	}
	if err := p.event.Publish(ctx, &r.RequestURI, &payload); err != nil {
		http.Error(w, err.Error(), http.StatusBadGateway)
		p.logger.Error(err.Error())
		return
	}
	p.target.ServeHTTP(w, r)
}

// Proxy represents HTTP Proxy.
type Proxy struct {
	address string
	target  *httputil.ReverseProxy
	event   session.Handler
	logger  logger.Logger
}

func NewProxy(address, targetUrl string, handler session.Handler, logger logger.Logger) (Proxy, error) {
	target, err := url.Parse(targetUrl)
	if err != nil {
		return Proxy{}, err
	}

	return Proxy{
		address: address,
		target:  httputil.NewSingleHostReverseProxy(target),
		event:   handler,
		logger:  logger,
	}, nil
}

func (p *Proxy) Listen() error {
	if err := http.ListenAndServe(p.address, nil); err != nil {
		return err
	}

	p.logger.Info("Server Exiting...")
	return nil
}

func (p *Proxy) ListenTLS(cert, key string) error {
	if err := http.ListenAndServeTLS(p.address, cert, key, nil); err != nil {
		return err
	}

	p.logger.Info("Server Exiting...")
	return nil
}
