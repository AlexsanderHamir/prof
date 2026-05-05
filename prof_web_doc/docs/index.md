# Prof documentation

Prof is a command-line tool for Go benchmarks. It runs `go test` with the profiling flags you choose, writes artifacts under `bench/<tag>/`, and can compare two tagged runs to surface function-level changes—including optional CI failure when regressions exceed a threshold.

## When to use Prof

Use Prof when you want repeatable benchmark runs, a fixed directory layout for binaries and reports, and diffs between baselines and candidates without scripting `pprof` yourself.

## In this documentation

| Article | Description |
| -------- | ----------- |
| [Install](install.md) | Requirements and installation. |
| [Quickstart](quickstart.md) | Collect, then compare, in a few steps. |
| [Working directory and paths](workspace.md) | Where Prof searches for tests and where files are written. |
| [Collect profiling data](collect.md) | `prof auto`, `prof manual`, and package grouping. |
| [Configure collection](configure.md) | `prof setup` and `config_template.json`. |
| [Compare runs](compare.md) | `prof track auto`, `prof track manual`, and report formats. |
| [Interactive TUI](tui.md) | `prof tui` and `prof tui track`. |
| [CI and regressions](ci.md) | Thresholds, exit codes, and JSON configuration. |
| [Optional tools](tools.md) | `prof tools benchstat` and `prof tools qcachegrind`. |

## Source and issue tracking

Product source and revision history live in the [Prof repository](https://github.com/AlexsanderHamir/prof). For deeper CI schema and examples, the repository also ships [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md) alongside the code.
