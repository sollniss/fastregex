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
bm, err := fastregex.NewByteMatcher(`\.Type$`)
if err != nil { return err }
bm.Match([]byte("github.com/target/pkg/internal/some.Type")) // true
```

## Benchmarks

Comparing `regexp.Regexp.MatchString` (Std) against `fastregex` (Fast) on realistic Go symbol strings.

<!-- BEGIN BENCHMARK CASES -->

| Name | Pattern | Input |
| --- | --- | --- |
| Exact | `^github\.com/target/pkg/internal/some\.Type$` | `github.com/target/pkg/internal/some.Type` |
| Prefix | `^github\.com/target/pkg` | `github.com/target/pkg/internal/some.Type` |
| Suffix | `some\.Type$` | `github.com/target/pkg/internal/some.Type` |
| Contains | `\.Method\(` | `(*github.com/target/pkg/internal/some.Type).Method()` |
| PrefixSuffix | `^github\.com/target/pkg.*\.Type$` | `github.com/target/pkg/internal/some.Type` |
| Regex | `^\(.*\)\.Method\(.*\)$` | `(*github.com/target/pkg/internal/some.Type).Method()` |
<!-- END BENCHMARK CASES -->

<!-- BEGIN BENCHMARKS -->

```
                       │      Std      │                Fast                 │
                       │    sec/op     │   sec/op     vs base                │
Readme/Exact-12           74.260n ± 2%   5.370n ± 2%  -92.77% (p=0.000 n=10)
Readme/Prefix-12          64.325n ± 2%   4.846n ± 2%  -92.47% (p=0.000 n=10)
Readme/Suffix-12         154.200n ± 2%   4.603n ± 3%  -97.02% (p=0.000 n=10)
Readme/Contains-12        141.00n ± 4%   12.57n ± 1%  -91.09% (p=0.000 n=10)
Readme/PrefixSuffix-12   500.750n ± 2%   8.672n ± 3%  -98.27% (p=0.000 n=10)
Readme/Regex-12            827.4n ± 2%   821.4n ± 2%        ~ (p=0.670 n=10)
geomean                    187.2n        14.85n       -92.07%
```
<!-- END BENCHMARKS -->
