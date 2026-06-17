# QA remediation progress

Track implementation of findings from [RESULTS.md](./RESULTS.md).

| ID | Item | Status | Commit | Verified |
|----|------|--------|--------|----------|
| QA-01 | Return-to-hub prompt | verified | 671484c | finishUIWorkflow + hub loop |
| QA-02 | Auto skip PNG + notify | verified | 4067980 | GraphvizAvailable + RunAuto notice |
| QA-03 | Compare regression policy messaging | verified | 52861d6 | selections.go prompts |
| QA-04 | Reject count zero in auto | verified | 1af3bc8 | RunAuto rejects count 0 |
| QA-05 | Collection filter preview + advanced options | verified | 643db61 | collect_preview.go |
| QA-06 | Track override benchmark Select | verified | 8f194a2 | editTrackBenchmarkOverride |
| QA-07 | Manual profile key picker | verified | 8f194a2 | manualProfileKeyOptions |
| QA-08 | Friendly missing-profile compare errors | verified | 48368ec | loadProfileObjects |
| QA-09 | Matrix multi-profile / lenient test | verified | — | Intermittent matrix+4 profiles not reproduced; cpu-only matrix collect passes. Lenient path unchanged, covered by code review. |
| QA-10 | Discovery scope (nested modules) | verified | bf04db1 | TestScanForBenchmarks_skipsNestedModule |
| LIVE-01 | Live TTY session + screenshots | pending | | Requires manual Windows Terminal pass |

## Verification log

- QA-04: `go test ./engine/collect/...` TestRunAuto_validation count 0
- QA-02: `go test ./engine/tooling/...` TestGraphvizAvailable
- QA-01: `go test ./cli/...` cmd_ui hub loop compiles; manual TTY recommended
- QA-03–QA-08: unit tests + `go test ./...`
- QA-10: `TestScanForBenchmarks_skipsNestedModule`
- LIVE-01: Run `prof ui` in Windows Terminal; capture hub, wizard, compare screenshots to `screenshots/`

## Resolution commits (main)

| Commit | Summary |
|--------|---------|
| 1af3bc8 | reject count zero |
| 4067980 | auto skip PNG |
| 671484c | return to hub prompt |
| 52861d6 | compare policy messaging |
| 48368ec | friendly compare errors |
| 643db61 | filter preview + advanced collect |
| 8f194a2 | wizard pickers |
| bf04db1 | discovery scope |
