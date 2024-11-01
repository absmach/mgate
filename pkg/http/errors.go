// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import "encoding/json"

type httpProxyError struct {
	statusCode int
	err        error
}

type HTTPProxyError interface {
	error
	MarshalJSON() ([]byte, error)
	StatusCode() int
}

var _ HTTPProxyError = (*httpProxyError)(nil)

func (hpe *httpProxyError) Error() string {
	return hpe.err.Error()
}

func (hpe *httpProxyError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Error string `json:"message"`
	}{
		Error: hpe.err.Error(),
	})
}

func (hpe *httpProxyError) StatusCode() int {
	return hpe.statusCode
}

func NewHTTPProxyError(statusCode int, err error) HTTPProxyError {
	return &httpProxyError{statusCode: statusCode, err: err}
}
