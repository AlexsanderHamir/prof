# Collect profiling data

To collect **without memorizing flags**, use **`prof ui`** (main menu → collect) or **`prof tui`** (collect-only prompts). The sections below document **`prof auto`** and **`prof manual`** for scripts, CI, and advanced cases.

## Commands

| Command | Purpose |
| -------- | -------- |
| `prof auto` | Run `go test` benchmarks and collect selected profile types. |
| `prof manual` | Ingest existing profile binaries (for example from `go test -cpuprofile=…`) and produce the same style of layout. |

## prof auto

Runs benchmarks and collects profiles you list. Required flags: `--benchmarks`, `--profiles`, `--count`, `--tag`.

```bash
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

Optional flags:

| Flag | Effect |
| ---- | ------ |
| `--group-by-package` | Emit additional grouped-by-package text reports (`*_grouped.txt`). |
| `--lenient-profiles` | Continue if a profile binary is missing after the bench run. |
| `--skip-png` | Do not fail the run when PNG generation fails (for example if Graphviz is missing). |

### Typical output layout (`prof auto`)

Layout is under `bench/<tag>/`. Structure can vary slightly by profile type and benchmark name; a representative tree:

```text
bench/baseline/
├── description.txt
├── bin/BenchmarkGenPool/
│   ├── BenchmarkGenPool_cpu.out
│   ├── BenchmarkGenPool_memory.out
│   └── …
├── text/BenchmarkGenPool/
│   ├── BenchmarkGenPool_cpu.txt
│   ├── BenchmarkGenPool_memory.txt
│   └── BenchmarkGenPool.txt
├── cpu_functions/BenchmarkGenPool/
│   └── … function-level snippets …
└── memory_functions/BenchmarkGenPool/
    └── …
```

## prof manual

Processes existing profile files (`.out`, `.prof`, or other paths your toolchain produced). Does not run `go test`. Required: `--tag` and one or more file paths.

```bash
prof manual --tag "external-profiles" cpu.prof memory.prof
```

With package grouping:

```bash
prof manual --tag "external-profiles" --group-by-package cpu.prof memory.prof
```

### Typical output layout (`prof manual`)

Profiles are grouped by file stem under the tag directory, for example:

```text
bench/external-profiles/
├── BenchmarkGenPool_cpu/
│   ├── BenchmarkGenPool_cpu.txt
│   └── functions/
└── memory/
    ├── memory.txt
    └── functions/
```

## Package grouping

With `--group-by-package`, Prof adds text that rolls up functions by import path, which helps when many packages appear in one profile. See grouped text under the `text/` area for your benchmark or profile stem.

## Next article

[Configure collection](configure.md)
