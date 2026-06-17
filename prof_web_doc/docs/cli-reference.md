# CLI reference

This page lists supported commands, profile and output identifiers, and flag-level defaults so you can script Prof without guessing. Subcommands also print `prof <cmd> -h`.

## Local documentation build

From the `prof_web_doc` directory (with [MkDocs](https://www.mkdocs.org/) and [Material](https://squidfunk.github.io/mkdocs-material/) installed):

```bash
cd prof_web_doc
mkdocs serve
```

Static site output is written to `prof_web_doc/site/` when you run `mkdocs build` (that directory is ignored by git in this repo).

## Command overview

| Command | Purpose |
| ------- | ------- |
| `prof ui` | Full-screen menu: collect, compare, tools, manage configuration. |
| `prof tui` | Terminal collect flow (multi-select benchmarks and profiles). |
| `prof tui track` | Terminal compare flow (pick two tags under `bench/`). |
| `prof auto` | Run `go test` benchmarks and collect listed profiles into `bench/<tag>/`. |
| `prof manual` | Ingest existing profile files into the same layout style (no `go test`). |
| `prof track auto` | Compare two tags created by `prof auto` (or compatible layout). |
| `prof track manual` | Compare two profile files by filesystem path. |
| `prof config init` | Create `prof.json` next to `go.mod`. |
| `prof config validate` | Load and validate `prof.json`; exit non-zero on error. |
| `prof config path` | Print resolved `prof.json` path. |
| `prof setup` | Hidden alias for `prof config init`. |
| `prof tools benchstat` | Run `benchstat` for two tags; writes under `bench/tools/`. |
| `prof tools qcachegrind` | Emit Callgrind files for QCacheGrind. |

## Profile types (`--profiles`, `--profile-type`)

These IDs are the ones `go test` integration supports (comma-separated for `--profiles`):

| ID | Role |
| -- | ---- |
| `cpu` | CPU profile |
| `memory` | Memory / allocs profile |
| `mutex` | Mutex profile |
| `block` | Block profile |

## Compare output formats

Values accepted by `prof track auto`, `prof track manual`, and the TUI compare flow:

| Value | Output |
| ----- | ------ |
| `summary` | Short text report |
| `detailed` | Full text report (default for `prof track auto` when the flag is omitted) |
| `summary-html` | HTML summary |
| `detailed-html` | HTML detailed |
| `summary-json` | JSON summary |
| `detailed-json` | JSON detailed |

`prof track manual` requires `--output-format` explicitly (no default).

## `prof auto`

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--benchmarks` | strings (repeatable, comma-separated) | Yes | n/a | Benchmark names to run (for example `BenchmarkGenPool`). |
| `--profiles` | strings | Yes | n/a | Profile IDs, comma-separated (for example `cpu,memory,mutex,block`). |
| `--tag` | string | Yes | n/a | Tag directory name under `bench/`. |
| `--count` | int | Yes | n/a | Number of benchmark iterations or runs `go test` should perform (must be positive). |
| `--group-by-package` | bool | No | `false` | Also write package-grouped text listings. |
| `--lenient-profiles` | bool | No | `false` | Skip missing profile binaries instead of failing the run. |
| `--skip-png` | bool | No | `false` | Succeed even when PNG generation fails (for example Graphviz missing). |

## `prof manual`

Positional arguments are one or more profile file paths to ingest.

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--tag` | string | Yes | n/a | Tag directory name under `bench/`. |
| `--group-by-package` | bool | No | `false` | Same as `prof auto`. |

## `prof track auto`

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--base` | string | Yes | n/a | Baseline tag name (directory under `bench/`). |
| `--current` | string | Yes | n/a | Current tag name. |
| `--bench-name` | string | Yes | n/a | Benchmark name (must match collect). |
| `--profile-type` | string | Yes | n/a | One of `cpu`, `memory`, `mutex`, `block`. |
| `--output-format` | string | No | `detailed` | See [Compare output formats](#compare-output-formats). |
| `--fail-on-regression` | bool | No | `false` | When combined with a positive `--regression-threshold`, exit with an error if the worst flat-time regression meets or exceeds the threshold. |
| `--regression-threshold` | float | No | `0` | Maximum allowed worst flat regression (percent). Values at or below `0` do not enable the CLI gate unless CI config applies. Set a positive value when using `--fail-on-regression`. |

## `prof track manual`

`--base` and `--current` are filesystem paths to the two profiles to compare (not tag names).

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--base` | string | Yes | n/a | Path to baseline profile file. |
| `--current` | string | Yes | n/a | Path to current profile file. |
| `--output-format` | string | Yes | n/a | See [Compare output formats](#compare-output-formats). |
| `--fail-on-regression` | bool | No | `false` | Same semantics as `prof track auto`. |
| `--regression-threshold` | float | No | `0` | Same semantics as `prof track auto`. |

## `prof tools benchstat`

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--base` | string | Yes | n/a | Baseline tag under `bench/`. |
| `--current` | string | Yes | n/a | Current tag. |
| `--bench-name` | string | Yes | n/a | Benchmark name. |

Requires `benchstat` on `PATH` (`go install golang.org/x/perf/cmd/benchstat@latest`).

## `prof tools qcachegrind`

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--tag` | string | Yes | n/a | Tag to read binaries from. |
| `--profiles` | strings | Yes | n/a | Profile IDs (command uses the first for generation). Pass one such as `cpu` when unsure. |
| `--bench-name` | string | Yes | n/a | Benchmark name. |

## Exit codes

Prof follows normal Go CLI conventions: exit code `0` on success, non-zero when a command returns an error (invalid flags, failed `go test`, missing `bench/` tags, parser errors, or a regression gate failure when configured).

There is no stable assignment of distinct integers per error type today. Treat any non-zero exit as failure and read stderr.

## Next steps

- Step-by-step: [Quickstart](quickstart.md)
- Deeper behavior: [Collect profiling data](collect.md), [Compare runs](compare.md), [CI and regressions](ci.md)

## Related

- [Troubleshooting](troubleshooting.md)
