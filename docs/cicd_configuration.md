# CI/CD configuration reference

Short guide (flags, gates, link here): [CI and regressions](https://alexsanderhamir.github.io/prof/ci/) in the Prof docs.

Field reference for the `track` section: [Configure collection — track](https://alexsanderhamir.github.io/prof/configure/#track).

This file is the **full** reference: `track` section in `prof.json`, ignores, thresholds, and GitHub Actions examples.

## Overview

The `track` section filters noisy functions, sets global and per-benchmark regression caps (`prof track` CLI flags win when provided), and can set `fail_on_improvement`.

Create or edit via `prof ui` → Manage configuration, or `prof config init` (writes minimal `prof.json` plus commented `prof.json.example`; add a `track` section to enable config-only gates).

## Configuration Structure

Track policy lives in `prof.json` under the `track` section:

```json
{
  "version": 1,
  "collection": {
    "defaults": {},
    "benchmarks": {}
  },
  "track": {
    "defaults": {
      "ignore_prefixes": ["runtime."],
      "max_regression_percent": 15.0
    },
    "benchmarks": {
      "BenchmarkName": {
        "max_regression_percent": 10.0
      }
    }
  }
}
```

## Global Configuration (track.defaults)

Global settings apply to all benchmarks unless overridden by benchmark-specific settings:

```json
"defaults": {
  "ignore_functions": [
    "runtime.gcBgMarkWorker",
    "runtime.systemstack",
    "testing.(*B).ResetTimer"
  ],
  "ignore_prefixes": [
    "runtime.",
    "reflect.",
    "testing."
  ],
  "min_change_percent": 5.0,
  "max_regression_percent": 20.0,
  "fail_on_improvement": false
}
```

### Global Settings Explained

| Setting                    | Description                                      | Default |
| -------------------------- | ------------------------------------------------ | ------- |
| `ignore_functions`         | Functions to ignore during comparison (exact)      | `[]`    |
| `ignore_prefixes`          | Function prefixes to ignore (e.g., "runtime.")   | `[]`    |
| `min_change_percent`       | Minimum change % for improvement gate / noise floor | `0.0`   |
| `max_regression_percent`   | Maximum acceptable regression % before fail      | disabled (`0`) |
| `fail_on_improvement`      | Whether to fail on performance improvements      | `false` |

## Benchmark-Specific Configuration

You can override global settings for specific benchmarks:

```json
"benchmarks": {
  "BenchmarkMyFunction": {
    "ignore_functions": ["BenchmarkMyFunction"],
    "min_change_percent": 3.0,
    "max_regression_percent": 10.0,
    "fail_on_improvement": true
  }
}
```

## Function Filtering

### Ignoring Specific Functions

Functions can be ignored by exact name:

```json
"ignore_functions": [
  "runtime.gcBgMarkWorker",
  "testing.(*B).ResetTimer",
  "myproject.BenchmarkFunction"
]
```

### Ignoring Function Prefixes

Functions can be ignored by package prefix:

```json
"ignore_prefixes": [
  "runtime.",
  "reflect.",
  "testing.",
  "syscall.",
  "internal/cpu."
]
```

This will ignore all functions from the `runtime`, `reflect`, `testing`, `syscall`, and `internal/cpu` packages.

## Threshold Configuration

### Minimum Change Threshold

Regressions below this percent do not fail the run (noise floor). When `fail_on_improvement` is true, improvements must exceed this magnitude to fail:

```json
"min_change_percent": 5.0
```

This prevents CI/CD from failing on minor fluctuations (e.g., 1–2% changes).

### Maximum Regression Threshold

When CLI regression flags are **not** set, the merged `track` policy uses this cap:

```json
"max_regression_percent": 15.0
```

If the worst flat regression meets or exceeds this percent (and the function is not ignored), the run fails.

### CLI vs configuration precedence

When **both** `--fail-on-regression` and a positive `--regression-threshold` are passed, the CLI threshold applies for that run and `track` thresholds are not used for the gate.

When those CLI flags are omitted, prof loads `prof.json` and applies the merged `track` policy (`track.defaults` plus `track.benchmarks[<name>]` overrides).

**Within `track` config**, benchmark-specific fields override `track.defaults` field-by-field (not “lowest threshold wins” across unrelated sources):

1. Start from `track.defaults`
2. Overlay non-empty fields from `track.benchmarks[<benchmark-name>]`

## Failing on Improvements

Sometimes you want to detect unexpected performance improvements:

```json
"fail_on_improvement": true
```

This is useful when:

- Performance improvements might indicate bugs
- You want to track all significant changes
- You're debugging unexpected behavior

## Complete Example

Here's a complete configuration example:

```json
{
  "version": 1,
  "collection": {
    "defaults": {
      "include_prefixes": ["github.com/myorg/myproject"],
      "ignore_functions": ["init", "TestMain"]
    }
  },
  "track": {
    "defaults": {
      "ignore_functions": ["runtime.gcBgMarkWorker", "testing.(*B).ResetTimer"],
      "ignore_prefixes": ["runtime.", "reflect.", "testing."],
      "min_change_percent": 5.0,
      "max_regression_percent": 20.0,
      "fail_on_improvement": false
    },
    "benchmarks": {
      "BenchmarkCriticalPath": {
        "min_change_percent": 1.0,
        "max_regression_percent": 5.0
      }
    }
  }
}
```

## CI/CD Integration

### GitHub Actions Example

```yaml
name: Performance Regression Check
on: [pull_request]

jobs:
  check:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: ">=1.24"

      - name: Install prof
        run: go install github.com/AlexsanderHamir/prof/cmd/prof@latest

      - name: Collect baseline
        run: |
          git fetch origin main --depth=1
          git checkout -qf origin/main
          prof auto --benchmarks "BenchmarkMyFunction" --profiles "cpu" --count 5 --tag baseline

      - name: Collect current
        run: |
          git checkout -
          prof auto --benchmarks "BenchmarkMyFunction" --profiles "cpu" --count 5 --tag PR

      - name: Check for regressions
        run: |
          prof track auto --base baseline --current PR \
            --profile-type cpu --bench-name "BenchmarkMyFunction" \
            --output-format summary
```

### Configuration File Location

The configuration file must be at your project root (same directory as `go.mod`):

```
your-project/
├── go.mod
├── prof.json  # ← CI/CD config goes here
├── cmd/
├── internal/
└── ...
```

## Complete Working Example

Here's a complete example that shows how to set up CI/CD performance tracking without requiring CLI flags:

### 1. Configuration File (`prof.json`)

```json
{
  "version": 1,
  "track": {
    "defaults": {
      "ignore_prefixes": ["runtime.", "reflect.", "testing."],
      "min_change_percent": 5.0,
      "max_regression_percent": 15.0,
      "fail_on_improvement": false
    },
    "benchmarks": {
      "BenchmarkMyFunction": {
        "min_change_percent": 3.0,
        "max_regression_percent": 10.0,
        "ignore_functions": ["setup", "teardown"]
      }
    }
  }
}
```

### 2. CI/CD Pipeline (`.github/workflows/performance.yml`)

```yaml
- name: Check for regressions
  run: |
    prof track auto --base baseline --current PR \
      --profile-type cpu --bench-name "BenchmarkMyFunction" \
      --output-format summary
```

Notice that no `--fail-on-regression` or `--regression-threshold` flags are needed. The tool will automatically use the thresholds from your configuration file.

## Best Practices

### 1. Start with track.defaults

Begin with defaults that apply to all benchmarks:

```json
"track": {
  "defaults": {
    "ignore_prefixes": ["runtime.", "reflect.", "testing."],
    "min_change_percent": 5.0,
    "max_regression_percent": 15.0
  }
}
```

### 2. CLI Flags vs Configuration

When using CI/CD configuration, the `--fail-on-regression` and `--regression-threshold` flags become optional:

**With CLI flags (overrides config for that run):**

```bash
prof track auto --base baseline --current PR \
  --profile-type cpu --bench-name "BenchmarkMyFunction" \
  --output-format summary --fail-on-regression --regression-threshold 5.0
```

**Without CLI flags (uses config only):**

```bash
prof track auto --base baseline --current PR \
  --profile-type cpu --bench-name "BenchmarkMyFunction" \
  --output-format summary
```

The second approach will use the thresholds defined in your `prof.json` file. This makes CI/CD pipelines cleaner and more maintainable.

### 3. Add benchmark-specific overrides

Only override `track.defaults` when necessary:

```json
"track": {
  "benchmarks": {
    "BenchmarkCriticalPath": {
      "min_change_percent": 1.0
    }
  }
}
```

### 4. Use Function Filtering Sparingly

Don't ignore too many functions - you might miss real regressions:

```json
"ignore_functions": [
  "runtime.gcBgMarkWorker",  // Known noisy function
  "testing.(*B).ResetTimer"  // Test infrastructure
]
```

### 5. Set Reasonable Thresholds

- `min_change_percent`: 5-10% for most cases
- `max_regression_percent`: 15-25% for most cases
- Critical paths: 1-5%

### 6. Monitor and Adjust

Review CI/CD failures and adjust thresholds based on:

- False positives (too sensitive)
- Missed regressions (not sensitive enough)
- Team feedback

## Troubleshooting

### Common Issues

1. **Configuration not loaded**: Ensure `prof.json` is at project root
2. **Functions still causing failures**: Check `ignore_functions` and `ignore_prefixes`
3. **Thresholds not working**: Verify `min_change_percent` and `max_regression_percent`
4. **Defaults vs benchmark settings**: `track.benchmarks[name]` overrides `track.defaults` field-by-field
5. **CLI flags vs config**: CLI gate applies only when `--fail-on-regression` and a positive `--regression-threshold` are both set; otherwise use `track` in `prof.json`

### Debug Information

Run `prof track auto` with logging enabled and inspect whether config loaded:

```bash
prof track auto --base baseline --current PR --bench-name "BenchmarkMyFunction"
```

When CLI gate flags are omitted, look for logs such as:

- "No CLI regression flags provided, using track configuration settings"
- "Performance regression below minimum threshold, not failing"

### Validation

Prof validates `prof.json` on load (`prof config validate`). Common validation errors:

- Negative thresholds
- Malformed JSON

## Migration from Command-Line

If you're currently using command-line flags:

### Before (Command-Line Only)

```bash
prof track auto --base baseline --current PR \
  --bench-name "BenchmarkMyFunction" \
  --fail-on-regression --regression-threshold 10.0
```

### After (With Configuration)

```json
{
  "version": 1,
  "track": {
    "defaults": {
      "max_regression_percent": 10.0
    }
  }
}
```

The configuration file provides the same functionality with more flexibility and better maintainability.
