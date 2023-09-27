package mqtt

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"testing"
	"time"

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
		address string
		target  string
		handler session.Handler
		logger  logger.Logger
	}

	cfg := config{
		logLevel: "info",
	}

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	tests := []struct {
		name    string
		args    args
		session *session.Session
		want    *Proxy
	}{
		{
			name: "successfully created new proxy",
			args: args{
				address: "localhost:8080",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
			},
			session: nil,
			want: &Proxy{
				address: "localhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
			},
		},
		{
			name: "incorrect proxy",
			args: args{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
			},
			session: nil,
			want: &Proxy{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
			},
		},
	}

	for _, tt := range tests {
		got := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger)
		assert.Equal(t, got.address, tt.want.address, fmt.Sprintf("%s: expected %s got %s\n", tt.name, tt.want.address, got.address))
		assert.Equal(t, got.target, tt.want.target, fmt.Sprintf("%s: expected %s got %s\n", tt.name, tt.want.target, got.target))
	}
}

func TestProxy_Listen(t *testing.T) {
	type args struct {
		address string
		target  string
		handler session.Handler
		logger  logger.Logger
		context context.Context
	}

	var cfg config

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "successfully listen",
			args: args{
				address: "localhost:8080",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr: nil,
		},
		{
			name: "incorrect listen",
			args: args{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr: nil,
		},
		{
			name: "successfully listen",
			args: args{
				address: "localhost",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr: nil, //Change back to a bool
		},
	}

	for _, tt := range tests {
		p := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger)
		err := p.Listen(tt.args.context)
		assert.Equal(t, err, tt.wantErr, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}

func Test_LisetenTLS(t *testing.T) {
	type args struct {
		address string
		target  string
		handler session.Handler
		logger  logger.Logger
		context context.Context
		config  *tls.Config
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "successfully listen",
			args: args{
				address: "localhost:8080",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
				context: ctx,
				config: &tls.Config{
					Certificates: []tls.Certificate{cert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    roots,
				},
			},
			wantErr: assert.AnError,
		},
		{
			name: "incorrect listen",
			args: args{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully listen",
			args: args{
				address: "localhost",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr: nil, //Change back to a bool
		},
	}

	for _, tt := range tests {
		p := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger)
		err := p.ListenTLS(tt.args.context, tt.args.config)
		assert.Equal(t, err, tt.wantErr, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}

func Test_ListenTLS(t *testing.T) {
	type args struct {
		address string
		target  string
		handler session.Handler
		logger  logger.Logger
		context context.Context
		config  *tls.Config
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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)

	defer cancel()

	tests := []struct {
		name    string
		args    args
		wantErr error
	}{
		{
			name: "successfully listen",
			args: args{
				address: "localhost:8080",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
				context: ctx,
				config: &tls.Config{
					Certificates: []tls.Certificate{cert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    roots,
				},
			},
			wantErr: assert.AnError,
		},
		{
			name: "incorrect listen",
			args: args{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr: assert.AnError,
		},
		{
			name: "successfully listen",
			args: args{
				address: "localhost",
				target:  "localhost:8085",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr: nil, //Change back to a bool
		},
	}

	for _, tt := range tests {
		p := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger)

		// Create a channel to signal that the function has completed
		done := make(chan struct{})

		go func() {
			// Run the function and capture any error
			err := p.ListenTLS(tt.args.context, tt.args.config)

			// Check if the context was canceled
			if tt.args.context.Err() == context.Canceled {
				// The context was canceled as expected
				assert.Error(t, err, "Context canceled, but an error should be returned.")
			} else {
				// The function completed before the context was canceled
				assert.NoError(t, err, "Expected no error.")
			}

			// Signal that the function has completed
			close(done)
		}()

		// Wait for the function to complete or the timeout to occur
		select {
		case <-done:
			// The function has completed or was canceled, continue with the next test case
		case <-time.After(5 * time.Second):
			t.Fatal("Test timed out waiting for function completion")
		}
	}
}
