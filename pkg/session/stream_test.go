package session

import (
	"context"
	"crypto/x509"
	"net"
	"os"
	"testing"
	"time"

	"github.com/mainflux/mainflux/logger"
)

type config struct {
	logLevel string
}

func TestStream(t *testing.T) {
	type args struct {
		ctx      context.Context
		inbound  net.Conn
		outbound net.Conn
		handler  Handler
		cert     x509.Certificate
	}

	cfg := config{
		logLevel: "info",
	}

	logger, _ := logger.New(os.Stdout, cfg.logLevel)

	outboundConn, _ := net.Dial("tcp", "golang.org:80")

	// listener, _ := net.Listen("tcp", ":8080")
	inboundConn, _ := net.Dial("tcp", "localhost:8080")

	tests := []struct {
		name        string
		args        args
		wantErr     bool
		timeoutSecs int
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
			wantErr:     false,
			timeoutSecs: 5,
		},
	}
	for _, tt := range tests {
		ctx, cancel := context.WithTimeout(tt.args.ctx, time.Duration(tt.timeoutSecs)*time.Second)

		defer cancel()

		errChan := make(chan error, 1)

		go func() {
			errChan <- Stream(ctx, tt.args.inbound, tt.args.outbound, tt.args.handler, tt.args.cert)
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
