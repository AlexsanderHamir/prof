# Codebase

Prof is a Go benchmark profiling tool designed to automate performance analysis and comparison. It wraps Go's built-in benchmarking and pprof tools to provide comprehensive profiling data collection, organization, and performance regression detection.

**[Interactive Architecture Graph](https://claude.ai/public/artifacts/3582cc33-8d87-447a-8cac-7e94c2b67f5b)** - Explore the component relationships visually

## Core Functionality

Prof provides three main capabilities:

1. **Automated Profiling**: Runs Go benchmarks with multiple profile types (CPU, memory, mutex, block)
2. **Data Organization**: Collects and organizes all profiling data in a structured directory hierarchy
3. **Performance Tracking**: Compares benchmark runs to detect performance regressions and improvements

## Package Architecture

### Package Structure

```
prof/
├── cmd/prof/          # Application entry point
├── cli/               # Command-line interface
├── engine/            # Core business logic
│   ├── benchmark/     # Benchmark execution
│   ├── collector/     # Profile collection
│   └── tracker/       # Performance analysis
│   └── version/       # Version management
├── parser/            # Profile data parsing
├── internal/          # Internal utilities (protected)
│   ├── args/          # Data structures
│   ├── config/        # Configuration management
│   ├── shared/        # Common utilities
└── tests/             # Test suite
```

## Package Details

### `cmd/prof` - Entry Point

**Location**: `cmd/prof/main.go`

The application entry point that initializes and starts the CLI. This is the minimal main function that delegates all functionality to the CLI package.

**Responsibilities**:

- Application startup
- Error handling and exit codes
- CLI delegation

### `cli` - Command Interface

**Location**: `cli/`

Handles all user interaction through a Cobra-based command-line interface. Acts as the orchestration layer that coordinates between different engine components.

**Key Files**:

- `api.go` - Command definitions and CLI setup
- `helpers.go` - CLI utility functions and output formatting

**Commands**:

- `auto` - Automated benchmark and profiling workflow
- `manual` - Manual profile file processing
- `track` - Performance comparison between runs
- `setup` - Configuration file generation
- `version` - Version information and updates

**Responsibilities**:

- Command parsing and validation
- User input handling
- Workflow orchestration
- Output formatting and reporting

### `engine` - Core Business Logic

The engine package contains all the core operational components of Prof.

#### `engine/benchmark` - Test Execution

**Responsibilities**:

- Directory structure setup
- Go benchmark execution with profiling flags
- Profile file management and organization
- Integration with collector for profile processing

**Key Functions**:

- `SetupDirectories()` - Creates output directory structure
- `RunBenchmark()` - Executes Go benchmarks with profiling
- `ProcessProfiles()` - Processes profile files after benchmark execution
- `CollectProfileFunctions()` - Collects function-level profiling data

#### `engine/collector` - Profile Collection

**Responsibilities**:

- Profile file processing (binary → text conversion)
- PNG visualization generation
- Function-level data extraction
- Manual profile file handling

**Key Functions**:

- `GetProfileTextOutput()` - Converts binary profiles to text format
- `GetPNGOutput()` - Generates profile visualizations
- `GetFunctionsOutput()` - Extracts individual function profiles
- `RunCollector()` - Handles manual profile processing workflow

#### `engine/tracker` - Performance Analysis

**Responsibilities**:

- Performance comparison between benchmark runs
- Regression and improvement detection
- Detailed performance reporting
- Statistical analysis of profile differences

**Key Functions**:

- `CheckPerformanceDifferences()` - Compares two benchmark runs
- `CheckPerformanceDifferencesManual()` - Manual profile comparison
- `DetectChange()` - Analyzes performance changes between runs

### `parser` - Profile Data Processing

**Location**: `parser/`

Handles parsing and processing of pprof text output into structured data that can be analyzed and compared.

**Responsibilities**:

- pprof text file parsing
- Function name extraction with filtering
- Profile data structure conversion
- Line-by-line profile analysis

**Key Functions**:

- `GetAllFunctionNames()` - Extracts function names from profiles
- `TurnLinesIntoObjects()` - Converts profile lines to structured data
- `ShouldKeepLine()` - Applies filtering rules to profile lines

### `internal` - Protected Utilities

The internal package contains utilities that are protected from external imports by Go's internal package convention.

#### `internal/config` - Configuration Management

**Responsibilities**:

- JSON configuration file handling
- Function filtering configuration
- Template generation

**Key Types**:

- `Config` - Main configuration structure
- `FunctionFilter` - Per-benchmark filtering rules

#### `internal/args` - Data Structures

**Responsibilities**:

- Shared data structures across packages
- Parameter passing between components

**Key Types**:

- `BenchArgs` - Benchmark execution parameters
- `CollectionArgs` - Profile collection parameters
- `LineFilterArgs` - Profile filtering parameters

#### `internal/shared` - Common Utilities

**Responsibilities**:

- File system operations
- Constants and shared values
- Common utility functions

**Key Constants**:

- Directory and file naming conventions
- Command names and status messages
- File permissions and extensions

#### `internal/version` - Version Management

**Responsibilities**:

- Version checking and comparison
- GitHub release API integration
- Update notifications

## Data Flow

### 1. Benchmark Execution Flow

```
CLI → Benchmark → Collector → Parser
```

1. **CLI** receives user command with benchmark parameters
2. **Benchmark** sets up directories and runs Go tests with profiling
3. **Collector** processes the generated profile files
4. **Parser** extracts function-level data from profiles

### 2. Performance Tracking Flow

```
CLI → Tracker → Parser
```

1. **CLI** receives track command with baseline and current tags
2. **Tracker** loads profile data from both runs
3. **Parser** converts profile text to comparable objects
4. **Tracker** analyzes differences and generates reports

## Configuration System

Prof uses a JSON-based configuration system for customizing function collection behavior.

### Configuration File Structure

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

### Configuration Options

- **`include_prefixes`**: Only collect functions matching these package prefixes
- **`ignore_functions`**: Skip specific function names even if they match prefixes

## Output Structure

Prof organizes all output in a structured directory hierarchy:

```
bench/
├── {tag}/
│   ├── bin/                    # Binary profile files
│   │   └── {benchmark}/
│   │       ├── {benchmark}_cpu.out
│   │       ├── {benchmark}_memory.out
│   │       └── ...
│   ├── text/                   # Text profile reports
│   │   └── {benchmark}/
│   │       ├── {benchmark}_cpu.txt
│   │       ├── {benchmark}_memory.txt
│   │       └── ...
│   ├── cpu_functions/          # Function-level CPU data
│   │   └── {benchmark}/
│   │       ├── function1.txt
│   │       ├── function2.txt
│   │       └── ...
│   └── memory_functions/       # Function-level memory data
│       └── {benchmark}/
│           ├── function1.txt
│           └── ...
```

## Dependencies

### External Dependencies

- **Cobra** (`github.com/spf13/cobra`) - CLI framework
- **Go toolchain** - `go test`, `go tool pprof`

### Internal Dependencies

The dependency graph follows a clean layered architecture:

- **CLI** depends on **Engine** and **Internal** packages
- **Engine** components depend on **Parser** and **Internal** packages
- **Parser** depends on **Internal** packages
- **Internal** packages have minimal cross-dependencies

## Testing Strategy

### Test Organization

- **Integration Tests** (`tests/`) - End-to-end workflow testing
- **Unit Tests** (per package) - Component-specific testing
- **Fixtures** - Test data and profile files for testing

### Test Coverage

- CLI command validation
- Benchmark execution workflows
- Profile parsing accuracy
- Configuration handling
- Error scenarios and edge cases

## Usage Examples

### Basic Benchmark Profiling

```bash
prof auto --benchmarks BenchmarkMyFunction --profiles cpu,memory --count 5 --tag baseline
```

### Manual Profile Processing

```bash
prof manual --tag manual-analysis cpu.out memory.out
```

### Performance Comparison

```bash
prof track auto --base-tag baseline --current-tag optimized --bench-name BenchmarkMyFunction --profile-type cpu --output-format summary
```

## Contributing Guidelines

When modifying the codebase:

1. **Respect Package Boundaries** - Keep engine, parser, and internal concerns separated
2. **Maintain Interfaces** - Preserve existing API contracts
3. **Add Tests** - Include tests for new functionality
4. **Update Documentation** - Keep this documentation current with changes
5. **Follow Conventions** - Use established naming and organization patterns

---

> For more detailed information about specific functions and APIs, refer to the source code documentation and inline comments.
