# Profiling Data Management

When performing complex profiling, developers often find themselves lost in a maze of repetitive commands and scattered files. You run `go test -bench=BenchmarkMyFunc -cpuprofile=cpu.out`, then `go tool pprof -top cpu.out > results.txt`, inspect a function with `go tool pprof -list=MyFunc cpu.out`, make modifications, run the benchmark again—and hours later, you're exhausted, have dozens of inconsistently named files scattered across directories, and can't remember which changes led to which results. Without systematic organization, you lose track of your optimization journey, lack accurate "before and after" snapshots to share with your team, and waste valuable time context-switching between profiling commands instead of focusing on actual performance improvements. Prof eliminates this chaos by capturing everything in one command and automatically organizing all profiling data—binary files, text reports, function-level analysis, and visualizations—into a structured, tagged hierarchy that preserves your optimization history and makes collaboration effortless.

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

By default, prof collects all functions shown in the text report of a profile. To customize this behavior, run:

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

This filtering helps focus profiling on relevant code paths while excluding test setup and initialization functions that may not be meaningful for performance analysis.

## Manual

The `manual` command processes existing profile files without running benchmarks - it only uses `pprof` to organize data you already have:

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

Use `track manual` when comparing external profile files by specifying their relative paths:

```bash
prof track manual --base path/to/base/report/cpu.txt \
                  --current path/to/current/report/cpu.txt \
                  --output-format "summary"

prof track manual --base path/to/base/report/cpu.txt \
                  --current path/to/current/report/cpu.txt \
                  --output-format "detailed"
```

## Result Examples

Prof's performance comparison provides two output formats to help you understand performance changes at different levels of detail.

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
