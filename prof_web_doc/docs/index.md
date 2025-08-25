# Profiling Data Management

When performing complex profiling, developers often find themselves lost in a maze of repetitive commands and scattered files. You run `go test -bench=BenchmarkMyFunc -cpuprofile=cpu.out`, then `go tool pprof -top cpu.out > results.txt`, inspect a function with `go tool pprof -list=MyFunc cpu.out`, make modifications, run the benchmark again—and hours later, you're exhausted, have dozens of inconsistently named files scattered across directories, and can't remember which changes led to which results. Without systematic organization, you lose track of your optimization journey, lack accurate "before and after" snapshots to share with your team, and waste valuable time context-switching between profiling commands instead of focusing on actual performance improvements. Prof eliminates this chaos by capturing everything in one command and automatically organizing all profiling data—binary files, text reports, function-level analysis, and visualizations—into a structured, tagged hierarchy that preserves your optimization history and makes collaboration effortless.

## Quick Reference

**Main Commands:**

- **`prof auto`**: Automated benchmark collection and profiling
- **`prof tui`**: Interactive benchmark collection
- **`prof tui track`**: Interactive performance comparison
- **`prof manual`**: Process existing profile files
- **`prof track auto`**: Compare performance between tags
- **`prof track manual`**: Compare external pprof files (`.out`/`.prof`)

**Directory Flexibility:**

- **Project root**: Run from anywhere in your Go project (recommended)
- **Configuration**: Configuration file (`config_template.json`) is always looked for at the project root
- **Global Search**: prof auto searches for the benchmark name globally, regardless of the directory you run it from. If you run it from a subdirectory, the **bench** directory will be created there instead of at the project root.

## Auto

The `auto` command wraps `go test` and `pprof` to run benchmarks, collect all profile types, and organize everything automatically:

```bash
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

This single command replaces dozens of manual steps and creates a complete, organized profiling dataset ready for analysis or comparison.

**Output Structure:**

```
bench/baseline/
├── description.txt                 # User documentation for this run
├── bin/BenchmarkGenPool/           # Binary profile files
│   ├── BenchmarkGenPool_cpu.out
│   ├── BenchmarkGenPool_memory.out
│   ├── BenchmarkGenPool_mutex.out
│   └── BenchmarkGenPool_block.out
├── text/BenchmarkGenPool/          # Text reports & benchmark output
│   ├── BenchmarkGenPool_cpu.txt
│   ├── BenchmarkGenPool_memory.txt
│   └── BenchmarkGenPool.txt
├── cpu_functions/BenchmarkGenPool/ # Function-level CPU profile data
│   ├── Put.txt
│   ├── Get.txt
│   └── getShard.txt
└── memory_functions/BenchmarkGenPool/ # Function-level memory profile data
    ├── Put.txt
    └── allocator.txt
