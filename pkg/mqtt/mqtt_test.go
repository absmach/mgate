package mqtt

import (
	"context"
	"crypto/tls"
	"net"
	"reflect"
	"testing"

	"github.com/mainflux/mproxy/pkg/logger"
	"github.com/mainflux/mproxy/pkg/session"
)

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
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
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

func TestNew(t *testing.T) {
	type args struct {
		address string
		target  string
		handler session.Handler
		logger  logger.Logger
	}

	expectedProxy := &Proxy{
        Field1: "expectedValue1",
        Field2: "expectedValue2",
    }

	tests := []struct {
		name string
		args args
		want *Proxy
	}{
		name: "testCreateNewInstance",
		args: args{
			address: "localhost",
			target: "localhost",
			handler: session.Handler,
			logger: logger.Logger,
		},
		want: expectedProxy
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.address, tt.args.target, tt.args.handler, tt.args.logger); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestProxy_accept(t *testing.T) {
	type args struct {
		ctx context.Context
		l   net.Listener
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
			tt.p.accept(tt.args.ctx, tt.args.l)
		})
	}
}

func TestProxy_handle(t *testing.T) {
	type args struct {
		ctx     context.Context
		inbound net.Conn
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
			tt.p.handle(tt.args.ctx, tt.args.inbound)
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

func TestProxy_close(t *testing.T) {
	type args struct {
		conn net.Conn
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
			tt.p.close(tt.args.conn)
		})
	}
}
