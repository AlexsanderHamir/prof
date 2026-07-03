# QA remediation progress

Track implementation of findings from [RESULTS.md](./RESULTS.md).

| ID | Item | Status | Commit | Verified |
|----|------|--------|--------|----------|
| QA-01 | Return-to-hub prompt | verified | 671484c | finishUIWorkflow + hub loop |
| QA-02 | Warn-and-continue on PNG / Graphviz missing | verified | cad1aa2, 6f4ee20 | Engine always warns; prelude notice in RunAuto prepare |
| QA-03 | Compare regression policy messaging | verified | 52861d6 | selections.go prompts |
| QA-04 | Reject count zero in auto | verified | 1af3bc8 | RunAuto rejects count 0 |
| QA-05 | Collection filter preview | verified | 643db61, fc0294f | collect_preview.go (output-options prompt removed in fc0294f) |
| QA-06 | Track override benchmark Select | verified | 8f194a2 | editTrackBenchmarkOverride |
| QA-07 | Manual profile key picker | verified | 8f194a2 | manualProfileKeyOptions |
| QA-08 | Friendly missing-profile compare errors | verified | 48368ec | loadProfileObjects |
| QA-09 | Matrix multi-profile / missing-profile tolerance | verified | cad1aa2, c4339c6 | profiles_test.go + engine warn-and-continue |
| QA-10 | Discovery scope (nested modules) | verified | bf04db1 | TestScanForBenchmarks_skipsNestedModule |
| LIVE-01 | Live TTY session + screenshots | pending | | Requires manual Windows Terminal pass |

## Verification log

- QA-04: `go test ./engine/collect/...` TestRunAuto_validation count 0
- QA-02: `go test ./engine/collect/...` TestProcessProfiles_continuesOnPNGFailure; Graphviz prelude in RunAuto
- QA-01: `go test ./cli/...` cmd_ui hub loop compiles; manual TTY recommended
- QA-03–QA-08: unit tests + `go test ./...`
- QA-09: `engine/collect/profiles_test.go`
- QA-10: `TestScanForBenchmarks_skipsNestedModule`
- LIVE-01: Run `prof ui` in Windows Terminal; capture hub, wizard, compare screenshots to `screenshots/`

## Resolution commits (main + polish/random-fixes)

| Commit | Summary |
|--------|---------|
| 1af3bc8 | reject count zero |
| 4067980 | Graphviz prelude notice in RunAuto (superseded by cad1aa2 warn-and-continue) |
| 671484c | return to hub prompt |
| 52861d6 | compare policy messaging |
| 48368ec | friendly compare errors |
| 643db61 | filter preview |
| 8f194a2 | wizard pickers |
| bf04db1 | discovery scope |
| 7998f7b | test closure + QA sandbox |
| cad1aa2–6f4ee20 | remove output-options prompt; warn-and-continue collect policy |
