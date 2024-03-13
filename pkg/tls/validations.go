// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"bytes"
	"crypto"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"golang.org/x/crypto/ocsp"
)

var (
	errParseIssuerCrt          = errors.New("failed to parse issuer certificate")
	errCreateOCSPReq           = errors.New("failed to create OCSP Request")
	errCreateOCSPHTTPReq       = errors.New("failed to create OCSP HTTP Request")
	errParseOCSPUrl            = errors.New("failed to parse OCSP server URL")
	errOCSPReq                 = errors.New("OCSP request failed")
	errOCSPReadResp            = errors.New("failed to read OCSP response")
	errParseOCSPRespForCert    = errors.New("failed to parse OCSP Response for Certificate")
	errParseCert               = errors.New("failed to parse Certificate")
	errRetrieveCRL             = errors.New("failed to retrieve CRL")
	errReadCRL                 = errors.New("failed to read CRL")
	errParseCRL                = errors.New("failed to parse CRL")
	errExpiredCRL              = errors.New("crl expired")
	errCRLSign                 = errors.New("failed to verify CRL signature")
	errOfflineCRLLoad          = errors.New("failed to load offline CRL file")
	errNoCRL                   = errors.New("neither offline crl file nor crl distribution points in certificate doesn't exists")
	errInvalidClientValidation = errors.New("invalid client validation method")
	errIssuerCert              = errors.New("neither the issuer certificate is present in the chain nor is the issuer certificate URL present in AIA")
	errNoOCSPURL               = errors.New("OCSP Server/Responder URL is not present AIA of certificate")
	errOCSPServerFailed        = errors.New("OCSP Server Failed")
	errOCSPUnknown             = errors.New("OCSP status unknown")
	errCertRevoked             = errors.New("certificate revoked")
	errRetrieveIssuerCrt       = errors.New("failed to retrieve issuer certificate")
	errReadIssuerCrt           = errors.New("failed to read issuer certificate")
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

type Validation struct {
	ValidationMethods                   []ValidateMethod `env:"CLIENT_CERT_VALIDATION_METHODS"             envDefault:""`
	OCSPDepth                           uint             `env:"OCSP_DEPTH"                                 envDefault:"0"`
	OCSPResponderURL                    url.URL          `env:"OCSP_RESPONDER_URL"                         envDefault:""`
	CRLDepth                            uint             `env:"CRL_DEPTH"                                  envDefault:"1"`
	OfflineCRLFile                      string           `env:"OFFLINE_CRL_FILE"                           envDefault:""`
	CRLDistributionPoints               url.URL          `env:"CRL_DISTRIBUTION_POINTS"                    envDefault:""`
	CRLDistributionPointsSignCheck      bool             `env:"CRL_DISTRIBUTION_POINTS_SIGN_CHECK"         envDefault:"false"`
	CRLDistributionPointsIssuerCertFile string           `env:"CRL_DISTRIBUTION_POINTS_ISSUER_CERT_FILE "  envDefault:""`
}

// Client certificate verification fails when there is partial certificates of either verifiedChains or rawCerts.
func (c *Config) verifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
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
	for _, method := range c.ClientValidation.ValidationMethods {
		switch method {
		case OCSP:
			return c.ocspVerificationVerifiedCerts(verifiedChains)
		case CRL:
			return c.crlVerificationVerifiedCerts(verifiedChains)
		}
	}
	return nil
}

