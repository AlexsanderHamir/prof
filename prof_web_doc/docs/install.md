# Install Prof

This page explains **how to install the `prof` binary**, what you need on your machine first, and **how to confirm the install**.

## Before you begin

| Requirement | Notes |
| ----------- | ----- |
| Go | 1.24.3 or newer (matches the `go.mod` in this project). |
| Module root | Your project’s `go.mod` directory when you run Prof. |
| Graphviz | Optional; for PNG call graphs. If missing, use `prof auto --skip-png` or skip PNG in `prof ui` / `prof tui`. |

## Install the binary

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

## Confirm the install

```bash
prof --help
```

You should see the root command help with subcommands (`ui`, `auto`, `track`, and others).

From the module root, run `prof ui` or follow [Quickstart](quickstart.md).

## Testing / verify

- `prof --help` exits **0** and prints usage.
- `prof --version` (if supported by your build) shows a version string; development builds may show `devel`.

## Next steps

- [Quickstart](quickstart.md) for your first collect and compare.
- [CLI reference](cli-reference.md) for every subcommand.

## Related

- [Troubleshooting](troubleshooting.md) · [Home](index.md)
