# Testing guide

This document describes where tests live, how to run them, and how to measure coverage.

## Required command

Always run the **full** suite before pushing (same as CI):

```bash
go test ./...
```

CI runs this on every push/PR with Graphviz installed (see [`.github/workflows/test.yaml`](.github/workflows/test.yaml)).

## Optional fast loops

These do **not** replace `go test ./...`:

| Goal | Command |
|------|---------|
| One package while editing | `go test ./engine/collect/ -count=1` |
| Edge tests only | `go test ./parser ./engine/collect -count=1 -run '^TestEdge'` |
| Debug: skip slow integration | `go test -short ./...` (see note below) |

`-short` skips subprocess integration tests in [`tests/blackbox_test.go`](tests/blackbox_test.go). It is for local debugging only; CI never passes `-short`.

## Three test layers

| Layer | Where | When to add |
|-------|-------|-------------|
| **Unit** | `*_test.go` next to code (`cli/`, `internal/`, `engine/`, `parser/`) | Pure logic, validation, stubs — no subprocess |
| **Fixture** | Same package (`TestEdge_*` in `parser/`, `engine/collect/`) or `tests/edgecases_*_test.go` | Committed `.out` fixtures, `tooling.FakeRunner`, one real `pprof` at most |
| **Integration** | `tests/` only (`blackbox_test.go`) | End-to-end `prof` binary + `go test -bench` wiring |

**Rule:** new behavior tests go in the package you changed. Add to `tests/` only when the full CLI subprocess path must be proven.

### Patterns to copy

| Change | Copy from |
|--------|-----------|
| CLI flags / wiring | [`cli/commands_test.go`](cli/commands_test.go) |
| Intent validation | [`internal/intent/collect_test.go`](internal/intent/collect_test.go) |
| Config / filters | [`internal/config/config_test.go`](internal/config/config_test.go) |
| Layout paths | [`internal/workspace/layout_test.go`](internal/workspace/layout_test.go) |
| Parser fixtures | [`parser/parser_test.go`](parser/parser_test.go), [`parser/edge_test.go`](parser/edge_test.go) |
| Collect (expand here) | [`engine/collect/manual_test.go`](engine/collect/manual_test.go), [`engine/collect/edge_test.go`](engine/collect/edge_test.go) |
| Edge invariants | [`parser/edge_test.go`](parser/edge_test.go), [`engine/collect/edge_test.go`](engine/collect/edge_test.go) |

## Coverage

### Run coverage (full suite + report)

Uses the Go toolchain only:

- [`go test -coverprofile`](https://pkg.go.dev/cmd/go#hdr-Testing_flags) with [`-coverpkg=./...`](https://go.dev/blog/integration-test-coverage) so integration tests in `tests/` count toward every package they exercise
- [`go tool cover`](https://pkg.go.dev/cmd/go#hdr-Show_coverage_results) for HTML and function-level totals
- [`scripts/coverreport/main.go`](scripts/coverreport/main.go) for a sorted per-package table (lowest coverage first)

**Linux / macOS:**

```bash
./scripts/test-cover.sh
./scripts/test-cover.sh -html    # also writes coverage.html
```

**Windows (PowerShell):**

```powershell
.\scripts\test-cover.ps1
.\scripts\test-cover.ps1 -WriteHtml
```

Artifacts (gitignored): `coverage.out`, `coverage.html`.

### Reading the numbers

| Metric | Meaning |
|--------|---------|
| **Total statement coverage** | From `go tool cover -func` on the merged profile — official module-wide percentage |
| **Per-package table** | Mean of function-level coverage in that package (sorted ascending — gaps at the top) |
| **`go test ./pkg -cover`** | Per-package % when only that package's tests run — useful for focused work |

There are **no merge-blocking coverage gates**. Run the script, pick the lowest package in the table, add a test next to that code, re-run.

### Baseline snapshot

Full suite (`./scripts/test-cover.sh` or `.ps1` with `-coverpkg=./...`):

- **Total statement coverage:** ~51.5% (from `go tool cover -func`)
- **Lowest packages (add tests here first):** `cmd/prof`, `engine/collect` (~29% function mean)

Re-run the script after adding tests and note the new total in your PR if coverage changed materially.

## Domain map

| Domain | Unit tests | Integration |
|--------|------------|-------------|
| `internal/config`, `internal/workspace` | `*_test.go` in package | — |
| `engine/collect` | `engine/collect/*_test.go`, `edge_test.go` (expand) | `TestAutoEndToEnd`, `TestManualCommand` |
| `parser` | `parser/*_test.go`, `parser/edge_test.go` | — |
| `cli`, `internal/intent`, `internal/tui` | stub `app.Services` | command validation in `tests/` |
| `engine/tooling` | `engine/tooling/*_test.go` | — |

## Fixtures

- Committed pprof binaries: [`tests/assets/fixtures/`](tests/assets/fixtures/)
- Regenerate: `go generate ./tests/...` (see [`tests/doc.go`](tests/doc.go))

## Related docs

- [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md) — architecture and edge-case catalog
- [CONTRIBUTING.md](CONTRIBUTING.md) — PR workflow
