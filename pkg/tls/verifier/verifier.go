package verifier

import (
	"crypto/x509"
)

type Verifier interface {
	VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error
}

type Validator func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

func NewValidator(verifiers []Verifier) Validator {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		for _, vm := range verifiers {
			if err := vm.VerifyPeerCertificate(rawCerts, verifiedChains); err != nil {
				return err
			}
		}
		return nil
	}
}
