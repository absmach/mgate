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
	errTLSdetails     = errors.New("failed to get TLS details of connection")
	errLoadCerts      = errors.New("failed to load certificates")
	errLoadCA         = errors.New("failed to load CA file")
	errAppendClientCA = errors.New("failed to append client root ca tls.Config")
	errAppendServerCA = errors.New("failed to append server root ca tls.Config")
)

// LoadTLS returns a TLS configuration that can be used in TLS servers.
func LoadTLS(c *Config) (*tls.Config, error) {
	certificate, err := loadCertificates(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, err
	}
	if certificate == nil {
		return nil, nil
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{*certificate},
	}

	// Loading Server CA file
	if _, err = appendCAs(&tlsConfig.RootCAs, c.ServerCAFile); err != nil {
		return nil, errors.Join(errAppendServerCA, err)
	}

	// Loading Client CA File
	appended, err := appendCAs(&tlsConfig.ClientCAs, c.ClientCAFile)
	if err != nil {
		return nil, errors.Join(errAppendClientCA, err)
	}

	if appended {
		tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
	}
	if c.Validator != nil {
		tlsConfig.VerifyPeerCertificate = c.Validator
	}
	return tlsConfig, nil
}

// LoadDTLS returns a DTLS configuration that can be used in DTLS servers.
func LoadDTLS(c *Config) (*dtls.Config, error) {
	certificate, err := loadCertificates(c.CertFile, c.KeyFile)
	if err != nil {
		return nil, err
	}
	if certificate == nil {
		return nil, nil
	}

	dtlsConfig := &dtls.Config{
		Certificates: []tls.Certificate{*certificate},
	}

	// Loading Server CA file
	if _, err = appendCAs(&dtlsConfig.RootCAs, c.ServerCAFile); err != nil {
		return nil, errors.Join(errAppendServerCA, err)
	}

	// Loading Client CA File
	appended, err := appendCAs(&dtlsConfig.ClientCAs, c.ClientCAFile)
	if err != nil {
		return nil, errors.Join(errAppendClientCA, err)
	}

	if appended {
		dtlsConfig.ClientAuth = dtls.RequireAndVerifyClientCert
	}
	if c.Validator != nil {
		dtlsConfig.VerifyPeerCertificate = c.Validator
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

func loadCertificates(certFile, keyFile string) (*tls.Certificate, error) {
	if certFile == "" || keyFile == "" {
		return nil, nil
	}
	cert, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		return nil, errors.Join(errLoadCerts, err)
	}
	return &cert, err
}

func appendCAs(certPool **x509.CertPool, caFile string) (bool, error) {
	ca, err := loadCertFile(caFile)
	if err != nil {
		return false, errors.Join(errLoadCA, err)
	}
	if len(ca) > 0 {
		if *certPool == nil {
			*certPool = x509.NewCertPool()
		}
		if !(*certPool).AppendCertsFromPEM(ca) {
			return false, errors.New("failed to append CA certificates")
		}
		return true, nil
	}
	return false, nil
}
