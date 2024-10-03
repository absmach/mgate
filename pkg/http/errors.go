// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import mgerrors "github.com/absmach/magistrala/pkg/errors"

type httpProxyError struct {
	statusCode int
	err        mgerrors.Error
}

type HTTPProxyError interface {
	mgerrors.Error
	StatusCode() int
}

var _ HTTPProxyError = (*httpProxyError)(nil)

func (hpe *httpProxyError) Error() string {
	return hpe.err.Error()
}

func (hpe *httpProxyError) Err() mgerrors.Error {
	return hpe.err
}

func (hpe *httpProxyError) Msg() string {
	return hpe.err.Msg()
}

func (hpe *httpProxyError) MarshalJSON() ([]byte, error) {
	return hpe.err.MarshalJSON()
}

func (hpe *httpProxyError) StatusCode() int {
	return hpe.statusCode
}

func NewHTTPProxyError(statusCode int, err error) HTTPProxyError {
	var merr mgerrors.Error
	var ok bool
	if merr, ok = err.(mgerrors.Error); !ok {
		merr = mgerrors.New(err.Error())
	}
	return &httpProxyError{statusCode: statusCode, err: merr}
}
