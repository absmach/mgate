package tls

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"

	"net"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	ca  = "../../certs/ca.crt"
	crt = "../../certs/ca.crt"
	key = "../../certs/ca.key"
)

func TestLoadTLSCfg(t *testing.T) {
	type args struct {
		ca  string
		crt string
		key string
	}

	cert, _ := tls.LoadX509KeyPair(crt, key)
	roots := x509.NewCertPool()
	caCertPEM, _ := os.ReadFile(ca)
	roots.AppendCertsFromPEM(caCertPEM)

	tests := []struct {
		name    string
		args    args
		want    *tls.Config
		wantErr error
	}{
		{
			name: "Successfully loaded config",
			args: args{
				ca:  ca,
				crt: crt,
				key: key,
			},
			want: &tls.Config{
				Certificates: []tls.Certificate{cert},
				ClientAuth:   tls.RequireAndVerifyClientCert,
				ClientCAs:    roots,
			},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		got, err := LoadTLSCfg(tt.args.ca, tt.args.crt, tt.args.key)
		require.Nil(t, err, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
		assert.Equal(t, got.Certificates, tt.want.Certificates, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want.Certificates, got.Certificates))
		assert.Equal(t, got.ClientAuth, tt.want.ClientAuth, fmt.Sprintf("%s: expected %s got %s\n", tt.name, tt.want.ClientAuth, got.ClientAuth))
		assert.Equal(t, got.ClientCAs.Equal(tt.want.ClientCAs), true, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want.ClientCAs, got.ClientCAs))
	}
}

func TestClientCert(t *testing.T) {
	type args struct {
		conn net.Conn
	}

	var d net.Dialer

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	tlsConn, _ := d.DialContext(ctx, "tcp", "golang.org:80")

	tests := []struct {
		name    string
		args    args
		want    x509.Certificate
		wantErr error
	}{
		{
			name: "Successfully loaded tcp client certificate",
			args: args{
				conn: tlsConn,
			},
			want:    x509.Certificate{},
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		got, err := ClientCert(tt.args.conn)
		require.Nil(t, err, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
		assert.Equal(t, got, tt.want, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.want, got))
	}
}
