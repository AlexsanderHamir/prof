# Interactive UI and TUI

This guide covers `prof ui` (full-screen menus) and `prof tui` (terminal collect). They call the same engines as `prof auto`; artifacts land under `bench/<tag>/`.

## Before you begin

- TTY: stdin and stdout must be a normal interactive terminal.
- Module root as cwd ([Working directory and paths](workspace.md)).

!!! important

    `prof ui` and `prof tui` need a normal terminal (stdin and stdout must be TTYs). For automation, use flag commands and `prof -h`. See [Troubleshooting](troubleshooting.md#prof-ui-or-prof-tui-fails-in-ci-or-ides).

## What is `prof ui`?

`prof ui` is the recommended first run: a Bubble Tea full-screen menu where you choose Collect Profiles, Create Configuration File, Documentation Site, or Quit.

### Start the UI

```bash
prof ui
```

After you pick an action, Survey-style prompts collect parameters (benchmarks, profiles, tags, and so on), equivalent to the flags documented in [Collect profiling data](collect.md).

## What is `prof tui`?

`prof tui` is a collect-only terminal flow: multi-select benchmarks and profiles, set count and tag, then options such as lenient profiles and skip PNG (same semantics as `prof auto`).

```bash
prof tui
```

Navigation: arrow keys; Space toggles in multi-select; type to filter long lists.

## Testing / verify

After finishing `prof tui` or the collect path in `prof ui`, confirm `bench/<tag>/` exists with `bin/` and `text/` populated.

## Next steps

- [Collect profiling data](collect.md) for exact flag meanings mirrored in the TUIs.

## Related

- [Quickstart](quickstart.md) · [Troubleshooting](troubleshooting.md) · [CLI reference](cli-reference.md)
