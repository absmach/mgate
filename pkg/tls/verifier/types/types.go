package types

import (
	"crypto/x509"
	"errors"
	"strings"
)

// Invalid client validation method
var ErrInvalidClientValidation = errors.New("invalid client validation method")

type Validation int

type ValidationMethod interface {
	VerifyVerifiedPeerCertificates(verifiedChains [][]*x509.Certificate) error
	VerifyRawPeerCertificates(peerCertificates []*x509.Certificate) error
}

const (
	OCSP Validation = iota + 1
	CRL
)

func ParseValidation(v string) (Validation, error) {
	v = strings.ToUpper(strings.TrimSpace(v))
	switch v {
	case "OCSP":
		return OCSP, nil
	case "CRL":
		return CRL, nil
	default:
		return 0, ErrInvalidClientValidation
	}
}
