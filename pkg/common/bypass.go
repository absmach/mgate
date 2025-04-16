// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
)

const errNotByPassed = "route - %s is not in bypass list"

var errByPassDisabled = errors.New("bypass disabled")

type BypassMatcher interface {
	ShouldBypass(r *http.Request) error
}

type bypassMatcher struct {
	enabled  bool
	patterns []*regexp.Regexp
}

var _ BypassMatcher = (*bypassMatcher)(nil)

func NewBypassMatcher(expressions []string) (BypassMatcher, error) {
	var patterns []*regexp.Regexp
	enabled := len(expressions) > 0
	for _, expr := range expressions {
		re, err := regexp.Compile(expr)
		if err != nil {
			return nil, err
		}
		patterns = append(patterns, re)
	}
	return &bypassMatcher{enabled: enabled, patterns: patterns}, nil
}

func (b *bypassMatcher) ShouldBypass(r *http.Request) error {
	if !b.enabled {
		return errByPassDisabled
	}
	for _, pattern := range b.patterns {
		if pattern.MatchString(r.URL.Path) {
			return nil
		}
	}
	return fmt.Errorf(errNotByPassed, r.URL.Path)
}
