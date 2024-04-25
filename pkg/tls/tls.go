// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"net"
	"os"

	"github.com/pion/dtls/v2"
)

var (
	errTLSdetails   = errors.New("failed to get TLS details of connection")
	errLoadCerts    = errors.New("failed to load certificates")
	errLoadServerCA = errors.New("failed to load Server CA")
	errLoadClientCA = errors.New("failed to load Client CA")
	errAppendCA     = errors.New("failed to append root ca tls.Config")
)

// LoadTLS returns a TLS configuration that can be used in TLS servers.
func LoadTLS(c *Config) (*tls.Config, error) {
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
		if c.Validator != nil {
			tlsConfig.VerifyPeerCertificate = c.Validator
		}
	}
	return tlsConfig, nil
}

// LoadDTLS returns a DTLS configuration that can be used in DTLS servers.
func LoadDTLS(c *Config) (*dtls.Config, error) {
	if c.CertFile == "" || c.KeyFile == "" {
		return nil, nil
	}

	dtlsConfig := &dtls.Config{}

	certificate, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, errors.Join(errLoadCerts, err)
	}
	dtlsConfig = &dtls.Config{
		Certificates: []tls.Certificate{certificate},
	}

	// Loading Server CA file
	rootCA, err := loadCertFile(c.ServerCAFile)
	if err != nil {
		return nil, errors.Join(errLoadServerCA, err)
	}
	if len(rootCA) > 0 {
		if dtlsConfig.RootCAs == nil {
			dtlsConfig.RootCAs = x509.NewCertPool()
		}
		if !dtlsConfig.RootCAs.AppendCertsFromPEM(rootCA) {
			return nil, errAppendCA
		}
	}

	// Loading Client CA File
	clientCA, err := loadCertFile(c.ClientCAFile)
	if err != nil {
		return nil, errors.Join(errLoadClientCA, err)
	}
	if len(clientCA) > 0 {
		if dtlsConfig.ClientCAs == nil {
			dtlsConfig.ClientCAs = x509.NewCertPool()
		}
		if !dtlsConfig.ClientCAs.AppendCertsFromPEM(clientCA) {
			return nil, errAppendCA
		}
		dtlsConfig.ClientAuth = dtls.RequireAndVerifyClientCert
		if c.Validator != nil {
			dtlsConfig.VerifyPeerCertificate = c.Validator
		}
	}
	return dtlsConfig, nil
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

// SecurityStatus returns log message from TLS config.
func SecurityStatus(c *tls.Config) string {
	if c == nil {
		return "no TLS"
	}
	ret := "TLS"
	// It is possible to establish TLS with client certificates only.
	if len(c.Certificates) == 0 {
		ret = "no server certificates"
	}
	if c.ClientCAs != nil {
		ret += " and " + c.ClientAuth.String()
	}
	return ret
}

func loadCertFile(certFile string) ([]byte, error) {
	if certFile != "" {
		return os.ReadFile(certFile)
	}
	return []byte{}, nil
}
