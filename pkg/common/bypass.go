// Copyright (c) Abstract Machines
// SPDX-License-Identifier: Apache-2.0

package common

import (
	"net/http"
	"regexp"
)

type BypassMatcher interface {
	ShouldBypass(r *http.Request) bool
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

func (b *bypassMatcher) ShouldBypass(r *http.Request) bool {
	if !b.enabled {
		return false
	}
	for _, pattern := range b.patterns {
		if pattern.MatchString(r.URL.Path) {
			return true
		}
	}
	return false
}
