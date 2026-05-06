# Troubleshooting

This page lists common failure modes when using Prof and what to change before opening an issue.

### prof ui or prof tui fails in CI or IDEs {#prof-ui-or-prof-tui-fails-in-ci-or-ides}

Menus never appear, or you see an error about stdin or stdout not being a TTY.

`prof ui`, `prof tui`, and `prof tui track` expect an interactive terminal.

Use non-interactive commands instead: `prof auto`, `prof track auto`, `prof track manual`, and `prof tools …`. See [Quickstart](quickstart.md) Path B and [CLI reference](cli-reference.md).

### Wrong directory or no bench folder {#wrong-directory-or-no-bench-folder}

Collect succeeds elsewhere, or `bench/` appears in the wrong repo.

Prof writes `bench/` under the current working directory, which must be your module root (where `go.mod` lives).

`cd` to the same directory you use for `go test`, then run Prof again. See [Working directory and paths](workspace.md).

### prof tui track says you need two tags

Message like "Need at least 2 tags to compare".

Compare needs at least two different tags under `bench/` from prior collects.

Run collect twice with two different `--tag` values (or use `prof ui` or `prof tui` twice), then compare.

### Graphviz or PNG errors {#graphviz-png-errors}

Failure when generating call-graph PNGs.

PNG generation uses Graphviz `dot` when installed; otherwise generation can error.

Install [Graphviz](https://graphviz.org/), or pass `--skip-png` on `prof auto`, or disable PNG in interactive flows. See [Collect profiling data](collect.md).

### benchstat not found {#benchstat-not-found}

`prof tools benchstat` fails because `benchstat` is missing.

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

Ensure `$(go env GOPATH)/bin` is on your `PATH`. See [Optional tools](tools.md).

### QCacheGrind not installed {#qcachegrind-not-installed}

`prof tools qcachegrind` cannot find or run QCacheGrind.

Install [QCacheGrind](https://kcachegrind.github.io/html/Home.html) (or KCachegrind) and ensure the binary is on `PATH`. See [Optional tools](tools.md).

### Invalid output format {#invalid-output-format}

Error mentioning `invalid output format`.

The value must be one of the formats listed in [CLI reference](cli-reference.md#compare-output-formats).

Pass a valid string (for example `summary` or `detailed-json`). For `prof track manual`, `--output-format` is required.

### Regression gate always passes or does not fail the build {#regression-gate-always-passes-or-does-not-fail-the-build}

You passed `--fail-on-regression` but the command still exits 0.

The CLI gate applies when both `--fail-on-regression` is set and `--regression-threshold` is greater than zero. A threshold of `0` does not activate that check (CI-only configuration may still apply if present; see [CI and regressions](ci.md)).

Set an explicit positive threshold, for example `--regression-threshold 5.0` for 5%.

### Compare shows no function changes

Log line like "No function changes detected between the two runs".

The two inputs may be identical, filters may remove differences, or the wrong benchmark or profile pair was selected.

Confirm tag names, `--bench-name`, and `--profile-type`, and that the two runs are actually different builds.

## Next steps

- [CLI reference](cli-reference.md) for flags and defaults
- [Compare runs](compare.md) for semantics of manual vs auto compare

## Related

- [Home](index.md)
