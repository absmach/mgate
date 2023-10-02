package session

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"net"
	"os"
	"testing"
	"time"

	"github.com/mainflux/mainflux/logger"
)

type config struct {
	logLevel string
}

var (
	ca = "../../certs/ca.crt"
)

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

	handle := newHandler(logger)

	outboundConn, _ := net.Dial("tcp", testURL)

	inboundConn, _ := net.Dial("tcp", testURL)

	roots := x509.NewCertPool()
	caCertPEM, _ := os.ReadFile(ca)
	block, _ := pem.Decode(caCertPEM)
	
	if block == nil {
		t.Fatalf("Failed to load certificate")
	}

	// Parse the certificate from the PEM block
	certificate, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		t.Fatalf("%v", err)
	}
	roots.AppendCertsFromPEM(caCertPEM)

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
				handler:  handle,
				cert:     *certificate,
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
