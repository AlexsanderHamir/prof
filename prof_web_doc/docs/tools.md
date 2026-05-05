# Optional tools

Subcommands under `prof tools` operate on data you already collected.

## prof tools benchstat

Runs Go’s [`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) on benchmark text from two tags.

```bash
prof tools benchstat --base baseline --current optimized --bench-name BenchmarkGenPool
```

**Prerequisite:**

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

**Output:** `bench/tools/benchstat/<BenchmarkName>_results.txt`

## prof tools qcachegrind

Builds callgrind data from a binary profile and can launch [QCacheGrind](https://kcachegrind.github.io/html/Home.html).

```bash
prof tools qcachegrind --tag optimized --profiles cpu --bench-name BenchmarkGenPool
```

**Prerequisite:** QCacheGrind installed (package name differs by OS; use your distribution or Homebrew).

**Output:** `bench/tools/qcachegrind/<BenchmarkName>_<profile>.callgrind`

## Combined workflow

```bash
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory" --count 10 --tag baseline
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory" --count 10 --tag optimized
prof track auto --base baseline --current optimized --profile-type cpu --bench-name BenchmarkGenPool --output-format summary
prof tools benchstat --base baseline --current optimized --bench-name BenchmarkGenPool
prof tools qcachegrind --tag optimized --profiles cpu --bench-name BenchmarkGenPool
```

## See also

[Compare runs](compare.md) · [Collect profiling data](collect.md)
