// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"fmt"
	"net/http"
)

var errNotAllowed = "origin - %s is not allowed"

type originChecker struct {
	enabled        bool
	allowedOrigins map[string]struct{}
}

var _ Checker = (*originChecker)(nil)

func NewOriginChecker(allowedOrigins []string) Checker {
	enabled := len(allowedOrigins) != 0
	ao := make(map[string]struct{})
	for _, allowedOrigin := range allowedOrigins {
		ao[allowedOrigin] = struct{}{}
	}

	return &originChecker{
		enabled:        enabled,
		allowedOrigins: ao,
	}
}

func (oc *originChecker) Check(r *http.Request) error {
	if !oc.enabled {
		return nil
	}
	origin := r.Header.Get("Origin")
	_, allowed := oc.allowedOrigins[origin]
	if allowed {
		return nil
	}
	return fmt.Errorf(errNotAllowed, origin)
}
