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
				address: "localhost:8080",
				target:  "localhost:8080",
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

	cfg := config{
		logLevel: "info",
	}

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		timeoutSecs int
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
			wantErr:     false,
			timeoutSecs: 5,
		},
		{
			name: "incorrect listen - missing port",
			args: args{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr:     true,
			timeoutSecs: 5,
		},
	}

	for _, tt := range tests {
		p := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger)

		ctx, cancel := context.WithTimeout(tt.args.context, time.Duration(tt.timeoutSecs)*time.Second)

		defer cancel()

		errChan := make(chan error, 1)

		go func() {
			errChan <- p.Listen(ctx)
		}()

		select {
		case err := <-errChan:
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected %v got %v\n", tt.name, tt.wantErr, err)
			}
		case <-ctx.Done():
			logger.Info("Listen completed successfully")
		}
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
		name        string
		args        args
		wantErr     bool
		timeoutSecs int
	}{
		{
			name: "successfully listen",
			args: args{
				address: "localhost:8000",
				target:  "localhost:8000",
				handler: h,
				logger:  logger,
				context: ctx,
				config: &tls.Config{
					Certificates: []tls.Certificate{cert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    roots,
				},
			},
			wantErr:     false,
			timeoutSecs: 5,
		},
		{
			name: "incorrect listen - missing port",
			args: args{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
				context: context.Background(),
				config: &tls.Config{
					Certificates: []tls.Certificate{cert},
					ClientAuth:   tls.RequireAndVerifyClientCert,
					ClientCAs:    roots,
				},
			},
			wantErr:     true,
			timeoutSecs: 5,
		},
		{
			name: "incorrect listen - missing certificates in config",
			args: args{
				address: "localhost",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
				context: context.Background(),
			},
			wantErr:     true,
			timeoutSecs: 5,
		},
	}

	for _, tt := range tests {
		p := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger)

		ctx, cancel := context.WithTimeout(tt.args.context, time.Duration(tt.timeoutSecs)*time.Second)

		defer cancel()

		errChan := make(chan error, 1)

		go func() {
			errChan <- p.ListenTLS(tt.args.context, tt.args.config)
		}()

		select {
		case err := <-errChan:
			if (err != nil) != tt.wantErr {
				t.Errorf("%s: expected %v got %v\n", tt.name, tt.wantErr, err)
			}
		case <-ctx.Done():
			logger.Info("Listen completed successfully")
		}
	}
}
