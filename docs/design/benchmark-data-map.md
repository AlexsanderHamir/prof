# Benchmark data map (map.json)

## Context

After collect, `.prof/<tag>/` holds hotspots, call trees, source lines, measurements, and raw profiles across **asymmetric paths**. LLM agents and tooling must discover what exists, what each artifact means, and which functions were extracted — without scanning the tree or reading hundreds of lines of hotspot text first.

## Decision

Emit one **machine-readable index** per benchmark at the end of each benchmark pipeline:

```text
.prof/<tag>/data_mapping/<BenchmarkName>/map.json
```

Paths inside JSON are **relative to `.prof/<tag>/`**. The map is built from data **already parsed during collect** (no extra `pprof` subprocesses). Map emit failures **warn and continue** — they do not fail a successful collect run.

Schema types and builder live in [`internal/datamap`](../../internal/datamap/). Layout paths: [`TagLayout.DataMapping`](../../internal/workspace/layout.go).

## Purpose enums

| Value | Section | Meaning |
| --- | --- | --- |
| `go_test_benchmem_results` | measurements | `go test -benchmem` transcript |
| `raw_pprof_binary` | profiles | Source `.out` binary |
| `flat_and_cumulative_ranking` | hotspots | `pprof -top` text |
| `caller_callee_context` | call_trees | `pprof -tree` text |
| `line_level_source_extract` | source_lines | Per-function `pprof -list` |
| `visual_call_graph` | call_graphs | Optional PNG |

## Recommended reading flow

1. **measurements** — confirm benchmark ran; headline ns/op and allocs
2. **hotspots** — find top symbols (`top_symbols` in map)
3. **call_trees** — understand call path (`hot_path_summary`)
4. **source_lines** — line-level detail for chosen symbols
5. **profiles** — raw binary only if deeper pprof queries needed

## Example (truncated)

```json
{
  "schema_version": 1,
  "tag": "baseline",
  "benchmark": "BenchmarkDataGeneration",
  "recommended_flow": ["measurements", "hotspots", "call_trees", "source_lines", "profiles"],
  "measurements": {
    "path": "measurements/BenchmarkDataGeneration/run.txt",
    "purpose": "go_test_benchmem_results",
    "summary": { "ns_per_op_median": 64894, "bytes_per_op": 88056, "allocs_per_op": 1750 }
  },
  "profiles": {
    "cpu": {
      "path": "profiles/BenchmarkDataGeneration/cpu.out",
      "purpose": "raw_pprof_binary"
    }
  },
  "hotspots": {
    "cpu": {
      "path": "hotspots/BenchmarkDataGeneration/cpu.txt",
      "purpose": "flat_and_cumulative_ranking",
      "producer": "go tool pprof -top",
      "top_symbols": [
        {
          "rank": 1,
          "symbol": "github.com/example/benchmarks/utils.(*DataGenerator).GenerateStrings",
          "flat_pct": 1.28,
          "cum_pct": 70.53
        }
      ]
    }
  },
  "call_trees": {
    "cpu": {
      "path": "call_trees/BenchmarkDataGeneration/cpu.txt",
      "purpose": "caller_callee_context",
      "producer": "go tool pprof -tree"
    }
  },
  "source_lines": {
    "cpu": {
      "dir": "source_lines/cpu/BenchmarkDataGeneration",
      "path_pattern": "source_lines/{profile}/{benchmark}/{output_stem}.txt",
      "purpose": "line_level_source_extract",
      "functions": {
        "GenerateStrings": {
          "path": "source_lines/cpu/BenchmarkDataGeneration/GenerateStrings.txt",
          "full_symbol": "github.com/example/benchmarks/utils.(*DataGenerator).GenerateStrings",
          "status": "ok"
        }
      }
    }
  },
  "provenance": {
    "tag": "baseline",
    "collection_mode": "auto",
    "profiles_requested": ["cpu", "memory"]
  },
  "status": {
    "benchmark_run": "ok",
    "profiles": { "cpu": "ok" },
    "source_lines": { "cpu": { "collected": 156, "skipped": 0, "failed": 0 } }
  }
}
```

## Invariants

- Same `FunctionListEntry` set as source_lines on disk (from `prof.json` filters).
- Relative paths only — no absolute machine paths in JSON.
- Auto and manual collect produce the same schema shape.
- Tag-level manifest is **out of scope** for v1.

## Consequences

**Positive:** Agents load one JSON file to triage before opening large text artifacts.

**Neutral:** map.json grows with function count; acceptable for v1.

## Out of scope (v2)

- Tag-level `manifest.json`
- Cross-tag comparison hooks
- `triage_hint` synthesis
- Hot line parsing from `-list` output
- Post-hoc disk scanning

## See also

- [Collect request flow — Step 4](../collect-request-flow.md#step-4--benchmark-data-map)
- [Parallel source_lines collection](./source-lines-parallelism.md)
