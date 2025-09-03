# Prof - Go Benchmark Profiling Made Simple

Prof automates Go performance profiling by collecting all pprof data in one command and enabling easy performance comparisons between benchmark runs.

[![GoDoc](https://godoc.org/github.com/AlexsanderHamir/prof?status.svg)](https://godoc.org/github.com/AlexsanderHamir/prof)
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Issues](https://img.shields.io/github/issues/AlexsanderHamir/prof)
![Last Commit](https://img.shields.io/github/last-commit/AlexsanderHamir/prof)
![Code Size](https://img.shields.io/github/languages/code-size/AlexsanderHamir/prof)
![Version](https://img.shields.io/github/v/tag/AlexsanderHamir/prof?sort=semver)
![Go Version](https://img.shields.io/badge/Go-1.24.3%2B-blue)

📖 [Documentation](https://alexsanderhamir.github.io/prof/) | ▶️ [Watch Demo Video](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/output.mp4) | ▶️ [Watch TUI Demo](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/tui_prof.mp4)

> The project is actively looking for contributors !!

## Why Prof?

**Before Prof:** Profiling a single benchmark with multiple profile types requires dozens of manual commands:

```bash
# Run benchmark
go test -bench=BenchmarkName -cpuprofile=cpu.out -memprofile=memory.out ...

# Generate reports for each profile type
go tool pprof -cum -top cpu.out
go tool pprof -cum -top memory.out

# Extract function-level data for each function of interest
go tool pprof -list=Function1 cpu.out > function1.txt
go tool pprof -list=Function2 cpu.out > function2.txt
# ... repeat for every function × every profile type
```

**With Prof:** One command collects everything and organizes it automatically.

### 🚀 **Supercharge Your Development Tools**

When you use Prof with AI-powered development tools like **Cursor**, you get a massive productivity boost:

- **80% faster completion** - Prof provides structured, organized profiling data that AI tools can instantly analyze
- **Much better optimization suggestions** - AI tools can see the complete performance picture across all profile types
- **Smarter code recommendations** - With detailed function-level profiling data, AI tools can suggest more targeted optimizations
- **Faster debugging** - AI tools can quickly identify performance bottlenecks using Prof's organized output format

Instead of AI tools struggling with raw pprof files, they get clean, structured data that enables them to provide superior performance insights and code improvements.

## Key Features

### 🚀 One Command Profiling

Collect CPU, memory, mutex, and block profiles in a single command:

```bash
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
```

### 📊 Performance Comparison

Easily compare performance between different versions or optimizations:

```bash
prof track auto --base "baseline" --current "optimized" --profile-type "cpu" --bench-name "BenchmarkName"
```

### 🔍 Regression Detection

Fail CI/CD pipelines on performance regressions with configurable thresholds:

```bash
prof track auto --base "baseline" --current "PR" --profile-type "cpu" --bench-name "BenchmarkName" --fail-on-regression --regression-threshold 5.0
```

**Enhanced CI/CD Support**: Configure function filtering, and custom thresholds to reduce noise and make CI/CD more reliable. See [CI/CD Configuration Guide](docs/cicd_configuration.md) for details.

### 📁 Organized Output

All profiling data is automatically organized under `bench/<tag>/` directories with clear structure.

### 📦 Package-Level Grouping

Organize profile data by package/module for better analysis and collaboration:

```bash
# Group profile data by package when collecting
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory" --count 5 --tag "baseline" --group-by-package

# Group profile data from existing files
prof manual --tag "external-profiles" --group-by-package cpu.prof memory.prof
```

When enabled, this creates additional `*_grouped.txt` files that organize functions by their package/module, making it easier to:

- Identify which packages consume the most resources
- Share package-level performance insights with team members
- Focus optimization efforts on specific modules

## Interactive TUI

Don't want to remember benchmark names or commands? Use the interactive terminal interface:

```bash
prof tui
```

**What it does:**

- 🔍 **Auto-discovers** all `BenchmarkXxx` functions in your project
- 📋 **Interactive selection** of benchmarks, profiles, count, and tag
- 🎯 **No typos** - everything is selected from menus
- 📁 **Same output** as `prof auto` - organized under `bench/<tag>/`

**TUI Track Mode:**
Compare existing benchmark data interactively:

```bash
prof tui track
```

## Installation

```bash
go install github.com/AlexsanderHamir/prof/cmd/prof@latest
```

## Quick Start

1. **Collect profiling data:**

```bash
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory,mutex,block" --count 10 --tag "baseline"
prof auto --benchmarks "BenchmarkName" --profiles "cpu,memory,mutex,block" --count 10 --tag "optimized"
```

2. **Compare performance:**

```bash
prof track auto --base "baseline" --current "optimized" --profile-type "cpu" --bench-name "BenchmarkName" --output-format "summary"
```

## Documentation

- 📚 **[Full Documentation](https://alexsanderhamir.github.io/prof/)** - Complete API reference and guides
- 🚀 **[Contributing Guidelines](./CONTRIBUTING.md)** - How to contribute to Prof
- 📋 **[Code of Conduct](./CODE_OF_CONDUCT.md)** - Community guidelines
- 🏗️ **[Codebase Design](./CODEBASE_DESIGN.md)** - Architecture and design decisions

## Requirements

- Go 1.24.3 or later
- Install graphviz
- A Go module (`go.mod`) at the repository root

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
