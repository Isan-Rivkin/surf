package vaultsearch

import (
	"regexp"
	"strings"
)

type Matcher interface {
	IsMatch(needle, haystack string) (bool, error)
}

type RegexMatcher struct {
	// lower capital letters of path some/Secret/Val -> some/secret/val
	LowerHaystack bool
	// lower capital Azure_user -> azure_user
	LowerNeedle bool
}

func NewDefaultRegexMatcher() Matcher {
	return &RegexMatcher{
		LowerHaystack: true,
		LowerNeedle: true,
	}
}

func (m *RegexMatcher) IsMatch(needle, haystack string) (bool, error) {
	if m.LowerHaystack {
		haystack = strings.ToLower(haystack)
	}
	if m.LowerNeedle {
		needle = strings.ToLower(needle)
	}

	matched, err := regexp.MatchString(needle, haystack)

	return matched, err
}
