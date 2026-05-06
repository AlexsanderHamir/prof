# Collect profiling data

Interactive collect: `prof ui` or `prof tui`. Below: `prof auto` and `prof manual` for flags and CI.

## Commands

| Command | Purpose |
| -------- | -------- |
| `prof auto` | Run benchmarks; collect profile types you list. |
| `prof manual` | Ingest existing profile binaries; same layout style. |

`prof auto` and `prof manual` run `go` and `go tool pprof` on your machine. The implementation centralizes those commands in `engine/tooling` so argv and supported profile names stay consistent.

## prof auto

Required: `--benchmarks`, `--profiles`, `--count`, `--tag`.

```bash
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

| Flag | Effect |
| ---- | ------ |
| `--group-by-package` | Extra grouped-by-package text (`*_grouped.txt`). |
| `--lenient-profiles` | Skip missing profile binaries instead of failing. |
| `--skip-png` | Do not fail if PNG generation fails (e.g. no Graphviz). |

### What collection stores

Each run writes a single tag directory, `bench/<tag>/`, under your [module root](index.md#terminology) (same cwd rules as [Working directory and paths](workspace.md)). That folder is the durable record of one experiment: which benchmarks ran, how many iterations you requested, and which profile types you enabled.

#### What is being collected

- Runtime profiles from `go test` for each benchmark and each profile type you list (`cpu`, `memory`, `mutex`, `block`). These answer where time or allocations went during that benchmark, not just the final `ns/op` line.
- Binary profile files (`.out`) so you or Prof can run `go tool pprof` again later without re-running the benchmark.
- Text renderings of each profile so you can skim, search, or diff results without an interactive session.
- Per-function extracts when your [configuration](configure.md) selects functions, so you can read `pprof -list`-style detail for hot symbols tied to that benchmark and profile.

#### How that helps

- Compare runs: `prof track` pairs two tags; consistent paths under each tag make baselines and candidates comparable.
- Investigate regressions: jump from a worse `benchstat` number to CPU or alloc stacks, then into specific functions, using files you already saved.
- Share context: zip `bench/<tag>/` or attach key `text/` files to an issue or PR so others see the same profile view you did.
- Re-open in pprof: point `go tool pprof` at `bin/<BenchmarkName>/<BenchmarkName>_<profile>.out` for ad-hoc queries on the stored binary.

### Artifact layout under `bench/<tag>/`

| Location | What you get | Typical use |
| -------- | ------------- | ------------- |
| `description.txt` | Short tag-level note (placeholder until you edit it). | Record *why* this run exists (branch, experiment, machine). |
| `bin/<BenchmarkName>/` | One `BenchmarkName_<profile>.out` per profile type collected. | Source of truth for `pprof`; required for regenerating text and PNGs. |
| `text/<BenchmarkName>/` | For each profile: `BenchmarkName_<profile>.txt` (flat listing). With `--group-by-package`, also `BenchmarkName_<profile>_grouped.txt` (package-oriented summary). | Read, grep, or diff stacks; grouped files help when flat output is too noisy. |
| `<profile>_functions/<BenchmarkName>/` | Per-function text files for symbols in scope, plus optional `BenchmarkName_<profile>.png` when Graphviz is available (or omit failure with `--skip-png`). | Deep dive on specific functions; optional flame-style PNG for presentations. |

Exact names and suffixes are defined in the implementation (`internal` constants and `engine/benchmark` path helpers); the table above matches the usual `prof auto` / `prof manual` layout.

## prof manual

Requires `--tag` and profile file paths. Does not run `go test`.

```bash
prof manual --tag "external-profiles" cpu.prof memory.prof
```

```bash
prof manual --tag "external-profiles" --group-by-package cpu.prof memory.prof
```

## Next article

[Configure collection](configure.md)
