package websocket

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	"github.com/gorilla/websocket"
	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/examples/simple"
	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/stretchr/testify/assert"
)

type config struct {
	logLevel string
}

var (
	ca  = "../../certs/ca.crt"
	crt = "../../certs/ca.crt"
	key = "../../certs/ca.key"
)

func TestNew(t *testing.T) {
	type args struct {
		target string
		path   string
		scheme string
		event  session.Handler
		logger logger.Logger
	}

	var cfg config

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	tests := []struct {
		name string
		args args
		want *Proxy
	}{
		{
			name: "New proxy",
			args: args{
				target: "target",
				path:   "path",
				scheme: "scheme",
				event:  h,
				logger: logger,
			},
			want: &Proxy{
				target: "target",
				path:   "path",
				scheme: "scheme",
				event:  h,
				logger: logger,
			},
		},
	}
	for _, tt := range tests {
		got := New(tt.args.target, tt.args.path, tt.args.scheme, tt.args.event, tt.args.logger)
		assert.Equal(t, got.target, tt.want.target, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want.target, got.target))
		assert.Equal(t, got.path, tt.want.path, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want.path, got.path))
		assert.Equal(t, got.scheme, tt.want.scheme, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want.scheme, got.scheme))
		assert.Equal(t, got.event, tt.want.event, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want.event, got.event))
	}
}

func Test_Handler(t *testing.T) {
	cfg := config{
		logLevel: "Info",
	}

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	pr := Proxy{
		target: "localhost",
		path:   ":8080",
		scheme: "https",
		event:  h,
		logger: logger,
	}

	tests := []struct {
		name string
		p    Proxy
		want http.Handler
	}{
		{
			name: "Successfully handled request",
			p: Proxy{
				target: "localhost",
				path:   ":8080",
				scheme: "https",
				event:  h,
				logger: logger,
			},
			want: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				cconn, err := upgrader.Upgrade(w, r, nil)
				if err != nil {
					pr.logger.Error("Error upgrading connection " + err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer cconn.Close()

				u := url.URL{Scheme: pr.scheme, Host: pr.target, Path: pr.path}
				proxyConn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
				if err != nil {
					pr.logger.Error("Error dialing to proxy " + err.Error())
					http.Error(w, err.Error(), http.StatusInternalServerError)
					return
				}
				defer proxyConn.Close()

				ctx, cancel := context.WithCancel(r.Context())
				defer cancel()

				go pr.pass(ctx, cconn)
				pr.pass(ctx, proxyConn)
			}),
		},
	}
	for _, tt := range tests {
		got := tt.p.Handler()
		assert.Equal(t, got, tt.want, fmt.Sprintf("%s: expected %v got %v", tt.name, tt.want, got))
	}
}

func TestProxy_Listen(t *testing.T) {
	s := httptest.NewServer(http.HandlerFunc(handler))
	defer s.Close()

	// Convert http://127.0.0.1 to ws://127.0.0.1
	// u := strings.Replace(s.URL, "http", "ws", 1)

	type args struct {
		wsPort string
	}

	cfg := config{
		logLevel: "info",
	}

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	tests := []struct {
		name    string
		p       Proxy
		args    args
		wantErr error
	}{
		{
			name: "Successfully listen",
			p: Proxy{
				target: "target",
				path:   "path",
				scheme: "scheme",
				event:  h,
				logger: logger,
			},
			args: args{
				wsPort: "8080",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		errorChan := make(chan error, 1)

		go func() {
			err := tt.p.Listen(tt.args.wsPort)

			if err != nil {
				errorChan <- err
			}
		}()

		select {
		case err := <-errorChan:
			t.Errorf("%s: unexpected error: %v\n", tt.name, err)
		case <-time.After(5 * time.Second):
			logger.Info("Listen completed successfully without errors")
		}
	}
}

func TestProxy_ListenTLS(t *testing.T) {
	type args struct {
		tlsCfg  *tls.Config
		crt     string
		key     string
		wssPort string
	}

	cert, _ := tls.LoadX509KeyPair(crt, key)
	roots := x509.NewCertPool()
	caCertPEM, _ := os.ReadFile(ca)
	roots.AppendCertsFromPEM(caCertPEM)

	cfg := config{
		logLevel: "info",
	}

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	tests := []struct {
		name    string
		p       Proxy
		args    args
		wantErr error
	}{
		{
			name: "Successfully listen",
			p: Proxy{
				target: "target",
				path:   "path",
				scheme: "scheme",
				event:  h,
				logger: logger,
			},
			args: args{
				tlsCfg: &tls.Config{
					Certificates: []tls.Certificate{cert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    roots,
				},
				crt:     crt,
				key:     key,
				wssPort: "8080",
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		errorChan := make(chan error, 1)

		go func() {
			err := tt.p.ListenTLS(tt.args.tlsCfg, tt.args.crt, tt.args.key, tt.args.wssPort)

			if err != nil {
				errorChan <- err
			}
		}()

		select {
		case err := <-errorChan:
			t.Errorf("%s: unexpected error: %v\n", tt.name, err)
		case <-time.After(5 * time.Second):
			logger.Info("Listen completed successfully without errors")
		}
	}
}
