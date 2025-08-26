# CI/CD Configuration Guide

This document explains how to configure Prof for CI/CD environments to reduce noise and make performance regression detection more reliable.

## Overview

Prof's CI/CD configuration allows you to:

- **Filter out noisy functions** that shouldn't cause CI/CD failures
- **Set different thresholds** for different benchmarks
- **Override command-line settings** with configuration files
- **Fail on unexpected improvements** if needed

## Configuration Structure

The CI/CD configuration is added to your existing `config_template.json` file under the `ci_config` section:

```json
{
  "function_collection_filter": {
    // ... existing function filtering ...
  },
  "ci_config": {
    "global": {
      // Global CI/CD settings
    },
    "benchmarks": {
      "BenchmarkName": {
        // Benchmark-specific CI/CD settings
      }
    }
  }
}
```

## Global Configuration

Global settings apply to all benchmarks unless overridden by benchmark-specific settings:

```json
"global": {
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
  "min_change_threshold": 5.0,
  "max_regression_threshold": 20.0,
  "fail_on_improvement": false,
}
```

### Global Settings Explained

| Setting                    | Description                                      | Default |
| -------------------------- | ------------------------------------------------ | ------- |
| `ignore_functions`         | Functions to ignore during CI/CD (exact matches) | `[]`    |
| `ignore_prefixes`          | Function prefixes to ignore (e.g., "runtime.")   | `[]`    |
| `min_change_threshold`     | Minimum change % to trigger CI/CD failure        | `0.0`   |
| `max_regression_threshold` | Maximum acceptable regression %                  | `∞`     |
| `fail_on_improvement`      | Whether to fail on performance improvements      | `false` |

## Benchmark-Specific Configuration

You can override global settings for specific benchmarks:

```json
"benchmarks": {
  "BenchmarkMyFunction": {
    "ignore_functions": ["BenchmarkMyFunction"],
    "min_change_threshold": 3.0,
    "max_regression_threshold": 10.0,
    "fail_on_improvement": true,
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

Only functions with changes ≥ this threshold will cause CI/CD failures:

```json
"min_change_threshold": 5.0
```

This prevents CI/CD from failing on minor fluctuations (e.g., 1-2% changes).

### Maximum Regression Threshold

This overrides command-line `--regression-threshold` settings:

```json
"max_regression_threshold": 15.0
```

If a function regresses by 15%, CI/CD will fail regardless of command-line settings.

### Command-Line Override Priority

1. Benchmark-specific `max_regression_threshold`
2. Global `max_regression_threshold`
3. Command-line `--regression-threshold`

The most restrictive (lowest) threshold wins.

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
  "function_collection_filter": {
    "*": {
      "include_prefixes": ["github.com/myorg/myproject"],
      "ignore_functions": ["init", "TestMain"]
    }
  },
  "ci_config": {
    "global": {
      "ignore_functions": ["runtime.gcBgMarkWorker", "testing.(*B).ResetTimer"],
      "ignore_prefixes": ["runtime.", "reflect.", "testing."],
      "min_change_threshold": 5.0,
      "max_regression_threshold": 20.0,
      "fail_on_improvement": false
    },
    "benchmarks": {
      "BenchmarkCriticalPath": {
        "min_change_threshold": 1.0,
        "max_regression_threshold": 5.0
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
├── config_template.json  # ← CI/CD config goes here
├── cmd/
├── internal/
└── ...
```

## Complete Working Example

Here's a complete example that shows how to set up CI/CD performance tracking without requiring CLI flags:

### 1. Configuration File (`config_template.json`)

```json
{
  "ci_config": {
    "global": {
      "ignore_prefixes": ["runtime.", "reflect.", "testing."],
      "min_change_threshold": 5.0,
      "max_regression_threshold": 15.0,
      "fail_on_improvement": false
    },
    "benchmarks": {
      "BenchmarkMyFunction": {
        "min_change_threshold": 3.0,
        "max_regression_threshold": 10.0,
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

### 1. Start with Global Configuration

Begin with global settings that apply to all benchmarks:

```json
"global": {
  "ignore_prefixes": ["runtime.", "reflect.", "testing."],
  "min_change_threshold": 5.0
}
```

### 2. CLI Flags vs Configuration

When using CI/CD configuration, the `--fail-on-regression` and `--regression-threshold` flags become optional:

**With CLI flags (overrides config):**

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

The second approach will use the thresholds defined in your `config_template.json` file. This makes CI/CD pipelines cleaner and more maintainable.

### 3. Add Benchmark-Specific Overrides

Only override global settings when necessary:

```json
"benchmarks": {
  "BenchmarkCriticalPath": {
    "min_change_threshold": 1.0  // More sensitive for critical paths
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

- `min_change_threshold`: 5-10% for most cases
- `max_regression_threshold`: 15-25% for most cases
- Critical paths: 1-5%

### 6. Monitor and Adjust

Review CI/CD failures and adjust thresholds based on:

- False positives (too sensitive)
- Missed regressions (not sensitive enough)
- Team feedback

## Troubleshooting

### Common Issues

1. **Configuration not loaded**: Ensure `config_template.json` is at project root
2. **Functions still causing failures**: Check `ignore_functions` and `ignore_prefixes`
3. **Thresholds not working**: Verify `min_change_threshold` and `max_regression_threshold`
4. **Global vs benchmark settings**: Benchmark-specific settings override global
5. **CLI flags vs config**: When using CI/CD config, `--fail-on-regression` and `--regression-threshold` are optional

### Debug Information

Prof logs configuration loading and filtering decisions:

```bash
prof track auto --base baseline --current PR --bench-name "BenchmarkMyFunction"
```

Look for logs like:

- "Applied CI/CD configuration filtering"
- "Function ignored by CI/CD config"
- "Performance regression below minimum threshold"

### Validation

Prof validates configuration on startup. Common validation errors:

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
  "ci_config": {
    "global": {
      "max_regression_threshold": 10.0
    }
  }
}
```

The configuration file provides the same functionality with more flexibility and better maintainability.
