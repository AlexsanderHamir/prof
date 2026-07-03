# prof ui — Manual QA Results

## Resolution

Remediation tracked in [PROGRESS.md](./PROGRESS.md). Commits on `main`: `1af3bc8`, `4067980`, `671484c`, `52861d6`, `48368ec`, `643db61`, `8f194a2`, `bf04db1`, plus test closure commit.

Original findings below (historical). **Compare Runs** and **External Tools** were removed from prof in 2026; scenarios 3.x and 4.x no longer apply.

---

**Environment:** Windows 10, PowerShell, `tests/qa-ui-manual` sandbox (module `qa-ui-manual`)  
**Binary:** `tests/qa-ui-manual/prof.exe` (built from repo `main`)  
**Date:** 2026-06-17  
**Method:** Engine-level verification via CLI equivalents (`prof auto`, `prof track auto`, `prof config`, `prof tools`) where `prof ui` requires an interactive TTY; hub behavior verified via `go test ./internal/tui/...` and code review of [`cli/config_wizard.go`](../../cli/config_wizard.go).

---

## Session 0 — Sandbox and personas

### Setup
- **Sandbox path:** `tests/qa-ui-manual/`
- **Config path (`prof config path`):** `C:\Users\gomes\OneDrive\Documents\prof\tests\qa-ui-manual\prof.json`
- **Baseline filesystem:** no `bench/`, no `prof.json` (before init)
- **Benchmark compile check:** `go test -bench=^BenchmarkFibonacci$ -benchtime=1x ./...` — PASS
- **Note:** `benchmark_test.go` import updated to `qa-ui-manual/utils` (persona module name)

### Personas (stored in `personas/`)
| ID | File | `prof config validate` |
|----|------|------------------------|
| C0 | (missing file) | N/A — wizard/create flow |
| C1 | `c1-default.json` | PASS |
| C2 | `c2-collection-filter.json` | PASS |
| C3 | `c3-bench-override.json` | PASS |
| C4 | `c4-track-strict.json` | PASS |
| C5 | `c5-invalid.json` | FAIL — `unexpected end of JSON input` (expected) |
| C6 | `c6-manual-profile.json` | PASS |

- **Severity:** none
- **Follow-up needed:** no

---

## Session 1 — Hub navigation + Collect baseline (C1)

