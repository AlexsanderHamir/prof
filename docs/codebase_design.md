# Prof Codebase Design

Prof is a Go benchmark profiling tool that automates performance analysis by wrapping Go's built-in benchmarking and pprof tools. This document explains the architecture, design decisions, and how the system works.

## 🎯 What Prof Does

Prof solves three main problems:

1. **Automates profiling**: Runs Go benchmarks with multiple profile types (CPU, memory, mutex, block)
2. **Organizes data**: Creates a structured directory hierarchy for all profiling outputs
3. **Tracks performance**: Compares benchmark runs to detect regressions and improvements

## 🏗️ Architecture Overview

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

### `cmd/prof/` - Application Entry Point

**File**: `main.go`
**Purpose**: Minimal main function that delegates to CLI

```go
func main() {
    if err := cli.Execute(); err != nil {
        // Handle fatal errors with visual indicators
        os.Exit(1)
    }
}
```

**Key Design**: Single responsibility - only handles startup and error exit. All business logic is delegated.

### `cli/` - Command Interface Layer

**Location**: `cli/`
**Framework**: Cobra CLI framework
**Purpose**: Orchestrates user commands and coordinates between engine components

#### Commands Available:

- **`auto`** - Automated benchmark execution with profiling
- **`manual`** - Process existing profile files
- **`track`** - Compare performance between runs
- **`setup`** - Generate configuration templates
- **`tui`** - Text-based user interface

#### Key Functions:

```go
// Creates the root command with all subcommands
func CreateRootCmd() *cobra.Command

// Handles benchmark execution workflow
func runBenchmarks(cmd *cobra.Command, args []string) error

// Manages performance tracking
func runTrackAuto(cmd *cobra.Command, args []string) error
```

**Design Pattern**: Command pattern with clear separation between command definition and execution logic.

### `engine/` - Core Business Logic

The engine package contains three main components that handle the core workflows:

#### `engine/benchmark/` - Test Execution Engine

**Purpose**: Manages Go benchmark execution and directory setup

**Key Functions**:

```go
// Creates the output directory structure
func SetupDirectories(tag string, benchmarks, profiles []string) error

// Runs a specific benchmark with profiling
func RunBenchmark(benchmarkName string, profiles []string, count int, tag string) error

// Processes generated profile files
func ProcessProfiles(benchmarkName string, profiles []string, tag string) error
```

**How It Works**:

1. Creates organized directory structure for outputs
2. Executes `go test -bench` with profiling flags
3. Moves generated profile files to organized locations
4. Triggers profile processing pipeline

#### `engine/collector/` - Profile Collection Engine

**Purpose**: Converts binary profiles to text and generates visualizations

**Key Functions**:

```go
// Converts binary profiles to text format
func GetProfileTextOutput(binaryFile, outputFile string) error

// Generates PNG visualizations
func GetPNGOutput(binaryFile, outputFile string) error

// Handles manual profile processing workflow
func RunCollector(files []string, tag string) error
```

**How It Works**:

1. Uses `go tool pprof` to convert binary profiles to text
2. Generates PNG visualizations for visual analysis
3. Applies function filtering based on configuration
4. Organizes outputs in structured directories

#### `engine/tracker/` - Performance Analysis Engine

**Purpose**: Compares benchmark runs to detect performance changes

**Key Functions**:

```go
// Compares auto-generated benchmark data
func CheckPerformanceDifferences(baselineTag, currentTag, benchName, profileType string) (*ProfileChangeReport, error)

// Compares manually specified profile files
func CheckPerformanceDifferencesManual(baselineProfile, currentProfile string) (*ProfileChangeReport, error)
```

**How It Works**:

1. Loads profile data from two different runs
2. Parses text profiles into comparable objects
3. Matches functions between runs
4. Calculates performance differences (regressions/improvements)
5. Generates detailed reports

### `parser/` - Profile Data Processing

**Purpose**: Converts pprof text output into structured data for analysis

**Key Functions**:

```go
// Converts profile lines to structured objects
func TurnLinesIntoObjects(profilePath string) ([]*LineObj, error)

// Extracts function names with filtering
func GetAllFunctionNames(filePath string, filter config.FunctionFilter) (names []string, err error)
```

**Data Structure**:

