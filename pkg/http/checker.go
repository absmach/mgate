// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

const errNotBypassFmt = "route - %s is not in bypass list"

type bypassChecker struct {
	enabled        bool
	byPassPatterns []*regexp.Regexp
}

type originChecker struct {
	enabled        bool
	allowedOrigins map[string]struct{}
}

var (
	errBypassDisabled = errors.New("bypass disabled")
	errNotAllowed     = "origin - %s is not allowed"

	_ Checker = (*originChecker)(nil)
	_ Checker = (*bypassChecker)(nil)
)

func NewBypassChecker(byPassPatterns []string) (Checker, error) {
	enabled := len(byPassPatterns) != 0
	var byp []*regexp.Regexp
	for _, expr := range byPassPatterns {
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		byp = append(byp, re)
	}

	return &bypassChecker{
		enabled:        enabled,
		byPassPatterns: byp,
	}, nil
}

func (bpc *bypassChecker) Check(r *http.Request) error {
	if !bpc.enabled {
		return errBypassDisabled
	}
	for _, pattern := range bpc.byPassPatterns {
		if pattern.MatchString(r.URL.Path) {
			return nil
		}
	}
	return fmt.Errorf(errNotBypassFmt, r.URL.Path)
}

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