```

## Auto - Configuration

By default, prof gathers code-level data for every function listed in a profile’s text report. To change this behavior, run:

```bash
prof setup
```

This creates a configuration file with the following structure:

```json
{
  "function_collection_filter": {
    "BenchmarkGenPool": {
      "include_prefixes": ["github.com/example/GenPool"],
      "ignore_functions": ["init", "TestMain", "BenchmarkMain"]
    }
  }
}
```

**Configuration Options:**

- `BenchmarkGenPool`: Replace it with your benchmark function name, or with `"*"` to apply for all benchmarks.
- `include_prefixes`: Only collect functions whose names start with these prefixes.
- `ignore_functions`: Exclude specific functions from collection, even if they match the include prefixes.

## TUI - Interactive Selection

The `tui` command provides an interactive terminal interface that automatically discovers benchmarks in your project and guides you through the selection process:

```bash
prof tui
```

**What it does:**

1. **Discovers benchmarks**: Automatically scans your Go module for `func BenchmarkXxx(b *testing.B)` functions in `*_test.go` files.
2. **Interactive selection**: Presents a menu where you can select:
   - Which benchmarks to run (multi-select from discovered list)
   - Which profiles to collect (cpu, memory, mutex, block)
   - Number of benchmark runs (count)
   - Tag name for organizing results

**Navigation:**

- **Page size**: Shows up to 20 benchmarks at once for readability
- **Scroll**: Use arrow keys (↑/↓) to navigate through the list
- **Multi-select**: Use spacebar to select/deselect benchmarks
- **Search**: Type to filter and find specific benchmarks quickly

## Manual

The `manual` command processes existing pprof files (`.out` or `.prof`) without running benchmarks - it uses `pprof` to convert them to text reports and organize the data:

```bash
prof manual --tag "external-profiles" BenchmarkGenPool_cpu.out memory.out block.out
```

This organizes your existing profile files into a flatter structure based on the profile filename:

**Manual Output Structure:**

```
bench/external-profiles/
├── BenchmarkGenPool_cpu/
│   ├── BenchmarkGenPool_cpu.txt    # Text report
│   └── functions/                  # Function-level profile data
│       ├── Put.txt
│       ├── Get.txt
│       └── getShard.txt
├── memory/
│   ├── memory.txt                  # Text report
│   └── functions/                  # Function-level profile data
│       └── allocator.txt
└── block/
    ├── block.txt                   # Text report
    └── functions/                  # Function-level profile data
        └── runtime.txt
```

## Manual - Configuration

The configuration works the same as auto configuration, except you should use profile file base names (without extensions) instead of benchmark names:

```json
{
  "function_collection_filter": {
    "BenchmarkGenPool_cpu": {
      "include_prefixes": ["github.com/example/GenPool"],
      "ignore_functions": ["init", "TestMain", "BenchmarkMain"]
    }
  }
}
```

For example, `BenchmarkGenPool_cpu.out` becomes `BenchmarkGenPool_cpu` in the configuration.

Use `*` if you want the config to be applied to all profile files.

# Performance Comparison

Prof's performance comparison automatically drills down from benchmark-level changes to show you exactly which functions changed. Instead of just reporting that performance improved or regressed, Prof pinpoints the specific functions responsible and shows you detailed before-and-after comparisons.

## Track Auto

Use `track auto` when comparing data collected with `prof auto`. Simply reference the tag names:

```bash
prof track auto --base "baseline" --current "optimized" \
                --profile-type "cpu" --bench-name "BenchmarkGenPool" \
                --output-format "summary"

prof track auto --base "baseline" --current "optimized" \
                --profile-type "cpu" --bench-name "BenchmarkGenPool" \
                --output-format "detailed"
```

## Track Manual

Use `track manual` when comparing external profile files by specifying their relative paths. **Note**: This command accepts pprof files (`.out` or `.prof`) directly, not text reports:

```bash
prof track manual --base path/to/base/BenchmarkGenPool_cpu.out \
                  --current path/to/current/BenchmarkGenPool_cpu.out \
                  --output-format "summary"

prof track manual --base path/to/base/BenchmarkGenPool_cpu.out \
                  --current path/to/current/BenchmarkGenPool_cpu.out \
                  --output-format "detailed"
```

## TUI Track - Interactive Performance Comparison

The `tui track` command provides an interactive interface for comparing performance between existing benchmark runs. This is a companion to the main `prof tui` command and requires that you have already collected benchmark data using either `prof tui` or `prof auto`.

```bash
prof tui track
```

**What it does:**

1. **Discovers existing data**: Scans the `bench/` directory for tags you've already collected
2. **Interactive selection**: Guides you through selecting:
   - Baseline tag (the "before" version)
   - Current tag (the "after" version)
   - Benchmark to compare
   - Profile type to analyze
   - Output format
   - Regression threshold settings

**Prerequisites:**

- Must have run `prof tui` or `prof auto` at least twice to create baseline and current tags
- Data must be organized under `bench/<tag>/` directories

## Output formats supported:

Prof's performance comparison provides multiple output formats to help you understand performance changes at different levels of detail and presentation.

- **summary**: High-level overview of all performance changes
- **detailed**: Comprehensive analysis for each changed function
- **summary-html**: HTML export of summary report
- **detailed-html**: HTML export of detailed report
- **summary-json**: JSON export of summary report
- **detailed-json**: JSON export of detailed report

### Summary Format

The summary format gives you a high-level overview of all performance changes, organized by impact:

```
==== Performance Tracking Summary ====
Total Functions Analyzed: 78
Regressions: 9
Improvements: 8
Stable: 61