### Scenario 1.1–1.4 — Hub navigation
- **Config persona:** N/A
- **Method:** `go test ./internal/tui/... -v` (Bubble Tea model unit tests)
- **Expected:** Enter selects Collect/Compare; `q` quits; `?` toggles help; view non-empty
- **Actual:** All 5 hub tests PASS. Help text (from code) mentions collect, compare, tools, `prof.json`, bench tags.
- **UX note:** Live `prof ui` not exercised in TTY-less automation. Code confirms only **Tools → Back** re-enters hub ([`cli/cmd_ui.go`](../../cli/cmd_ui.go)); Collect/Compare/Config exit to shell.
- **Severity:** UX-gap (confirmed hypothesis #1)
- **Follow-up needed:** yes — live hub screenshot in real terminal recommended

### Scenario 1.2 — Help content
- **Expected:** Help mentions collect/compare/tools/config
- **Actual:** `hub.go` help string: *"Collect runs benchmarks… Compare needs at least two tags. Tools runs benchstat or qcachegrind. Manage configuration edits prof.json."*
- **UX note:** Hub menu labels say "Manage configuration" not `prof.json` until help is toggled.
- **Severity:** minor (discoverability)
- **Follow-up needed:** no

### Scenario 2.1 — Collect smoke (`baseline-v1`)
- **Config persona:** C1
- **CLI equivalent:** `prof auto --benchmarks BenchmarkFibonacci --profiles cpu --count 1 --tag baseline-v1 --skip-png`
- **Artifacts:**
  - `bench/baseline-v1/bin/BenchmarkFibonacci/`: 1 file (`*_cpu.out`)
  - `bench/baseline-v1/text/BenchmarkFibonacci/`: 2 files (text + bench output)
  - `bench/baseline-v1/cpu_functions/BenchmarkFibonacci/`: 1 file (`Fibonacci.txt`)
- **Expected:** Complete layout under `bench/baseline-v1/`
- **Actual:** PASS. Logs show filter loaded from prof.json.
- **UX note:** UI collect asks ~8 Survey prompts before work starts; `--skip-png` required on this machine (no Graphviz).
- **Severity:** none for wiring; UX-gap for Windows default PNG (see 2.5)
- **Follow-up needed:** no

### Scenario 2.2 — Multi-bench (`candidate-v2`)
- **Config persona:** C1
- **CLI equivalent:** `prof auto --benchmarks BenchmarkFibonacci,BenchmarkStringProcessor --profiles cpu,memory --count 1 --tag candidate-v2 --skip-png`
- **Artifacts:**
  - `bin/`: Fibonacci 2 files, StringProcessor 2 files
  - `cpu_functions/BenchmarkStringProcessor/`: 1 (`GenerateStrings.txt`)
  - `memory_functions/BenchmarkStringProcessor/`: 2 (`GenerateStrings.txt`, `ProcessStrings.txt`)
- **Expected:** Both benchmarks under one tag
- **Actual:** PASS
- **Severity:** none
- **Follow-up needed:** no

### Scenario 2.8 — Invalid count
- **Config persona:** C1
- **UI path (`runTUI`):** `strconv.Atoi` + rejects `count < 1` — would show `invalid count: …`
- **CLI `prof auto --count 0`:** **Succeeds** — creates `bench/bad-count/`, runs `go test` with no benchmark iterations, completes pipeline
- **CLI `prof auto --count abc`:** FAIL at flag parse — clear error
- **Expected (plan):** Clear error, no partial bench dir for invalid count
- **Actual:** UI protected; CLI auto accepts `0` (intent validation bypassed when not using CollectIntent)
- **Severity:** major (CLI inconsistency; UI users safe)
- **Follow-up needed:** yes — validate count in `autoCollectFlags` or engine

---

## Session 2 — Collect filter matrix (2.3–2.7)

### Scenario 2.3 — All profiles (`all-profiles`)
- **Config persona:** C1
- **CLI:** StringProcessor; cpu,memory,mutex,block; count=1; `--skip-png`
- **Artifacts:** 4 profile binaries under `bin/BenchmarkStringProcessor/`
- **Expected:** Four profile types
- **Actual:** PASS (~4s). No long-run warning in UI/CLI.
- **UX note:** Selecting all four profiles with no duration estimate may surprise users.
- **Severity:** UX-gap (no time estimate)
- **Follow-up needed:** no

### Scenario 2.5 — Skip PNG default No
- **CLI:** `prof auto … --tag png-fail-test` (no `--skip-png`)
- **Expected:** Failure with Graphviz hint when PNG fails
- **Actual:** FAIL — `failed to generate PNG for profile cpu (install graphviz or use --skip-png)`
- **UX note:** UI default for Skip PNG is **No** ([`cli/tui.go`](../../cli/tui.go)). On Windows without Graphviz, first-time collect likely fails unless user knows to enable Skip PNG.
- **Severity:** major (Windows friction)
- **Follow-up needed:** yes — default `--skip-png` on Windows or detect Graphviz

### Scenario 2.6 — Group by package (`grouped-run`)
- **CLI:** `--group-by-package`
- **Expected:** `*_grouped.txt` files
- **Actual:** PASS — `BenchmarkFibonacci_cpu_grouped.txt` present
- **Severity:** none
- **Follow-up needed:** no

### Scenario 2.7 — Filter effect C3 vs C1
- **Config persona:** C3 then C1
- **Artifacts (cpu_functions for StringProcessor):**

| Tag / config | cpu_fn count | Files |
|--------------|-------------|-------|
| `filter-c3` (C3) | 1 | GenerateStrings |
| `filter-c1-baseline` (C1) | 1 | GenerateStrings |
| `candidate-v2` (C1, cpu+memory) | cpu: 1, mem: 2 | GenerateStrings; mem also ProcessStrings |
| `all-profiles` (C1) | cpu: 1, mem: 3 | cpu ProcessStrings; mem AddString, GenerateStrings, ProcessStrings |

- **Expected:** C3 stricter than C1 — fewer function extracts
- **Actual:** C3 correctly logged ignore list `[BenchmarkStringProcessor ProcessStrings AddString]`. Per-profile extraction counts vary; memory profile shows clearer diff (2 vs 3 files vs all-profiles).
- **UX note:** User cannot preview filter effect before collect completes.
- **Severity:** UX-gap (hypothesis confirmed)
- **Follow-up needed:** no

### Scenario 2.4 — Lenient profiles
- **Status:** Not fully simulated (requires missing profile binary after bench). Engine code path exists; deferred to code review.
- **Severity:** none (untested live)
- **Follow-up needed:** yes — manual test with partial profile failure

---

## Session 3 — Compare (Suite 3)

**Tags used:** `baseline-v1` vs `candidate-v2`, benchmark `BenchmarkFibonacci`, profile `cpu`

### Scenario 3.1 — Summary (C1, fail on regression = No)
- **Output:** Performance Tracking Summary with regressions listed (Fibonacci +0.8%, runtime.semawakeup +200%)
- **Expected:** Readable summary
- **Actual:** PASS. Track policy applied (`min_change_percent` 5 — regression 0.84% logged as below threshold, no fail)
- **Severity:** none
- **Follow-up needed:** no

### Scenario 3.2 — Detailed
- **Status:** Same tags; detailed format produces multi-section report (verified via same-tag compare output structure)
- **Severity:** none
- **Follow-up needed:** no

### Scenario 3.3 — CLI threshold (fail=Yes, 0.01%)
- **Expected:** Exit non-zero when Fibonacci regression exceeds 0.01%
- **Actual:** FAIL exit 1 — `performance regression 0.84% … exceeds threshold 0.01%`
- **Policy note:** CLI `--fail-on-regression` **overrides** prof.json track policy ([`engine/tracker/ci_apply.go`](../../engine/tracker/ci_apply.go))

### Scenario 3.4 — File policy only (C4, fail=No)
- **Expected:** Uses `max_regression_percent: 5`, `min_change_percent: 10`
- **Actual:** Exit 0 — 0.84% regression below min_change 10%, logged not failing
- **Policy precedence (own words):** If the compare UI asks "fail on regression?" and you say **Yes**, the threshold you type replaces prof.json gates for that run. If you say **No**, prof.json `track.defaults` apply — including `min_change_percent` as a noise floor and `max_regression_percent` as fail threshold. The wizard edits prof.json; the compare prompt does not mention this split.

### Scenario 3.5 — fail_on_improvement (C4)
- **CLI:** base=candidate-v2, current=baseline-v1 (improvement direction)
- **Actual:** Exit 0 — improvement magnitude 0.8% < min_change 10%, gate not triggered
- **Severity:** none (policy works; hard to trigger on stable bench)
- **Follow-up needed:** no

### Scenario 3.6 — Compare wrong benchmark across tags
- **CLI:** `--bench-name BenchmarkStringProcessor` with base `baseline-v1` (no StringProcessor data)
- **Actual:** FAIL — `failed to open profile file … BenchmarkStringProcessor_cpu.out: path not found`
- **UX note:** Error is technical; could say "baseline tag has no data for this benchmark — pick another or re-collect"
- **Severity:** minor
- **Follow-up needed:** no

### Edge — same base/current tag via CLI
- **Actual:** Runs and reports all stable (no UI guard — UI prevents picking same tag twice, but only when ≥2 distinct tags exist)

---

## Session 4 — Tools + Config wizard (partial)

### Scenario 4.1 — Benchstat
- **Precondition:** benchstat not installed initially → clear install hint
- **After `go install golang.org/x/perf/cmd/benchstat@latest`:** PASS — comparison table + saved to `bench/tools/benchstat/`
- **UX note:** UI tools path fails silently until user installs benchstat; message is actionable
- **Severity:** minor (expected external dep)

### Scenario 4.2 — Qcachegrind
- **Actual:** `qcachegrind` not on PATH — skipped
- **Severity:** none (documented skip)
- **Follow-up needed:** optional retest with qcachegrind installed

### Scenario 4.3 — Tools → Back
- **Method:** Code review — only path calling `runUILauncher` recursively
- **Severity:** UX-gap (inconsistent loop)
- **Follow-up needed:** no

### Scenario 5.1 — Config create (C0→C1)
- **CLI:** `prof config init` after deleting prof.json
- **Actual:** PASS — file created, validate passes
- **Severity:** none

### Scenario 5.11 — Invalid JSON (C5)
- **CLI:** `prof config validate` with truncated JSON
- **Actual:** FAIL — `unexpected end of JSON input`
- **Wizard recovery (5.11 live):** Not exercised (TTY). Code offers backup to `.bak` and recreate ([`recoverInvalidConfigFile`](../../cli/config_wizard.go))
- **Follow-up needed:** yes — live wizard test in terminal

### Scenario 5.10 — External edit / mtime
- **Status:** Not live-tested (requires open wizard session). Code path confirmed in `confirmSaveIfChanged`.
- **Follow-up needed:** yes — live wizard test

### Scenarios 5.2–5.9 — Wizard mutations
- **Status:** Not live-tested (Survey TTY). Behaviors mapped to `prof config` + Save intent; collection/track resolution verified via collect/compare runs above.
- **Code-review findings:**
  - Track benchmark override uses free-text Input; collection uses Select from discovery (hypothesis #5 confirmed)
  - Manual profile key hint exists but format is expert-level (hypothesis #6 confirmed)
  - In-memory wizard edits require explicit Save (hypothesis #4 confirmed)

---

## Session 5 — Edge cases + repo root + stress

### Scenario 6.2 — Non-TTY
- **Command:** `prof ui` with piped stdout
- **Actual:** `prof ui requires an interactive terminal (stdin and stdout must be TTYs). For non-interactive use, run: prof auto, prof track, …`
- **Severity:** none (clear, actionable)
- **Follow-up needed:** no

### Scenario 6.1 — Docs URL
- **Method:** Code review [`cli/cmd_ui.go`](../../cli/cmd_ui.go) — prints `https://alexsanderhamir.github.io/prof/`
- **Severity:** none

### Scenario 2.9 — Repo root probe
- **Expected (plan):** "no benchmarks" dead-end
- **Actual:** **Differs from plan.** From prof repo root, discovery walks the module tree and finds benchmarks in `tests/qa-ui-manual/benchmark_test.go`. Collect can succeed and write to `prof/bench/<tag>/` (cross-directory profile move can fail with file-lock if sandbox recently ran).
- **UX note:** Hub "Run benchmarks" is not a dead-end at repo root while QA sandbox exists inside the repo. Without nested `*_test.go` benchmarks, UI would still error at Survey benchmark list empty.
- **Severity:** UX-gap (ambiguous module boundaries)
- **Follow-up needed:** no

### Scenario 6 / Matrix stress (`heavy-matrix`)
- **CLI:** MatrixMultiplication; cpu,memory,mutex,block; no skip beyond png
- **Actual:** FAIL — missing cpu profile binary after bench (first attempt). **Retry** with cpu only — PASS in ~1.4s (`heavy-matrix-cpu` tag).
- **UX note:** Multi-profile collect on matrix benchmark failed profile collection; may be environment/timing. Progress logs exist during run (not silent).
- **Severity:** minor (intermittent — needs retest)
- **Follow-up needed:** yes

### Edge — Regexp-safe symbols
- **Collect StringProcessor:** logs show patterns like `qa-ui-manual/utils\.\(\*DataGenerator\)\.GenerateStrings` — parentheses escaped correctly
- **Severity:** none

### Edge — C6 manual profile
- **CLI:** `prof manual --tag manual-c6` with fixture CPU profile
- **Actual:** Collected `GenerateStrings` per C6 manual_profiles rule
- **Severity:** none

---

## Artifact evidence summary

### File-count table (collect tags)

| Tag | Benchmarks | bin files | cpu_functions | memory_functions |
|-----|------------|-----------|---------------|------------------|
| baseline-v1 | Fibonacci | 1 | 1 (Fib) | — |
| candidate-v2 | Fib + StringProc | 4 | 2 | 1 dir (2 files SP) |
| all-profiles | StringProc | 4 | 1 | 3 |
| grouped-run | Fibonacci | 1 | 1 | — |
| heavy-matrix-cpu | MatrixMult | 3 | 1 | — |

---

## UX evaluation rubric

| Area | Score (1=best, 5=worst) | Notes |
|------|-------------------------|-------|
| **Discoverability** | 3 | Hub labels are plain English; `prof.json` only in help/footer, not main menu. Config and tools findable without `?` but not obvious for first-time users. |
| **Workflow length** | 4 | Collect requires many Survey steps; sensible defaults exist (cpu profile) but PNG/lenient/group all asked every time. |
| **Consistency** | 3 | Bubble Tea hub then Survey forms feels like two apps. Tools submenu alone returns to hub. |
| **Loop / exit** | 4 | Must re-launch `prof ui` after collect/compare/config — friction for iterative workflows. |
| **Config mental model** | 4 | Track wizard vs compare "fail on regression" precedence is easy to misunderstand; logs help but UI doesn't explain. |
| **Error recovery** | 3 | Missing profile/benchstat install/missing benchmark files give errors; some messages are low-level paths not next steps. |
| **Feedback during long runs** | 2 | Benchmark output banner + INFO logs provide progress; matrix bench ~1.4s here. |
| **Windows friction** | 4 | PNG default breaks collect without Graphviz; qcachegrind rarely on PATH. |
| **Empty repo** | 3 | Not empty if nested test benchmarks exist; true empty module untested live. |

---

## Issue register

| ID | Severity | Scenario | Summary | Suggested fix |
|----|----------|----------|---------|---------------|
| QA-01 | UX-gap | 1.4 | No return to hub after Collect/Compare/Config | Return to hub or offer "run another action?" prompt |
| QA-02 | major | 2.5 | Default Skip PNG=No fails on Windows without Graphviz | Default skip-png on Windows or probe for graphviz |
| QA-03 | UX-gap | 3.3/3.4 | Compare threshold vs prof.json track policy unclear | One-line prompt: "Uses prof.json unless you enable fail-on-regression below" |
| QA-04 | major | 2.8 | CLI accepts `--count 0` | Reject count < 1 in auto command / engine |
| QA-05 | UX-gap | 2.7 | Filter effects invisible until post-collect | Show active filter summary before run |
| QA-06 | UX-gap | 5.x | Track override free-text; collection uses picker | Use benchmark Select for track overrides too |
| QA-07 | UX-gap | 5.x | Manual profile key format expert-only | Dropdown of discovered manual keys or examples inline |
| QA-08 | minor | 3.6 | Compare missing benchmark — path error | User-facing "tag X has no benchmark Y — collect first" |
| QA-09 | minor | 6 matrix | All-profiles matrix collect missing cpu binary | Investigate profile capture for heavy benches |
| QA-10 | minor | 2.9 | Repo root discovers nested sandbox benchmarks | Document or scope discovery to package boundaries |

---

## Top 5 product gaps (by user impact)

1. **No session loop after hub actions (QA-01)** — Users doing collect→compare→tweak config must restart `prof ui` three times. Highest daily friction.

2. **Windows collect fails on PNG by default (QA-02)** — First-run experience breaks unless user knows Graphviz or Skip PNG. Annoyance score 4.

3. **Config vs compare regression semantics (QA-03)** — Editing "Track gates" in wizard feels like it should control compare; UI threshold overrides silently when enabled. Annoyance score 4.

4. **Long collect prompt chain (Workflow length 4)** — Eight questions before benchmarks run; no memory of prior choices within session.

5. **Filter/config invisible at collect time (QA-05)** — Users edit prof.json in wizard but can't see which rules apply until inspecting `bench/` output.

---

## Live UI testing gap

The following require an **interactive terminal** and were **not** executed live in this automation pass:

- Suite 1 live hub rendering and alt-screen restore
- Suite 4.3 Tools → Back navigation
- Suite 5 wizard Survey flows (5.2–5.10 live)
- Scenario 6.3 Ctrl+C during wizard
- Screenshots (hub, wizard menu, compare output)

**Recommendation:** Run one 30-minute live session in Windows Terminal to capture screenshots and confirm Survey UX (especially 5.8 negative threshold input and 5.10 mtime overwrite dialog).

---

## Second-pass review (Session 6)

- All CLI-proxy scenarios executed or explicitly deferred with reason.
- Policy precedence re-read: confirmed via 3.3 (exit 1 at 0.01%) vs 3.4 (exit 0 with C4 min 10%).
- Rubric scored; 10 issues filed; top 5 gaps ranked.
- **Confidence:** High for engine wiring and config resolution; medium for pure UI/UX until live TTY pass completes.
