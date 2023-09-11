package mqtt

import (
	"context"
	"crypto/tls"
	"net"
	"os"
	"reflect"
	"testing"

	mflog "github.com/mainflux/mainflux/logger"
	"github.com/mainflux/mproxy/examples/simple"
	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
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
		name string
		args args
		want *Proxy
	}{
		{
			name: "successfully created new proxy",
			args: args{
				address: "localhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
			},
			want: expectedProxy,
		},
		{
			name: "incorrect proxy",
			args: args{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
			},
			want: &Proxy{
				address: "unlocalhost",
				target:  "localhost",
				handler: h,
				logger:  logger,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_Listen(t *testing.T) {
	type fields struct {
		address string
		target  string
		handler session.Handler
		logger  logger.Logger
		dialer  net.Dialer
	}
	type args struct {
		ctx context.Context
	}
	var cfg config

	logger, _ := mflog.New(os.Stdout, cfg.logLevel)

	h := simple.New(logger)
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "successfully started proxy",
			fields: fields{
				address: "localhost:8080",
				target:  "localhost:8080",
				handler: h,
				logger:  logger,
				dialer:  net.Dialer{},
			},
			args:    args{ctx: context.Background()},
			wantErr: false,
		},
		{
			name: "incorrect proxy",
			fields: fields{
				address: "localhost",
				target:  "localhost",
				handler: nil,
				logger:  nil,
				dialer:  net.Dialer{},
			},
			args:    args{ctx: context.Background()},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := Proxy{
				address: tt.fields.address,
				target:  tt.fields.target,
				handler: tt.fields.handler,
				logger:  tt.fields.logger,
				dialer:  tt.fields.dialer,
			}
			if err := p.Listen(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.Listen() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestProxy_ListenTLS(t *testing.T) {
	type args struct {
		ctx    context.Context
		tlsCfg *tls.Config
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
			if err := tt.p.ListenTLS(tt.args.ctx, tt.args.tlsCfg); (err != nil) != tt.wantErr {
				t.Errorf("Proxy.ListenTLS() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
