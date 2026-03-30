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
