# Contributing

Thanks for your interest in improving **`prof`**! We welcome contributions of all kinds — bug fixes, features, tests, and documentation.

Before starting, check [issues](https://github.com/AlexsanderHamir/prof/issues) to avoid duplication or to pick something up.

## Codebase overview

**Architecture:** CLI → [`internal/app` composition root](internal/app/services.go) → `engine/*` → `parser` / `internal` helpers.

**Principles:**

- Single responsibility per package
- Stable interfaces at the `app.Services` boundary (easier testing and alternate backends)
- JSON-driven filters and CI behavior where possible
- **Errors**: propagate `error`; wrap with `%w`; optional lenience only via documented flags (`--skip-png`, `--lenient-profiles`) — see [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md#error-handling-project-rules).

**Structure (matches the repo tree):**

- `cmd/prof/` – Program entry (`main`)
- `cli/` – Cobra commands, flags, interactive TUI
- `internal/app/` – Interfaces and default wiring into engines
- `engine/benchmark/` – `go test` orchestration and `bench/<tag>/` layout
- `engine/collector/` – Profile ingestion, text/PNG/function outputs, manual flow
- `engine/tracker/` – Compare runs, reports, CI-style filtering
- `engine/tools/` – Optional tooling (benchstat, qcachegrind)
- `parser/` – pprof decoding, aggregation, line/package reports (`Pipeline`)
- `internal/` – Shared config types (`Config`, `FunctionFilter`, …), command wire types (`BenchArgs`, `CollectionArgs`), constants, filesystem helpers (`LoadFromFile`, `FindGoModuleRoot`, …). *Not split into nested `internal/config` packages—everything lives here as `.go` files.*
- `tests/` – Integration and blackbox checks

📖 Diagrams, command → file map, and sharp edges: [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md).

## Quick Start

**Requirements:** Go 1.24.3+, Git

```bash
git clone https://github.com/AlexsanderHamir/prof.git
cd prof
go mod tidy

go test ./...

go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

go build -o prof ./cmd/prof
```

## Contribution guidelines

1. **Tests and linting** — Run before pushing:

   ```bash
   go test ./...
   golangci-lint run
   ```

2. **Code style** — Idiomatic Go; small functions; exported symbols commented when non-obvious.

3. **Commits** — One logical change per commit; descriptive messages (`feat:`, `fix:`, `docs:`, `refactor:`).

4. **Tests** — Prefer tests for bug fixes and new behavior; integration tests when touching CLI/output layout.

5. **Documentation** — Update README, CODEBASE_DESIGN, or CLI help when user-visible behavior changes.

6. **Pull requests** — Clear summary; reference issues (`Closes #123`).

7. **Reviews** — Discussion and iterative feedback welcome.
