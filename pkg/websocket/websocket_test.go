package websocket

import (
	"context"
	"crypto/tls"
	"net/http"
	"reflect"
	"testing"

	"github.com/gorilla/websocket"
	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
)

func TestNew(t *testing.T) {
	type args struct {
		target string
		path   string
		scheme string
		event  session.Handler
		logger logger.Logger
	}
	tests := []struct {
		name string
		args args
		want *Proxy
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.target, tt.args.path, tt.args.scheme, tt.args.event, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_Handler(t *testing.T) {
	tests := []struct {
		name string
		p    Proxy
		want http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.Handler(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Proxy.Handler() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_handle(t *testing.T) {
	tests := []struct {
		name string
		p    Proxy
		want http.Handler
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.p.handle(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Proxy.handle() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_pass(t *testing.T) {
	type args struct {
		ctx context.Context
		in  *websocket.Conn
	}
	tests := []struct {
		name string
		p    Proxy
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.p.pass(tt.args.ctx, tt.args.in)
		})
	}
}

func TestProxy_Listen(t *testing.T) {
	type args struct {
		wsPort string
	}
	tests := []struct {
		name    string
		p       Proxy
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.Listen(tt.args.wsPort); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.Listen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProxy_ListenTLS(t *testing.T) {
	type args struct {
		tlsCfg  *tls.Config
		crt     string
		key     string
		wssPort string
	}
	tests := []struct {
		name    string
		p       Proxy
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.p.ListenTLS(tt.args.tlsCfg, tt.args.crt, tt.args.key, tt.args.wssPort); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.ListenTLS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
