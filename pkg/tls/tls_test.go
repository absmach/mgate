package tls_test

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"testing"
	"time"

	mptls "github.com/absmach/mproxy/pkg/tls"
)

func createTempFile(content []byte, t *testing.T) string {
	tmpfile, err := os.CreateTemp("", "test")
	if err != nil {
		t.Fatalf("Failed to create temp file: %s", err)
	}

	if _, err := tmpfile.Write(content); err != nil {
		t.Fatalf("Failed to write to temp file: %s", err)
	}

	if err := tmpfile.Close(); err != nil {
		t.Fatalf("Failed to close temp file: %s", err)
	}

	return tmpfile.Name()
}

func generateDummyCert(t *testing.T) ([]byte, []byte) {
	priv, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("Failed to generate private key: %s", err)
	}

	cert := &x509.Certificate{
		SerialNumber: big.NewInt(1),
		Subject: pkix.Name{
			Organization: []string{"Test Org"},
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
	}

	certBytes, err := x509.CreateCertificate(rand.Reader, cert, cert, &priv.PublicKey, priv)
	if err != nil {
		t.Fatalf("Failed to create certificate: %s", err)
	}

	certPEM := pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certBytes})
	keyBytes, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		t.Fatalf("Failed to marshal private key: %s", err)
	}
	keyPEM := pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes})

	return certPEM, keyPEM
}

func TestLoadTLSCfg(t *testing.T) {
	certPEM, keyPEM := generateDummyCert(t)

	caFile := createTempFile(certPEM, t)
	defer os.Remove(caFile)

	certFile := createTempFile(certPEM, t)
	defer os.Remove(certFile)

	keyFile := createTempFile(keyPEM, t)
	defer os.Remove(keyFile)

	tests := []struct {
		name    string
		ca      string
		crt     string
		key     string
		wantErr bool
	}{
		{"ValidConfig", caFile, certFile, keyFile, false},
		{"InvalidCAFile", "invalid_ca.pem", certFile, keyFile, true},
		{"InvalidCertFile", caFile, "invalid_cert.pem", keyFile, true},
		{"InvalidKeyFile", caFile, certFile, "invalid_key.pem", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := mptls.LoadTLSCfg(tt.ca, tt.crt, tt.key)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadTLSCfg() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
