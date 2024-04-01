package utils

import (
	"crypto/tls"
	"crypto/x509"
)

type ValidationMethod interface {
	VerifyVerifiedPeerCertificates(verifiedChains [][]*x509.Certificate) error
	VerifyRawPeerCertificates(peerCertificates []*x509.Certificate) error
}

func SecurityStatus(c *tls.Config) string {
	if c == nil || c.Certificates == nil || len(c.Certificates) == 0 {
		return "no TLS"
	}
	if c.ClientCAs != nil {
		return c.ClientAuth.String()
	}
	return "TLS"
}
