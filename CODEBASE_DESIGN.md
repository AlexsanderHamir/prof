# Prof Codebase Design

Prof is a Go benchmark profiling tool that automates performance analysis by wrapping Go's built-in benchmarking and pprof tools.

## 🎯 Purpose

Prof solves three core problems:

- **Automates profiling**: Runs Go benchmarks with multiple profile types (CPU, memory, mutex, block)
- **Organizes data**: Creates structured directory hierarchies for profiling outputs
- **Tracks performance**: Compares benchmark runs to detect regressions and improvements

## 🏗️ Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   CLI Layer     │───▶│  Engine Layer   │───▶│  Parser Layer   │
│   (User Input)  │    │ (Business Logic)│    │ (Data Processing)│
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Internal      │    │   Configuration │    │   File System   │
│   (Utilities)   │    │   (JSON Config) │    │   (Output Org)  │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 📦 Package Structure

### `cmd/prof/main.go` - Entry Point

Minimal main function that delegates to CLI:

```go
func main() {
    if err := cli.Execute(); err != nil {
        os.Exit(1)
    }
}
```

### `cli/` - Command Interface

**Framework**: Cobra CLI  
**Commands**:

- `auto` - Automated benchmark execution with profiling
- `manual` - Process existing profile files
- `track` - Compare performance between runs
- `setup` - Generate configuration templates
- `tui` - Text-based user interface

### `engine/` - Core Business Logic

#### `engine/benchmark/` - Test Execution

- Creates organized directory structures
- Executes `go test -bench` with profiling flags
- Manages profile file organization

#### `engine/collector/` - Profile Processing

- Converts binary profiles to text format
- Generates PNG visualizations
- Applies function filtering

#### `engine/tracker/` - Performance Analysis

- Compares benchmark runs
- Detects performance changes
- Generates detailed reports

### `parser/` - Data Processing

Converts pprof text output into structured data:

```go
type LineObj struct {
    FnName         string
    Flat           float64
    FlatPercentage float64
    SumPercentage  float64
    Cum            float64
    CumPercentage  float64
}
```

### `internal/` - Utilities

#### `internal/config/` - Configuration Management

JSON-based function filtering:

```json
{
  "function_collection_filter": {
    "BenchmarkName": {
      "include_prefixes": ["github.com/myorg/myproject"],
      "ignore_functions": ["init", "TestMain"]
    }
  }
}
```

#### `internal/args/` - Data Structures

Shared parameter structures for component communication.

#### `internal/shared/` - Common Utilities

File system operations and shared constants.

## 🔄 Workflows

### Automated Benchmark Flow

```
User → CLI → Benchmark Engine → Collector → Parser → Output Files
```

### Performance Tracking Flow

```
User → CLI → Tracker → Parser → Comparison → Report
```

## 📁 Output Structure

```
bench/
├── {tag}/                         # Run identifier
│   ├── bin/                       # Binary profiles
│   ├── text/                      # Text reports
│   ├── cpu_functions/             # Function-level data
│   └── memory_functions/
```

## ⚙️ Configuration

### Function Filtering

- **`include_prefixes`**: Restrict to specific packages
- **`ignore_functions`**: Skip specific functions (supports wildcards)
- **Global filters**: Apply to all benchmarks using `*` key

## � Testing Strategy

### Test Types

- **Integration Tests**: End-to-end workflow testing
- **Unit Tests**: Component-specific testing
- **Blackbox Tests**: CLI validation and output verification

### Coverage Areas

- CLI command validation
- Benchmark execution
- Profile parsing accuracy
- Error handling scenarios

## 🔌 Dependencies

### External

- **Cobra**: CLI framework
- **Go Toolchain**: Built-in profiling tools
- **Survey**: Interactive prompts

### Internal

Layered dependency structure with clear boundaries between components.

## 🎨 Design Principles

1. **Single Responsibility**: Each package has one clear purpose
2. **Interface Segregation**: Well-defined component interfaces
3. **Configuration Over Code**: Behavior controlled through config files
4. **Comprehensive Error Handling**: Descriptive errors and graceful degradation