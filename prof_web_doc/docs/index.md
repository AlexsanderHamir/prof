# Prof documentation

Prof is a command-line tool for Go benchmarks. It runs `go test` with the profiling flags you choose, writes artifacts under `bench/<tag>/`, and can compare two tagged runs to surface function-level changes—including optional CI failure when regressions exceed a threshold.

**You do not have to memorize subcommands or flags for everyday use.** Start with **[Interactive UI and TUI](tui.md)** (`prof ui`, or `prof tui` / `prof tui track` for a single workflow). Use `prof auto`, `prof track`, and the rest when you want stable, copy-pastable commands—for example in CI or Makefiles.

## When to use Prof

Use Prof when you want repeatable benchmark runs, a fixed directory layout for binaries and reports, and diffs between baselines and candidates without scripting `pprof` yourself.

## Interactive UI and TUI versus flags

- **Menus (default path):** Run **`prof ui`** for a full-screen main menu (Bubble Tea), then prompts for collect, compare, tools, or setup. Use **`prof tui`** for collect-only prompts or **`prof tui track`** for compare-only prompts. These require a normal interactive terminal (stdin and stdout must be TTYs).
- **Flags (when you choose them):** Use **`prof auto`**, **`prof track`**, **`prof tools`**, and so on when you want scriptable, reproducible one-liners—not because they are the only way to use Prof. Run `prof -h` and `prof <command> -h` for flag reference.
- **Completion:** Run **`prof completion <shell>`** to print a completion script for bash, zsh, fish, or PowerShell.

## In this documentation

| Article | Description |
| -------- | ----------- |
| [Install](install.md) | Requirements, installation, completion scripts. |
| [Quickstart](quickstart.md) | **`prof ui`** first, then the same workflow with flags for CI. |
| [Working directory and paths](workspace.md) | Where Prof searches for tests and where files are written. |
| [Collect profiling data](collect.md) | `prof auto`, `prof manual`, and package grouping (plus interactive collect via UI/TUI). |
| [Configure collection](configure.md) | `prof setup` and `config_template.json`. |
| [Compare runs](compare.md) | `prof track auto`, `prof track manual`, and report formats (plus interactive compare via UI/TUI). |
| [Interactive UI and TUI](tui.md) | **`prof ui`**, **`prof tui`**, **`prof tui track`**, and how they map to flag commands. |
| [CI and regressions](ci.md) | Thresholds, exit codes, and JSON configuration. |
| [Optional tools](tools.md) | `prof tools benchstat` and `prof tools qcachegrind` (also available from **`prof ui`**). |

## Source and issue tracking

Product source and revision history live in the [Prof repository](https://github.com/AlexsanderHamir/prof). For deeper CI schema and examples, the repository also ships [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md) alongside the code.