func (c *Config) verifyRawPeerCertificates(rawCerts [][]byte) error {
	var peerCertificates []*x509.Certificate
	for _, rawCert := range rawCerts {
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return errors.Join(errParseCert, err)
		}
		peerCertificates = append(peerCertificates, cert)
	}

	for _, method := range c.ClientValidation.ValidationMethods {
		switch method {
		case OCSP:
			if err := c.ocspVerificationRawCerts(peerCertificates); err != nil {
				return err
			}
		case CRL:
			if err := c.crlVerificationRawCerts(peerCertificates); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) ocspVerificationRawCerts(peerCertificates []*x509.Certificate) error {
	for i, peerCertificate := range peerCertificates {
		issuer := retrieveIssuerCert(peerCertificate.Issuer, peerCertificates)
		if err := c.ocspVerify(peerCertificate, issuer); err != nil {
			return err
		}
		if i+1 == int(c.ClientValidation.OCSPDepth) {
			return nil
		}
	}
	return nil
}

func (c *Config) ocspVerificationVerifiedCerts(verifiedPeerCertificateChains [][]*x509.Certificate) error {
	for _, verifiedChain := range verifiedPeerCertificateChains {
		for i := range verifiedChain {
			cert := verifiedChain[i]
			issuer := cert
			if i+1 < len(verifiedChain) {
				issuer = verifiedChain[i+1]
			}
			if err := c.ocspVerify(cert, issuer); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) ocspVerify(peerCertificate, issuerCert *x509.Certificate) error {
	opts := &ocsp.RequestOptions{Hash: crypto.SHA256}
	var err error

	if !isRootCA(peerCertificate) {
		if issuerCert == nil {
			if len(peerCertificate.IssuingCertificateURL) < 1 {
				return fmt.Errorf("%w common name %s  and serial number %x", errIssuerCert, peerCertificate.Subject.CommonName, peerCertificate.SerialNumber)
			}
			issuerCert, err = retrieveIssuingCertificate(peerCertificate.IssuingCertificateURL[0])
			if err != nil {
				return err
			}
		}
	} else {
		issuerCert = peerCertificate
	}

	buffer, err := ocsp.CreateRequest(peerCertificate, issuerCert, opts)
	if err != nil {
		return errors.Join(errCreateOCSPReq, err)
	}

	ocspURL := ""
	ocspURLHost := ""
	if c.ClientValidation.OCSPResponderURL.String() == "" {
		if len(peerCertificate.OCSPServer) < 1 {
			return fmt.Errorf("%w common name %s and serial number %x", errNoOCSPURL, peerCertificate.Subject.CommonName, peerCertificate.SerialNumber)
		}
		ocspURL = peerCertificate.OCSPServer[0]
		ocspParsedURL, err := url.Parse(peerCertificate.OCSPServer[0])
		if err != nil {
			return errors.Join(errParseOCSPUrl, err)
		}
		ocspURLHost = ocspParsedURL.Host
	} else {
		ocspURLHost = c.ClientValidation.OCSPResponderURL.Host
		ocspURL = c.ClientValidation.OCSPResponderURL.String()
	}

	httpRequest, err := http.NewRequest(http.MethodPost, ocspURL, bytes.NewBuffer(buffer))
	if err != nil {
		return errors.Join(errCreateOCSPHTTPReq, err)
	}
	httpRequest.Header.Add("Content-Type", "application/ocsp-request")
	httpRequest.Header.Add("Accept", "application/ocsp-response")
	httpRequest.Header.Add("host", ocspURLHost)

	httpClient := &http.Client{}
	httpResponse, err := httpClient.Do(httpRequest)
	if err != nil {
		return errors.Join(errOCSPReq, err)
	}
	defer httpResponse.Body.Close()
	output, err := io.ReadAll(httpResponse.Body)
	if err != nil {
		return errors.Join(errOCSPReadResp, err)
	}
	ocspResponse, err := ocsp.ParseResponseForCert(output, peerCertificate, issuerCert)
	if err != nil {
		return errors.Join(errParseOCSPRespForCert, err)
	}
	switch ocspResponse.Status {
	case ocsp.Good:
		return nil
	case ocsp.Revoked:
		return fmt.Errorf("%w command name %s and serial number %x revoked at %v", errCertRevoked, peerCertificate.Subject.CommonName, peerCertificate.SerialNumber, ocspResponse.RevokedAt)
	case ocsp.ServerFailed:
		return errOCSPServerFailed
	case ocsp.Unknown:
		fallthrough
	default:
		return errOCSPUnknown
	}
}

func (c *Config) crlVerificationVerifiedCerts(verifiedPeerCertificateChains [][]*x509.Certificate) error {
	offlineCRL, err := c.loadOfflineCRL()
	if err != nil {
		return err
	}
	for _, verifiedChain := range verifiedPeerCertificateChains {
		for i := range verifiedChain {
			cert := verifiedChain[i]
			issuer := cert
			if i+1 < len(verifiedChain) {
				issuer = verifiedChain[i+1]
			}

			crl, err := c.getCRLFromDistributionPoint(cert, issuer)
			if err != nil {
				return err
			}
			switch {
			case crl == nil && offlineCRL != nil:
				crl = offlineCRL
			case crl == nil && offlineCRL == nil:
				return errNoCRL
			}

			if err := c.crlVerify(cert, crl); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) crlVerificationRawCerts(peerCertificates []*x509.Certificate) error {
	offlineCRL, err := c.loadOfflineCRL()
	if err != nil {
		return err
	}
	for i, peerCertificate := range peerCertificates {
		issuerCert := retrieveIssuerCert(peerCertificate.Issuer, peerCertificates)
		crl, err := c.getCRLFromDistributionPoint(peerCertificate, issuerCert)
		if err != nil {
			return err
		}
		switch {
		case crl == nil && offlineCRL != nil:
			crl = offlineCRL
		case crl == nil && offlineCRL == nil:
			return errNoCRL
		}

		if err := c.crlVerify(peerCertificate, crl); err != nil {
			return err
		}
		if i+1 == int(c.ClientValidation.CRLDepth) {
			return nil
		}
	}
	return nil
}

func (c *Config) crlVerify(peerCertificate *x509.Certificate, crl *x509.RevocationList) error {
	for _, revokedCertificate := range crl.RevokedCertificateEntries {
		if revokedCertificate.SerialNumber.Cmp(peerCertificate.SerialNumber) == 0 {
			return errCertRevoked
		}
	}
	return nil
}

func (c *Config) loadOfflineCRL() (*x509.RevocationList, error) {
	offlineCRLBytes, err := loadCertFile(c.ClientValidation.OfflineCRLFile)
	if err != nil {
		return nil, errors.Join(errOfflineCRLLoad, err)
	}
	if len(offlineCRLBytes) == 0 {
		return nil, nil
	}
	offlineCRL, err := parseVerifyCRL(offlineCRLBytes, nil, false)
	if err != nil {
		return nil, err
	}
	return offlineCRL, nil
}

func (c *Config) getCRLFromDistributionPoint(cert, issuer *x509.Certificate) (*x509.RevocationList, error) {
	switch {
	case len(cert.CRLDistributionPoints) > 0:
		return retrieveCRL(cert.CRLDistributionPoints[0], issuer, true)
	default:
		if c.ClientValidation.CRLDistributionPoints.String() == "" {
			return nil, nil
		}
		var crlIssuerCrt *x509.Certificate
		var err error
		if c.ClientValidation.CRLDistributionPointsSignCheck {
			if crlIssuerCrt, err = c.loadCRLIssuerCert(); err != nil {
				return nil, err
			}
		}
		return retrieveCRL(c.ClientValidation.CRLDistributionPoints.String(), crlIssuerCrt, c.ClientValidation.CRLDistributionPointsSignCheck)
	}
}

func (c *Config) loadCRLIssuerCert() (*x509.Certificate, error) {
	crlIssuerCertBytes, err := loadCertFile(c.ClientValidation.OfflineCRLFile)
	if err != nil {
		return nil, errors.Join(errOfflineCRLLoad, err)
	}
	if len(crlIssuerCertBytes) == 0 {
		return nil, nil
	}
	crlIssuerCert, err := x509.ParseCertificate(crlIssuerCertBytes)
	if err != nil {
		return nil, err
	}
	return crlIssuerCert, nil
}

func retrieveIssuerCert(issuerSubject pkix.Name, certs []*x509.Certificate) *x509.Certificate {
	for _, cert := range certs {
		if cert.Subject.SerialNumber != "" && issuerSubject.SerialNumber != "" && cert.Subject.SerialNumber == issuerSubject.SerialNumber {
			return cert
		}
		if (cert.Subject.SerialNumber == "" || issuerSubject.SerialNumber == "") && cert.Subject.String() == issuerSubject.String() {
			return cert
		}
	}
	return nil
}

func retrieveIssuingCertificate(issuingCertificateURL string) (*x509.Certificate, error) {
	resp, err := http.Get(issuingCertificateURL)
	if err != nil {
		return nil, errors.Join(errRetrieveIssuerCrt, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(errReadIssuerCrt, err)
	}
	issCert, err := x509.ParseCertificate(body)
	if err != nil {
		return nil, errors.Join(errParseIssuerCrt, err)
	}
	return issCert, nil
}

func isRootCA(cert *x509.Certificate) bool {
	if cert.IsCA {
		// Check AuthorityKeyId and SubjectKeyId are same.
		if len(cert.AuthorityKeyId) > 0 && len(cert.SubjectKeyId) > 0 && bytes.Equal(cert.AuthorityKeyId, cert.SubjectKeyId) {
			return true
		}
		// Alternatively, check Issuer and Subject are same.
		if cert.Issuer.String() == cert.Subject.String() {
			return true
		}
	}
	return false
}

func retrieveCRL(crlDistributionPoints string, issuerCert *x509.Certificate, checkSign bool) (*x509.RevocationList, error) {
	resp, err := http.Get(crlDistributionPoints)
	if err != nil {
		return nil, errors.Join(errRetrieveCRL, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Join(errReadCRL, err)
	}
	return parseVerifyCRL(body, issuerCert, checkSign)
}

func parseVerifyCRL(clrB []byte, issuerCert *x509.Certificate, checkSign bool) (*x509.RevocationList, error) {
	crl, err := x509.ParseRevocationList(clrB)
	if err != nil {
		return nil, errors.Join(errParseCRL, err)
	}

	if checkSign {
		if err := crl.CheckSignatureFrom(issuerCert); err != nil {
			return nil, errors.Join(errCRLSign, err)
		}
	}

	if crl.NextUpdate.Before(time.Now()) {
		return nil, errExpiredCRL
	}
	return crl, nil
}
