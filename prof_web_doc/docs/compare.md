# Compare runs

Interactive compare: **`prof ui`** or **`prof tui track`**. Below: **`prof track`** for flags and CI.

## Commands

| Command | Use when |
| -------- | -------- |
| `prof track auto` | Both runs live under `bench/<tag>/` from collect. |
| `prof track manual` | You have two profile **file paths**; no tag layout required. |

## prof track auto

Required: `--base`, `--current`, `--profile-type`, `--bench-name`. Optional: `--output-format`, `--fail-on-regression`, `--regression-threshold`.

`--base` / `--current` are **tag directory names** under `bench/`, not file paths.

```bash
prof track auto --base "baseline" --current "optimized" \
  --profile-type "cpu" --bench-name "BenchmarkGenPool" \
  --output-format "summary"
```

## prof track manual

Required: `--base`, `--current`, `--output-format`. Here `--base` / `--current` are **paths** to profile binaries (e.g. `.out`, `.prof`).

```bash
prof track manual --base "path/to/baseline_cpu.out" \
  --current "path/to/candidate_cpu.out" \
  --output-format "summary"
```

## Output formats

| Value | Description |
| ----- | ----------- |
| `summary` / `detailed` | Text; default for `track auto` is `detailed` if omitted (check `prof track auto -h`). |
| `summary-html` / `detailed-html` | HTML. |
| `summary-json` / `detailed-json` | JSON. |

## Regression gate (short)

`--fail-on-regression` with `--regression-threshold` uses worst **flat-time** regression %. More rules: [CI and regressions](ci.md).

## Next article

[Interactive UI and TUI](tui.md)
