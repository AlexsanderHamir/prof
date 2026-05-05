# Configure collection

Template: **`prof setup`** or **`prof ui`** → **Create configuration template** → writes `config_template.json` next to `go.mod`.

```bash
prof setup
```

## function_collection_filter

Controls per-function text extracts. Keys: **benchmark names** (`prof auto`) or **profile file base names** without extension (`prof manual`).

```json
{
  "function_collection_filter": {
    "BenchmarkGenPool": {
      "include_prefixes": ["github.com/example/myproject"],
      "ignore_functions": ["init", "TestMain", "BenchmarkMain"]
    }
  }
}
```

| Field | Meaning |
| ----- | ------- |
| Key | Benchmark name, or `"*"` for all benchmarks (`prof auto`). For `prof manual`, use the file stem (e.g. `BenchmarkGenPool_cpu` for `BenchmarkGenPool_cpu.out`). |
| `include_prefixes` | If set, only functions whose full name starts with one of these. |
| `ignore_functions` | Short names excluded even when prefixes match. |

For `prof manual`, keys are **profile file base names**, not the Go benchmark name.

## Next article

[Compare runs](compare.md)
