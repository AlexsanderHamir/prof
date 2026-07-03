# Contributing to prof

**Contributing to prof** means opening a pull request with a focused change that passes the same checks CI runs on `main`.

## Before you begin

- **Go 1.24.3+** and **Git**
- Skim [open issues](https://github.com/AlexsanderHamir/prof/issues) to avoid duplicate work or pick something up
- When you need package layout, call flow, or invariants, read [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md) first

## 1. Set up your environment

```bash
git clone https://github.com/AlexsanderHamir/prof.git
cd prof
go mod tidy

go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```

## 2. Make your change

Run the CLI from the repo root while you edit. Put flags and subcommands **after** `--` so `go run` does not consume them:

```bash
go run ./cmd/prof -- version
```

After the first compile, `go run` uses the build cache. Use `go build -o prof ./cmd/prof` only when you need a binary on disk.

Follow the conventions in [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md#before-you-begin) (layering, subprocess routing, error handling). `golangci-lint` enforces the import and `exec` rules at verify time.

## 3. Verify locally

Run both commands from the repository root before every push:

```bash
go test ./...
golangci-lint run
```

All packages must pass. For fast loops while editing, fixture regeneration, coverage scripts, and where to add tests, see [TESTING.md](TESTING.md).

## 4. Open a pull request

1. One logical change per commit (`feat:`, `fix:`, `docs:`, `refactor:`).
2. Write a clear summary; use `Closes #123` when an issue applies.
3. Expect review and iterative feedback.

Coding agents: [AGENTS.md](AGENTS.md).

## Maintainer: cut a release

Releases are manual. Open **Actions → Release → Run workflow** on the branch to ship (usually `main`).

| Outcome | Detail |
| --- | --- |
| Version bump | Next patch from the latest `v*` tag via [`svu`](https://github.com/caarlos0/svu); no tag yet → `v0.1.0` |
| Failure | Errors when there are no new commits since the last tag, or the computed tag already exists |
| Success | Multi-platform binaries, checksums, annotated tag, GitHub release with auto-generated notes |

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@VX.Y.Z
```

## Task recipes

### Add a profile kind

1. Register in [`engine/tooling/catalog.go`](engine/tooling/catalog.go) (`DefaultCatalog`).
2. Confirm [`engine/collect/constants.go`](engine/collect/constants.go) and [`internal/app/profiles.go`](internal/app/profiles.go) pick it up.
3. Run `go test ./...`; update [`engine/tooling`](engine/tooling) or [`engine/collect`](engine/collect) tests if argv or behavior changed.

Use [`tooling.Runner`](engine/tooling/runner.go) in production code so tests can inject [`tooling.FakeRunner`](engine/tooling/fake_runner.go).

### Update documentation

| You changed | Edit |
| --- | --- |
| User install, usage, troubleshooting | [readme.md](readme.md), [prof_web_doc/](prof_web_doc/) |
| Architecture or invariants | [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md) |
| CLI help | [`cli/`](cli/) command definitions |
| Agent tool playbooks | [docs/agents/](docs/agents/) |

Define each workflow once; link instead of copying. User docs: [prof_web_doc/docs/index.md](prof_web_doc/docs/index.md). Tone: [Microsoft Writing Style Guide](https://learn.microsoft.com/en-us/style-guide/welcome/).

## Further reading

| Doc | Use when |
| --- | --- |
| [CODEBASE_DESIGN.md](CODEBASE_DESIGN.md) | Navigating or changing the codebase |
| [TESTING.md](TESTING.md) | Writing or debugging tests |
| [readme.md](readme.md) | User-facing behavior |
| [CODE_OF_CONDUCT.md](CODE_OF_CONDUCT.md) | Community standards |
