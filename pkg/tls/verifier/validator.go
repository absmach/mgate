// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"crypto/x509"
	"errors"
)

var (
	errParseCert = errors.New("failed to parse Certificate")
	errClientCrt = errors.New("client certificate not received")
)

type Validator func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error

func NewValidator(verifiers []Verifier) Validator {
	return func(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
		switch {
		case len(verifiedChains) > 0:
			for _, vm := range verifiers {
				if err := vm.VerifyVerifiedPeerCertificates(verifiedChains); err != nil {
					return err
				}
			}
			return nil
		case len(rawCerts) > 0:
			var peerCertificates []*x509.Certificate
			peerCertificates, err := parseCertificates(rawCerts)
			if err != nil {
				return err
			}
			for _, vm := range verifiers {
				if err := vm.VerifyRawPeerCertificates(peerCertificates); err != nil {
					return err
				}
			}
			return nil
		default:
			return errClientCrt
		}
	}
}

func parseCertificates(rawCerts [][]byte) ([]*x509.Certificate, error) {
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
