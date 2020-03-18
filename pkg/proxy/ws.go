package proxy

import (
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mproxy/pkg/mqtt"
)

var upgrader = websocket.Upgrader{
	// Timeout for WS upgrade request handshake
	HandshakeTimeout: 10 * time.Second,
	// Paho JS client expecting header Sec-WebSocket-Protocol:mqtt in Upgrade response during handshake.
	Subprotocols: []string{"mqtt"},
	// Allow CORS
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Handle - proxies HTTP traffic
func (p wsProxy) Handler() http.Handler {
	return p.handle()
}

func (p wsProxy) handle() http.Handler {
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

func (p wsProxy) pass(in *websocket.Conn) {
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
	c := wrappWSConn(in)
	s := wrappWSConn(srv)

	defer s.Close()
	defer c.Close()

	session := mqtt.NewSession(c, s, p.event, p.logger)
	err = session.Stream()
	errc <- err
	p.logger.Warn("Broken connection for client: " + session.Client.ID + " with error: " + err.Error())

}
