# Quickstart

This guide gets you from **installed `prof`** to **one tagged collect** and **one compare between two tags** in about **10 minutes**, using either menus or flags.

## Before you begin

- **Go** 1.24.3 or newer ([Install Prof](install.md) lists full prerequisites).
- **Module root**: your shell’s cwd is the directory that contains `go.mod`.
- **Benchmarks**: you have at least one benchmark in `*_test.go` (replace `BenchmarkExample` below with a real name from your module).
- **Terminal**: `prof ui` needs a normal TTY; for non-interactive shells, use the flag flow below.

## What is a tag?

A **tag** is a short label for one profiling run. Prof writes all artifacts for that run under `bench/<tag>/`. You compare two tags (for example `baseline` and `candidate`) to see what changed between runs.

## Path A — Use menus (default)

**Step 1: Start the UI**

```bash
prof ui
```

**Step 2:** In the menu, run **Run benchmarks and collect profiles** once with tag `baseline`, then change code (or stay on the same commit for a dry run), run again with tag `candidate`.

**Step 3:** Choose **Compare two tagged runs**, pick `baseline` and `candidate`, your benchmark, and profile type `cpu`.

**Verify:** You should see compare output in the terminal and directories `bench/baseline/` and `bench/candidate/` with `bin/` and `text/` populated.

If the UI does not start, your environment may not expose a TTY—use Path B or see [Troubleshooting](troubleshooting.md#prof-ui-or-prof-tui-fails-in-ci-or-ides).

## Path B — Use flags (CI or scripts)

**Step 1: Collect baseline**

```bash
prof auto --benchmarks "BenchmarkExample" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

**Step 2: Collect after your change**

```bash
prof auto --benchmarks "BenchmarkExample" --profiles "cpu,memory,mutex,block" --count 10 --tag "candidate"
```

**Step 3: Compare**

```bash
prof track auto --base "baseline" --current "candidate" \
  --profile-type "cpu" --bench-name "BenchmarkExample" \
  --output-format "summary"
```

**Verify:**

- On disk: `bench/baseline/` and `bench/candidate/` each contain `bin/BenchmarkExample/` and `text/BenchmarkExample/`.
- In the terminal: `prof track auto` prints a report; with `--output-format summary` the report is shorter than the default `detailed` format ([Compare runs](compare.md)).

## Windows (PowerShell)

Quoting is the same for these examples. Ensure your cwd is the folder that contains `go.mod`:

```powershell
Set-Location path\to\your\module
prof ui
```

## If something fails

See [Troubleshooting](troubleshooting.md) (wrong cwd, missing tags, Graphviz, regression exit codes).

## Next steps

- [Configure collection](configure.md) for per-function extracts and CI rules.
- [CI and regressions](ci.md) to fail pipelines on regressions.
- [CLI reference](cli-reference.md) for every flag and default.

## Related

- [Collect profiling data](collect.md) · [Compare runs](compare.md) · [Interactive UI and TUI](tui.md)
