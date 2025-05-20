## Desired Brainstorming

1. Analysis folder containing AI analysis on all the files inside of the text folder
   - Don't make AI do all the analysis job, get it to flag potential issues.

## Version Features

### Version 0.0.1 - Initial Benchmark Organization

This version focuses on establishing the basic structure for benchmark analysis:

- **Benchmark Collection**: Gathers benchmarks based on benchmark names
- **Profile Management**: Collects and organizes selected profiles for each benchmark
- **File Organization**:
  - Moves .out files into a dedicated bin directory
  - Extracts top nodes from pprof into text files
- **Directory Structure**: Implements a clear hierarchy: bench/tag/text/bin/description.txt
- **Benchmark Tagging**: Enables tagging for benchmark comparisons

### Version 0.0.2 - Function Profile Analysis

This version introduces detailed function-level profiling analysis:

- **Profile Function Analysis**: Scans profile text files to create detailed function profiles using the pprof list command
- **Organized Storage Structure**:
  - Creates a `profile_functions` directory inside the bench folder
  - Supports different profile types (e.g., `cpu_functions`, `memory_functions`)
  - Each function gets its own file, named after the function
- **Benchmark-Specific Organization**:
  - Separate folders for each benchmark (e.g., `BenchmarkGenPool`, `BenchmarkSyncPool`)
- **Flexible Function Filtering**:
  - Optional user-defined function prefixes (e.g., `github.com/AlexsanderHamir/GenPool`)
  - Configurable function ignore list
- **Automated Profile Generation**:
  - Uses `go tool pprof -list` to generate detailed function profiles
  - Organizes output files in a structured hierarchy (e.g., `cpu_functions/GenPool/RetrieveOrCreate.txt`)

## 0.1.2 - AI help


