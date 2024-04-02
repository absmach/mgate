package verifier

import (
	"crypto/x509"
)

type Verifier interface {
	VerifyVerifiedPeerCertificates(verifiedChains [][]*x509.Certificate) error
	VerifyRawPeerCertificates(peerCertificates []*x509.Certificate) error
}
