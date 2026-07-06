# Interactive collect request flow

**Content type:** Explanation — internal call chain for one workflow, not a user tutorial.

This page traces what happens **inside prof** when you run interactive benchmark collection: the Survey prompts from `prof tui`, or **Run Benchmarks & Collect Profiles** in `prof ui`. It complements [CODEBASE_DESIGN.md](../CODEBASE_DESIGN.md) (package map) and the user guides in [prof_web_doc/docs/tui.md](../prof_web_doc/docs/tui.md) and [prof_web_doc/docs/collect.md](../prof_web_doc/docs/collect.md).

## How to use this page

1. **See which command you entered** → [Scope](#scope)
2. **Map a Survey prompt to code** → [Prompt → code mapping](#prompt--code-mapping)
3. **Follow execution after you confirm the last prompt** → [Engine pipeline](#engine-pipeline)
4. **Know what lands on disk** → [Output layout (example)](#output-layout-example)
5. **Find the right file to edit** → [Where to change behavior](#where-to-change-behavior)

## Before you begin

- You know Go module layout and that prof writes under `.prof/<tag>/`.
- For installing and running prof as a user, read [readme.md](../readme.md) and [prof_web_doc/docs/tui.md](../prof_web_doc/docs/tui.md).
- For package boundaries and invariants, read [CODEBASE_DESIGN.md](../CODEBASE_DESIGN.md).

## Scope

This page covers **interactive collect only**:

| Entry | First code |
| --- | --- |
| `prof tui` | [`cli/tui.go`](../cli/tui.go) `runTUI` |
| `prof ui` → **Run Benchmarks & Collect Profiles** | [`internal/tui/hub.go`](../internal/tui/hub.go) `RunMainMenu` → [`cli/cmd_ui.go`](../cli/cmd_ui.go) → `runTUI` |

`prof auto` skips Survey and [`internal/intent`](../internal/intent), but both paths call the same engine entry point:

| Path | Presentation | Engine |
| --- | --- | --- |
| `prof tui` / `prof ui` collect | Survey → `CollectIntent` → `intent.RunValidated` | `app.Services.Collect.RunAuto` → `collect.RunAuto` |
| `prof auto` | Cobra flags in [`cli/cmd_collect.go`](../cli/cmd_collect.go) | Same `collect.RunAuto` |

**Note:** Only `prof ui` returns to the Bubble Tea hub after collect. [`finishUIWorkflow`](../cli/cmd_ui.go) prints errors to stderr, then asks **Return to main menu?** via `promptReturnToHub`. `prof tui` exits when `runTUI` returns.

## Running example

The walkthrough below uses these choices (from a typical interactive session):

- **Benchmark:** `BenchmarkMatrixMultiplication`
- **Profiles:** `cpu`, `memory`
- **Count:** `5`
- **Tag:** `Baseline`

## Entry paths

```mermaid
flowchart LR
  profUI["prof ui"] --> hub["tui.RunMainMenu"]
  hub -->|MainCollect| runTUI["cli.runTUI"]
  profTUI["prof tui"] --> runTUI
  runTUI --> intent["intent.CollectIntent"]
  intent --> appSvc["app.Services.Collect.RunAuto"]
  appSvc --> engine["collect.RunAuto"]
```

**Key files**

- Hub menu: [`internal/tui/hub.go`](../internal/tui/hub.go) — Bubble Tea full-screen menu; `MainCollect` dispatches to `runTUI` from [`cli/cmd_ui.go`](../cli/cmd_ui.go).
- Survey wizard: [`cli/tui.go`](../cli/tui.go) — all collect prompts and `CollectIntent` construction.
- Intent boundary: [`internal/intent/collect.go`](../internal/intent/collect.go), [`internal/intent/kind.go`](../internal/intent/kind.go) (`RunValidated`).

## Prompt → code mapping

Each Survey step maps to a function, validation rule, and field on [`CollectIntent`](../internal/intent/collect.go):

| Prompt (as shown) | Code | `CollectIntent` field |
| --- | --- | --- |
| Select benchmarks to run | `svc.Collect.DiscoverBenchmarks(cwd)` → [`scanForBenchmarks`](../engine/collect/discovery.go) | `Benchmarks` |
| Collection filters: none / from prof.json | [`printCollectionFilterPreview`](../cli/collect_preview.go) | *(preview only — not stored on intent)* |
| Select profiles | `svc.Collect.SupportedProfiles()` | `Profiles` |
| Number of runs (count) | `strconv.Atoi` in `runTUI` | `Count` |
| Tag name | Survey input | `Tag` |

**Prompt effects**

| Prompt | Effect |
| --- | --- |
| Select benchmarks | Regex scan of `*_test.go` under cwd; skips dot-prefixed dirs, `vendor`, `bench` (legacy), `tests`, and nested `go.mod` trees |
| Collection filters line | Read-only preview via `config.Load` + `ResolveCollectionFilter`; does not block the run |
| Select profiles | Profile IDs from [`engine/tooling/catalog.go`](../engine/tooling/catalog.go) |
| Number of runs | Rejects count `< 1` in `runTUI` before intent validation |
| Tag name | Trimmed tag becomes `.prof/<tag>/` via [`workspace.TagLayout`](../internal/workspace/layout.go) |

`CollectIntent.Run` copies fields into `app.CollectAutoOptions` ([`internal/app/dto.go`](../internal/app/dto.go)) before calling `collect.RunAuto`.

### After the last prompt

1. `collect.Normalize()` trims the tag and drops empty benchmark/profile entries.
2. `intent.RunValidated(collect, svc)` calls `Validate()` then `Run()`.
3. `CollectIntent.Run` calls `svc.Collect.RunAuto` ([`internal/app/defaults.go`](../internal/app/defaults.go)), which delegates to [`collect.RunAuto`](../engine/collect/entry.go).

## Engine pipeline

Once `RunAuto` runs, the same pipeline executes for every selected benchmark. The flow below is the internal request path after your Survey answers are committed.

```mermaid
flowchart TB
  runAuto["RunAuto"] --> prep["Preparing setup + prelude warns"]
  prep --> loop["for each benchmark"]
  loop --> s1["1 Running benchmark go test + move"]
  s1 --> s2["2 Collecting profiles text + PNG"]
  s2 --> s3["3 Collecting function profiles parser + pprof"]
  s3 --> loop
  loop --> done["session.Success"]
```

### 1. Validate and prepare

[`collect.RunAuto`](../engine/collect/entry.go):

- Rejects empty benchmarks/profiles and count `< 1`.
- Loads optional `prof.json` via [`config.Load`](../internal/config/load.go). Missing config is non-fatal; collection proceeds with empty filters.
- Skips [`config.PrintAutoConfiguration`](../internal/config/load.go) on an interactive TTY (options were already confirmed in Survey).
- On an interactive TTY, runs a **Preparing** stage ([`PhasePrepare`](../internal/termui/progress.go)) that creates the tag layout and emits prelude warnings (missing `prof.json`, Graphviz unavailable notice) indented under that stage. Non-TTY keeps separate `slog.Info` lines and runs [`setupDirectories`](../engine/collect/layout.go) before the benchmark loop.

### 2. Create output layout

[`setupDirectories`](../engine/collect/layout.go) (inside **Preparing** on TTY, or before config print on non-TTY):

- Resolves `.prof/<tag>/` with [`workspace.CleanOrCreateTag`](../internal/workspace/tag.go).
- Creates `profiles/<benchmark>/`, `measurements/<benchmark>/`, `hotspots/<benchmark>/`, `source_lines/<profile>/<benchmark>/`, and `notes.txt`.

### 3–5. Per-benchmark progress (TTY)

[`runBenchAndGetProfiles`](../engine/collect/pipeline.go) orchestrates **three user-visible steps** per benchmark via [`termui.Session`](../internal/termui/progress.go). Artifact moves after `go test` are part of step 1 (no separate spinner).

| Step | Label (TTY stderr) | Internal work |
| --- | --- | --- |
| 1 | `Running benchmark 1/2: BenchmarkX (count=5)…` | [`runBenchmark`](../engine/collect/gotest.go): `go test`, write `measurements/.../run.txt`, [`moveProfileFiles`](../engine/collect/artifacts.go) |
| 2 | `Collecting profiles for BenchmarkX (cpu, memory)…` | [`processProfiles`](../engine/collect/profiles.go): hotspots + call graphs |
| 3 | `Collecting function profiles for BenchmarkX…` | [`collectProfileFunctions`](../engine/collect/pipeline.go): parser + per-function `pprof -list` |

On an interactive TTY after Survey:

- **Stderr:** a **persistent stage log** — each step shows a spinner while running, then a `✓` line that stays on screen; the next step appears below. Warnings from `Session.Warn` print indented (`    warning: …`) under the **active** stage and remain after that stage completes.
- **Stdout:** stays clean for Survey/hub.
- **No** per-function `Collected function` lines or stage `slog.Info` spam.
- One faint success line (`Session.Success`) after all benchmarks finish.

Example (two benchmarks, no Graphviz):

```text
✓ Preparing
    warning: No prof.json found; …
    warning: Graphviz not found; …
✓ Running benchmark 1/2: BenchmarkStringProcessor (count=5)
✓ Collecting profiles for BenchmarkStringProcessor (cpu, memory)
    warning: PNG skipped for BenchmarkStringProcessor/cpu: …
✓ Collecting function profiles for BenchmarkStringProcessor
⠋ Running benchmark 2/2: BenchmarkFibonacci (count=5)…
```

Non-TTY (CI, piped `prof auto`): no spinners; stage `slog.Info` / `slog.Warn` unchanged; success still logged via `Session.Success` → `slog.Info` for [`tests/run.go`](../tests/run.go).

Recoverable issues on TTY route through `Session.Warn` under the active stage (missing profile binary and skipped PNG under **Collecting profiles**; per-function list skip under **Collecting function profiles**; prelude issues under **Preparing**).

#### Step 1 — Run benchmark (`go test` + move)

For `BenchmarkMatrixMultiplication`, [`runBenchmark`](../engine/collect/gotest.go):

- Locates the package directory containing the benchmark function.
- Builds `go test -run=^$ -bench=^BenchmarkMatrixMultiplication$ -benchmem -count=5` plus profile flags from the tooling catalog (`cpu`, `memory`).
- Runs the command in the benchmark package directory via [`tooling.Runner`](../engine/tooling/runner.go).
- Writes combined benchmark output to `measurements/<benchmark>/run.txt`; moves profile binaries (`.out`) into `.prof/baseline/profiles/BenchmarkMatrixMultiplication/`. Failures return combined output in the error.

#### Step 2 — Process profiles

[`processProfiles`](../engine/collect/profiles.go) runs per profile (`cpu`, then `memory`):

| Step | Output | Notes for this example |
| --- | --- | --- |
| Stat binary | — | Missing `.out` logs a warning and skips that profile instead of failing |
| Hotspot summary | `hotspots/.../cpu.txt` (and `memory.txt`) | Via `go tool pprof -top` |
| Call tree | `call_trees/.../cpu.txt` and `cpu.json` | Via `go tool pprof -tree` plus in-process call-graph JSON |
| PNG | `call_graphs/<profile>/.../cpu.png` | PNG failure logs a warning; run still succeeds if hotspot summaries were produced |

Resolved function filters for each benchmark come from `config.ResolveCollectionFilter` (same rules previewed during the Survey step).

#### Step 3 — Per-function extracts

[`collectProfileFunctions`](../engine/collect/pipeline.go):

- For each successfully processed profile, [`parser.GetFunctionListEntriesV2`](../parser/) reads the binary and applies config filters.
- `go tool pprof -list` output is written under `source_lines/cpu/BenchmarkMatrixMultiplication/` and `source_lines/memory/BenchmarkMatrixMultiplication/`.

When all benchmarks finish, prof logs collection success and returns.

## Output layout (example)

All paths come from [`workspace.TagLayout`](../internal/workspace/layout.go). For the running example:

```text
.prof/
└── Baseline/
    ├── notes.txt
    ├── profiles/BenchmarkMatrixMultiplication/
    │   ├── cpu.out
    │   └── memory.out
    ├── measurements/BenchmarkMatrixMultiplication/
    │   └── run.txt
    ├── hotspots/BenchmarkMatrixMultiplication/
    │   ├── cpu.txt
    │   └── memory.txt
    ├── call_trees/BenchmarkMatrixMultiplication/
    │   ├── cpu.txt
    │   ├── cpu.json
    │   ├── memory.txt
    │   └── memory.json
    ├── source_lines/cpu/BenchmarkMatrixMultiplication/
    │   └── <function>.txt
    └── source_lines/memory/BenchmarkMatrixMultiplication/
        └── <function>.txt
```

PNG files, when generated, live under `call_graphs/<profile>/<benchmark>/`.

## Where to change behavior

| You want to change… | Start here |
| --- | --- |
| Survey prompts or defaults | [`cli/tui.go`](../cli/tui.go), [`cli/collect_preview.go`](../cli/collect_preview.go) |
| Collect progress UI (TTY vs non-TTY) | [`internal/termui/progress.go`](../internal/termui/progress.go), [`engine/collect/pipeline.go`](../engine/collect/pipeline.go) |
| Hub menu labels or actions | [`internal/tui/hub.go`](../internal/tui/hub.go), [`cli/cmd_ui.go`](../cli/cmd_ui.go) |
| Intent validation rules | [`internal/intent/collect.go`](../internal/intent/collect.go) |
| Benchmark discovery rules | [`engine/collect/discovery.go`](../engine/collect/discovery.go) |
| `go test` argv or profile flags | [`engine/collect/gotest.go`](../engine/collect/gotest.go), [`engine/tooling/catalog.go`](../engine/tooling/catalog.go) |
| Artifact paths or tag lifecycle | [`internal/workspace/layout.go`](../internal/workspace/layout.go), [`engine/collect/layout.go`](../engine/collect/layout.go) |
| Missing profile / PNG handling | [`engine/collect/profiles.go`](../engine/collect/profiles.go) |
| Per-function file list and filters | [`parser/`](../parser/), [`internal/config/filter.go`](../internal/config/filter.go) |

## Layering

`cli`, `internal/tui`, and `internal/intent` never import `engine/*` directly. They pass DTOs through [`internal/app`](../internal/app) only ([CODEBASE_DESIGN.md](../CODEBASE_DESIGN.md) layering rule). The interactive path uses `CollectIntent`; the flag path uses [`cli/cmd_collect.go`](../cli/cmd_collect.go) — both converge on `collect.RunAuto`.

## Common failure points

| Symptom | Layer | Code / flag |
| --- | --- | --- |
| No benchmarks in multi-select | Discovery | [`scanForBenchmarks`](../engine/collect/discovery.go) — empty result errors in `runTUI` |
| Invalid count | CLI | `runTUI` before intent; `CollectIntent.Validate` |
| Missing profile binary after bench | Engine | Warn and skip profile ([`profiles.go`](../engine/collect/profiles.go)); fails only if zero profiles processed |
| PNG / Graphviz missing | Engine | Prelude notice in [`entry.go`](../engine/collect/entry.go); per-profile PNG failure warns in [`profiles.go`](../engine/collect/profiles.go) |
| Tag dir not empty | Workspace | [`CleanOrCreateTag`](../internal/workspace/tag.go) during `setupDirectories` |

See [CODEBASE_DESIGN.md — Edge-case catalog](../CODEBASE_DESIGN.md#edge-case-catalog) for the full contributor table.

## See also

- [CODEBASE_DESIGN.md](../CODEBASE_DESIGN.md) — package map, invariants, edge-case catalog
- [prof_web_doc/docs/tui.md](../prof_web_doc/docs/tui.md) — user-facing UI and TUI guide
- [prof_web_doc/docs/collect.md](../prof_web_doc/docs/collect.md) — `prof auto` / `prof manual` flags and artifact reference
- [docs/agents/README.md](./agents/README.md) — agent playbooks
