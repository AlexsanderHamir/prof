# Optional tools

`prof tools` works on data under `bench/`. Same flows from **`prof ui`** → **Tools**.

## prof tools benchstat

[`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) across two tags.

```bash
prof tools benchstat --base baseline --current optimized --bench-name BenchmarkGenPool
```

Install: `go install golang.org/x/perf/cmd/benchstat@latest`

Output: `bench/tools/benchstat/<BenchmarkName>_results.txt`

## prof tools qcachegrind

Callgrind + [QCacheGrind](https://kcachegrind.github.io/html/Home.html).

```bash
prof tools qcachegrind --tag optimized --profiles cpu --bench-name BenchmarkGenPool
```

Needs QCacheGrind installed. Output: `bench/tools/qcachegrind/<BenchmarkName>_<profile>.callgrind`

## See also

[Compare runs](compare.md) · [Collect profiling data](collect.md)
