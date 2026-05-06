# Configure collection

This guide explains `config_template.json`: how to generate it with `prof setup`, how `function_collection_filter` selects per-function extracts, and where `ci_config` fits for regression rules in CI.

## Before you begin

- You run commands from the module root (next to `go.mod`).
- You understand what a tag and benchmark name are ([Home](index.md#terminology)).

## Create the template

```bash
prof setup
```

Or in `prof ui`, choose Create configuration template. That writes `config_template.json` next to `go.mod`.

## `function_collection_filter`

Controls per-function text extracts. Keys are benchmark names (`prof auto`) or profile file base names without extension (`prof manual`).

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

| Field | Type | Required | Description |
| ----- | ---- | --------- | ----------- |
| Key (benchmark) | string | Yes | Benchmark name, or `"*"` for all benchmarks (`prof auto`). For `prof manual`, use the file stem (for example `BenchmarkGenPool_cpu` for `BenchmarkGenPool_cpu.out`). |
| `include_prefixes` | array of string | No | If set, only functions whose full name starts with one of these prefixes. |
| `ignore_functions` | array of string | No | Short names excluded even when prefixes match. |

For `prof manual`, keys are profile file base names, not the Go benchmark name.

## `ci_config` (regression and CI behavior)

Prof can read regression thresholds, ignores, noise floors, and per-benchmark caps from `ci_config` inside the same JSON file. That keeps CI policy next to your extract filters.

Do not duplicate the full schema here; it evolves with releases. Use the canonical document on GitHub:

- [CI/CD configuration](https://github.com/AlexsanderHamir/prof/blob/main/docs/cicd_configuration.md) (full JSON schema, field semantics, and GitHub Actions examples)

This site’s [CI and regressions](ci.md) explains how `prof track` flags interact with `ci_config` at runtime.

## Testing / verify

After `prof setup`, confirm `config_template.json` exists beside `go.mod`. Run a small `prof auto` collect and check that `<profile>_functions/<BenchmarkName>/` contains files when your filter matches hot symbols.

## Next steps

- [Compare runs](compare.md)
- [CI and regressions](ci.md)

## Related

- [Collect profiling data](collect.md) · [CLI reference](cli-reference.md)
