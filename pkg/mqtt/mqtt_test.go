package mqtt

import (
	"context"
	"fmt"
	"os"
	"testing"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/examples/simple"
	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
	"github.com/stretchr/testify/assert"
)

type config struct {
	logLevel string
}

func TestNew(t *testing.T) {
	type args struct {
		address string
		target  string
		handler session.Handler
		logger  logger.Logger
	}

	var cfg config

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)

	expectedProxy := &Proxy{
		address: "localhost",
		target:  "localhost",
		handler: h,
		logger:  logger,
	}

	tests := []struct {
		name    string
		args    args
		session *session.Session
		want    *Proxy
	}{
		{
			name: "successfully created new proxy",
			args: args{
				address: "localhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
			},
			session: nil,
			want:    expectedProxy,
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
