# CI and regressions

This guide explains how to fail a job when performance regresses using `prof track` flags and optional `track` policy in `prof.json`, without using interactive menus.

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

## JSON in `prof.json`

Add a `track` section for ignores, noise floors, per-benchmark caps, and related policy. When CLI gate flags are omitted, track config applies. CLI flags override when provided. See the canonical schema:

- [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md)

Edit via `prof ui` → Manage configuration, or `prof config init`.

## Testing / verify

- Expect pass: run `prof track auto` on two tags with no meaningful regression; exit code `0`.
- Expect fail: deliberately worsen the hot path, tighten `--regression-threshold`, and confirm a non-zero exit when the worst regression exceeds your cap.
- Formats: add `--output-format summary-json` (or another valid format) when your CI needs machine-readable output ([CLI reference](cli-reference.md#compare-output-formats)).

## Next steps

- [Optional tools](tools.md) for `benchstat` across tags in review workflows.
- [Configure collection](configure.md) for `collection` filters and track policy.

## Related

- [Compare runs](compare.md) · [CLI reference](cli-reference.md) · [Troubleshooting](troubleshooting.md)
