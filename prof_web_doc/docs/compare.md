# Compare runs

This guide covers how to diff two profiling runs so you can see per-function flat and cumulative time changes, optional HTML or JSON reports, and optional CI-style failure when regressions exceed a threshold.

Interactive flows live in [Interactive UI and TUI](tui.md). This page focuses on `prof track` for scripts and automation.

## Before you begin

- Auto compare expects data under `bench/<tag>/` from `prof auto` (or a compatible layout). See [Collect profiling data](collect.md).
- Manual compare expects two file paths you can read from disk (typically profile artifacts you copied or generated outside Prof).
- You know the benchmark name and profile type (`cpu`, `memory`, `mutex`, or `block`) you want to compare.

## What is `prof track auto`?

`prof track auto` loads the binary profile pair for one benchmark and one profile type from two tag directories under `bench/`, builds a function-level diff, prints or writes a report, and optionally fails the process if the worst flat-time regression exceeds a threshold.

## What is `prof track manual`?

`prof track manual` does the same comparison logic using two explicit filesystem paths instead of resolving paths from tag layout. Use it when profiles were not produced by `prof auto` but you still want the same report formats.

## Commands

| Command | Use when |
| ------- | -------- |
| `prof track auto` | Both runs live under `bench/<tag>/` from collect. |
| `prof track manual` | You have two profile file paths; no tag layout required. |

## `prof track auto`

`--base` and `--current` are tag directory names under `bench/` (not full paths to files).

```bash
prof track auto --base "baseline" --current "optimized" \
  --profile-type "cpu" --bench-name "BenchmarkGenPool" \
  --output-format "summary"
```

### Flags

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--base` | string | Yes | n/a | Baseline tag name (`bench/<base>/`). |
| `--current` | string | Yes | n/a | Current tag name. |
| `--bench-name` | string | Yes | n/a | Benchmark name (must match collect). |
| `--profile-type` | string | Yes | n/a | `cpu`, `memory`, `mutex`, or `block`. |
| `--output-format` | string | No | `detailed` | Text, HTML, or JSON report style ([CLI reference](cli-reference.md#compare-output-formats)). |
| `--fail-on-regression` | bool | No | `false` | Enable regression gate together with a positive threshold. |
| `--regression-threshold` | float | No | `0` | Worst flat regression percent above which the command fails when the gate is enabled. |

## `prof track manual`

`--base` and `--current` are paths to profile files on disk (for example files under `bench/<tag>/bin/<BenchmarkName>/`).

```bash
prof track manual --base "path/to/baseline_cpu.out" \
  --current "path/to/candidate_cpu.out" \
  --output-format "summary"
```

`--output-format` is required for `prof track manual` (there is no default).

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--base` | string | Yes | n/a | Filesystem path to baseline profile. |
| `--current` | string | Yes | n/a | Filesystem path to current profile. |
| `--output-format` | string | Yes | n/a | Report format ([CLI reference](cli-reference.md#compare-output-formats)). |
| `--fail-on-regression` | bool | No | `false` | Same as `prof track auto`. |
| `--regression-threshold` | float | No | `0` | Same as `prof track auto`. |

## Output formats

| Value | Description |
| ----- | ----------- |
| `summary` or `detailed` | Text to stdout |
| `summary-html` or `detailed-html` | HTML files |
| `summary-json` or `detailed-json` | JSON files |

## Regression gate

`--fail-on-regression` with a positive `--regression-threshold` compares the worst flat-time regression percent against your cap. Flat time is time in the function itself excluding callees.

Full rules, `config_template.json` integration, and links to the CI schema: [CI and regressions](ci.md).

## Testing / verify

- Success: command exits `0`; stdout or generated files contain a report with function-level rows.
- Invalid format: non-zero exit and a message listing valid format names ([Troubleshooting](troubleshooting.md#invalid-output-format)).
- Regression gate: non-zero exit when the worst regression meets or exceeds the threshold ([Troubleshooting](troubleshooting.md#regression-gate-always-passes-or-does-not-fail-the-build)).

## Next steps

- [CI and regressions](ci.md) for `ci_config` and Actions examples.
- [Optional tools](tools.md) for `benchstat` and QCacheGrind on the same tags.

## Related

- [CLI reference](cli-reference.md) · [Troubleshooting](troubleshooting.md) · [Collect profiling data](collect.md)
