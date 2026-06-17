# Configure collection and track policy

This guide explains `prof.json`: how to create it, how `collection` selects per-function extracts, and how `track` configures regression gates for CI.

## Before you begin

- You run commands from the module root (next to `go.mod`).
- You understand what a tag and benchmark name are ([Home](index.md#terminology)).

## Create the configuration file

```bash
prof config init
```

Or in `prof ui`, choose **Manage prof.json configuration**. That creates or edits `prof.json` next to `go.mod`.

## Benchmark discovery

`prof auto` and `prof ui` discover benchmarks by scanning `*_test.go` files under your module root. Directories named `tests/`, `bench/`, and `vendor/`, plus nested directories that contain their own `go.mod` (separate Go modules), are skipped so fixtures and QA sandboxes under `tests/` do not appear in your benchmark list.

`prof setup` is a hidden alias for `prof config init`.

## File shape (version 1)

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

## `collection` (prof auto / prof manual)

Controls per-function text extracts and grouped package reports.

| Section | Key meaning |
| ------- | ----------- |
| `defaults` | Applies to all benchmarks unless overridden |
| `benchmarks` | Per-benchmark rules for `prof auto` (benchmark name as key) |
| `manual_profiles` | Per-file rules for `prof manual` (file stem as key, e.g. `BenchmarkFoo_cpu`) |

| Field | Description |
| ----- | ----------- |
| `include_prefixes` | If set, only functions whose full name starts with one of these prefixes |
| `ignore_functions` | Short names excluded even when prefixes match |

## `track` (prof track)

Regression policy when `--fail-on-regression` is not passed. CLI flags override track config when provided.

See [CI and regressions](ci.md) and the full reference in [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md).

## CLI helpers

```bash
prof config path      # print resolved prof.json path
prof config validate  # load and validate; exit 1 on error
```

## Testing / verify

After `prof config init`, confirm `prof.json` exists beside `go.mod`. Run a small `prof auto` collect and check that `<profile>_functions/<BenchmarkName>/` contains files when your filter matches hot symbols.

## Next steps

- [Compare runs](compare.md)
- [CI and regressions](ci.md)

## Related

- [Collect profiling data](collect.md) · [CLI reference](cli-reference.md)
