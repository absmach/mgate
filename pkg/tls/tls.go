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

	"golang.org/x/crypto/ocsp"
)

var (
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

// LoadTLSCfg return a TLS configuration that can be used in TLS servers
func LoadTLSCfg(serverCA, clientCA, crt, key string) (*tls.Config, Security, error) {
	tlsConfig := &tls.Config{}
	secure := WithoutTLS
	if crt != "" || key != "" {
		certificate, err := tls.LoadX509KeyPair(crt, key)
		if err != nil {
			return nil, secure, errors.Join(errLoadCerts, err)
		}
		tlsConfig = &tls.Config{
			ClientAuth:   tls.RequireAndVerifyClientCert,
			Certificates: []tls.Certificate{certificate},
		}

		// Loading Server CA file
		rootCA, err := loadCertFile(serverCA)
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
			secure = WithTLS
		}

		// Loading Client CA File
		clientCA, err := loadCertFile(clientCA)
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
			tlsConfig.VerifyPeerCertificate = verifyPeerCertificate
			secure = WithmTLS
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

func verifyPeerCertificate(rawCerts [][]byte, verifiedChains [][]*x509.Certificate) error {
	if len(rawCerts) == 0 {
		return errClientCrt
	}
	for _, rawCert := range rawCerts {
		peerCertificate, err := x509.ParseCertificate(rawCert)
		if err != nil {
			return errors.Join(errParseCert, err)
		}
		if err := ocspVerify(peerCertificate); err != nil {
			return err
		}
	}
	return nil
}

func ocspVerify(peerCertificate *x509.Certificate) error {
	opts := &ocsp.RequestOptions{Hash: crypto.SHA256}
	issuerCert := peerCertificate
	var err error
	if !isRootCA(peerCertificate) {
		if len(peerCertificate.IssuingCertificateURL) < 1 {
			return fmt.Errorf("certificate issuer URL is not present AIA of certificate with common name %s  and serial number %x", peerCertificate.Subject.CommonName, peerCertificate.SerialNumber)
		}
		issuerCert, err = retrieveIssuingCertificate(peerCertificate.IssuingCertificateURL[0])
		if err != nil {
			return err
		}
	}

	buffer, err := ocsp.CreateRequest(peerCertificate, issuerCert, opts)
	if err != nil {
		return errors.Join(errCreateOCSPReq, err)
	}
	if len(peerCertificate.OCSPServer) < 1 {
		return fmt.Errorf("OCSP Server/Responder URL is not present AIA of certificate with common name %s and serial number %x", peerCertificate.Subject.CommonName, peerCertificate.SerialNumber)
	}
	httpRequest, err := http.NewRequest(http.MethodPost, peerCertificate.OCSPServer[0], bytes.NewBuffer(buffer))
	if err != nil {
		return errors.Join(errCreateOCSPHTTPReq, err)
	}
	ocspURL, err := url.Parse(peerCertificate.OCSPServer[0])
	if err != nil {
		return errors.Join(errParseOCSPUrl, err)
	}
	httpRequest.Header.Add("Content-Type", "application/ocsp-request")
	httpRequest.Header.Add("Accept", "application/ocsp-response")
	httpRequest.Header.Add("host", ocspURL.Host)

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
