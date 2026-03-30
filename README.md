# fastregex

A thin wrapper around [`regexp.Regexp`](https://pkg.go.dev/regexp#Regexp) that
auto-detects simple pattern shapes and dispatches to faster string operations
instead of running the full regex automaton.

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

| Name         | Pattern                                        | Input                                                  |
| ------------ | ---------------------------------------------- | ------------------------------------------------------ |
| Exact        | `^github\.com/target/pkg/internal/some\.Type$` | `github.com/target/pkg/internal/some.Type`             |
| Prefix       | `^github\.com/target/pkg`                      | `github.com/target/pkg/internal/some.Type`             |
| Suffix       | `some\.Type$`                                  | `github.com/target/pkg/internal/some.Type`             |
| Contains     | `\.Method\(`                                   | `(*github.com/target/pkg/internal/some.Type).Method()` |
| PrefixSuffix | `^github\.com/target/pkg.*\.Type$`             | `github.com/target/pkg/internal/some.Type`             |
| Regex        | `^\(.*\)\.Method\(.*\)$`                       | `(*github.com/target/pkg/internal/some.Type).Method()` |

<!-- END BENCHMARK CASES -->

<!-- BEGIN BENCHMARKS -->

```
                       │      Std      │                Fast                 │
                       │    sec/op     │   sec/op     vs base                │
Readme/Exact-12           74.870n ± 2%   5.287n ± 3%  -92.94% (p=0.000 n=10)
Readme/Prefix-12          62.810n ± 2%   4.606n ± 3%  -92.67% (p=0.000 n=10)
Readme/Suffix-12         153.850n ± 2%   5.494n ± 2%  -96.43% (p=0.000 n=10)
Readme/Contains-12        138.25n ± 1%   12.60n ± 2%  -90.89% (p=0.000 n=10)
Readme/PrefixSuffix-12   496.950n ± 1%   8.444n ± 4%  -98.30% (p=0.000 n=10)
Readme/Regex-12            831.0n ± 2%   840.7n ± 3%        ~ (p=0.190 n=10)
geomean                    185.9n        15.12n       -91.87%
```

<!-- END BENCHMARKS -->
