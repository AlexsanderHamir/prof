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

## Shell completion (optional)

Prof can emit completion scripts for bash, zsh, fish, and PowerShell:

```bash
prof completion -h
```

Example (bash): save the script and `source` it from your shell config, following your platform’s conventions for completion files.

## Next step

[Quickstart](quickstart.md)
