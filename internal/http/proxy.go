package http

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"

	"github.com/mainflux/mainflux/logger"
)

// Proxy - struct that holds HTTP proxy info
type Proxy struct {
	host         string
	port         string
	ReverseProxy *httputil.ReverseProxy
	logger       logger.Logger
}

// New - creates new HTTP proxy
func New(host, port string, logger logger.Logger) *Proxy {
	url := url.URL{
		Host: fmt.Sprintf("%s:%s", host, port),
	}
	return &Proxy{
		host:         host,
		port:         port,
		ReverseProxy: httputil.NewSingleHostReverseProxy(&url),
	}
}

// Handle - proxies HTTP traffic
func (p *Proxy) Handle(w http.ResponseWriter, r *http.Request) {
	// Note that ServeHttp is non blocking and uses a go routine under the hood
	p.ReverseProxy.ServeHTTP(w, r)
}
