# Changelog

All notable changes to this project are documented in this file.

## Unreleased

### Changed

- **CLI / `prof auto`:** Missing profile binaries and PNG generation default to **strict** failure (clear error). Use `--lenient-profiles` to skip missing `.out` files after a bench run, and `--skip-png` when Graphviz is not available. The interactive TUI continues to pass `--skip-png`-equivalent behavior internally for smoother local use.
- **Tracking:** Invalid `--output-format` errors list every supported format. HTML/JSON report writers now **return errors** to the CLI instead of logging and exiting successfully.
- **Docs:** [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md) and [CONTRIBUTING.md](CONTRIBUTING.md) now match the real package layout (`internal/app`, `engine/*`, `parser`, unified `internal`).

### Added

- **Parser:** Stable path-based aliases without the `V2` suffix: `TurnLinesIntoObjects`, `GetAllFunctionNames`, `OrganizeProfileByPackage` (wrappers around the existing `*V2` APIs).
- **`internal/repofs`:** Encapsulates `go.mod` root discovery and tag directory cleanup; package `internal` re-exports the same behavior via `FindGoModuleRoot` / `CleanOrCreateTag`.

### Migration

- If `prof auto` fails on PNG generation, install [Graphviz](https://graphviz.org/) or add `--skip-png` to your command or scripts.
- Integration tests under `tests/` append `--skip-png` so CI does not require Graphviz.
