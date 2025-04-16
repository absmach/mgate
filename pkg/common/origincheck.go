// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"fmt"
	"log/slog"
	"net/http"
)

type OriginChecker interface {
	CheckOrigin(r *http.Request) bool
}

type originChecker struct {
	enabled        bool
	logger         *slog.Logger
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
		logger:         logger,
		allowedOrigins: ao,
	}
}

func (o *originChecker) CheckOrigin(r *http.Request) bool {
	if !o.enabled {
		return true
	}
	origin := r.Header.Get("Origin")
	_, allowed := o.allowedOrigins[origin]
	if !allowed {
		o.logger.Debug(fmt.Sprintf("Blocked connection from origin: %s", origin))
	}
	return allowed
}
