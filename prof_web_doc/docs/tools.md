# Optional tools

This guide covers **`prof tools`**: commands that read **`bench/<tag>/`** and produce extra artifacts such as **`benchstat`** summaries or **QCacheGrind** callgrind files. The same flows are available from **`prof ui` → Tools**.

## Before you begin

- You have already collected data into `bench/` ([Collect profiling data](collect.md)).
- For **`benchstat`**, the `benchstat` binary is on your `PATH` ([Troubleshooting](troubleshooting.md#benchstat-not-found)).
- For **QCacheGrind**, the GUI is installed ([Troubleshooting](troubleshooting.md#qcachegrind-not-installed)).

## `prof tools benchstat`

[`benchstat`](https://pkg.go.dev/golang.org/x/perf/cmd/benchstat) compares benchmark text results across two tags.

```bash
prof tools benchstat --base baseline --current optimized --bench-name BenchmarkGenPool
```

Install helper:

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

**Output:** `bench/tools/benchstat/<BenchmarkName>_results.txt`

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--base` | string | Yes | — | Baseline tag under `bench/`. |
| `--current` | string | Yes | — | Current tag. |
| `--bench-name` | string | Yes | — | Benchmark name. |

## `prof tools qcachegrind`

Generates Callgrind input for [QCacheGrind](https://kcachegrind.github.io/html/Home.html) from stored binary profiles.

```bash
prof tools qcachegrind --tag optimized --profiles cpu --bench-name BenchmarkGenPool
```

**Output:** `bench/tools/qcachegrind/<BenchmarkName>_<profile>.callgrind`

| Flag | Type | Required | Default | Description |
| ---- | ---- | --------- | ------- | ----------- |
| `--tag` | string | Yes | — | Tag to read from. |
| `--profiles` | strings | Yes | — | Profile IDs; the command uses the **first** ID you pass—pass a single profile such as `cpu` when unsure. |
| `--bench-name` | string | Yes | — | Benchmark name. |

## Testing / verify

- After **`benchstat`**, open `bench/tools/benchstat/<BenchmarkName>_results.txt` and confirm it references both tags.
- After **`qcachegrind`**, confirm `.callgrind` files exist and open in QCacheGrind.

## Next steps

- [Compare runs](compare.md) for statistical diff of profiles rather than `benchstat` alone.
- [Collect profiling data](collect.md) if `bench/` is missing.

## Related

- [CLI reference](cli-reference.md) · [Troubleshooting](troubleshooting.md) · [Interactive UI and TUI](tui.md)
