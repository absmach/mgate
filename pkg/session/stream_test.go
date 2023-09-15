package session

import (
	"context"
	"crypto/x509"
	"fmt"
	"net"
	"testing"

	"github.com/stretchr/testify/require"
	// "github.com/mainflux/mainflux/logger"
)

func TestStream(t *testing.T) {
	type args struct {
		ctx      context.Context
		inbound  net.Conn
		outbound net.Conn
		handler  Handler
		cert     x509.Certificate
	}

	// type Handler struct {
	// 	logger logger.Logger
	// }

	// h := &Handler{
	// 	logger: nil,
	// }

	outboundConn, _ := net.Dial("tcp", "golang.org:80")

	// listener, _ := net.Listen("tcp", ":8080")
	inboundConn, _ := net.Dial("tcp", "localhost:8080")

	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "successfully stream",
			args: args{
				ctx:      context.Background(),
				inbound:  inboundConn,
				outbound: outboundConn,
				handler:  nil,
				cert:     x509.Certificate{},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := Stream(tt.args.ctx, tt.args.inbound, tt.args.outbound, tt.args.handler, tt.args.cert); (err != nil) != tt.wantErr {
				t.Errorf("Stream() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
	for _, tt := range tests {
		err := Stream(tt.args.ctx, tt.args.inbound, tt.args.outbound, tt.args.handler, tt.args.cert)
		require.Nil(t, err, fmt.Sprintf("%s: expected %v got %v\n", tt.name, tt.wantErr, err))
	}
}
