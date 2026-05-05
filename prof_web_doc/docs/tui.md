# Interactive UI and TUI

Same engines as `prof auto` / `prof track`; output under `bench/<tag>/`. Details: [Collect profiling data](collect.md), [Compare runs](compare.md).

!!! important

    **`prof ui`**, **`prof tui`**, and **`prof tui track`** need a normal terminal (stdin and stdout must be TTYs). For automation, use flag commands and `prof -h`.

## prof ui (recommended)

```bash
prof ui
```

Bubble Tea full-screen menu; Survey prompts after you choose **Run benchmarks and collect profiles**, **Compare two tagged runs**, **Tools** (benchstat / qcachegrind), **Create configuration template**, **Show documentation URL**, or **Quit**.

## prof tui (collect only)

```bash
prof tui
```

Multi-select benchmarks and profiles, count, tag, then group-by-package, lenient profiles, skip PNG (same as `prof auto`).

**Navigation:** arrows; Space toggles in multi-select; type to filter long lists.

## prof tui track (compare only)

```bash
prof tui track
```

Needs **two tags** under `bench/` (from any collect path). Prompts: baseline, current, benchmark, profile type, output format, optional regression gate.

## Next steps

[Collect profiling data](collect.md) · [Compare runs](compare.md) · [CI and regressions](ci.md)
