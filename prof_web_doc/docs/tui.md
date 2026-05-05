# Interactive TUI

Use the TUI when you prefer menus to typing benchmark names and flags. The TUI calls the same engines as the non-interactive commands and writes the same `bench/<tag>/` layout.

## Collect: prof tui

```bash
prof tui
```

1. Discovers `Benchmark*` functions in the module.
2. Lets you choose benchmarks (multi-select), profile types, run count, and tag.
3. Runs collection equivalent to `prof auto` with your selections.

**Navigation (typical):**

- Arrow keys move the selection.
- Space toggles selection for multi-select lists.
- Typing filters long benchmark lists (page size is capped for readability).

## Compare: prof tui track

```bash
prof tui track
```

1. Lists existing tags under `bench/`.
2. Prompts for baseline tag, current tag, benchmark, profile type, output format, and optional regression options.

**Prerequisite:** at least two tags already exist from `prof auto` or `prof tui`.

## Next article

[CI and regressions](ci.md)
