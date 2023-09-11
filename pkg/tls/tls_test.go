package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"net"
	"reflect"
	"testing"
	"time"
)

var (
	ca  = "../../test/certs/ca.crt"
	crt = "../../test/certs/server.crt"
	key = "../../test/certs/server.key"
)

func TestLoadTLSCfg(t *testing.T) {
	type args struct {
		ca  string
		crt string
		key string
	}

	cert, _ := tls.LoadX509KeyPair(crt, key)
	roots := x509.NewCertPool()

	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr bool
	}{
		{
			name: "Successfully loaded config",
			args: args{
				ca:  "../../test/certs/ca.crt",
				crt: "../../test/certs/server.crt",
				key: "../../test/certs/server.key",
			},
			want: &tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    roots,
			},
			wantErr: true,
		},
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

	var d net.Dialer

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	conn, _ := d.DialContext(ctx, "tcp", "localhost:8080")

	tests := []struct {
		name    string
		args    args
		want    x509.Certificate
		wantErr bool
	}{
		{
			name: "Successfully loaded client certificate",
			args: args{
				conn: conn,
			},
			want:    x509.Certificate{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ClientCert(tt.args.conn)
			if (err != nil) != tt.wantErr {
				t.Errorf("ClientCerteeee() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("ClientCert() = %v, want %v", got, tt.want)
			}
		})
	}
}
