# Working directory and paths

Prof uses the **current working directory** to find the Go module and to write `bench/`. Run from the same place you run `go test` for that module (usually the [module root](index.md#terminology)).

- **Benchmark discovery:** module-relative to cwd.
- **Output:** `bench/<tag>/` under cwd ([terminology](index.md#terminology)).
- **Config:** `prof setup` writes `config_template.json` next to `go.mod` at the module root—keep cwd aligned with that root.

## Next article

[Collect profiling data](collect.md)