⚠️  Top Regressions (worst first):
• internal/cache.getShard: +200.0% (0.030s → 0.090s)
• internal/hash.Spread: +180.0% (0.050s → 0.140s)
• pool/acquire: +150.0% (0.020s → 0.050s)
• encoding/json.Marshal: +125.0% (0.080s → 0.180s)
• sync.Pool.Get: +100.0% (0.010s → 0.020s)

✅ Top Improvements (best first):
• compress/gzip.NewWriter: -100.0% (0.020s → 0.000s)
• internal/metrics.resetCounters: -100.0% (0.010s → 0.000s)
• encoding/json.Unmarshal: -95.0% (0.100s → 0.005s)
• net/url.ParseQuery: -90.0% (0.050s → 0.005s)
• pool/isFull: -85.0% (0.020s → 0.003s)
```

### Detailed Format

The detailed format provides comprehensive analysis for each changed function, including impact assessment and action recommendations:

```
📊 Summary: 78 total functions | 🔴 9 regressions | 🟢 8 improvements | ⚪ 61 stable
📋 Report Order: Regressions first (worst → best), then Improvements (best → worst), then Stable

║ ║ ║ ║ ║ ║ ║      PERFORMANCE CHANGE REPORT

Function: github.com/Random/Pool/pool.getShard
Analysis Time: 2025-07-23 15:51:59 PDT
Change Type: REGRESSION
⚠️ Performance regression detected

║ ║ ║ ║ ║      FLAT TIME ANALYSIS

Before:        0.030000s
After:         0.090000s
Delta:         +0.060000s
Change:        +200.00%
Impact:        Function is 200.00% SLOWER

║ ║ ║ ║ ║      CUMULATIVE TIME ANALYSIS

Before:        0.030s
After:         0.100s
Delta:         +0.070s
Change:        +233.33%

║ ║ ║ ║ ║      IMPACT ASSESSMENT

