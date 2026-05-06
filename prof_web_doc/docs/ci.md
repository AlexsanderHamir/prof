# CI and regressions

This guide explains how to fail a job when performance regresses using `prof track` flags and optional `ci_config` in `config_template.json`, without using interactive menus.

## Before you begin

- Two comparable runs exist (tags under `bench/` for `prof track auto`, or two profile files for `prof track manual`). See [Collect profiling data](collect.md).
- You have read [Compare runs](compare.md) for `--output-format` and manual vs auto behavior.

## What is flat time?

Flat time is the time spent in the function itself, excluding callees. The regression gate compares the worst flat-time regression percent across functions when you enable failure mode.

Example: baseline flat 100 ms, current 110 ms, about +10% regression in flat time for that function.

## CLI flags

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--fail-on-regression` | bool | No | `false` | When set with a positive `--regression-threshold`, exit non-zero if the worst regression meets or exceeds the threshold. |
| `--regression-threshold` | float | No | `0` | Maximum allowed worst flat regression (percent). Must be greater than zero for the CLI gate to apply. |

```bash
prof track auto --base baseline --current pr-branch \
  --profile-type cpu --bench-name "BenchmarkMyHotPath" \
  --fail-on-regression --regression-threshold 5.0
```

If you pass `--fail-on-regression` but leave the threshold at `0`, the CLI gate does not activate. See [Troubleshooting](troubleshooting.md#regression-gate-always-passes-or-does-not-fail-the-build).

## JSON in `config_template.json`

Add a `ci_config` section for ignores, noise floors, per-benchmark caps, and related policy. Precedence (tightest wins): per-benchmark `max_regression_threshold`, then global `max_regression_threshold`, then `--regression-threshold` when CLI flags are in effect. See the canonical schema for exact rules:

- [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md)

## Testing / verify

- Expect pass: run `prof track auto` on two tags with no meaningful regression; exit code `0`.
- Expect fail: deliberately worsen the hot path, tighten `--regression-threshold`, and confirm a non-zero exit when the worst regression exceeds your cap.
- Formats: add `--output-format summary-json` (or another valid format) when your CI needs machine-readable output ([CLI reference](cli-reference.md#compare-output-formats)).

## Next steps

- [Optional tools](tools.md) for `benchstat` across tags in review workflows.
- [Configure collection](configure.md) to add `ci_config` next to filters.

## Related

- [Compare runs](compare.md) · [CLI reference](cli-reference.md) · [Troubleshooting](troubleshooting.md)
