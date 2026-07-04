# Quickstart

This guide gets you from an installed `prof` binary to a tagged collect in about 10 minutes, using either menus or flags.

## Before you begin

- Go 1.24.3 or newer ([Install Prof](install.md) lists full prerequisites).
- Module root: your shell cwd is the directory that contains `go.mod`.
- Benchmarks: you have at least one benchmark in `*_test.go` (replace `BenchmarkExample` below with a real name from your module).
- Terminal: `prof ui` needs a normal TTY; for non-interactive shells, use the flag flow below.

## What is a tag?

A tag is a short label for one profiling run. Prof writes all artifacts for that run under `bench/<tag>/`.

## Path A: menus (default)

### Step 1: start the UI

```bash
prof ui
```

### Step 2: collect

In the menu, choose Collect Profiles, pick your benchmarks and profile types, and enter a tag such as `baseline`.

Verify: you should see `bench/baseline/` with `profiles/`, `measurements/`, and `hotspots/` populated.

If the UI does not start, your environment may not expose a TTY. Use Path B or see [Troubleshooting](troubleshooting.md#prof-ui-or-prof-tui-fails-in-ci-or-ides).

## Path B: flags (CI or scripts)

```bash
prof auto --benchmarks "BenchmarkExample" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

Verify:

- On disk: `bench/baseline/` contains `profiles/BenchmarkExample/`, `measurements/BenchmarkExample/`, and `hotspots/BenchmarkExample/`.

## If something fails

See [Troubleshooting](troubleshooting.md) (wrong cwd, missing tags, Graphviz).

## Next steps

- [Configure collection](configure.md) for per-function extracts.
- [CLI reference](cli-reference.md) for every flag and default.

## Related

- [Collect profiling data](collect.md) · [Interactive UI and TUI](tui.md)
