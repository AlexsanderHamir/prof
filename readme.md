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

We welcome contributions of all kinds—bug fixes, new features, tests, and documentation improvements. Before getting started, make sure to review the [issues](https://github.com/AlexsanderHamir/prof/issues) to avoid duplicated effort.

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

### Instructions

1. **Run Tests and Linters Locally:**
   GitHub Actions will automatically run unit tests and lint checks on each pull request, but you should run them locally first to ensure your code is clean and passes all checks:

   - Run all tests: `go test ./...`
   - Run the linter: `golangci-lint run`

2. **Follow the Code Style:**
   Use idiomatic Go. Keep functions small, testable, and documented. Use descriptive names and avoid unnecessary abstractions.

3. **Keep Commits Atomic and Meaningful:**
   Structure commits logically, each should represent a focused change. Avoid mixing formatting, refactoring, and feature implementation in a single commit.

4. **Add Tests for New Features or Fixes:**
   All non-trivial changes should be accompanied by appropriate test coverage. If you're unsure how to test something, feel free to ask in the PR or open an issue first.

5. **Document Any User-Facing Changes:**
   If your contribution affects the CLI, config file, or output format, update the relevant parts of the documentation (README, CLI help, or usage examples).

6. **Open a Pull Request:**
   Once your changes are ready and tested locally, open a PR with a clear description of what's changed and why. Link to any relevant issues.

7. **Be Open to Feedback:**
   Reviews are meant to maintain code quality and project direction. We're happy to help iterate on PRs to get them merged.

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
