# Install Prof

## Prerequisites

| Requirement | Notes |
| ----------- | ----- |
| Go | 1.24.3+ (`go.mod` in repo). |
| Module root | Your project’s `go.mod` when you run Prof. |
| Graphviz | Optional; for PNG call graphs. If missing, use `prof auto --skip-png` or skip PNG in **`prof ui`** / **`prof tui`**. |

## Install the binary

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

```bash
prof --help
```

From the module root, run **`prof ui`** or follow [Quickstart](quickstart.md).

## Next step

[Quickstart](quickstart.md)