```go
type LineObj struct {
    FnName         string  // Function name
    Flat           float64 // Flat time/memory usage
    FlatPercentage float64 // Percentage of total
    SumPercentage  float64 // Cumulative percentage
    Cum            float64 // Cumulative time/memory
    CumPercentage  float64 // Cumulative percentage
}
```

**How It Works**:

1. Reads pprof text output line by line
2. Parses numeric values and function names using regex
3. Applies filtering rules (prefixes, ignore functions)
4. Creates structured objects for comparison

### `internal/` - Protected Utilities

#### `internal/config/` - Configuration Management

**Purpose**: Manages JSON-based configuration for function filtering

**Configuration Structure**:

```json
{
  "function_collection_filter": {
    "BenchmarkName": {
      "include_prefixes": ["github.com/myorg/myproject"],
      "ignore_functions": ["init", "TestMain"]
    },
    "*": {
      "include_prefixes": ["github.com/myorg"],
      "ignore_functions": ["runtime.*"]
    }
  }
}
```

**Key Features**:

- Per-benchmark filtering rules
- Global filtering with `*` wildcard
- Prefix-based inclusion (package paths)
- Function name exclusion

#### `internal/args/` - Data Structures

**Purpose**: Shared data structures for parameter passing between components

**Key Types**:

```go
type BenchArgs struct {
    Benchmarks []string
    Profiles   []string
    Count      int
    Tag        string
}

type CollectionArgs struct {
    Tag             string
    Profiles        []string
    BenchmarkName   string
    BenchmarkConfig config.FunctionFilter
}
```

#### `internal/shared/` - Common Utilities

**Purpose**: File system operations and shared constants

**Key Functions**:

```go
// Creates or cleans directories
func CleanOrCreateDir(dir string) error

// Finds Go module root
func FindGoModuleRoot() (string, error)

// File operations with proper permissions
func GetScanner(filePath string) (*bufio.Scanner, *os.File, error)
```

## 🔄 Data Flow

### 1. Automated Benchmark Workflow

```
User Command → CLI → Benchmark Engine → Collector → Parser → Output Files
     ↓              ↓           ↓           ↓         ↓
  prof auto    Setup Dirs   Run Tests   Process   Extract
  --benchmarks --profiles   --count     Profiles  Functions
```

**Step-by-Step**:

1. **CLI**: Validates user input and loads configuration
2. **Benchmark**: Creates directories and runs `go test -bench`
3. **Collector**: Converts binary profiles to text and PNG
4. **Parser**: Extracts function-level data with filtering
5. **Output**: Organized directory structure with all data

### 2. Performance Tracking Workflow

```
User Command → CLI → Tracker → Parser → Comparison → Report
     ↓              ↓         ↓         ↓           ↓
  prof track   Load Data   Parse    Match      Generate
  --base       --current   Files    Functions  Report
```

**Step-by-Step**:

1. **CLI**: Validates tracking parameters
2. **Tracker**: Loads profile data from both runs
3. **Parser**: Converts text to comparable objects
4. **Comparison**: Matches functions and calculates differences
5. **Report**: Generates detailed performance analysis

## 📁 Output Organization

Prof creates a structured directory hierarchy:

```
bench/
├── {tag}/                          # Benchmark run identifier
│   ├── bin/                        # Binary profile files
│   │   └── {benchmark}/
│   │       ├── {benchmark}_cpu.out
│   │       ├── {benchmark}_memory.out
│   │       └── {benchmark}_block.out
│   ├── text/                       # Text profile reports
│   │   └── {benchmark}/
│   │       ├── {benchmark}_cpu.txt
│   │       ├── {benchmark}_memory.txt
│   │       └── {benchmark}_block.txt
│   ├── cpu_functions/              # Function-level CPU data
│   │   └── {benchmark}/
│   │       ├── function1.txt
│   │       └── function2.txt
│   ├── memory_functions/           # Function-level memory data
│   │   └── {benchmark}/
│   │       ├── function1.txt
│   │       └── function2.txt
│   └── {profile}_functions/        # Other profile types
```

**Design Benefits**:

- **Organized**: Clear separation by profile type and data format
- **Tagged**: Multiple runs can coexist without conflicts
- **Structured**: Consistent naming and organization
- **Accessible**: Easy to navigate and find specific data

## ⚙️ Configuration System

### Function Filtering

