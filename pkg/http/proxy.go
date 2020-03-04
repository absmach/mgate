package http

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/pkg/events"
)

// Proxy - struct that holds HTTP proxy info
type Proxy struct {
	host         string
	port         string
	ReverseProxy *httputil.ReverseProxy
	event        events.Event
	logger       logger.Logger
}

// New - creates new HTTP proxy
func New(host, port, scheme string, event events.Event, logger logger.Logger) *Proxy {
	url := url.URL{
		Host:   fmt.Sprintf("%s:%s", host, port),
		Scheme: scheme,
	}
	return &Proxy{
		host:         host,
		port:         port,
		event:        event,
		logger:       logger,
		ReverseProxy: httputil.NewSingleHostReverseProxy(&url),
	}
}

// Handle - proxies HTTP traffic
func (p *Proxy) Handle(w http.ResponseWriter, r *http.Request) {
	// Note that ServeHttp is non blocking and uses a go routine under the hood
	p.ReverseProxy.ServeHTTP(w, r)
}
