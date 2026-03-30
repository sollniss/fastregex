package fastregex

import (
	"bytes"
	"regexp"
	"regexp/syntax"
)

// NewByteMatcher compiles the regular expression and returns a [ByteMatcher] that
// uses the fastest possible comparison for the given pattern shape.
func NewByteMatcher(expr string) (ByteMatcher, error) {
	re, err := regexp.Compile(expr)
	if err != nil {
		return nil, err
	}
	return NewByteMatcherFromRegexp(re), nil
}

// ByteMatcher matches a byte slice.
// It either holds a fast-path implementation for the specific regular expression,
// or the original [regexp.Regexp].
type ByteMatcher interface {
	Match(b []byte) bool
}

type (
	exactByteMatcher        []byte
	prefixByteMatcher       []byte
	suffixByteMatcher       []byte
	containsByteMatcher     []byte
	prefixSuffixByteMatcher struct{ prefix, suffix []byte }
)

// All of these should get inlined.

func (m exactByteMatcher) Match(b []byte) bool    { return bytes.Equal(b, []byte(m)) }
func (m prefixByteMatcher) Match(b []byte) bool   { return bytes.HasPrefix(b, []byte(m)) }
func (m suffixByteMatcher) Match(b []byte) bool   { return bytes.HasSuffix(b, []byte(m)) }
func (m containsByteMatcher) Match(b []byte) bool { return bytes.Contains(b, []byte(m)) }
func (m prefixSuffixByteMatcher) Match(b []byte) bool {
	// Without the len check `^ab.*ab$` would match "ab".
	return len(b) >= len(m.prefix)+len(m.suffix) &&
		bytes.HasPrefix(b, m.prefix) &&
		bytes.HasSuffix(b, m.suffix)
}

// NewByteMatcherFromRegexp analyses the compiled regex and selects the best strategy.
func NewByteMatcherFromRegexp(re *regexp.Regexp) ByteMatcher {
	syn, err := syntax.Parse(re.String(), syntax.Perl)
	if err != nil {
		return re
	}
	syn = syn.Simplify()

	info := extractLiteral[[]byte](syn)
	if len(info.literal) == 0 {
		return re
	}

	if len(info.suffix) > 0 {
		if info.hasBegin && info.hasEnd {
			return prefixSuffixByteMatcher{prefix: info.literal, suffix: info.suffix}
		}
		return re
	}

	constLeft := info.hasBegin && !info.openLeft
	constRight := info.hasEnd && !info.openRight

	switch {
	case constLeft && constRight:
		return exactByteMatcher(info.literal)
	case constLeft:
		return prefixByteMatcher(info.literal)
	case constRight:
		return suffixByteMatcher(info.literal)
	default:
		return containsByteMatcher(info.literal)
	}
}