Severity:      CRITICAL
Recommendation: Critical regression! Immediate investigation required.
```

# CI/CD: Fail on regressions

**Understanding the regression threshold:**

The `--regression-threshold` flag sets a percentage limit on performance regressions. When enabled with `--fail-on-regression`, the command will exit with a non-zero status code if any function's **flat time** regression exceeds this threshold.

**Flat time regression calculation:**

```
Flat regression % = (current_time - baseline_time) / baseline_time × 100
```

**Example:** If a function took 100ms in baseline and 110ms in current run:

- Flat regression = (110 - 100) / 100 × 100 = +10%
- With `--regression-threshold 5.0`, this would fail the build
- With `--regression-threshold 15.0`, this would pass

**Note:** The threshold applies to **flat time** (time spent directly in the function), not cumulative time (time including all called functions). Flat time gives a more direct measure of the function's own performance impact.

## CI/CD Configuration-Based Approach

Prof now supports a configuration-based approach for CI/CD that eliminates the need for command-line flags and provides more flexibility.

### Configuration Structure

Add a `ci_config` section to your existing `config_template.json` file:

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

### Global Configuration

```json
"global": {
  "ignore_functions": ["runtime.gcBgMarkWorker", "testing.(*B).ResetTimer"],
  "ignore_prefixes": ["runtime.", "reflect.", "testing."],
  "min_change_threshold": 5.0,
  "max_regression_threshold": 20.0,
  "fail_on_improvement": false
}
```

### Benchmark-Specific Configuration

```json
"benchmarks": {
  "BenchmarkMyFunction": {
    "min_change_threshold": 3.0,
    "max_regression_threshold": 10.0
  }
}
```

### Function Filtering

**Ignore specific functions:**

```json
"ignore_functions": ["runtime.gcBgMarkWorker", "testing.(*B).ResetTimer"]
```

**Ignore function prefixes:**

```json
"ignore_prefixes": ["runtime.", "reflect.", "testing."]
```

### Threshold Configuration

- `min_change_threshold`: Minimum change % to trigger CI/CD failure
- `max_regression_threshold`: Maximum acceptable regression %
- Command-line flags are optional when using configuration

### Complete Example

```json
{
  "ci_config": {
    "global": {
      "ignore_prefixes": ["runtime.", "reflect.", "testing."],
      "min_change_threshold": 5.0,
      "max_regression_threshold": 20.0
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

### CI/CD Integration

With configuration-based CI/CD, you no longer need `--fail-on-regression` or `--regression-threshold` flags:

```bash
prof track auto --base baseline --current PR \
  --profile-type cpu --bench-name "BenchmarkMyFunction" \
```

**Example GitHub Actions:**

```yaml
- name: Check for regressions
  run: |
    prof track auto --base baseline --current PR \
      --profile-type cpu --bench-name "BenchmarkMyFunction" \
```

**Configuration File Location:** Must be at project root (same directory as `go.mod`).

# Prof Tools

Prof provides additional tools that can easily operate on the collected data for enhanced analysis and visualization.

## Tools Overview

The `prof tools` command provides access to specialized analysis tools:

```bash
prof tools [command] [flags]
```

Available tools:

- **`benchstat`**: Statistical analysis of benchmark results
- **`qcachegrind`**: Visual call graph analysis

## Benchstat Tool

Runs Go's official `benchstat` command on collected benchmark data.

### Usage

```bash
prof tools benchstat --base <baseline-tag> --current <current-tag> --bench-name <benchmark-name>
```

### Example

```bash
prof tools benchstat --base baseline --current optimized --bench-name BenchmarkGenPool
```

### Prerequisites

```bash
go install golang.org/x/perf/cmd/benchstat@latest
```

### Output

Results are saved to `bench/tools/benchstats/{benchmark_name}_results.txt`

## QCacheGrind Tool

Generates call graph data from binary profile files and launches the QCacheGrind visualizer.

### Usage

```bash
prof tools qcachegrind --tag <tag> --profiles <profile-type> --bench-name <benchmark-name>
```

### Example

```bash
prof tools qcachegrind --tag optimized --profiles cpu --bench-name BenchmarkGenPool
```

### Prerequisites

**Ubuntu/Debian:**

```bash
sudo apt-get install qcachegrind
```

**macOS:**

```bash
brew install qcachegrind
```

### Output

Callgrind files are saved to `bench/tools/qcachegrind/{benchmark_name}_{profile_type}.callgrind`

## Tool Output Organization

```
bench/
├── baseline/
├── optimized/
└── tools/
    ├── benchstats/
    │   └── BenchmarkGenPool_results.txt
    └── qcachegrind/
        └── BenchmarkGenPool_cpu.callgrind
```

## Integration with Existing Workflow

1. **Collect data**: Use `prof auto` or `prof tui`
2. **Compare performance**: Use `prof track`
3. **Deep analysis**: Use `prof tools`
4. **Visual exploration**: Use QCacheGrind for interactive call graph analysis

## Best Practices

**Combine tools for comprehensive analysis:**

```bash
# Collect data
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory" --count 10 --tag baseline
prof auto --benchmarks "BenchmarkGenPool" --profiles "cpu,memory" --count 10 --tag optimized

# Compare performance
prof track auto --base baseline --current optimized --bench-name BenchmarkGenPool

# Statistical validation
prof tools benchstat --base baseline --current optimized --bench-name BenchmarkGenPool

# Deep analysis
prof tools qcachegrind --tag optimized --profiles cpu --bench-name BenchmarkGenPool
```
