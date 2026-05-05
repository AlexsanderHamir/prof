# Compare runs

## Commands

| Command | When to use it |
| -------- | ---------------- |
| `prof track auto` | Both runs were collected with `prof auto` and stored under `bench/<tag>/`. |
| `prof track manual` | You have two profile **files** on disk and want a report without the `bench/` tag layout. |

## prof track auto

Required flags: `--base`, `--current`, `--profile-type`, `--bench-name`. Optional: `--output-format`, `--fail-on-regression`, `--regression-threshold`.

```bash
prof track auto --base "baseline" --current "optimized" \
  --profile-type "cpu" --bench-name "BenchmarkGenPool" \
  --output-format "summary"
```

```bash
prof track auto --base "baseline" --current "optimized" \
  --profile-type "cpu" --bench-name "BenchmarkGenPool" \
  --output-format "detailed"
```

- `--base` and `--current` are **tag names** (directory names under `bench/`), not file paths.

## prof track manual

Required flags: `--base`, `--current`, `--output-format`.

Each of `--base` and `--current` must be a **filesystem path** to a profile file for that run (the same binary profile format `go test` produces, such as `.out` or `.prof`). Despite the flag names, these are not tag labels.

```bash
prof track manual --base "path/to/baseline_cpu.out" \
  --current "path/to/candidate_cpu.out" \
  --output-format "summary"
```

**Note:** The built-in command description may refer to “text files”; the implementation loads standard pprof profile binaries. Prefer the file types you captured with `go test` or `prof auto`.

## Output formats

| Value | Description |
| ----- | ----------- |
| `summary` | High-level list of regressions, improvements, and stable functions. |
| `detailed` | Per-function sections with flat and cumulative metrics. |
| `summary-html` | Summary as HTML. |
| `detailed-html` | Detailed report as HTML. |
| `summary-json` | Summary as JSON. |
| `detailed-json` | Detailed report as JSON. |

Default for `track auto` if omitted is `detailed` (see CLI help for your version).

## Regression gating (summary)

To fail the process when the worst **flat-time** regression exceeds a percentage, use `--fail-on-regression` with `--regression-threshold`. Details and JSON-based CI rules are in [CI and regressions](ci.md).

## Next article

[Interactive TUI](tui.md)
