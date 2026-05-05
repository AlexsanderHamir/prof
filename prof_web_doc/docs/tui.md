# Interactive UI and TUI

Prof is designed so **menus are the default path** and **memorizing flags is optional**. Use the commands in this article from a normal interactive terminal; for automation, use `prof auto`, `prof track`, and the other subcommands directly (see [Collect profiling data](collect.md), [Compare runs](compare.md), and the root help: `prof -h`).

The interactive flows call the **same engines** as the flag commands and write the same `bench/<tag>/` layout.

## Main menu: prof ui (recommended)

```bash
prof ui
```

Requires an interactive terminal (stdin and stdout must be TTYs). If you are automating Prof, use `prof auto`, `prof track`, and other subcommands instead.

The **main menu** is a **Bubble Tea** full-screen UI (arrow keys, Enter, `?` for help, `q` or Esc to quit). After you pick an action, Prof uses **prompted questions** (Survey) for details—same information you would pass as flags.

The menu offers:

1. **Run benchmarks and collect profiles** — same prompts as `prof tui` (benchmarks, profiles, count, tag, then optional group-by-package, lenient profiles, skip PNG).
2. **Compare two tagged runs** — same flow as `prof tui track`.
3. **Tools** — guided `benchstat` or `qcachegrind` using tags and benchmarks discovered under `bench/`.
4. **Create configuration template** — same as `prof setup`, after a short confirmation.
5. **Show documentation URL** — prints the hosted docs link.
6. **Quit** — exit without doing anything.

## Collect only: prof tui

```bash
prof tui
```

1. Discovers `Benchmark*` functions in the module.
2. Lets you choose benchmarks (multi-select), profile types, run count, tag, then group-by-package, lenient profiles, and whether to tolerate PNG generation failures (same options as `prof auto`).
3. Runs collection equivalent to `prof auto` with your selections.

**Navigation (typical):**

- Arrow keys move the selection.
- Space toggles selection for multi-select lists.
- Typing filters long benchmark lists (page size is capped for readability).

## Compare only: prof tui track

```bash
prof tui track
```

1. Lists existing tags under `bench/`.
2. Prompts for baseline tag, current tag, benchmark, profile type, output format, and optional regression options.

**Prerequisite:** at least two tags already exist from `prof auto`, `prof tui`, or **Collect** from `prof ui`.

## Next article

[CI and regressions](ci.md)
