#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "$0")"

COUNT="${1:-10}"

echo "Extracting benchmark cases table..."
CASES=$(go test -run='^TestBenchmarkTable$' -v 2>/dev/null | grep '^|')

BENCH_FILE="$(mktemp)"
trap 'rm -f "$BENCH_FILE"' EXIT

echo "Running BenchmarkReadme (count=${COUNT})..."
go test -bench=BenchmarkReadme -benchmem -count="$COUNT" -timeout=10m | tee "$BENCH_FILE"

echo ""
echo "Generating benchstat comparison..."
RESULT=$(go run golang.org/x/perf/cmd/benchstat@latest -col /impl "$BENCH_FILE")

# Extract only the sec/op table (first table up to the first blank line after the data).
SECOP=$(echo "$RESULT" | awk '
	/sec\/op/ { found=1; if (prev) print prev; print; next }
	found && /^$/ { exit }
	found     { print }
	{ prev=$0 }
')

echo ""
echo "$SECOP"

README="README.md"
BEGIN="<!-- BEGIN BENCHMARKS -->"
END="<!-- END BENCHMARKS -->"

if [ ! -f "$README" ]; then
	echo "WARNING: $README not found, skipping update."
	exit 0
fi

BEGIN_CASES="<!-- BEGIN BENCHMARK CASES -->"
END_CASES="<!-- END BENCHMARK CASES -->"

CASES="$CASES" awk -v begin="$BEGIN_CASES" -v end="$END_CASES" '
	$0 ~ begin { print; printf "\n%s\n", ENVIRON["CASES"]; skip=1; next }
	$0 ~ end   { skip=0 }
	!skip       { print }
' "$README" > "${README}.tmp" && mv "${README}.tmp" "$README"

# Replace benchmark results between markers.
SECOP="$SECOP" awk -v begin="$BEGIN" -v end="$END" '
	$0 ~ begin { print; printf "\n```\n%s\n```\n", ENVIRON["SECOP"]; skip=1; next }
	$0 ~ end   { skip=0 }
	!skip       { print }
' "$README" > "${README}.tmp" && mv "${README}.tmp" "$README"

echo ""
echo "Updated $README."
