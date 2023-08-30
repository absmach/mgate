package tls

import (
	"crypto/tls"
	"crypto/x509"
	"net"
	"reflect"
	"testing"
)

func TestLoadTLSCfg(t *testing.T) {
	type args struct {
		ca  string
		crt string
		key string
	}
	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadTLSCfg(tt.args.ca, tt.args.crt, tt.args.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTLSCfg() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadTLSCfg() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClientCert(t *testing.T) {
	type args struct {
		conn net.Conn
	}
	tests := []struct {
		name    string
		args    args
		want    x509.Certificate
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ClientCert(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientCert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientCert() = %v, want %v", got, tt.want)
			}
		})
	}
}
