# Prof Benchmark Tool

A CLI tool for benchmarking Go code with profile analysis and AI-powered insights.

## Building from Source

1. Ensure you have Python 3.8+ installed
2. Install build dependencies:
   ```bash
   pip install -e .
   ```
3. Build the binary:
   ```bash
   python build.py
   ```
4. Install the binary:
   ```bash
   cp dist/prof ~/bin/
   ```

## Usage

```bash
prof -benchmarks "[BenchmarkName]" -profiles "[cpu,memory]" -tag "test1" -count 1 -analyze
```

### Options

- `-benchmarks`: Comma-separated list of benchmark types (e.g., "[BenchmarkGenPool,BenchmarkSyncPool]")
- `-profiles`: Comma-separated list of profile types (e.g., "[cpu,memory,mutex]")
- `-tag`: Tag for the benchmark run (e.g., "test1")
- `-count`: Number of benchmark iterations (e.g., 5)
- `-analyze`: (Optional) Run AI analysis on the benchmark results
- `-benchmark-config`: (Optional) JSON-like string containing benchmark-specific configurations

### Example

```bash
prof -benchmarks "[BenchmarkGenPool]" -profiles "[cpu,memory]" -tag "test1" -count 1 -analyze
```

## Development

To modify the tool:

1. Make your changes to the source files
2. Rebuild using `python build.py`
3. Reinstall the binary using `cp dist/prof ~/bin/`
