package fastregex

import (
	"fmt"
	"regexp"
	"testing"
)

func TestCorrectness(t *testing.T) {
	patterns := []string{
		// Exact
		`^a$`,
		`^\(\*\.\)$`,
		`^$`,
		// Prefix
		`^a`,
		`^aa`,
		`^a.*`,
		`^aa.*`,
		// Suffix
		`a$`,
		`a$`,
		`^.*a$`,
		`^.*aa$`,
		`.*aa$`,
		// Contains
		`a`,
		`aa`,
		`^.*a.*$`,
		`^.*aa.*$`,
		`.*a.*`,
		`.*aa.*`,
		`.*a`,
		`.*aa`,
		`a.*`,
		`aa.*`,
		// PrefixSuffix
		`^a.*b$`,
		`^a.*a$`,
		`^a.a$`,
		`^a?a$`,
		// Regex fallback (must NOT be optimised)
		`.*`,
		`^\d+$`,
		`^(b|c)$`,
		`^(bb|cc)$`,
		`(b|c)`,
		`(bb|cc)`,
		`a.*b$`,
		`^a.*b`,
		`a.*a$`,
		`^a.*a`,
		`^[a-b]+$`,
		`^a.b$`,
		`a?b`,
		`a+b`,
	}

	inputs := []string{
		"",
		"a",
		"aaa",
		"AAA",
		"aAa",
		"(*.)",
		"aaa123aaa",
		"A",
		"b",
		"c",
		"bbb",
		"bbbccc",
		"123",
		"ab",
		"axb",
		"axxxxb",
		"abcxyz",
	}

	for _, pat := range patterns {
		std := regexp.MustCompile(pat)
		fastStr := NewStringMatcherFromRegexp(std)
		fastByte := NewByteMatcherFromRegexp(std)

		for _, input := range inputs {
			wantStr := std.MatchString(input)
			gotStr := fastStr.MatchString(input)
			if gotStr != wantStr {
				t.Errorf("pattern=%q input=%q: MatchString=%t, regexp=%t (%T)",
					pat, input, gotStr, wantStr, fastStr)
			}

			wantByte := std.Match([]byte(input))
			gotByte := fastByte.Match([]byte(input))
			if gotByte != wantByte {
				t.Errorf("pattern=%q input=%q: Match=%t, regexp=%t (%T)",
					pat, input, gotByte, wantByte, fastByte)
			}
		}
	}
}

func TestMatchKind(t *testing.T) {
	type matchKindCase struct {
		stringType string
		byteType   string
		patterns   []string
	}

	tests := []matchKindCase{
		{
			stringType: fmt.Sprintf("%T", exactStringMatcher("")),
			byteType:   "fastregex.exactByteMatcher",
			patterns:   []string{`^a$`, `^\(\*\.\)$`},
		},
		{
			stringType: fmt.Sprintf("%T", prefixStringMatcher("")),
			byteType:   "fastregex.prefixByteMatcher",
			patterns:   []string{`^a`, `^a.*`, `^a.*$`},
		},
		{
			stringType: fmt.Sprintf("%T", suffixStringMatcher("")),
			byteType:   "fastregex.suffixByteMatcher",
			patterns:   []string{`a$`, `^.*a$`, `.*a$`},
		},
		{
			stringType: fmt.Sprintf("%T", containsStringMatcher("")),
			byteType:   "fastregex.containsByteMatcher",
			patterns:   []string{`a`, `^.*a.*$`, `.*a.*`},
		},
		{
			stringType: fmt.Sprintf("%T", prefixSuffixStringMatcher{}),
			byteType:   "fastregex.prefixSuffixByteMatcher",
			patterns:   []string{`^a.*b$`},
		},
		{
			stringType: fmt.Sprintf("%T", &regexp.Regexp{}),
			byteType:   "*regexp.Regexp",
			patterns:   []string{`^.*a.*b.*$`, `^a.*b`, `a.*b$`, `(?i)a`, `^a.b$`, `^a+b$`, `^a?b$`},
		},
	}

	for _, tc := range tests {
		for _, p := range tc.patterns {
			t.Run("String/"+p, func(t *testing.T) {
				m, err := NewStringMatcher(p)
				if err != nil {
					t.Fatal(err)
				}
				got := fmt.Sprintf("%T", m)
				if got != tc.stringType {
					t.Errorf("pattern=%q: got type %s, want %s", p, got, tc.stringType)
				}
			})
			t.Run("Byte/"+p, func(t *testing.T) {
				m, err := NewByteMatcher(p)
				if err != nil {
					t.Fatal(err)
				}
				got := fmt.Sprintf("%T", m)
				if got != tc.byteType {
					t.Errorf("pattern=%q: got type %s, want %s", p, got, tc.byteType)
				}
			})
		}
	}
}

var benchmarkCases = []struct {
	name    string
	pattern string
	input   string
}{
	{"Exact", `^github\.com/target/pkg/internal/some\.Type$`, "github.com/target/pkg/internal/some.Type"},
	{"Prefix", `^github\.com/target/pkg`, "github.com/target/pkg/internal/some.Type"},
	{"Suffix", `some\.Type$`, "github.com/target/pkg/internal/some.Type"},
	{"Contains", `\.Method\(`, "(*github.com/target/pkg/internal/some.Type).Method()"},
	{"PrefixSuffix", `^github\.com/target/pkg.*\.Type$`, "github.com/target/pkg/internal/some.Type"},
	{"Regex", `^\(.*\)\.Method\(.*\)$`, "(*github.com/target/pkg/internal/some.Type).Method()"},
}

func TestBenchmarkTable(t *testing.T) {
	fmt.Println("| Name | Pattern | Input |")
	fmt.Println("| --- | --- | --- |")
	for _, tc := range benchmarkCases {
		fmt.Printf("| %s | `%s` | `%s` |\n", tc.name, tc.pattern, tc.input)
	}
}

func BenchmarkReadme(b *testing.B) {
	for _, tc := range benchmarkCases {
		std := regexp.MustCompile(tc.pattern)
		fast := MustCompile(tc.pattern)
		b.Run(tc.name, func(b *testing.B) {
			b.Run("impl=Std", func(b *testing.B) {
				for b.Loop() {
					_ = std.MatchString(tc.input)
				}
			})
			b.Run("impl=Fast", func(b *testing.B) {
				for b.Loop() {
					_ = fast.MatchString(tc.input)
				}
			})
		})
	}
}
