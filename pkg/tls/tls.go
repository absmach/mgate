// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"net"
	"os"
	"strings"

	"github.com/absmach/mproxy/pkg/tls/verifier"
	"github.com/caarlos0/env/v10"
)

var (
	errTLSdetails   = errors.New("failed to get TLS details of connection")
	errLoadCerts    = errors.New("failed to load certificates")
	errLoadServerCA = errors.New("failed to load Server CA")
	errLoadClientCA = errors.New("failed to load Client CA")
	errAppendCA     = errors.New("failed to append root ca tls.Config")
)

type Security int

const (
	WithoutTLS Security = iota + 1
	WithTLS
	WithmTLS
	WithmTLSVerify
)

func (s Security) String() string {
	switch s {
	case WithTLS:
		return "with TLS"
	case WithmTLS:
		return "with mTLS"
	case WithmTLSVerify:
		return "with mTLS and validation of client certificate revocation status"
	case WithoutTLS:
		fallthrough
	default:
		return "without TLS"
	}
}

type Config struct {
	CertFile     string `env:"CERT_FILE"                                  envDefault:""`
	KeyFile      string `env:"KEY_FILE"                                   envDefault:""`
	ServerCAFile string `env:"SERVER_CA_FILE"                             envDefault:""`
	ClientCAFile string `env:"CLIENT_CA_FILE"                             envDefault:""`
	Verifier     verifier.Config
}

func (c *Config) SecurityStatus() string {
	if c.CertFile == "" && c.KeyFile == "" {
		return "TLS"
	}

	if c.ClientCAFile != "" {
		if len(c.Verifier.ValidationMethods) > 0 {
			validations := []string{}
			for _, v := range c.Verifier.ValidationMethods {
				validations = append(validations, v.String())
			}
			return fmt.Sprintf("mTLS with client verification %s", strings.Join(validations, ", "))
		}
		return "mTLS"
	}

	return "no TLS"
}

func (c *Config) EnvParse(opts env.Options) error {
	return env.ParseWithOptions(c, opts)
}

// Load return a TLS configuration that can be used in TLS servers.
func (c *Config) Load() (*tls.Config, error) {
	if c.CertFile == "" || c.KeyFile == "" {
		return nil, nil
	}

	tlsConfig := &tls.Config{}

	certificate, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, errors.Join(errLoadCerts, err)
	}
	tlsConfig = &tls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	// Loading Server CA file
	rootCA, err := loadCertFile(c.ServerCAFile)
	if err != nil {
		return nil, errors.Join(errLoadServerCA, err)
	}
	if len(rootCA) > 0 {
		if tlsConfig.RootCAs == nil {
			tlsConfig.RootCAs = x509.NewCertPool()
		}
		if !tlsConfig.RootCAs.AppendCertsFromPEM(rootCA) {
			return nil, errAppendCA
		}
	}

	// Loading Client CA File
	clientCA, err := loadCertFile(c.ClientCAFile)
	if err != nil {
		return nil, errors.Join(errLoadClientCA, err)
	}
	if len(clientCA) > 0 {
		if tlsConfig.ClientCAs == nil {
			tlsConfig.ClientCAs = x509.NewCertPool()
		}
		if !tlsConfig.ClientCAs.AppendCertsFromPEM(clientCA) {
			return nil, errAppendCA
		}
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
		if len(c.Verifier.ValidationMethods) > 0 {
			tlsConfig.VerifyPeerCertificate = c.Verifier.VerifyPeerCertificate
		}
	}
	return tlsConfig, nil
}

func (c *Config) Security() Security {
	if c.CertFile == "" && c.KeyFile == "" {
		return WithoutTLS
	}

	if c.ClientCAFile != "" {
		if len(c.Verifier.ValidationMethods) > 0 {
			return WithmTLSVerify
		}
		return WithmTLS
	}

	return WithTLS
}

// ClientCert returns client certificate.
func ClientCert(conn net.Conn) (x509.Certificate, error) {
	switch connVal := conn.(type) {
	case *tls.Conn:
		if err := connVal.Handshake(); err != nil {
			return x509.Certificate{}, err
		}
		state := connVal.ConnectionState()
		if state.Version == 0 {
			return x509.Certificate{}, errTLSdetails
		}
		if len(state.PeerCertificates) == 0 {
			return x509.Certificate{}, nil
		}
		cert := *state.PeerCertificates[0]
		return cert, nil
	default:
		return x509.Certificate{}, nil
	}
}

func loadCertFile(certFile string) ([]byte, error) {
	if certFile != "" {
		return os.ReadFile(certFile)
	}
	return []byte{}, nil
}
