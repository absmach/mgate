// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package http

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

const errNotByPassed = "route - %s is not in bypass list"

var (
	errNotAllowed     = "origin - %s is not allowed"
	errByPassDisabled = errors.New("bypass disabled")
)

type Checkers interface {
	ShouldBypass(r *http.Request) error
	CheckOrigin(r *http.Request) error
}

type checkers struct {
	allowedOrigins map[string]struct{}
	byPassPatterns []*regexp.Regexp
}

var _ Checkers = (*checkers)(nil)

func NewCheckers(allowedOrigins []string, byPassPatterns []string) (Checkers, error) {
	var byp []*regexp.Regexp
	for _, expr := range byPassPatterns {
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		byp = append(byp, re)
	}

	ao := make(map[string]struct{})
	for _, allowedOrigin := range allowedOrigins {
		ao[allowedOrigin] = struct{}{}
	}

	return &checkers{
		allowedOrigins: ao,
		byPassPatterns: byp,
	}, nil
}

func (c *checkers) CheckOrigin(r *http.Request) error {
	if len(c.allowedOrigins) == 0 {
		return nil
	}
	origin := r.Header.Get("Origin")
	_, allowed := c.allowedOrigins[origin]
	if allowed {
		return nil
	}
	return fmt.Errorf(errNotAllowed, origin)
}

func (c *checkers) ShouldBypass(r *http.Request) error {
	if len(c.byPassPatterns) == 0 {
		return errByPassDisabled
	}
	for _, pattern := range c.byPassPatterns {
		if pattern.MatchString(r.URL.Path) {
			return nil
		}
	}
	return fmt.Errorf(errNotByPassed, r.URL.Path)
}
