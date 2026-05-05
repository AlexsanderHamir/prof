# Collect profiling data

Interactive collect: **`prof ui`** or **`prof tui`**. Below: **`prof auto`** and **`prof manual`** for flags and CI.

## Commands

| Command | Purpose |
| -------- | -------- |
| `prof auto` | Run benchmarks; collect profile types you list. |
| `prof manual` | Ingest existing profile binaries; same layout style. |

## prof auto

Required: `--benchmarks`, `--profiles`, `--count`, `--tag`.

```bash
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

| Flag | Effect |
| ---- | ------ |
| `--group-by-package` | Extra grouped-by-package text (`*_grouped.txt`). |
| `--lenient-profiles` | Skip missing profile binaries instead of failing. |
| `--skip-png` | Do not fail if PNG generation fails (e.g. no Graphviz). |

### Layout (typical)

Under `bench/<tag>/`: `description.txt`, `bin/<bench>/`, `text/<bench>/`, `*_functions/<bench>/` (details vary by profile).

## prof manual

Requires `--tag` and profile file paths. Does not run `go test`.

```bash
prof manual --tag "external-profiles" cpu.prof memory.prof
```

```bash
prof manual --tag "external-profiles" --group-by-package cpu.prof memory.prof
```

## Next article

[Configure collection](configure.md)
