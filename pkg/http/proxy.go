package http

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/pkg/events"
	"github.com/mainflux/mproxy/pkg/mqtt"
)

// Proxy - struct that holds HTTP proxy info
type Proxy struct {
	host    string
	port    string
	path    string
	scheme  string
	event   events.Event
	logger  logger.Logger
	session mqtt.Session
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:   1048562,
	WriteBufferSize:  1048562,
	HandshakeTimeout: 10 * time.Second,
	// Paho JS client expecting header Sec-WebSocket-Protocol:mqtt in Upgrade response during handshake.
	Subprotocols: []string{"mqtt"},
	// Allow CORS
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// New - creates new HTTP proxy
func New(host, port, path, scheme string, event events.Event, logger logger.Logger) *Proxy {
	return &Proxy{
		host:    host,
		port:    port,
		path:    path,
		scheme:  scheme,
		event:   event,
		logger:  logger,
		session: mqtt.Session{},
	}
}

// Handle - proxies HTTP traffic
func (p *Proxy) Handler() http.Handler {
	return p.handle()
}

func (p *Proxy) handle() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cconn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			p.logger.Error("Error upgrading connection " + err.Error())
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		go p.pass(cconn)
	})
}

func (p Proxy) pass(in *websocket.Conn) {
	defer in.Close()

	url := url.URL{
		Scheme: p.scheme,
		Host:   fmt.Sprintf("%s:%s", p.host, p.port),
		Path:   p.path,
	}

	srv, _, err := websocket.DefaultDialer.Dial(url.String(), nil)

	if err != nil {
		p.logger.Error("Unable to connect to broker, reason: " + err.Error())
		return
	}

	errc := make(chan error, 1)
	c := NewConn(in)
	s := NewConn(srv)

	defer s.Close()
	defer c.Close()

	session := mqtt.NewSession(c, s, p.event, p.logger)
	err = session.Stream()
	errc <- err
	p.logger.Error("Streaming error:" + err.Error())

}
