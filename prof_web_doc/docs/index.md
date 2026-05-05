# Prof documentation

Prof runs Go benchmarks with `go test` profiling, writes output under `bench/<tag>/`, and compares two tags for regressions. Use **`prof ui`** for menus; use **`prof auto`**, **`prof track`**, and flags when you need scripts or CI (`prof -h`).

## Terminology

| Term | Meaning |
| ---- | ------- |
| **Module root** | Directory with your `go.mod`; run Prof from here (same as for `go test`). |
| **Tag** | Label for one run; data lives in `bench/<tag>/`. |
| **Baseline / current** | Two tags you compare (before vs after). |
| **Profile type** | One of `cpu`, `memory`, `mutex`, `block`. |

## Articles

| Article | Purpose |
| ------- | ------- |
| [Install Prof](install.md) | Install binary, completion, try `prof ui`. |
| [Quickstart](quickstart.md) | First successful collect and compare. |
| [Working directory and paths](workspace.md) | Where benchmarks are found; where files go. |
| [Collect profiling data](collect.md) | `prof auto`, `prof manual`. |
| [Compare runs](compare.md) | `prof track auto`, `prof track manual`. |
| [Configure collection](configure.md) | `prof setup`, `config_template.json`. |
| [Interactive UI and TUI](tui.md) | `prof ui`, `prof tui`, `prof tui track`. |
| [CI and regressions](ci.md) | Gates, exit codes, link to full CI config. |
| [Optional tools](tools.md) | `prof tools` (also from **`prof ui`**). |

## Source

[Prof on GitHub](https://github.com/AlexsanderHamir/prof). Full CI schema: [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md).
