package websockets_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/absmach/mproxy/pkg/session/mocks"
	ws "github.com/absmach/mproxy/pkg/websockets"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

var (
	handler = new(mocks.Handler)
	logger  = new(slog.Logger)
)

func createMockServer(t *testing.T) *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		upgrader := websocket.Upgrader{}
		conn, err := upgrader.Upgrade(w, r, nil)
		assert.Nil(t, err, fmt.Sprintf("Unexpected error upgrading connection: %v", err))
		defer conn.Close()
	}))

	return server
}

func TestHandlerSuccess(t *testing.T) {
	mockServer := createMockServer(t)
	defer mockServer.Close()
	mockServerURL := "ws" + strings.TrimPrefix(mockServer.URL, "http")

	proxy, err := ws.NewProxy("ws://example.com", mockServerURL, logger, handler)
	assert.Nil(t, err, fmt.Sprintf("Unexpected error creating proxy: %v", err))
	testServer := httptest.NewServer(http.HandlerFunc(proxy.Handler))
	defer testServer.Close()
	proxyURL := "ws" + strings.TrimPrefix(testServer.URL, "http") + "/test"

	cases := []struct {
		desc             string
		url              string
		header           http.Header
		authConnectErr   error
		authSubscribeErr error
		subscribeErr     error
		status           int
	}{
		{
			desc:             "successfull connection with authorization in query",
			url:              proxyURL + "?authorization=valid",
			header:           http.Header{},
			authConnectErr:   nil,
			authSubscribeErr: nil,
			subscribeErr:     nil,
			status:           http.StatusSwitchingProtocols,
		},
		{
			desc: "successfull connection with authorization in header",
			url:  proxyURL,
			header: http.Header{
				"Authorization": []string{"valid-token"},
			},
			authConnectErr:   nil,
			authSubscribeErr: nil,
			subscribeErr:     nil,
			status:           http.StatusSwitchingProtocols,
		},
		{
			desc:   "unsuccesful connection with no authorization",
			url:    proxyURL,
			header: http.Header{},
			status: http.StatusUnauthorized,
		},
		{
			desc: "unsuccesful connection with failed session auth connect",
			url:  proxyURL,
			header: http.Header{
				"Authorization": []string{"valid-token"},
			},
			authConnectErr:   fmt.Errorf("failed auth connect"),
			authSubscribeErr: nil,
			subscribeErr:     nil,
			status:           http.StatusUnauthorized,
		},
		{
			desc: "unsuccesful connection with failed session auth subscribe",
			url:  proxyURL,
			header: http.Header{
				"Authorization": []string{"valid-token"},
			},
			authConnectErr:   nil,
			authSubscribeErr: fmt.Errorf("failed auth subscribe"),
			subscribeErr:     nil,
			status:           http.StatusUnauthorized,
		},
		{
			desc: "unsuccesful connection with failed session subscribe",
			url:  proxyURL,
			header: http.Header{
				"Authorization": []string{"valid-token"},
			},
			authConnectErr:   nil,
			authSubscribeErr: nil,
			subscribeErr:     fmt.Errorf("failed subscribe"),
			status:           http.StatusBadRequest,
		},
	}
	for _, tc := range cases {
		sessionCall := handler.On("AuthConnect", mock.Anything).Return(tc.authConnectErr)
		sessionCall1 := handler.On("AuthSubscribe", mock.Anything, mock.Anything).Return(tc.authSubscribeErr)
		sessionCall2 := handler.On("Subscribe", mock.Anything, mock.Anything).Return(tc.subscribeErr)
		_, res, _ := websocket.DefaultDialer.Dial(tc.url, tc.header)
		assert.Equal(t, tc.status, res.StatusCode, fmt.Sprintf("%s expected status code %d but got %d", tc.desc, tc.status, res.StatusCode))
		sessionCall.Unset()
		sessionCall1.Unset()
		sessionCall2.Unset()
	}
}

// func TestListen(t *testing.T) {
// 	proxy, err := ws.NewProxy("localhost:8080", "ws://127.0.0.1", logger, handler)
// 	assert.NoError(t, err)
// 	go func() {
// 		err := proxy.Listen()
// 		assert.Nil(t, err, fmt.Sprintf("Unexpected error listening: %v", err))
// 	}()
// 	time.Sleep(100 * time.Millisecond)

// 	req, err := http.NewRequest("GET", "http://example.com", nil)
// 	assert.Nil(t, err, fmt.Sprintf("Unexpected error creating request: %v", err))
// 	req.Header.Set("Authorization", "valid-token")
// 	rr := httptest.NewRecorder()
// 	handler := http.HandlerFunc(proxy.Handler)
// 	handler.ServeHTTP(rr, req)
// 	assert.Equal(t, http.StatusBadGateway, rr.Code, "Expected status code 502 but got %d", rr.Code)

// 	// err = proxy.Shutdown(context.Background())
// 	// assert.Nil(t, err, fmt.Sprintf("Unexpected error shutting down: %v", err))

// }

// func TestListenTLS(t *testing.T) {
// 	proxy, err := ws.NewProxy("localhost:8080", "wss://127.0.0.1", logger, handler)
// 	assert.NoError(t, err)
// 	go func() {
// 		err := proxy.ListenTLS("cert.pem", "key.pem")
// 		assert.Nil(t, err, fmt.Sprintf("Unexpected error listening: %v", err))
// 	}()

// }
