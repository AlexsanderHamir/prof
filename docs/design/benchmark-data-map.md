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
2. **hotspots** — open linked `hotspots/*.txt` for flat/cum rankings (`profile_cost_columns` explains columns)
3. **call_trees** — open linked `call_trees/*.txt` for caller/callee context
4. **source_lines** — line-level detail for chosen symbols (path index only in map.json)
5. **profiles** — raw binary only if deeper pprof queries needed

## Mapping vs metrics

`map.json` is an **artifact index**, not a copy of profile sample data.

| Section | map.json holds | Metrics live in |
| --- | --- | --- |
| `hotspots` | Path to `-top` text + column glossary | `hotspots/<benchmark>/<profile>.txt` |
| `source_lines.functions` | Path, `full_symbol`, `status` | Prior hotspots read; line detail in linked `.txt` |
| `profiles` | Path + profile total (orientation) | Raw `.out` binary |

Do not expect `flat`/`cum` on `source_lines` entries — agents reach source_lines after choosing a symbol from hotspots.

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
      "hotspots_metrics_note": "flat/cum rankings and sample values are in the hotspots text at path..."
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

## Profile cost columns (flat / cum)

`go tool pprof -top` prints five metric columns. Read them in `hotspots/*.txt`; `profile_cost_columns` in map.json explains each column:

| Column | Meaning |
| --- | --- |
| `flat` | Cost in this function's own code only (excludes callees). CPU: time in the function body; memory: bytes allocated there. |
| `flat%` | `flat` as % of total profile samples. |
| `sum%` | Running sum of `flat%` reading the `-top` table top to bottom. |
| `cum` | Cost in this function plus all functions it called. CPU: seconds; memory: bytes including callees. |
| `cum%` | `cum` as % of total profile samples. |

Each emitted `map.json` includes `profile_cost_columns` and `profile_cost_triage` so agents can interpret the hotspots text file:

> High flat: optimize this function's body. High cum but low flat: work is mostly in callees — check call_trees or child symbols.

## Sample units and display fields

`flat` / `cum` / `total_samples` on the **profiles** section are raw profile totals in the pprof sample unit (nanoseconds for CPU, bytes for heap profiles). Per-function metrics are **not** duplicated in map.json — read `hotspots/*.txt` instead.

Display strings on **profiles** (`total_display`, `total_seconds`) match the `go tool pprof -top` header. Implementation: [`internal/pprofscale`](../../internal/pprofscale/).

## Invariants

- Same `FunctionListEntry` set as source_lines on disk (from `prof.json` filters).
- Relative paths only — no absolute machine paths in JSON.
- Auto and manual collect produce the same schema shape.
- Tag-level manifest is **out of scope** for v1.

## Consequences

**Positive:** Agents load one JSON file to triage before opening large text artifacts.

**Neutral:** map.json size scales with function count (path index only); much smaller than duplicating per-function metrics.

## Out of scope (v2)

- Tag-level `manifest.json`
- Cross-tag comparison hooks
- `triage_hint` synthesis
- Hot line parsing from `-list` output
- Post-hoc disk scanning

## See also

- [Collect request flow — Step 4](../collect-request-flow.md#step-4--benchmark-data-map)
- [Parallel source_lines collection](./source-lines-parallelism.md)