Prof uses a sophisticated filtering system to control which functions are analyzed:

```json
{
  "function_collection_filter": {
    "BenchmarkGenPool": {
      "include_prefixes": ["github.com/myorg/pool"],
      "ignore_functions": ["runtime.*", "testing.*"]
    }
  }
}
```

**Filter Types**:

- **`include_prefixes`**: Only collect functions from specific packages
- **`ignore_functions`**: Skip specific function names (supports wildcards)
- **Global filters**: Apply to all benchmarks using `*` key

### Configuration Benefits

- **Performance**: Reduces data collection overhead
- **Focus**: Concentrates analysis on relevant code
- **Flexibility**: Different rules per benchmark
- **Maintainability**: Easy to update without code changes

## 🧪 Testing Strategy

### Test Organization

- **Integration Tests** (`tests/`): End-to-end workflow testing
- **Unit Tests** (per package): Component-specific testing
- **Blackbox Tests**: CLI command validation and output verification

### Test Coverage Areas

- CLI command validation and error handling
- Benchmark execution workflows
- Profile parsing accuracy
- Configuration handling
- Error scenarios and edge cases
- Performance regression detection

### Test Data

- **Fixtures**: Sample profile files and benchmark outputs
- **Mock Data**: Simulated performance data for testing
- **Edge Cases**: Various profile formats and error conditions

## 🔌 Dependencies

### External Dependencies

- **Cobra** (`github.com/spf13/cobra`): CLI framework for command management
- **Go Toolchain**: `go test`, `go tool pprof` for profiling
- **Survey** (`github.com/AlecAivazis/survey/v2`): Interactive prompts for TUI

### Internal Dependencies

The dependency graph follows clean layered architecture:

```
CLI → Engine → Parser → Internal
 ↓      ↓       ↓        ↓
Cobra  Business Data   Utils
       Logic   Processing
```

**Dependency Rules**:

- Higher layers can depend on lower layers
- Lower layers cannot depend on higher layers
- Internal packages are protected from external imports
- Clear interfaces between major components

## 🎨 Design Principles

### 1. Single Responsibility

Each package has one clear purpose:

- **CLI**: User interaction and command orchestration
- **Engine**: Business logic and workflow management
- **Parser**: Data transformation and filtering
- **Internal**: Utility functions and shared resources

### 2. Interface Segregation

Components communicate through well-defined interfaces:

- Clear function signatures
- Consistent error handling
- Predictable data structures

### 3. Configuration Over Code

Behavior is controlled through configuration files:

- Function filtering rules
- Output formatting options
- Performance thresholds

### 4. Error Handling

Comprehensive error handling throughout the system:

- Descriptive error messages
- Proper error propagation
- Graceful degradation where possible

## 🚀 Performance Considerations

### Memory Management

- **Streaming**: Processes large profiles line by line
- **Filtering**: Reduces memory usage through early filtering
- **Cleanup**: Proper file handling and resource cleanup

### Scalability

- **Parallel Processing**: Can handle multiple profile types simultaneously
- **Efficient Parsing**: Regex-based parsing for fast text processing
- **Structured Output**: Organized data for quick access and analysis

## 🔧 Development Workflow

### Adding New Features

1. **Identify Layer**: Determine which architectural layer needs changes
2. **Maintain Boundaries**: Respect package dependencies and interfaces
3. **Add Tests**: Include comprehensive testing for new functionality
4. **Update Documentation**: Keep this document current with changes

### Code Organization

- **New Commands**: Add to appropriate CLI package
- **New Engines**: Create in engine package with clear interfaces
- **New Parsers**: Extend parser package for new data formats
- **New Utilities**: Place in internal packages for reuse

### Testing Guidelines

- **Unit Tests**: Test individual functions and methods
- **Integration Tests**: Test component interactions
- **End-to-End Tests**: Test complete user workflows
- **Performance Tests**: Ensure changes don't introduce regressions

## 📚 Further Reading

- **Go Profiling**: [Go pprof documentation](https://pkg.go.dev/runtime/pprof)
- **Cobra CLI**: [Cobra framework documentation](https://github.com/spf13/cobra)
- **Go Testing**: [Go testing package documentation](https://pkg.go.dev/testing)

---

This documentation provides a comprehensive understanding of Prof's architecture. For specific implementation details, refer to the source code and inline comments in each package.
