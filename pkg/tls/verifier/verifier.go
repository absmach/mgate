// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"crypto/x509"
	"errors"
	"strings"

	"github.com/absmach/mproxy/pkg/tls/verifier/crl"
	"github.com/absmach/mproxy/pkg/tls/verifier/ocsp"
)

var (
	errParseCert               = errors.New("failed to parse Certificate")
	errInvalidClientValidation = errors.New("invalid client validation method")
	errClientCrt               = errors.New("client certificate not received")
)

type ValidateMethod int

const (
	OCSP ValidateMethod = iota + 1
	CRL
)

func (v ValidateMethod) String() string {
	switch v {
	case OCSP:
		return "OCSP"
	case CRL:
		return "CRL"
	default:
		return ""
	}
}

func ParseValidateMethod(v string) (ValidateMethod, error) {
	v = strings.ToUpper(strings.TrimSpace(v))
	switch v {
	case "OCSP":
		return OCSP, nil
	case "CRL":
		return CRL, nil
	default:
		return 0, errInvalidClientValidation
	}
}

type Config struct {
	ValidationMethods []ValidateMethod `env:"CLIENT_CERT_VALIDATION_METHODS"             envDefault:""`
	OCSP              ocsp.Config
	CRL               crl.Config
}

// Client certificate verification fails when there is partial certificates of either verifiedChains or rawCerts.
func (c *Config) VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	switch {
	case len(verifiedChains) > 0:
		return c.verifyVerifiedPeerCertificates(verifiedChains)
	case len(rawCerts) > 0:
		return c.verifyRawPeerCertificates(rawCerts)
	default:
		return errClientCrt
	}
}

func (c *Config) verifyVerifiedPeerCertificates(verifiedChains [][]*x509.Certificate) error {
	for _, method := range c.ValidationMethods {
		switch method {
		case OCSP:
			return c.OCSP.VerificationVerifiedCerts(verifiedChains)
		case CRL:
			return c.CRL.VerificationVerifiedCerts(verifiedChains)
		}
	}
	return nil
}

func (c *Config) verifyRawPeerCertificates(rawCerts [][]byte) error {
	var peerCertificates []*x509.Certificate
	peerCertificates, err := c.parseCertificates(rawCerts)
	if err != nil {
		return err
	}
	for _, method := range c.ValidationMethods {
		switch method {
		case OCSP:
			if err := c.OCSP.VerificationRawCerts(peerCertificates); err != nil {
				return err
			}
		case CRL:
			if err := c.CRL.VerificationRawCerts(peerCertificates); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) parseCertificates(rawCerts [][]byte) ([]*x509.Certificate, error) {
	var certs []*x509.Certificate
	for _, rawCert := range rawCerts {
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return nil, errors.Join(errParseCert, err)
		}
		certs = append(certs, cert)
	}
	return certs, nil
}
