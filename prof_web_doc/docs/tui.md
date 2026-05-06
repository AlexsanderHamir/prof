# Interactive UI and TUI

This guide covers **`prof ui`** (full-screen menus), **`prof tui`** (terminal collect), and **`prof tui track`** (terminal compare). They call the same engines as `prof auto` and `prof track`; artifacts still land under `bench/<tag>/`.

## Before you begin

- **TTY required:** stdin and stdout must be a normal interactive terminal.
- **Module root** as cwd ([Working directory and paths](workspace.md)).

!!! important

    `prof ui`, `prof tui`, and `prof tui track` need a normal terminal (stdin and stdout must be TTYs). For automation, use flag commands and `prof -h`. See [Troubleshooting](troubleshooting.md#prof-ui-or-prof-tui-fails-in-ci-or-ides).

## What is `prof ui`?

`prof ui` is the **recommended first experience**: a Bubble Tea full-screen menu where you choose **Run benchmarks and collect profiles**, **Compare two tagged runs**, **Tools** (`benchstat` / QCacheGrind), **Create configuration template**, **Show documentation URL**, or **Quit**.

### Start the UI

```bash
prof ui
```

After you pick an action, Survey-style prompts collect parameters (benchmarks, profiles, tags, and so on)—equivalent to the flags documented in [Collect profiling data](collect.md) and [Compare runs](compare.md).

## What is `prof tui`?

`prof tui` is a **collect-only** terminal flow: multi-select benchmarks and profiles, set count and tag, then options such as group-by-package, lenient profiles, and skip PNG (same semantics as `prof auto`).

```bash
prof tui
```

**Navigation:** arrow keys; **Space** toggles in multi-select; type to filter long lists.

## What is `prof tui track`?

`prof tui track` is a **compare-only** flow. You need **at least two tags** under `bench/` from any prior collect. Prompts cover baseline, current, benchmark, profile type, output format, and optional regression gate.

```bash
prof tui track
```

## Testing / verify

- **Collect:** After finishing `prof tui` or the collect path in `prof ui`, confirm `bench/<tag>/` exists with `bin/` and `text/` populated.
- **Compare:** You should see report text in the terminal or the same files/HTML/JSON outputs as `prof track` ([Compare runs](compare.md)).

## Next steps

- [Collect profiling data](collect.md) for exact flag meanings mirrored in the TUIs.
- [Compare runs](compare.md) for output formats and regression options.
- [CI and regressions](ci.md) when moving the same flows to pipelines.

## Related

- [Quickstart](quickstart.md) · [Troubleshooting](troubleshooting.md) · [CLI reference](cli-reference.md)
