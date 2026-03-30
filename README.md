# fastregex

A thin wrapper around [`regexp.Regexp`](https://pkg.go.dev/regexp#Regexp) that auto-detects simple pattern shapes and dispatches to faster string operations instead of running the full regex automaton.

## Recognised fast paths

| Pattern shape               | Strategy                                   |
| --------------------------- | ------------------------------------------ |
| `^literal$`                 | `string == literal`                        |
| `^literal` / `^literal.*$`  | `strings.HasPrefix`                        |
| `literal$` / `^.*literal$`  | `strings.HasSuffix`                        |
| `literal` / `^.*literal.*$` | `strings.Contains`                         |
| `^literal.*literal$`        | `strings.HasPrefix` && `strings.HasSuffix` |
| everything else             | `regexp.Regexp.MatchString`                |

Equivalent `[]byte` matchers using the `bytes` package are also provided.

## Usage

```go
// String matching
m := fastregex.MustCompile(`^github\.com/target/pkg`)
m.MatchString("github.com/target/pkg/internal/some.Type") // true

// Byte matching
bm := fastregex.MustNewByteMatcher(`\.Type$`)
bm.Match([]byte("github.com/target/pkg/internal/some.Type")) // true
```

## Benchmarks

Comparing `regexp.Regexp.MatchString` (Std) against `fastregex` (Fast) on realistic Go symbol strings.

<!-- BEGIN BENCHMARK CASES -->

| Name | Pattern | Input |
| --- | --- | --- |
| Exact | `^github.com/target/pkg/internal/some.Type$` | `github.com/target/pkg/internal/some.Type` |
| Prefix | `^github.com/target/pkg` | `github.com/target/pkg/internal/some.Type` |
| Suffix | `some.Type$` | `github.com/target/pkg/internal/some.Type` |
| Contains | `.Method(` | `(*github.com/target/pkg/internal/some.Type).Method()` |
| PrefixSuffix | `^github.com/target/pkg.*.Type$` | `github.com/target/pkg/internal/some.Type` |
| Regex | `^(.*).Method(.*)$` | `(*github.com/target/pkg/internal/some.Type).Method()` |
<!-- END BENCHMARK CASES -->

<!-- BEGIN BENCHMARKS -->

```
                       │      Std      │                Fast                 │
                       │    sec/op     │   sec/op     vs base                │
Readme/Exact-12           74.005n ± 2%   5.344n ± 2%  -92.78% (p=0.000 n=10)
Readme/Prefix-12          62.990n ± 3%   4.611n ± 3%  -92.68% (p=0.000 n=10)
Readme/Suffix-12         155.700n ± 2%   4.653n ± 3%  -97.01% (p=0.000 n=10)
Readme/Contains-12        137.70n ± 2%   12.50n ± 2%  -90.92% (p=0.000 n=10)
Readme/PrefixSuffix-12   494.100n ± 2%   8.534n ± 3%  -98.27% (p=0.000 n=10)
Readme/Regex-12            819.5n ± 2%   820.6n ± 1%        ~ (p=0.425 n=10)
geomean                    185.3n        14.69n       -92.07%
```
<!-- END BENCHMARKS -->
