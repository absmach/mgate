package session

import (
	"context"
	"crypto/x509"
	"net"
	"testing"

	"github.com/eclipse/paho.mqtt.golang/packets"
)

func TestStream(t *testing.T) {
	type args struct {
		ctx      context.Context
		inbound  net.Conn
		outbound net.Conn
		handler  Handler
		cert     x509.Certificate
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Stream(tt.args.ctx, tt.args.inbound, tt.args.outbound, tt.args.handler, tt.args.cert); (err != nil) != tt.wantErr {
				t.Errorf("Stream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_stream(t *testing.T) {
	type args struct {
		ctx  context.Context
		dir  direction
		r    net.Conn
		w    net.Conn
		h    Handler
		errs chan error
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stream(tt.args.ctx, tt.args.dir, tt.args.r, tt.args.w, tt.args.h, tt.args.errs)
		})
	}
}

func Test_authorize(t *testing.T) {
	type args struct {
		ctx context.Context
		pkt packets.ControlPacket
		h   Handler
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := authorize(tt.args.ctx, tt.args.pkt, tt.args.h); (err != nil) != tt.wantErr {
				t.Errorf("authorize() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_notify(t *testing.T) {
	type args struct {
		ctx context.Context
		pkt packets.ControlPacket
		h   Handler
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			notify(tt.args.ctx, tt.args.pkt, tt.args.h)
		})
	}
}

func Test_wrap(t *testing.T) {
	type args struct {
		ctx context.Context
		err error
		dir direction
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := wrap(tt.args.ctx, tt.args.err, tt.args.dir); (err != nil) != tt.wantErr {
				t.Errorf("wrap() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
