// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package coap

import (
	"encoding/json"

	"github.com/plgd-dev/go-coap/v3/message/codes"
)

type coapProxyError struct {
	statusCode codes.Code
	err        error
}

type COAPProxyError interface {
	error
	MarshalJSON() ([]byte, error)
	StatusCode() codes.Code
}

var _ COAPProxyError = (*coapProxyError)(nil)

func (cpe *coapProxyError) Error() string {
	return cpe.err.Error()
}

func (cpe *coapProxyError) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Error string `json:"message"`
	}{
		Error: cpe.err.Error(),
	})
}

func (cpe *coapProxyError) StatusCode() codes.Code {
	return cpe.statusCode
}

func NewCOAPProxyError(statusCode codes.Code, err error) COAPProxyError {
	return &coapProxyError{statusCode: statusCode, err: err}
}
