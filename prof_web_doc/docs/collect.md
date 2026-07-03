# Collect profiling data

This guide explains how to capture benchmark profiles into `bench/<tag>/` using `prof auto` (runs `go test`) or `prof manual` (ingests existing profile files), so you can compare runs or open them in `pprof` later.

## Before you begin

- Module root as cwd; benchmarks discoverable from there ([Working directory and paths](workspace.md)).
- Go and `go test` work for your package.
- Optional: [Graphviz](https://graphviz.org/) for PNG call graphs; otherwise use `--skip-png` or expect PNG-related failures unless skipped.

## What is a profile run?

A run is one labeled experiment: benchmarks executed (or files ingested), profiles of selected types written under `bench/<tag>/`, plus text listings and optional per-function extracts.

## Commands

| Command | Purpose |
| ------- | ------- |
| `prof auto` | Run benchmarks via `go test`; collect profile types you list. |
| `prof manual` | Ingest existing profile binaries; same layout style (no `go test`). |

`prof auto` and `prof manual` run `go` and `go tool pprof` on your machine. The implementation centralizes those commands in `engine/tooling` so argv and supported profile names stay consistent.

## `prof auto`

### Minimal example

```bash
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

### Flags

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--benchmarks` | strings | Yes | n/a | Benchmark names to run. |
| `--profiles` | strings | Yes | n/a | Comma-separated profile IDs: `cpu`, `memory`, `mutex`, `block`. |
| `--tag` | string | Yes | n/a | Output directory `bench/<tag>/`. |
| `--count` | int | Yes | n/a | Number of runs; must be positive. |
| `--group-by-package` | bool | No | `false` | Extra grouped-by-package text (`*_grouped.txt`). |
| `--lenient-profiles` | bool | No | `false` | Skip missing profile binaries instead of failing. |
| `--skip-png` | bool | No | `false` | Do not fail the run when PNG generation fails (for example no Graphviz). |

### What collection stores

Each run writes a single tag directory, `bench/<tag>/`, under your [module root](index.md#terminology) (same cwd rules as [Working directory and paths](workspace.md)). That folder is the durable record of one experiment: which benchmarks ran, how many iterations you requested, and which profile types you enabled.

#### What is being collected

- Runtime profiles from `go test` for each benchmark and each profile type you list (`cpu`, `memory`, `mutex`, `block`). These answer where time or allocations went during that benchmark, not only the final `ns/op` line.
- Binary profile files (`.out`) so you or Prof can run `go tool pprof` again later without re-running the benchmark.
- Text renderings of each profile so you can skim, search, or diff results without an interactive session.
- Per-function extracts when your [configuration](configure.md) selects functions, so you can read `pprof -list`-style detail for hot symbols tied to that benchmark and profile.

#### How that helps

- Open profiles in `pprof`: binary and text files under each tag share predictable paths.
- Share context: zip `bench/<tag>/` or attach key `text/` files to an issue or PR so others see the same profile view you did.
- Re-open in pprof: point `go tool pprof` at `bin/<BenchmarkName>/<BenchmarkName>_<profile>.out` for ad-hoc queries on the stored binary.

### Artifact layout under `bench/<tag>/` { #artifact-layout-under-benchtag }

| Location | What you get | Typical use |
| -------- | ------------- | ----------- |
| `description.txt` | Short tag-level note (placeholder until you edit it). | Record why this run exists (branch, experiment, machine). |
| `bin/<BenchmarkName>/` | One `BenchmarkName_<profile>.out` per profile type collected. | Source of truth for `pprof`; required for regenerating text and PNGs. |
| `text/<BenchmarkName>/` | For each profile: `BenchmarkName_<profile>.txt` (flat listing). With `--group-by-package`, also `BenchmarkName_<profile>_grouped.txt` (package-oriented summary). | Read, grep, or diff stacks; grouped files help when flat output is too noisy. |
| `<profile>_functions/<BenchmarkName>/` | Per-function text files for symbols in scope, plus optional `BenchmarkName_<profile>.png` when Graphviz is available (or omit failure with `--skip-png`). | Deep dive on specific functions; optional call-graph PNG for presentations. |

Exact names and suffixes are defined in the implementation (`internal` constants and `engine/benchmark` path helpers); the table above matches the usual `prof auto` and `prof manual` layout.

## `prof manual` { #prof-manual }

Requires `--tag` and one or more profile file paths as positional arguments. Does not run `go test`.

```bash
prof manual --tag "external-profiles" cpu.prof memory.prof
```

```bash
prof manual --tag "external-profiles" --group-by-package cpu.prof memory.prof
```

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--tag` | string | Yes | n/a | Output directory `bench/<tag>/`. |
| `--group-by-package` | bool | No | `false` | Same as `prof auto`. |

Per-file collection filters use `collection.manual_profiles` in `prof.json`. Keys are profile file stems (e.g. `BenchmarkFoo_cpu` for `BenchmarkFoo_cpu.out`). See [Configure — manual profile overrides](configure.md#collection-manual-profiles).

## Testing / verify

After `prof auto`, you should see `bench/<tag>/bin/<BenchmarkName>/` containing `BenchmarkName_<profile>.out` for each profile you requested, and matching files under `text/<BenchmarkName>/`.

If `go test` fails, Prof exits non-zero. Fix the test failure first. For PNG or Graphviz issues, see [Troubleshooting](troubleshooting.md#graphviz-png-errors).

## Next steps

- [Configure collection](configure.md) for `collection` in `prof.json`.

## Related

- [CLI reference](cli-reference.md) · [Troubleshooting](troubleshooting.md) · [Interactive UI and TUI](tui.md)
