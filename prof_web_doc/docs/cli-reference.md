# CLI reference

This page lists supported commands, profile identifiers, and flag-level defaults so you can script Prof without guessing. Subcommands also print `prof <cmd> -h`.

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
| `prof ui` | Full-screen menu: collect profiles, create configuration, documentation link. |
| `prof tui` | Terminal collect flow (multi-select benchmarks and profiles). |
| `prof auto` | Run `go test` benchmarks and collect listed profiles into `bench/<tag>/`. |
| `prof manual` | Ingest existing profile files into the same layout style (no `go test`). |
| `prof config init` | Create minimal `prof.json` and commented `prof.json.example` next to `go.mod`. |
| `prof config validate` | Load and validate `prof.json`; exit non-zero on error. |
| `prof config path` | Print resolved `prof.json` path. |
| `prof setup` | Hidden alias for `prof config init`. |

## Profile types (`--profiles`)

These IDs are the ones `go test` integration supports (comma-separated for `--profiles`):

| ID | Role |
| -- | ---- |
| `cpu` | CPU profile |
| `memory` | Memory / allocs profile |
| `mutex` | Mutex profile |
| `block` | Block profile |

## `prof auto`

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--benchmarks` | strings (repeatable, comma-separated) | Yes | n/a | Benchmark names to run (for example `BenchmarkGenPool`). |
| `--profiles` | strings | Yes | n/a | Profile IDs, comma-separated (for example `cpu,memory,mutex,block`). |
| `--tag` | string | Yes | n/a | Tag directory name under `bench/`. |
| `--count` | int | Yes | n/a | Number of benchmark iterations or runs `go test` should perform (must be positive). |

## `prof manual`

Positional arguments are one or more profile file paths to ingest.

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--tag` | string | Yes | n/a | Tag directory name under `bench/`. |

## Exit codes

Prof follows normal Go CLI conventions: exit code `0` on success, non-zero when a command returns an error (invalid flags, failed `go test`, missing paths, parser errors).

There is no stable assignment of distinct integers per error type today. Treat any non-zero exit as failure and read stderr.

## Next steps

- Step-by-step: [Quickstart](quickstart.md)
- Deeper behavior: [Collect profiling data](collect.md)

## Related

- [Troubleshooting](troubleshooting.md)
