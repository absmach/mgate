// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"log/slog"
	"net/http"
)

const errNotAllowed = "origin - %s is not allowed"

type OriginChecker interface {
	CheckOrigin(r *http.Request) error
}

type originChecker struct {
	enabled        bool
	allowedOrigins map[string]struct{}
}

var _ OriginChecker = (*originChecker)(nil)

func NewOriginChecker(logger *slog.Logger, allowedOrigins []string) OriginChecker {
	enabled := len(allowedOrigins) > 0
	ao := make(map[string]struct{})
	for _, allowedOrigin := range allowedOrigins {
		ao[allowedOrigin] = struct{}{}
	}
	return &originChecker{
		enabled:        enabled,
		allowedOrigins: ao,
	}
}

func (o *originChecker) CheckOrigin(r *http.Request) error {
	if !o.enabled {
		return nil
	}
	origin := r.Header.Get("Origin")
	_, allowed := o.allowedOrigins[origin]
	if allowed {
		return nil
	}
	return fmt.Errorf(errNotAllowed, origin)
}
