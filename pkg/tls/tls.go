// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package tls

import (
	"bytes"
	"crypto"
	"bytes"
	"crypto"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"errors"
	"fmt"
	"io"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/caarlos0/env/v10"
	"golang.org/x/crypto/ocsp"
)

var (
	errTLSdetails              = errors.New("failed to get TLS details of connection")
	errLoadCerts               = errors.New("failed to load certificates")
	errLoadServerCA            = errors.New("failed to load Server CA")
	errLoadClientCA            = errors.New("failed to load Client CA")
	errAppendCA                = errors.New("failed to append root ca tls.Config")
	errClientCrt               = errors.New("client certificate not received")
	errRetrieveIssuerCrt       = errors.New("failed to retrieve issuer certificate")
	errReadIssuerCrt           = errors.New("failed to read issuer certificate")
	errParseIssuerCrt          = errors.New("failed to parse issuer certificate")
	errCreateOCSPReq           = errors.New("failed to create OCSP Request")
	errCreateOCSPHTTPReq       = errors.New("failed to create OCSP HTTP Request")
	errParseOCSPUrl            = errors.New("failed to parse OCSP server URL")
	errOCSPReq                 = errors.New("OCSP request failed")
	errOCSPReadResp            = errors.New("failed to read OCSP response")
	errParseOCSPRespForCert    = errors.New("failed to parse OCSP Response for Certificate")
	errParseCert               = errors.New("failed to parse Certificate")
	errInvalidClientValidation = errors.New("invalid client validation method")
	errRetrieveCRL             = errors.New("failed to retrieve CRL")
	errReadCRL                 = errors.New("failed to read CRL")
	errParseCRL                = errors.New("failed to parse CRL")
	errExpiredCRL              = errors.New("crl expired")
	errCertRevoked             = errors.New("certificate revoked")
	errCRLSign                 = errors.New("failed to verify CRL signature")
	errOfflineCRLLoad          = errors.New("failed to load offline CRL file")
	errNoCRL                   = errors.New("neither offline crl file nor crl distribution points in certificate doesn't exists")
)

type Security int

const (
	WithoutTLS Security = iota + 1
	WithTLS
	WithmTLS
	WithmTLSVerify
)

func (s Security) String() string {
	switch s {
	case WithTLS:
		return "with TLS"
	case WithmTLS:
		return "with mTLS"
	case WithmTLSVerify:
		return "with mTLS and validation of client certificate revocation status"
	case WithoutTLS:
		fallthrough
	default:
		return "without TLS"
	}
}
	errTLSdetails           = errors.New("failed to get TLS details of connection")
	errParseRoot            = errors.New("failed to parse root certificate")
	errLoadCerts            = errors.New("failed to load certificates")
	errLoadServerCA         = errors.New("failed to load Server CA")
	errLoadClientCA         = errors.New("failed to load Client CA")
	errAppendCA             = errors.New("failed to append root ca tls.Config")
	errClientCrt            = errors.New("client certificate not received")
	errRetrieveIssuerCrt    = errors.New("failed to retrieve issuer certificate")
	errReadIssuerCrt        = errors.New("failed to read issuer certificate")
	errParseIssuerCrt       = errors.New("failed to parse issuer certificate")
	errCreateOCSPReq        = errors.New("failed to create OCSP Request")
	errCreateOCSPHTTPReq    = errors.New("failed to create OCSP HTTP Request")
	errParseOCSPUrl         = errors.New("failed to parse OCSP server URL")
	errOCSPReq              = errors.New("OCSP request failed")
	errOCSPReadResp         = errors.New("failed to read OCSP response")
	errParseOCSPRespForCert = errors.New("failed to parse OCSP Response for Certificate")
	errParseCert            = errors.New("failed to parse Certificate")
)

type Security int

const (
	WithoutTLS Security = iota
	WithTLS
	WithmTLS
)

