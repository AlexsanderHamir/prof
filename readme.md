# Prof - Go Benchmark Data Collection Tool

Prof is a CLI tool that automates Go performance profiling workflows by collecting and organizing all pprof-generated data, including detailed function-level performance information, guided by custom configuration.

## Table of Contents

- [Key Features](#key-features)

  - [Automated Data Collection](#automated-data-collection)
  - [Why It Matters](#why-it-matters)

- [Performance Change Tracking](#performance-change-tracking)
- [Installation](#installation)
- [Quick Start](#quick-start)
- [Usage](#usage)

  - [Collecting Benchmark Data](#collecting-benchmark-data)
  - [Configuration Setup](#configuration-setup)
  - [Performance Tracking](#performance-tracking)

- [What Prof Collects For You](#what-prof-collects-for-you)
- [Supported Profile Types](#supported-profile-types)
- [Commands Reference](#commands-reference)

  - [Main Command (Data Collection)](#main-command-data-collection)
  - [Subcommands](#subcommands)

- [Requirements](#requirements)
- [Examples](#examples)

  - [Basic Data Collection](#basic-data-collection)
  - [Collect Data for Multiple Benchmarks](#collect-data-for-multiple-benchmarks)
  - [Performance Comparison (Analysis Feature)](#performance-comparison-analysis-feature)

- [Troubleshooting](#troubleshooting)

  - [Common Issues](#common-issues)

## Key Features

### Automated Data Collection

**Prof automates collection and organization of data from these standard `pprof` workflows:**

1. **Benchmark execution with profiling:**

   ```bash
   go test -bench=BenchmarkName -cpuprofile=cpu.out -memprofile=memory.out -mutexprofile=mutex.out -blockprofile=block.out -count=N
   ```

2. **Text reports for top functions (per profile):**

   ```bash
   go tool pprof -cum -edgefraction=0 -nodefraction=0 -top cpu.out
   # (repeated for each profile type)
   ```

3. **Graphical visualizations (PNG):**

   ```bash
   go tool pprof -png cpu.out > cpu_graph.png
   # (repeated for each profile type)
   ```

4. **Function-level profiling data:**

   ```bash
   go tool pprof -list=FunctionName1 cpu.out > function1.txt
   go tool pprof -list=FunctionName2 cpu.out > function2.txt
   # ... repeated across all specified functions and profile types
   # (Prof will auto-collect all functions if no configuration is provided)
   ```

### Why It Matters

Without Prof, a typical profiling session might require dozens of manual steps.

**Example: 1 benchmark, 2 profile types, 10 functions of interest**

| Task                        | Commands Required |
| --------------------------- | ----------------- |
| Benchmark execution         | 1                 |
| Top functions (per profile) | 2                 |
| Function-level reports      | 20                |
| **Total**                   | **23 commands**   |

**With Prof: ✅ Just one command to collect and structure everything — no manual repetition. Easily search any function or profile by name.**

## Performance Change Tracking

- Compare performance between different benchmark runs at the profile level.
- Detect regressions and improvements with detailed reporting.

[image example](./summary_example.png)

## Installation

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

Or clone and build from source:

```bash
git clone https://github.com/AlexsanderHamir/prof.git
cd prof
go build -o prof cmd/prof/main.go
```

## Quick Start

1. **Set up configuration** (optional - controls which functions to collect data for):

```bash
prof setup --create-template
```

2. **Run benchmarks and collect all profiling data**:

```bash
prof --benchmarks "[BenchmarkMyFunction]" --profiles "[cpu,memory]" --count 5 --tag "v1.0"
```

3. **Compare performance between runs**:

```bash
prof track --base-tag v1.0 --current-tag v1.1 --bench BenchmarkMyFunction --profile-type cpu
```

## Usage

### Collecting Benchmark Data

```bash
prof --benchmarks "[BenchmarkFunc1,BenchmarkFunc2]" \
     --profiles "[cpu,memory,mutex,block]" \
     --count 10 \
     --tag "experiment-1"
```

**Parameters:**

- `--benchmarks`: List of benchmark functions to run (in brackets)
- `--profiles`: Types of profiles to collect (`cpu`, `memory`, `mutex`, `block`)
- `--count`: Number of benchmark iterations
- `--tag`: Identifier for this benchmark run

### Configuration Setup

Generate a configuration template file:

```bash
prof setup --create-template
```

Example configuration (controls which functions to collect detailed data for):

```json
{
  "function_collection_filter": {
    "BenchmarkMyPool": {
      "include_prefixes": [
        "github.com/myorg/myproject",
        "github.com/myorg/myproject/internal"
      ],
      "ignore_functions": ["init", "TestMain"]
    }
  }
}
```

**Options:**

- `include_prefixes`: Only collect detailed data for functions with these prefixes.
- `ignore_functions`: Skip these specific function names even if includes a specified prefix.

**Without configuration**: Prof collects data for all functions (which can be a lot of files).

### Performance Tracking

Compare two benchmark runs:

```bash
prof track --base-tag baseline \
           --current-tag experiment \
           --bench BenchmarkMyFunction \
           --profile-type cpu \
           --format detailed
```

## What Prof Collects For You

Prof organizes all collected data in this structure:

```
bench/
└── your-tag/
    ├── bin/                          # Binary profile files (for further analysis)
    │   └── BenchmarkName/
    │       ├── BenchmarkName_cpu.out
    │       ├── BenchmarkName_memory.out
    │       └── BenchmarkName.test
    ├── text/                         # Text profile outputs
    │   └── BenchmarkName/
    │       ├── BenchmarkName.txt     # Raw benchmark output
    │       ├── BenchmarkName_cpu.txt # pprof text output
    │       └── BenchmarkName_memory.txt
    ├── cpu_functions/                # Function-level profiling data
    │   └── BenchmarkName/
    │       ├── function1.txt         # pprof -list=function1 output
    │       ├── function2.txt         # pprof -list=function2 output
    │       └── BenchmarkName_cpu.png # Profile visualization
    └── memory_functions/
        └── BenchmarkName/
            ├── function1.txt
            └── BenchmarkName_memory.png
```

## Supported Profile Types

Prof accepts these profile types:

- **cpu**: CPU profiling (execution time)
- **memory**: Memory allocation profiling
- **mutex**: Mutex contention profiling
- **block**: Blocking operations profiling

## Commands Reference

### Main Command (Data Collection)

```bash
prof --benchmarks "[list]" --profiles "[list]" --count N --tag "name"
```

### Subcommands

- `prof setup --create-template`: Generate configuration template
- `prof track`: Compare performance between runs
- `prof version`: Show version information

## Requirements

- Go 1.24.3 or later
- Access to `go test` and `go tool pprof` commands
- Must be run from where the desired benchmarks are located.

## Examples

### Basic Data Collection

```bash
# Collect CPU and memory profiling data
prof --benchmarks "[BenchmarkStringProcessor]" \
     --profiles "[cpu,memory]" \
     --count 5 \
     --tag "baseline"
```

### Collect Data for Multiple Benchmarks

```bash
prof --benchmarks "[BenchmarkPool,BenchmarkCache,BenchmarkQueue]" \
     --profiles "[cpu,memory,mutex]" \
     --count 10 \
     --tag "v2.0"
```

### Performance Comparison (Analysis Feature)

```bash
# Compare collected data between two runs
prof track --base-tag baseline \
           --current-tag v2.0 \
           --bench BenchmarkPool \
           --profile-type cpu \
           --format summary
```

## Troubleshooting

### Common Issues

1. **"go: cannot find main module"**

   - Ensure you're running prof from within a Go project directory
   - Check that `go.mod` exists in your project

2. **"Profile file not found"**

   - Verify benchmark names are correct (must start with `Benchmark`)
   - Ensure benchmarks actually run and complete successfully
   - Run the command from where the benchmark is located

3. **Too many function files generated**

   - Use configuration to filter which functions to collect data for
   - Add `include_prefixes` to focus on your project's functions only

4. **Empty or small profile files**
   - Increase benchmark iterations (`--count`)
   - Ensure benchmark has sufficient work to generate meaningful profiles
