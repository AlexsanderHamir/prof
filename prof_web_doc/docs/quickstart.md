# Quickstart

Assume the [module root](index.md#terminology) as cwd and benchmarks in `*_test.go`. Terms: [index](index.md#terminology).

## Use menus (default)

```bash
prof ui
```

Collect with one tag, change code, run again with another tag, then **Compare two tagged runs**. Narrower flows: [Interactive UI and TUI](tui.md) (`prof tui`, `prof tui track`).

## Use flags (CI or scripts)

1. Baseline:

   ```bash
   prof auto --benchmarks "BenchmarkExample" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
   ```

2. After changes:

   ```bash
   prof auto --benchmarks "BenchmarkExample" --profiles "cpu,memory,mutex,block" --count 10 --tag "candidate"
   ```

3. Compare:

   ```bash
   prof track auto --base "baseline" --current "candidate" --profile-type "cpu" --bench-name "BenchmarkExample" --output-format "summary"
   ```

## Output

- Artifacts: `bench/<tag>/` — [Collect profiling data](collect.md).
- Compare report: stdout (more formats in [Compare runs](compare.md)).

## Next steps

- [Configure collection](configure.md)
- [CI and regressions](ci.md)
