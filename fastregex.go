package fastregex

import "regexp/syntax"

// Compile is an alias for [NewStringMatcher].
func Compile(expr string) (StringMatcher, error) {
	return NewStringMatcher(expr)
}

// MustCompile is like [Compile] but panics if the expression cannot be parsed.
// It simplifies safe initialisation of global variables holding compiled matchers.
func MustCompile(expr string) StringMatcher {
	m, err := Compile(expr)
	if err != nil {
		panic("fastmatch: Compile(" + expr + "): " + err.Error())
	}
	return m
}

// literalInfo holds the result of decomposing a regex syntax tree.
type literalInfo[T string | []byte] struct {
	literal   T
	suffix    T    // non-empty only for `^prefix.*suffix$` patterns.
	hasBegin  bool // ^ anchor present
	hasEnd    bool // $ anchor present
	openLeft  bool // .* between ^ and literal (anything can precede)
	openRight bool // .* between literal and $ (anything can follow)
}

// extractLiteral attempts to decompose a simplified syntax tree into
// an optional begin anchor, a literal string, and an optional end
// anchor, with optional .* between anchors and the literal.
func extractLiteral[T string | []byte](re *syntax.Regexp) literalInfo[T] {
	switch re.Op {
	case syntax.OpLiteral:
		// Bare literal with no anchors: "hello" → contains.
		if re.Flags&syntax.FoldCase != 0 {
			return literalInfo[T]{}
		}
		return literalInfo[T]{literal: T(string(re.Rune))}

	case syntax.OpConcat:
		return extractFromConcat[T](re.Sub)

	default:
		return literalInfo[T]{}
	}
}

// extractFromConcat examines a concatenation's sub-expressions to find
// an optional ^, a literal core, and an optional $, with optional
// dot-star (.*) between anchors and the literal.
func extractFromConcat[T string | []byte](subs []*syntax.Regexp) literalInfo[T] {
	if len(subs) == 0 {
		return literalInfo[T]{}
	}

	var info literalInfo[T]
	i := 0

	// Optional leading `^`.
	if i < len(subs) &&
		(subs[i].Op == syntax.OpBeginLine ||
			subs[i].Op == syntax.OpBeginText) {
		info.hasBegin = true
		i++
	}

	// Optional leading `.*`.
	if i < len(subs) && isDotStar(subs[i]) {
		info.openLeft = true
		i++
	}

	// Find single literal.
	if i >= len(subs) || subs[i].Op != syntax.OpLiteral {
		return literalInfo[T]{}
	}
	if subs[i].Flags&syntax.FoldCase != 0 {
		return literalInfo[T]{}
	}
	info.literal = T(string(subs[i].Rune))
	if len(info.literal) == 0 {
		return literalInfo[T]{}
	}
	i++

	if i < len(subs) && isDotStar(subs[i]) {
		i++
		// If the next node is a literal,
		// this could be the ^prefix.*suffix$ pattern.
		if i < len(subs) && subs[i].Op == syntax.OpLiteral &&
			subs[i].Flags&syntax.FoldCase == 0 &&
			string(subs[i].Rune) != "" {

			info.suffix = T(string(subs[i].Rune))
			i++
		} else {
			// Just a `.*`.
			info.openRight = true
		}
	}

	// Optional trailing `$`.
	if i < len(subs) &&
		(subs[i].Op == syntax.OpEndLine ||
			subs[i].Op == syntax.OpEndText) {
		info.hasEnd = true
		i++
	}

	// Leftover nodes mean complex structure.
	if i != len(subs) {
		return literalInfo[T]{}
	}

	return info
}

// isDotStar reports whether re matches `.*`.
func isDotStar(re *syntax.Regexp) bool {
	if re.Op != syntax.OpStar || len(re.Sub) != 1 {
		return false
	}
	sub := re.Sub[0].Op
	return sub == syntax.OpAnyChar || sub == syntax.OpAnyCharNotNL
}
