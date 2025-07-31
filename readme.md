# Prof - Go Benchmark Profiling Made Simple

Prof automates Go performance profiling by collecting all pprof data in one command and enabling easy performance comparisons between benchmark runs.

[![GoDoc](https://godoc.org/github.com/AlexsanderHamir/prof?status.svg)](https://godoc.org/github.com/AlexsanderHamir/prof)
[![Go Report Card](https://goreportcard.com/badge/github.com/AlexsanderHamir/prof)](https://goreportcard.com/report/github.com/AlexsanderHamir/prof)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Issues](https://img.shields.io/github/issues/AlexsanderHamir/prof)
![Last Commit](https://img.shields.io/github/last-commit/AlexsanderHamir/prof)
![Code Size](https://img.shields.io/github/languages/code-size/AlexsanderHamir/prof)
![Version](https://img.shields.io/github/v/tag/AlexsanderHamir/prof?sort=semver)
![Go Version](https://img.shields.io/badge/Go-1.24.3%2B-blue)

📖 [Documentation](https://alexsanderhamir.github.io/prof/) | ▶️ [Watch Demo Video](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/output.mp4)

## Benchmark Comparison Summary View:

This view highlights regressions, improvements, and stable functions.

![Summary of benchmark performance changes](./summary_example.png)

## Benchmark Comparison Detailed View:

This view provides a breakdown of performance metrics per function.

![Function-level performance comparison](./detailed_example.png)

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

2. **Compare performance between tags:**

```bash
prof track auto --base "baseline" --current "optimized" --profile-type "cpu" --bench-name "BenchmarkName" --output-format "summary"
```

## Requirements

- Go 1.24.3 or later
- Install graphviz
- Run from within your Go project directory (where benchmarks are located)

## Troubleshooting

**"go: cannot find main module"**

- Run prof from within a Go project directory with `go.mod`

**"Profile file not found"**

- Verify benchmark names are correct
- Ensure benchmarks run successfully and generate profiles

**Too many function files**

- Use configuration to filter functions with `include_prefixes`

**Configuration Not Taking Effect**

- Make sure the config file is located in the current working directory, the one you're running the command from.

## Contributing

We welcome contributions of all kinds—bug fixes, new features, tests, and documentation improvements. Before getting started, make sure to review the [issues](https://github.com/AlexsanderHamir/prof/issues) to avoid duplicated effort, and see the [contribution guidelines](./CONTRIBUTING.md) for more information.

### Quick Start

**Requirements:** Go 1.24.3+, Git

```bash
# Clone the repository
git clone https://github.com/AlexsanderHamir/prof.git
cd prof

# Run tests
go test -v ./...

# Check for linter issues (first-time install if needed)
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
golangci-lint run

# Build the binary for local testing
go build -o prof ./cmd/prof
```

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
