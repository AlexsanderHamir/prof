# Configure collection and track policy

This guide explains `prof.json`: how to create it, how `collection` selects per-function extracts, and how `track` configures regression gates for CI.

## Before you begin

- You run commands from the module root (next to `go.mod`). See [Working directory and paths](workspace.md#profjson).
- You understand what a tag and benchmark name are ([Home](index.md#terminology)).

## Create the configuration file

```bash
prof config init
```

Or in `prof ui`, choose **Manage prof.json configuration**. That creates or edits `prof.json` next to `go.mod`.

`prof setup` is a hidden alias for `prof config init`.

## Generated files { #generated-files }

`prof config init` writes two files beside `go.mod`:

| File | Loaded by prof? | Purpose |
| ---- | ----------------- | ------- |
| `prof.json` | Yes | Your active config (valid JSON, no comments) |
| `prof.json.example` | No | Commented reference with optional sections and doc links |

**Generated `prof.json` is minimal** — only a version field until you add sections:

```json
{
  "version": 1
}
```

Copy sections from `prof.json.example` into `prof.json` when you want collection filters or track policy. The example file shows every optional key with inline comments.

## Benchmark discovery

`prof auto` and `prof ui` discover benchmarks by scanning `*_test.go` files under your module root. Directories named `tests/`, `bench/`, and `vendor/`, plus nested directories that contain their own `go.mod` (separate Go modules), are skipped so fixtures and QA sandboxes under `tests/` do not appear in your benchmark list.

## Full shape (version 1)

When you copy optional sections from `prof.json.example`, a typical project file looks like:

```json
{
  "version": 1,
  "collection": {
    "defaults": {
      "include_prefixes": ["github.com/example/myproject"],
      "ignore_functions": ["init", "BenchmarkMain"]
    },
    "benchmarks": {
      "BenchmarkGenPool": {
        "include_prefixes": ["github.com/example/myproject/pkg/pool"]
      }
    },
    "manual_profiles": {
      "BenchmarkGenPool_cpu": {
        "include_prefixes": ["github.com/example/myproject/pkg/pool"]
      }
    }
  },
  "track": {
    "defaults": {
      "ignore_prefixes": ["runtime.", "reflect.", "testing."],
      "min_change_percent": 5.0,
      "max_regression_percent": 15.0
    }
  }
}
```

## Collection { #collection }

Controls per-function text extracts and grouped package reports after [Collect profiling data](collect.md). Output lands under `bench/<tag>/<profile>_functions/<BenchmarkName>/` when filters match.

| Section | Key meaning |
| ------- | ----------- |
| `defaults` | Applies to all benchmarks unless overridden |
| `benchmarks` | Per-benchmark rules for `prof auto` (benchmark name as key) |
| `manual_profiles` | Per-file rules for `prof manual` (file stem as key, e.g. `BenchmarkFoo_cpu`) |

**Override precedence:** `defaults` → per-benchmark or per-manual-profile entry (field-by-field merge).

| Field | Description |
| ----- | ----------- |
| `include_prefixes` | If set, only functions whose **full pprof symbol contains** one of these substrings (usually your module import path) |
| `ignore_functions` | **Short** function names excluded even when `include_prefixes` matches (e.g. `init`, `BenchmarkMain`) |

If `include_prefixes` is empty, every function in the profile is eligible (often too broad). If set, a function must match a prefix **and** not appear in `ignore_functions`.

### Per-benchmark overrides { #collection-benchmarks }

Use `collection.benchmarks` to override filters for one benchmark run by `prof auto`. The key is the benchmark name exactly as passed to `--benchmarks`:

```json
"benchmarks": {
  "BenchmarkGenPool": {
    "include_prefixes": ["github.com/example/myproject/pkg/pool"],
    "ignore_functions": ["BenchmarkHelper"]
  }
}
```

Unset fields inherit from `collection.defaults`.

### Manual profile overrides { #collection-manual-profiles }

Use `collection.manual_profiles` for profiles ingested by `prof manual`. The key is the profile file **stem** under `bench/<tag>/bin/`, e.g. `BenchmarkFoo_cpu` for `BenchmarkFoo_cpu.out`:

```json
"manual_profiles": {
  "BenchmarkFoo_cpu": {
    "include_prefixes": ["github.com/example/myproject/pkg/foo"]
  }
}
```

See [Collect profiling data — prof manual](collect.md#prof-manual).

## Track { #track }

Regression policy for `prof track` and UI compare. When you omit `--fail-on-regression` and `--regression-threshold`, prof uses the merged `track` policy from `prof.json`. When both CLI flags are set with a positive threshold, **CLI overrides config** for that run.

Add a `track` section to `prof.json` if you want config-only gates — a freshly generated file has no `track` block until you copy it from `prof.json.example`.

| Field | Description |
| ----- | ----------- |
| `ignore_prefixes` | Skip functions whose full name **starts with** one of these strings (e.g. `runtime.`) |
| `ignore_functions` | Skip functions by **exact** full symbol name |
| `min_change_percent` | Noise floor: regressions below this percent do not fail the run; also minimum magnitude when `fail_on_improvement` is true |
| `max_regression_percent` | Fail when worst flat regression meets or exceeds this percent (`0` = disabled) |
| `fail_on_improvement` | Fail on significant speedups (unusual; default `false`) |

See [Compare runs — regression gate](compare.md#regression-gate) and [CI and regressions](ci.md). Full schema and Actions examples: [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md).

### Per-benchmark track overrides { #track-benchmarks }

Use `track.benchmarks` to tighten or relax policy for one benchmark. Keys are benchmark names; unset fields inherit from `track.defaults`:

```json
"benchmarks": {
  "BenchmarkCriticalPath": {
    "max_regression_percent": 5.0,
    "min_change_percent": 1.0
  }
}
```

## CLI helpers

```bash
prof config path      # print resolved prof.json path
prof config validate  # load and validate; exit 1 on error
```

## Testing / verify

After `prof config init`, confirm `prof.json` and `prof.json.example` exist beside `go.mod`. Add a `collection` section, run a small `prof auto` collect, and check that `<profile>_functions/<BenchmarkName>/` contains files when your filter matches hot symbols.

## Next steps

- [Compare runs](compare.md)
- [CI and regressions](ci.md)

## Related

- [Collect profiling data](collect.md) · [CLI reference](cli-reference.md)
