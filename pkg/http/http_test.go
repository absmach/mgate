package http_test

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/absmach/mproxy/pkg/session/mocks"

	mhttp "github.com/absmach/mproxy/pkg/http"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	validUrl    = "http://example.com"
	validAddess = "localhost:8080"
	valid       = "valid"
	invalid     = "invalid"
)

var (
	han *mocks.Handler
	log *slog.Logger
)

func newProxy(address, url string) (mhttp.Proxy, error) {
	han = new(mocks.Handler)
	log = new(slog.Logger)
	return mhttp.NewProxy(address, url, han, log)
}

func TestNewProxy(t *testing.T) {
	cases := []struct {
		desc    string
		address string
		url     string
		err     error
	}{
		{
			desc:    "create proxy with valid",
			address: validAddess,
			url:     validUrl,
			err:     nil,
		},
		{
			desc:    "create proxy with invalid url",
			address: validAddess,
			url:     "0000",
			err:     nil,
		},
	}
	for _, c := range cases {
		_, err := newProxy(c.address, c.url)
		assert.Equal(t, c.err, err, fmt.Sprintf("%s: expected %s got %s\n", c.desc, c.err, err))

	}
}

func TestHandler(t *testing.T) {
	proxy, err := newProxy(validAddess, validUrl)
	assert.Nil(t, err, fmt.Sprintf("expected nil got %s\n", err))
	request := httptest.NewRequest(http.MethodGet, "http://example.com", nil)

	cases := []struct {
		desc           string
		auth           func()
		authConnectErr error
		authPublishErr error
		code           int
	}{
		{
			desc: "successful request with username and password and basic auth",
			auth: func() {
				request.SetBasicAuth("username", "password")
			},
			code: http.StatusOK,
		},
		{
			desc: "successful request with token",
			auth: func() {
				request.Header.Set("Authorization", valid)
			},
			code: http.StatusOK,
		},
		{
			desc: "request without authorization token",
			auth: func() {
				request.Header.Set("Authorization", "")
			},
			code: http.StatusBadGateway,
		},
	}
	for _, tc := range cases {
		tc.auth()
		responseRecorder := httptest.NewRecorder()
		sessionCall := han.On("AuthConnect", mock.Anything).Return(tc.authConnectErr)
		sessionCall1 := han.On("Publish", mock.Anything, mock.Anything, mock.Anything).Return(tc.authPublishErr)
		proxy.Handler(responseRecorder, request)
		assert.Equal(t, tc.code, responseRecorder.Code, fmt.Sprintf("%s: expected %d got %d\n", tc.desc, tc.code, responseRecorder.Code))
		sessionCall.Unset()
		sessionCall1.Unset()

	}
}

func TestListen(t *testing.T) {
	proxy, err := newProxy(validAddess, validUrl)
	assert.Nil(t, err, fmt.Sprintf("expected nil got %s\n", err))

	t.Run("Listen", func(t *testing.T) {
		go func() {
			err := proxy.Listen()
			assert.Nil(t, err, fmt.Sprintf("expected nil got %s\n", err))
		}()
	})
}

func TestListenTLS(t *testing.T) {
	proxy, err := newProxy(validAddess, validUrl)
	assert.Nil(t, err, fmt.Sprintf("expected nil got %s\n", err))

	t.Run("ListenTLS", func(t *testing.T) {
		go func() {
			err := proxy.ListenTLS("cert", "key")
			assert.Nil(t, err, fmt.Sprintf("expected nil got %s\n", err))
		}()
	})
}
