# Install

## Prerequisites

| Requirement | Notes |
| ----------- | ----- |
| Go | Version 1.24.3 or later (see the repository `go.mod`). |
| Module root | A `go.mod` file at the repository root Prof runs against. |
| Graphviz | Optional for some visualizations; install if you rely on PNG call graphs. Without Graphviz, consider `prof auto --skip-png` if generation fails. |

## Install the binary

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

Verify:

```bash
prof --help
```

## Next step

[Quickstart](quickstart.md)
