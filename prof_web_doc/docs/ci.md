# CI and regressions

## Command-line gating

`prof track auto` (and `prof track manual`) can exit with a non-zero status when regressions exceed a limit.

| Flag | Purpose |
| ---- | -------- |
| `--fail-on-regression` | Enable failure based on thresholds. |
| `--regression-threshold` | Maximum allowed **flat-time** regression percent for the worst offender before exit code is non-zero. |

Flat time is time attributed directly to the function, not including callees.

**Example calculation:** baseline flat 100 ms, current flat 110 ms → approximately +10% flat regression.

## JSON configuration (recommended for pipelines)

For stable CI, add a `ci_config` section to `config_template.json` at the module root. You can set global ignores, minimum change noise floors, and per-benchmark regression caps. Command-line flags can be omitted or combined with file-based rules depending on your setup.

**Priority (most restrictive wins for regression caps):** benchmark-specific `max_regression_threshold`, then global `max_regression_threshold`, then `--regression-threshold`.

## Authoritative reference

The repository maintains the full schema, worked examples, and GitHub Actions patterns:

[CI/CD configuration (repository)](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md)

## Minimal workflow example

```bash
prof track auto --base baseline --current pr-branch \
  --profile-type cpu --bench-name "BenchmarkMyHotPath" \
  --fail-on-regression --regression-threshold 5.0
```

## Next article

[Optional tools](tools.md)
