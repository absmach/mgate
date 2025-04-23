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

var errByPassDisabled = errors.New("bypass disabled")

type byPassChecker struct {
	enabled        bool
	byPassPatterns []*regexp.Regexp
}

var _ Checker = (*byPassChecker)(nil)

func NewByPassChecker(byPassPatterns []string) (Checker, error) {
	enabled := len(byPassPatterns) != 0
	var byp []*regexp.Regexp
	for _, expr := range byPassPatterns {
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		byp = append(byp, re)
	}

	return &byPassChecker{
		enabled:        enabled,
		byPassPatterns: byp,
	}, nil
}

func (bpc *byPassChecker) Check(r *http.Request) error {
	if !bpc.enabled {
		return errByPassDisabled
	}
	for _, pattern := range bpc.byPassPatterns {
		if pattern.MatchString(r.URL.Path) {
			return nil
		}
	}
	return fmt.Errorf(errNotByPassed, r.URL.Path)
}
