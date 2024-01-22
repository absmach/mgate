package websocket_test

import (
	"log/slog"
	"net/http/httptest"
	"testing"

	"github.com/absmach/mproxy/pkg/mqtt/websocket"
	"github.com/absmach/mproxy/pkg/session/mocks"
)

func newProxy(target, path, scheme string) *websocket.Proxy {
	handler := new(mocks.Handler)
	interceptor := new(mocks.Interceptor)
	logger := new(slog.Logger)
	return websocket.New(target, path, scheme, handler, interceptor, logger)
}

func TestHandler(t *testing.T) {
	cases := []struct {
		desc   string
		target string
		path   string
		scheme string
	}{
		{
			desc:   "handler with valid target",
			target: "localhost:8080",
			path:   "/",
			scheme: "ws",
		},
	}
	for _, c := range cases {
		proxy := newProxy(c.target, c.path, c.scheme)
		responseRecorder := httptest.NewRecorder()
		request := httptest.NewRequest("GET", "http://example.com", nil)
		proxy.Handler().ServeHTTP(responseRecorder, request)
	}
}
