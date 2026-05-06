# Working directory and paths

This page explains **where Prof reads your module**, **where it writes `bench/`**, and **how paths are laid out** under each tag so compare and tools can find data.

## Before you begin

- You have a `go.mod` at the directory you treat as the **module root** (usually the repo root for the module).
- You run Prof with that directory as your **current working directory** (`cwd`), the same way you run `go test ./...` for that module.

## What Prof uses from cwd

Prof uses your **current working directory** to find the Go module and to write `bench/`. Run from the same place you run `go test` for that module (usually the [module root](index.md#terminology)).

- **Benchmark discovery** is relative to cwd (same rules as `go test`).
- **Output** goes to `bench/<tag>/` under cwd ([terminology](index.md#terminology)).
- **`prof setup`** writes `config_template.json` next to `go.mod` at the module root—keep cwd aligned with that root when you expect that file to be found.

## Directory layout under `bench/<tag>/`

Using Prof creates a `bench/` tree next to your module, one folder per run (tag). That layout is how compare, tools, and `go tool pprof` find the same data.

| Path | What it is |
| ---- | ---------- |
| `bench/<tag>/` | One labeled run: profiles, text extracts, and optional PNGs for that tag. |
| `bench/<tag>/bin/<BenchmarkName>/` | Binary profiles (`.out`)—the durable source for `go tool pprof`. |
| `bench/<tag>/text/<BenchmarkName>/` | Human-readable profile listings (flat and optional package-grouped). |
| `bench/<tag>/<profile>_functions/<BenchmarkName>/` | Per-function extracts when configured; optional call-graph PNGs if Graphviz is installed. |
| `config_template.json` | Next to `go.mod` after `prof setup` or the template action in `prof ui`; filters which functions get extra extracts. |

Details on collection flags and behavior: [Collect profiling data](collect.md). Configuration keys: [Configure collection](configure.md).

## Testing / verify

From the module root, after a successful collect, you should see a new directory `bench/<your-tag>/` with at least `bin/<BenchmarkName>/` and `text/<BenchmarkName>/` populated for the profiles you enabled.

If `bench/` never appears, see [Troubleshooting](troubleshooting.md#wrong-directory-or-no-bench-folder).

## Next steps

- [Collect profiling data](collect.md)
- [Configure collection](configure.md)

## Related

- [Quickstart](quickstart.md) · [CLI reference](cli-reference.md)
