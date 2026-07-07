# Parallel source_lines collection

## Context

Step 3 of the auto collect pipeline writes per-function `go tool pprof -list` extracts under `.prof/<tag>/source_lines/<profile>/<benchmark>/`. Each extract is an independent subprocess that reads the same profile binary and writes a unique output file.

Previously, prof ran these subprocesses sequentially: every profile kind in order, then every filtered function in order. On profiles with many in-scope symbols, this step dominated wall-clock time—especially on Windows where process spawn latency is high.

## Decision

Collect `source_lines` with **bounded parallelism** at two layers:

1. **Profile kinds** (`cpu`, `memory`, …) run concurrently inside [`collectProfileFunctions`](../../engine/collect/pipeline.go).
2. **Per-function `-list` calls** run concurrently inside [`getFunctionsOutput`](../../engine/collect/artifacts_pprof.go), shared by auto and manual collection.

Both layers use a fixed worker pool ([`parallelFor`](../../engine/collect/parallel.go)) with worker count:

```
min(jobCount, GOMAXPROCS, 8)
```

Each worker still invokes the same `go tool pprof -list=<pattern>` argv as before. Filters, output paths, filenames, and warn-and-continue semantics are unchanged.

## Alternatives considered

| Alternative | Why not chosen |
| --- | --- |
| Unbounded goroutine per function | Risk of memory and disk contention when dozens of `pprof` processes load the same binary |
| `golang.org/x/sync/errgroup` | Adds a dependency; indexed error aggregation still requires a full result buffer |
| In-process `-list` via `github.com/google/pprof` | Output formatting may diverge from the CLI; higher implementation and maintenance cost |
| User-facing `--workers` flag | No evidence yet that the fixed cap is wrong for typical machines |

## Consequences

**Positive**

- Wall-clock time for step 3 drops roughly with worker count (up to the cap) for large function lists, and across profile kinds when multiple profiles are collected.
- Auto and manual paths both benefit from function-level parallelism via shared `getFunctionsOutput`.

**Neutral / trade-offs**

- Up to eight concurrent `go tool pprof` processes per profile (and concurrent profiles multiply subprocess count). Disk and CPU contention are possible on very slow storage; the cap limits blast radius.
- Interactive warnings for skipped per-function lists remain capped at three plus a summary line; ordering is by function index within a profile, not subprocess completion order.

## Invariants

- Same `FunctionListEntry` set after `prof.json` filters.
- Same `.txt` paths and `pprof -list` subprocess argv.
- Per-function failures warn and continue; they do not fail the collect run.
- Profile-level failures (missing binary parse, mkdir error) still fail the benchmark step in profile slice order.

## Verification

```bash
go test ./engine/collect/ ./engine/tooling/ -count=1
go test ./...
```

Optional local equivalence check: run `prof auto` before and after, then `diff -r` the `source_lines/` trees—they should match when the same filters and benchmarks are used.

## See also

- [Collect request flow — Step 3](../collect-request-flow.md#step-3--per-function-extracts)
- [CODEBASE_DESIGN.md — Profile pipelines](../../CODEBASE_DESIGN.md#profile-pipelines)
