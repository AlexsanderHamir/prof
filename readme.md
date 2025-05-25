# Go Benchmark Profiler

This tool simplifies complex performance analysis by consolidating multiple pprof commands into a single step. It automatically collects all relevant profiling data, organizes it, makes it searchable within your workspace, and enhances the process with AI-powered insights.

## Table of Contents

| ------------------------------------------- |
| [Features](#features) |
| [Usage](#usage) |
| [Directory Structure](#directory-structure) |
| [Configuration](#configuration) |
| [Installation](#installation) |
| [AI Analysis](#ai-analysis) |
| [Contributing](#contribution) |
| [License](#license) |

## Features

1. **Automatic Profile Extraction**
   Collects all the info you would see if you ran `go tool pprof profile.out top` (including all nodes) for each profile you requested.

2. **Line-Level Source Mapping by Default**
   Collects for all functions in the profile by default. You can limit this by specifying function prefixes to include and specific functions to exclude.

3. **Searchable, File-Based Reports**
   All the profiling data is saved to your workspace, making it easy to search. Instead of running multiple commands to inspect different functions, just use Command + P (in VSCode or similar editors) and search by function name.

4. **AI Analysis**
   Each profile is analyzed using AI with data extracted from `go tool pprof profile.out top` (including all nodes). There are no default prompts — you provide your own prompt to guide and customize the analysis output.

## Usage

> ⚠️ Always run commands from the directory where your benchmark code is located.

### Step 1: Create a Template Configuration

Generate a starter config file with:

```bash
prof setup --create-template
```

### Step 2: Run Benchmarks and Collect Profiles

This command runs the selected benchmarks, collects the specified profiles, and stores everything in a directory named after the tag:

```bash
prof -benchmarks "[BenchmarkGenPool, BenchmarkSyncPool]" -profiles "[cpu,memory,mutex]" -tag "test1" -count 1
```

What it does:

1. Runs each selected benchmark the specified number of times (`-count`).
2. Collects all selected profiles (e.g., CPU, memory, mutex).
3. Creates a directory named `test1` to store results.
4. Extracts and saves line-level code mapping for all functions in each profile.

## Directory Structure

When you run a benchmark analysis, a new directory is created inside bench (named according to your `-tag` parameter) with the following structure:

```
test1/
├── bin/                   # Binary files
│   ├── BenchmarkGenPool/
│   └── BenchmarkSyncPool/
├── cpu_functions/         # CPU profile line-level function mappings
│   ├── BenchmarkGenPool/
│   └── BenchmarkSyncPool/
├── memory_functions/       # Memory profile line-level function mappings
│   ├── BenchmarkGenPool/
│   └── BenchmarkSyncPool/
├── mutex_functions/        # Mutex profile line-level function mappings
│   ├── BenchmarkGenPool/
│   └── BenchmarkSyncPool/
├── text/                  # Profile reports
│   ├── BenchmarkGenPool/
│   │   ├── BenchmarkGenPool.txt        # Benchmark results
│   │   ├── BenchmarkGenPool_cpu.txt    # CPU profile analysis
│   │   ├── BenchmarkGenPool_memory.txt # Memory profile analysis
│   │   └── BenchmarkGenPool_mutex.txt  # Mutex profile analysis
│   └── BenchmarkSyncPool/
│       ├── BenchmarkSyncPool.txt
│       ├── BenchmarkSyncPool_cpu.txt
│       ├── BenchmarkSyncPool_memory.txt
│       └── BenchmarkSyncPool_mutex.txt
└── description.txt        # A file for you to describe what you're doing, what has changed and how it impacted performance.
```

## Configuration

The configuration file (`config_template.json`) controls how the profiler interacts with the AI service and manages benchmark analysis. Here's a detailed breakdown of each section:

### API Configuration

```json
{
  "api_key": "your-api-key-here", // Your OpenAI API key
  "base_url": "https://api.openai.com/v1" // OpenAI API endpoint
}
```

### Model Settings

The `model_config` section controls how the AI analyzes your profiles:

```json
"model_config": {
    "model": "gpt-4-turbo-preview",        // AI model to use for analysis
    "max_tokens": 4096,                    // Maximum response length
    "temperature": 0.7,                    // Creativity level (0.0-1.0)
    "top_p": 1.0,                         // Response diversity (0.0-1.0)
    "general_analyze_prompt_location": "path/to/your/system_prompt.txt" // Custom analysis prompt
}
```

### Benchmark Configurations

The `benchmark_configs` section lets you customize analysis for each benchmark:

```json
"benchmark_configs": {
    "BenchmarkGenPool": {                  // Name of your benchmark function
        "prefixes": [                      // Package prefixes to include in analysis
            "github.com/example/GenPool",
            "github.com/example/GenPool/internal",
            "github.com/example/GenPool/pkg"
        ],
        "ignore": "init,TestMain,BenchmarkMain"  // Functions to exclude from analysis
    }
}
```

#### Key Configuration Options:

1. **Prefixes**:

   - List of package prefixes to include in the analysis
   - Only functions from these packages will be analyzed in detail
   - Helps focus the analysis on relevant code
   - Example: `"github.com/your-project/core"`

2. **Ignore**:
   - Comma-separated list of function names to exclude
   - Useful for excluding setup/teardown code
   - Example: `"init,TestMain,BenchmarkMain,setup,teardown"`

### Example Use Cases:

1. **Basic Configuration**:

```json
{
  "api_key": "your-api-key-here",
  "base_url": "https://api.openai.com/v1",
  "model_config": {
    "model": "gpt-4-turbo-preview",
    "max_tokens": 4096,
    "temperature": 0.7
  },
  "benchmark_configs": {
    "BenchmarkMyFunction": {
      "prefixes": ["github.com/myproject"],
      "ignore": "init,TestMain"
    }
  }
}
```

2. **Advanced Configuration** (Multiple Benchmarks):

```json
{
  "api_key": "your-api-key-here",
  "base_url": "https://api.openai.com/v1",
  "model_config": {
    "model": "gpt-4-turbo-preview",
    "max_tokens": 4096,
    "temperature": 0.7,
    "general_analyze_prompt_location": "./prompts/custom_analysis.txt"
  },
  "benchmark_configs": {
    "BenchmarkOptimized": {
      "prefixes": [
        "github.com/myproject/optimized",
        "github.com/myproject/core"
      ],
      "ignore": "setup,teardown,TestMain"
    },
    "BenchmarkStandard": {
      "prefixes": ["github.com/myproject/standard"],
      "ignore": "init,TestMain"
    }
  }
}
```

### Tips for Effective Configuration:

1. **Start Small**: Begin with a basic configuration and expand as needed
2. **Use Specific Prefixes**: Narrow down the analysis scope to relevant packages
3. **Exclude Noise**: Use the `ignore` option to exclude setup/teardown code
4. **Custom Prompts**: Create custom analysis prompts for specific use cases
5. **Multiple Benchmarks**: Configure each benchmark separately for targeted analysis

## AI Analysis

The profiler uses AI to analyze your benchmark profiles individually and provide insights. It does not analyze line-level code mappings—its focus is on interpreting each profile as a whole.

### Contribution

Bring your skills, share your ideas, and contribute your code.

## Installation

### Docker-based Installation (Recommended)

The easiest way to install and use the Go Benchmark Profiler is through Docker. This method requires no Python installation and works consistently across different operating systems.

1. **Prerequisites**:

   - Install [Docker](https://docs.docker.com/get-docker/)
   - Make sure Docker daemon is running

2. **Installation Steps**:

   ```bash
   # Clone the repository
   git clone https://github.com/yourusername/go-benchmark-profiler.git
   cd go-benchmark-profiler

   # Run the installation script
   chmod +x install.sh
   sudo ./install.sh
   ```

3. **Verify Installation**:
   ```bash
   prof --help
   ```

The installation script will:

- Build a Docker image with all required dependencies
- Create a `prof` command wrapper in `/usr/local/bin`
- Set up persistent configuration storage

### Manual Installation (Alternative)

If you prefer to install without Docker, you'll need:

- Python 3.11 or later
- Go 1.21 or later
- pip (Python package manager)

Then follow these steps:

1. Clone the repository
2. Create a virtual environment: `python -m venv venv`
3. Activate the virtual environment:
   - On Unix/macOS: `source venv/bin/activate`
   - On Windows: `.\venv\Scripts\activate`
4. Install dependencies: `pip install -r requirements.txt`
5. Make the prof script executable: `chmod +x prof`
6. Add the script to your PATH or use it with `./prof`

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
