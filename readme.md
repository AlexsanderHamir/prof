# Prof

CLI for Go benchmark profiling: run benchmarks with multiple pprof types in one shot, write a consistent `bench/<tag>/` layout, and compare runs for regressions or local iteration.

[![GoDoc](https://godoc.org/github.com/AlexsanderHamir/prof?status.svg)](https://godoc.org/github.com/AlexsanderHamir/prof)
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Version](https://img.shields.io/github/v/tag/AlexsanderHamir/prof?sort=semver)](https://github.com/AlexsanderHamir/prof/releases)
![Go Version](https://img.shields.io/badge/Go-1.24.3%2B-blue)

[Documentation](https://alexsanderhamir.github.io/prof/) · [Demo (CLI)](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/output.mp4) · [Demo (TUI)](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/tui_prof.mp4)

## What you get

Without Prof, you chain `go test` with several profile flags, then repeat `go tool pprof` (and often per-function `-list`) for each artifact. Prof wraps that: one command drives collection, text summaries, and optional package-grouped views under a predictable directory tree.

Output is plain text and standard pprof binaries—easy to diff, archive, or feed into editors and automation.

## Features

| Area | Summary |
|------|---------|
| **Collect** | CPU, memory, mutex, and block profiles in one command (`prof auto`). |
| **Compare** | Diff two tagged runs (`prof track auto`) with several output formats. |
| **CI** | Optional fail-on-regression with thresholds; function filters and JSON config—see [CI/CD configuration](docs/cicd_configuration.md). |
| **Layout** | Artifacts under `bench/<tag>/`; optional `--group-by-package` for `*_grouped.txt` reports. |
| **TUI** | `prof ui` opens a menu (collect, compare, tools, setup); `prof tui` / `prof tui track` are focused prompts for collect and compare. |

**Collect example:**

```bash
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

**Compare example:**

```bash
prof track auto --base "baseline" --current "optimized" --profile-type "cpu" --bench-name "BenchmarkName"
```

**Regression gate example:**

```bash
prof track auto --base "baseline" --current "PR" --profile-type "cpu" --bench-name "BenchmarkName" \
  --fail-on-regression --regression-threshold 5.0
```

**Package grouping:**

```bash
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory" --count 5 --tag "baseline" --group-by-package
prof manual --tag "external-profiles" --group-by-package cpu.prof memory.prof
```

## Interactive TUI

```bash
prof ui             # recommended: main menu, then prompts
prof tui            # collect only: benchmarks, profiles, count, tag
prof tui track      # compare existing runs only
```

Same engines and output layout as the non-interactive commands.

## Installation

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

## Quick start

Interactive (from your module root, with benchmarks in `_test.go`):

```bash
prof ui
```

Same workflow with flags (for scripts and CI):

```bash
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory,mutex,block" --count 10 --tag "optimized"
prof track auto --base "baseline" --current "optimized" --profile-type "cpu" --bench-name "BenchmarkName" --output-format "summary"
```

Shell tab completion: `prof completion bash` (or `zsh`, `fish`, `powershell`) — see `prof completion -h`.

## Documentation

- [Full docs](https://alexsanderhamir.github.io/prof/) — API and guides  
- [Contributing](CONTRIBUTING.md)  
- [Code of conduct](CODE_OF_CONDUCT.md)  
- [Codebase design](CODEBASE_DESIGN.md)  

## Requirements

- Go 1.24.3+
- [Graphviz](https://graphviz.org/) (for pprof graph generation where used)
- A `go.mod` at the repository root

## License

[MIT](LICENSE)
