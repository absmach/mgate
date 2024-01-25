package mqtt_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"testing"

	"github.com/absmach/mproxy/pkg/mqtt"
	"github.com/absmach/mproxy/pkg/session/mocks"
	"github.com/stretchr/testify/assert"
)

func newProxy(address, target string) *mqtt.Proxy {
	handler := new(mocks.Handler)
	interceptor := new(mocks.Interceptor)
	return mqtt.New(address, target, handler, interceptor, nil)
}

var tlsConfig = &tls.Config{}

func TestListen(t *testing.T) {
	cases := []struct {
		desc    string
		address string
		target  string
		err     error
	}{
		{
			desc:    "listen with valid address",
			address: "localhost:8080",
			target:  "localhost:8080",
			err:     nil,
		},
		// {
		// 	desc:    "listen with invalid address",
		// 	address: "0000",
		// 	target:  "localhost:8080",
		// 	err:     nil,
		// },
	}
	for _, c := range cases {
		proxy := newProxy(c.address, c.target)
		go func() {
			err := proxy.Listen(context.Background())
			assert.Nil(t, err, fmt.Sprintf("expected nil got %s\n", err))
		}()
	}
}

func TestListenTLS(t *testing.T) {
	cases := []struct {
		desc    string
		address string
		target  string
		err     error
	}{
		{
			desc:    "listen with valid address",
			address: "localhost:8080",
			target:  "localhost:8080",
			err:     nil,
		},
		// {
		// 	desc:    "listen with invalid address",
		// 	address: "0000",
		// 	target:  "localhost:8080",
		// 	err:     nil,
		// },
	}
	for _, c := range cases {
		
		proxy := newProxy(c.address, c.target)
		go func() {
			err := proxy.ListenTLS(context.Background(), tlsConfig)
			assert.Nil(t, err, fmt.Sprintf("expected nil got %s\n", err))
		}()
	}
}
