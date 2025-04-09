package common

import (
	"net/http"
	"regexp"
)

type BypassMatcher struct {
	enabled  bool
	patterns []*regexp.Regexp
}

func NewBypassMatcher(expressions []string) (*BypassMatcher, error) {
	var patterns []*regexp.Regexp
	var enabled bool = false
	if len(expressions) > 0 {
		enabled = true
		for _, expr := range expressions {
			re, err := regexp.Compile(expr)
			if err != nil {
				return nil, err
			}
			patterns = append(patterns, re)
		}
	}
	return &BypassMatcher{enabled: enabled, patterns: patterns}, nil
}

func (b *BypassMatcher) ShouldBypass(r *http.Request) bool {
	if b.enabled {
		for _, pattern := range b.patterns {
			if pattern.MatchString(r.URL.Path) {
				return true
			}
		}
	}
	return false
}
