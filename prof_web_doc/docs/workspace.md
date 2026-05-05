# Working directory and paths

## Benchmark discovery

Prof resolves the Go module from the working directory and discovers benchmarks from that context. Run commands from the directory that matches how you normally run `go test` for that module (often the module root).

## Output location

Collected data is written under a `bench/` directory relative to the **current working directory** when you invoke `prof`.

Example:

```text
bench/<tag>/
```

Tags are arbitrary labels you pass with `--tag` (for example `baseline`, `pr-123`).

## Configuration file

`prof setup` writes `config_template.json` beside `go.mod` at the **module root**, not necessarily your shell’s current directory if they differ.

**Recommendation:** run Prof from the module root so benchmark discovery, output paths, and configuration resolution stay aligned.

## Next article

[Collect profiling data](collect.md)
