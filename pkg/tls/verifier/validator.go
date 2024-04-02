// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package verifier

import (
	"crypto/x509"
	"errors"

	"github.com/absmach/mproxy/pkg/tls/verifier/types"
	"github.com/absmach/mproxy/pkg/tls/verifier/validation"
	"github.com/caarlos0/env/v10"
)

var (
	errParseCert = errors.New("failed to parse Certificate")
	errClientCrt = errors.New("client certificate not received")
)

type Validator interface {
	// VerifyPeerCertificate is used to verify certificate using verifiers.
	VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error
}

type config struct {
	validationMethods []types.ValidationMethod
}

var _ Validator = (*config)(nil)

func New(opts env.Options) (Validator, error) {
	vms, err := validation.NewValidationMethods(opts)
	if err != nil {
		return nil, err
	}
	return &config{vms}, nil
}

// Client certificate verification fails when there is partial certificates of either verifiedChains or rawCerts.
func (c *config) VerifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	switch {
	case len(verifiedChains) > 0:
		return c.verifyVerifiedPeerCertificates(verifiedChains)
	case len(rawCerts) > 0:
		return c.verifyRawPeerCertificates(rawCerts)
	default:
		return errClientCrt
	}
}

func (c *config) verifyVerifiedPeerCertificates(verifiedChains [][]*x509.Certificate) error {
	for _, vm := range c.validationMethods {
		if err := vm.VerifyVerifiedPeerCertificates(verifiedChains); err != nil {
			return err
		}
	}
	return nil
}

func (c *config) verifyRawPeerCertificates(rawCerts [][]byte) error {
	var peerCertificates []*x509.Certificate
	peerCertificates, err := c.parseCertificates(rawCerts)
	if err != nil {
		return err
	}
	for _, vm := range c.validationMethods {
		if err := vm.VerifyRawPeerCertificates(peerCertificates); err != nil {
			return err
		}
	}
	return nil
}

func (c *config) parseCertificates(rawCerts [][]byte) ([]*x509.Certificate, error) {
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
