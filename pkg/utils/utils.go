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
	if c == nil {
		return "no TLS"
	}
	ret := "TLS"
	// It is possible to establish TLS with client certificates only.
	if c.Certificates == nil || len(c.Certificates) == 0 {
		ret = "no server certificates"
	}
	if c.ClientCAs != nil {
		ret += " and " + c.ClientAuth.String()
	}
	return ret
}
