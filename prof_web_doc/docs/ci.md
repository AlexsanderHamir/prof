# CI and regressions

`prof track auto` / `prof track manual` can exit **non-zero** when regressions exceed a threshold.

| Flag | Purpose |
| ---- | -------- |
| `--fail-on-regression` | Turn on threshold-based failure. |
| `--regression-threshold` | Max allowed **flat-time** regression % for the worst function before non-zero exit. |

**Flat time** = time in the function itself, excluding callees. Example: baseline flat 100 ms, current 110 ms → about +10%.

## JSON in `config_template.json`

Add a `ci_config` section for ignores, noise floors, per-benchmark caps. **Precedence (tightest wins):** per-benchmark `max_regression_threshold`, then global `max_regression_threshold`, then `--regression-threshold`.

Full schema and Actions examples: [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md).

## Example

```bash
prof track auto --base baseline --current pr-branch \
  --profile-type cpu --bench-name "BenchmarkMyHotPath" \
  --fail-on-regression --regression-threshold 5.0
```

## Next article

[Optional tools](tools.md)
