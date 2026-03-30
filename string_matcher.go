// Package fastregex provides a thin wrapper around [regexp.Regexp] that
// auto-detects simple pattern shapes and dispatches to faster string
// operations instead of running the full regex automaton.
//
// Detection uses [regexp/syntax] to inspect the parsed AST.
//
// Recognised fast paths:
//
//	^literal$          → string equality (==)
//	^literal           → strings.HasPrefix
//	^literal.*$        → strings.HasPrefix
//	^.*literal$        → strings.HasSuffix
//	literal$           → strings.HasSuffix
//	literal            → strings.Contains
//	^.*literal.*$      → strings.Contains
//	^literal.*literal$ → strings.HasPrefix && strings.HasSuffix
//
// Everything else falls back to [regexp.Regexp.MatchString].
package fastregex

import (
	"regexp"
	"regexp/syntax"
	"strings"
)

// NewStringMatcher compiles the regular expression and returns a [StringMatcher] that
// uses the fastest possible comparison for the given pattern shape.
func NewStringMatcher(expr string) (StringMatcher, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return NewStringMatcherFromRegexp(re), nil
}

// StringMatcher matches a string.
// It either holds a fast-path implementation for the specific regular expression,
// or the original [regexp.Regexp].
type StringMatcher interface {
	MatchString(s string) bool
}

type (
	exactStringMatcher        string
	prefixStringMatcher       string
	suffixStringMatcher       string
	containsStringMatcher     string
	prefixSuffixStringMatcher struct{ prefix, suffix string }
)

// All of these should get inlined.

func (m exactStringMatcher) MatchString(s string) bool    { return s == string(m) }
func (m prefixStringMatcher) MatchString(s string) bool   { return strings.HasPrefix(s, string(m)) }
func (m suffixStringMatcher) MatchString(s string) bool   { return strings.HasSuffix(s, string(m)) }
func (m containsStringMatcher) MatchString(s string) bool { return strings.Contains(s, string(m)) }
func (m prefixSuffixStringMatcher) MatchString(s string) bool {
	// Without the len check `^ab.*ab$` would match "ab".
	return len(s) >= len(m.prefix)+len(m.suffix) &&
		strings.HasPrefix(s, m.prefix) &&
		strings.HasSuffix(s, m.suffix)
}

// NewStringMatcherFromRegexp analyses the compiled regex and selects the best strategy.
func NewStringMatcherFromRegexp(re *regexp.Regexp) StringMatcher {
	syn, err := syntax.Parse(re.String(), syntax.Perl)
	if err != nil {
		return re
	}
	syn = syn.Simplify()

	info := extractLiteral[string](syn)
	if info.literal == "" {
		return re
	}

	if info.suffix != "" {
		if info.hasBegin && info.hasEnd {
			return prefixSuffixStringMatcher{prefix: info.literal, suffix: info.suffix}
		}
		return re
	}

	constLeft := info.hasBegin && !info.openLeft
	constRight := info.hasEnd && !info.openRight

	switch {
	case constLeft && constRight:
		return exactStringMatcher(info.literal)
	case constLeft:
		return prefixStringMatcher(info.literal)
	case constRight:
		return suffixStringMatcher(info.literal)
	default:
		return containsStringMatcher(info.literal)
	}
}