func (s Security) String() string {
	switch s {
	case WithTLS:
		return "with TLS"
	case WithmTLS:
		return "with mTLS"
	case WithoutTLS:
		fallthrough
	default:
		return "without TLS"
	}
}

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
	v = strings.TrimSpace(v)
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
	CertFile                            string           `env:"CERT_FILE"                                  envDefault:""`
	KeyFile                             string           `env:"KEY_FILE"                                   envDefault:""`
	ServerCAFile                        string           `env:"SERVER_CA_FILE"                             envDefault:""`
	ClientCAFile                        string           `env:"CLIENT_CA_FILE"                             envDefault:""`
	ClientCertValidationMethods         []ValidateMethod `env:"CLIENT_CERT_VALIDATION_METHODS"             envDefault:""`
	OCSPDepth                           uint             `env:"OCSP_DEPTH"                                 envDefault:"0"`
	OCSPResponderURL                    url.URL          `env:"OCSP_RESPONDER_URL"                          envDefault:""`
	CRLDepth                            uint             `env:"CRL_DEPTH"                                  envDefault:"1"`
	OfflineCRLFile                      string           `env:"OFFLINE_CRL_FILE"                           envDefault:""`
	CRLDistributionPoints               url.URL          `env:"CRL_DISTRIBUTION_POINTS"                    envDefault:""`
	CRLDistributionPointsSignCheck      bool             `env:"CRL_DISTRIBUTION_POINTS_SIGN_CHECK"         envDefault:"false"`
	CRLDistributionPointsIssuerCertFile string           `env:"CRL_DISTRIBUTION_POINTS_ISSUER_CERT_FILE "  envDefault:""`
}

func (c *Config) EnvParse(opts env.Options) error {
	return env.ParseWithOptions(c, opts)
}

// Load return a TLS configuration that can be used in TLS servers
func (c *Config) Load() (*tls.Config, Security, error) {
	tlsConfig := &tls.Config{}
	secure := WithoutTLS
	if c.CertFile != "" || c.KeyFile != "" {
		certificate, err := tls.LoadX509KeyPair(c.CertFile, c.KeyFile)
		if err != nil {
			return nil, secure, errors.Join(errLoadCerts, err)
		}
		tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{certificate},
		}
		secure = WithTLS

		// Loading Server CA file
		rootCA, err := loadCertFile(c.ServerCAFile)
		if err != nil {
			return nil, secure, errors.Join(errLoadServerCA, err)
		}
		if len(rootCA) > 0 {
			if tlsConfig.RootCAs == nil {
				tlsConfig.RootCAs = x509.NewCertPool()
			}
			if !tlsConfig.RootCAs.AppendCertsFromPEM(rootCA) {
				return nil, secure, errAppendCA
			}
		}

		// Loading Client CA File
		clientCA, err := loadCertFile(c.ClientCAFile)
		if err != nil {
			return nil, secure, errors.Join(errLoadClientCA, err)
		}
		if len(clientCA) > 0 {
			if tlsConfig.ClientCAs == nil {
				tlsConfig.ClientCAs = x509.NewCertPool()
			}
			if !tlsConfig.ClientCAs.AppendCertsFromPEM(clientCA) {
				return nil, secure, errAppendCA
			}
			secure = WithmTLS
			tlsConfig.ClientAuth = tls.RequireAndVerifyClientCert
			if len(c.ClientCertValidationMethods) > 0 {
				tlsConfig.VerifyPeerCertificate = c.verifyPeerVerifiedCertificate
				secure = WithmTLSVerify
			}
		}
	}
	return tlsConfig, secure, nil
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

func loadCertFile(certFile string) ([]byte, error) {
	if certFile != "" {
		return os.ReadFile(certFile)
	}
	return []byte{}, nil
}

func (c *Config) verifyPeerRawCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(rawCerts) == 0 {
		return errClientCrt
	}
	var peerCertificates []*x509.Certificate
	for _, rawCert := range rawCerts {
		cert, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return errors.Join(errParseCert, err)
		}
		peerCertificates = append(peerCertificates, cert)
	}

	for _, method := range c.ClientCertValidationMethods {
		switch method {
		case OCSP:
			if err := c.ocspVerifications(peerCertificates); err != nil {
				return err
			}
		case CRL:
			if err := c.crlVerifications(peerCertificates); err != nil {
				return err
			}
		}
	}
	return nil
}

func (c *Config) verifyPeerVerifiedCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(verifiedChains) == 0 {
		return errClientCrt
	}
	for _, method := range c.ClientCertValidationMethods {
		switch method {
		case OCSP:
			for _, verifiedChain := range verifiedChains {
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
		case CRL:
			offlineCRL, err := c.loadOfflineCRL()
			if err != nil {
				return err
			}
			for _, verifiedChain := range verifiedChains {
				for i := range verifiedChain {
					cert := verifiedChain[i]
					issuer := cert
					if i+1 < len(verifiedChain) {
						issuer = verifiedChain[i+1]
					}
					crl := offlineCRL
					if len(cert.CRLDistributionPoints) > 0 {
						crl, err = retrieveCRL(cert.CRLDistributionPoints[0], issuer, true)
						if err != nil {
							return err
						}
					} else {
						if c.CRLDistributionPoints.String() != "" {
							var crlIssuerCrt *x509.Certificate
							if c.CRLDistributionPointsSignCheck {
								if crlIssuerCrt, err = c.loadCRLIssuerCert(); err != nil {
									return err
								}
							}
							crl, err = retrieveCRL(c.CRLDistributionPoints.String(), crlIssuerCrt, c.CRLDistributionPointsSignCheck)
							if err != nil {
								return err
							}
						}
					}
					if err := c.crlVerify(cert, crl); err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (c *Config) ocspVerifications(peerCertificates []*x509.Certificate) error {

	for i, peerCertificate := range peerCertificates {
		issuer := retrieveIssuerCert(peerCertificate.Issuer, peerCertificates)
		if err := c.ocspVerify(peerCertificate, issuer); err != nil {
			return err
		}
		if i+1 == int(c.OCSPDepth) {
			return nil
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
				return fmt.Errorf("neither the issuer certificate is present in the chain nor is the issuer certificate URL present in AIA, certificate common name %s  and serial number %x", peerCertificate.Subject.CommonName, peerCertificate.SerialNumber)
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
	if c.OCSPResponderURL.String() == "" {
		if len(peerCertificate.OCSPServer) < 1 {
			return fmt.Errorf("OCSP Server/Responder URL is not present AIA of certificate with common name %s and serial number %x", peerCertificate.Subject.CommonName, peerCertificate.SerialNumber)
		}
		ocspURL = peerCertificate.OCSPServer[0]
		url, err := url.Parse(peerCertificate.OCSPServer[0])
		if err != nil {
			return errors.Join(errParseOCSPUrl, err)
		}
		ocspURLHost = url.Host
	} else {
		ocspURLHost = c.OCSPResponderURL.Host
		ocspURL = c.OCSPResponderURL.String()
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
		return fmt.Errorf("certificate with command name %s and serial number %x revoked at %v", peerCertificate.Subject.CommonName, peerCertificate.SerialNumber, ocspResponse.RevokedAt)
	case ocsp.ServerFailed:
		return fmt.Errorf("OCSP Server Failed")
	case ocsp.Unknown:
		fallthrough
	default:
		return fmt.Errorf("OCSP status unknown")
	}
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

func (c *Config) crlVerifications(peerCertificates []*x509.Certificate) error {
	offlineCRL, err := c.loadOfflineCRL()
	if err != nil {
		return err
	}
	for i, peerCertificate := range peerCertificates {
		crl := offlineCRL
		issuerCert := retrieveIssuerCert(peerCertificate.Issuer, peerCertificates)
		if len(peerCertificate.CRLDistributionPoints) > 0 {
			crl, err = retrieveCRL(peerCertificate.CRLDistributionPoints[0], issuerCert, true)
			if err != nil {
				return err
			}
		} else {
			if c.CRLDistributionPoints.String() != "" {
				var crlIssuerCrt *x509.Certificate
				if c.CRLDistributionPointsSignCheck {
					if crlIssuerCrt, err = c.loadCRLIssuerCert(); err != nil {
						return err
					}
				}
				crl, err = retrieveCRL(c.CRLDistributionPoints.String(), crlIssuerCrt, c.CRLDistributionPointsSignCheck)
				if err != nil {
					return err
				}
			}
		}
		if crl == nil {
			return errNoCRL
		}

		if err := c.crlVerify(peerCertificate, crl); err != nil {
			return err
		}
		if i+1 == int(c.CRLDepth) {
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
	offlineCRLBytes, err := loadCertFile(c.OfflineCRLFile)
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

func (c *Config) loadCRLIssuerCert() (*x509.Certificate, error) {
	crlIssuerCertBytes, err := loadCertFile(c.OfflineCRLFile)
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
