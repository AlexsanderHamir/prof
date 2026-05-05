# Quickstart

This walkthrough assumes your module root is the current directory and benchmarks already exist in `*_test.go` files.

**Memorizing `prof auto` and `prof track` flags is optional.** The steps below start with menus; flag-only steps are for scripts, CI, or when you explicitly want command-line control.

## Option A: Interactive (recommended)

From the module root:

```bash
prof ui
```

`prof ui` opens a **full-screen menu** (Bubble Tea). Pick **Run benchmarks and collect profiles** to capture a baseline, run `prof ui` again after your code changes and collect with a new tag, then choose **Compare two tagged runs** when you have at least two tags under `bench/`. The follow-up prompts use the same engines as `prof auto` and `prof track auto`.

For a narrower flow without the main menu, see [Interactive UI and TUI](tui.md) (`prof tui`, `prof tui track`).

## Option B: Commands with flags (scripts and CI)

Use this path when you need stable, repeatable commands—for example in pipelines.

### 1. Collect a baseline

```bash
prof auto --benchmarks "BenchmarkExample" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

### 2. Collect a second run

After you change code, collect again with a different tag:

```bash
prof auto --benchmarks "BenchmarkExample" --profiles "cpu,memory,mutex,block" --count 10 --tag "candidate"
```

### 3. Compare the two tags

```bash
prof track auto --base "baseline" --current "candidate" --profile-type "cpu" --bench-name "BenchmarkExample" --output-format "summary"
```

## Results

- Profile binaries and text reports are under `bench/<tag>/`. See [Collect profiling data](collect.md) for the layout.
- The track command prints a summary or detailed report to stdout (and optional HTML or JSON when you select those formats). See [Compare runs](compare.md).

## Next steps

- [Configure collection](configure.md) to limit which functions are extracted.
- [CI and regressions](ci.md) to fail a pipeline on regressions.
