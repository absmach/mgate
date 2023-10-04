package websockets

import (
	"context"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
)

var upgrader = websocket.Upgrader{}

type Proxy struct {
	targetConn *websocket.Conn
	address    string
	event      session.Handler
	logger     logger.Logger
}

func (p *Proxy) handler(w http.ResponseWriter, r *http.Request) {
	token := r.URL.Query()["authorization"][0]
	topic := r.URL.Path
	s := session.Session{Password: []byte(token)}
	ctx := session.NewContext(context.Background(), &s)
	if err := p.event.AuthConnect(ctx); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			p.logger.Error(err.Error())
			return
		}
	}
	if err := p.event.AuthSubscribe(ctx, &[]string{topic}); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		if _, err := w.Write([]byte(err.Error())); err != nil {
			p.logger.Error(err.Error())
			return
		}
	}
	inConn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		p.logger.Warn(err.Error())
		return
	}
	defer inConn.Close()

	go p.stream(ctx, topic, inConn, p.targetConn, true)
	go p.stream(ctx, topic, p.targetConn, inConn, false)
}

func (p *Proxy) stream(ctx context.Context, topic string, src, dest *websocket.Conn, upstream bool) error {
	for {
		messageType, payload, err := src.ReadMessage()
		if err != nil {
			return err
		}
		if upstream {
			if err := p.event.AuthPublish(ctx, &topic, &payload); err != nil {
				return err
			}
		}
		if err := dest.WriteMessage(messageType, payload); err != nil {
			return err
		}
	}
}

func NewProxy(address, target string) (*Proxy, error) {
	targetConn, _, err := websocket.DefaultDialer.Dial(target, nil)
	if err != nil {
		return nil, err
	}
	return &Proxy{targetConn: targetConn}, nil
}

func (p *Proxy) Listen() error {
	return http.ListenAndServe(p.address, http.HandlerFunc(p.handler))
}
