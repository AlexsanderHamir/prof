# Configure collection

You can create the template with **`prof setup`** or from **`prof ui`** → **Create configuration template** (same result).

## Create the template

From the module root:

```bash
prof setup
```

This writes `config_template.json` next to `go.mod`.

## function_collection_filter

Use this object to control which functions get per-function text extracts. Keys are **benchmark names** for `prof auto`, or **profile file base names** (no extension) for `prof manual`.

Example:

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

| Field | Description |
| ----- | ----------- |
| Key name | Benchmark function name, or `"*"` to apply one block to all benchmarks (`prof auto`). For `prof manual`, use the stem of the file (for example `BenchmarkGenPool_cpu` for `BenchmarkGenPool_cpu.out`). |
| `include_prefixes` | If non-empty, only functions whose full name starts with one of these prefixes are collected. |
| `ignore_functions` | Short names to exclude even when prefixes match. |

**Note:** For `prof manual`, map keys to each profile file’s base name, not the Go benchmark name.

## Next article

[Compare runs](compare.md)
