# Working directory and paths

This page explains where Prof reads your module, where it writes `.prof/`, and how paths are laid out under each tag so `pprof` and your own tooling can find data.

## Before you begin

- You have a `go.mod` at the directory you treat as the module root (usually the repo root for the module).
- You run Prof with that directory as your current working directory (`cwd`), the same way you run `go test ./...` for that module.

## What Prof uses from cwd

Prof uses your current working directory to find the Go module and to write `.prof/`. Run from the same place you run `go test` for that module (usually the [module root](index.md#terminology)).

- Benchmark discovery is relative to cwd (same rules as `go test`).
- Output goes to `.prof/<tag>/` under cwd ([terminology](index.md#terminology)).
- `prof config init` writes minimal `prof.json` and commented `prof.json.example` next to `go.mod` at the module root. Keep cwd aligned with that root when you expect those files to be found.

## Directory layout under `.prof/<tag>/`

Using Prof creates a `.prof/` tree next to your module, one folder per run (tag). Domains describe the data they hold (`domain/<BenchmarkName>/artifact`).

| Path | What it is |
| ---- | ---------- |
| `.prof/<tag>/` | One labeled run: profiles, measurements, hotspots, and optional extracts for that tag. |
| `.prof/<tag>/profiles/<BenchmarkName>/` | Raw pprof profile binaries (`.out`); durable source for `go tool pprof`. |
| `.prof/<tag>/measurements/<BenchmarkName>/` | `go test` benchmark run stats (`run.txt`: ns/op, allocs). |
| `.prof/<tag>/hotspots/<BenchmarkName>/` | Function-ranked stack summaries per profile (`cpu.txt`, `memory.txt`). |
| `.prof/<tag>/call_trees/<BenchmarkName>/` | Call-tree text (`pprof -tree`) per profile. |
| `.prof/<tag>/source_lines/<profile>/<BenchmarkName>/` | Per-function `pprof -list` extracts when configured. |
| `.prof/<tag>/call_graphs/<profile>/<BenchmarkName>/` | Optional Graphviz PNG call graphs when installed. |
| `.prof/<tag>/data_mapping/<BenchmarkName>/map.json` | Machine-readable index of artifacts for this benchmark (paths, semantics, top symbols, function inventory). |
| `.prof/<tag>/notes.txt` | Short tag-level note (placeholder until you edit it). |
| `prof.json` | Active config next to `go.mod` after `prof config init` or **Manage configuration** in `prof ui`. |
| `prof.json.example` | Commented reference (not loaded); copy optional sections into `prof.json`. See [Configure — generated files](configure.md#generated-files). |

Details on collection flags and behavior: [Collect profiling data](collect.md). Configuration keys: [Configure collection](configure.md).

## Configuration files { #profjson }

Both files live beside `go.mod` at the module root:

- **`prof.json`** — active config (valid JSON). Created minimal by `prof config init`; add a `collection` section as needed.
- **`prof.json.example`** — commented reference with doc links; not loaded by prof.

## Testing / verify

From the module root, after a successful collect, you should see a new directory `.prof/<your-tag>/` with at least `profiles/<BenchmarkName>/`, `measurements/<BenchmarkName>/`, and `hotspots/<BenchmarkName>/` populated for the profiles you enabled.

If `.prof/` never appears, see [Troubleshooting](troubleshooting.md#wrong-directory-or-no-prof-folder).

## Next steps

- [Collect profiling data](collect.md)
- [Configure collection](configure.md)

## Related

- [Quickstart](quickstart.md) · [CLI reference](cli-reference.md)
