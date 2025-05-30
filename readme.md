# Go Benchmark Profiler

This tool simplifies complex performance analysis by consolidating multiple pprof commands into a single step. It automatically collects all relevant profiling data, organizes it, makes it searchable within your workspace, and enhances the process with AI-powered insights.

[Example Profile Analysis Video](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/prof.mp4)

## Table of Contents

[Features](#features)

[Usage](#usage)

[Directory Structure](#directory-structure)

[Configuration](#configuration)

[Installation](#installation)

[AI Analysis](#ai-analysis)

[Contributing](#contribution)

[License](#license)

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

Use the following command to run benchmarks, collect profiles, and store results:

```bash
prof -benchmarks "[BenchmarkGenPool, BenchmarkSyncPool]" -profiles "[cpu,memory,mutex]" -tag "test1" -count 1
```

This command:

1. Runs each specified benchmark (`-benchmarks`) the given number of times (`-count`).
2. Collects the selected profiles (e.g., CPU, memory, mutex).
3. Saves results in a directory named after the tag (`test1`).
4. Extracts and stores line-level code mappings for all functions in each profile.

## Directory Structure

When you run a benchmark analysis, a new directory is created inside `bench/` (named according to your `-tag` parameter) with the following structure:

```
bench/
└── test1/                # Directory named after your -tag parameter
    ├── bin/              # Binary files
    │   ├── BenchmarkGenPool/
    │   └── BenchmarkSyncPool/
    ├── cpu_functions/    # CPU profile line-level function mappings
    │   ├── BenchmarkGenPool/
    │   └── BenchmarkSyncPool/
    ├── memory_functions/ # Memory profile line-level function mappings
    │   ├── BenchmarkGenPool/
    │   └── BenchmarkSyncPool/
    ├── mutex_functions/  # Mutex profile line-level function mappings
    │   ├── BenchmarkGenPool/
    │   └── BenchmarkSyncPool/
    ├── text/            # Profile reports
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
    └── description.txt  # A file for you to describe what you're doing, what has changed and how it impacted performance.
```

## Configuration

The configuration file (`config_template.json`) controls how the profiler interacts with the AI service and manages benchmark analysis. Here's a detailed breakdown of each section:

### API Configuration

```json
{
  "api_key": "your-api-key-here",
  "base_url": "https://api.openai.com/v1"
}
```

### Model Settings

The `model_config` section controls how the AI analyzes your profiles:

```json
"model_config": {
    "model": "gpt-4-turbo-preview",
    "max_tokens": 4096,
    "temperature": 0.7,
    "top_p": 1.0,
    "general_analyze_prompt_location": "path/to/your/system_prompt.txt"
}
```

### Benchmark Configurations

The `benchmark_configs` section lets you customize analysis for each benchmark:

```json
"benchmark_configs": {
    "BenchmarkGenPool": {
        "prefixes": [
            "github.com/example/GenPool",
            "github.com/example/GenPool/internal",
            "github.com/example/GenPool/pkg"
        ],
        "ignore": "init,TestMain,BenchmarkMain"
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

## AI Analysis

The profiler uses AI to analyze benchmark profiles, providing insights into performance patterns and bottlenecks. Each profile (CPU, memory, mutex) is analyzed individually using data from `text/Benchmark_profile.txt`, with results saved in `bench/tag/AI`.

### Usage

Enable AI analysis by adding the `-general_analyze` flag:

```bash
prof -benchmarks "[BenchmarkGenPool]" -profiles "[cpu,memory]" -tag "test1" -general_analyze
```

### Customization

1. **Custom Prompts**

   - Set `general_analyze_prompt_location` in your config (file location)
   - Create tailored prompts for specific analysis needs (e.g., performance aspects, baseline comparisons)

2. **Model Settings**
   ```json
   "model_config": {
       "model": "gpt-4-turbo-preview",
       "max_tokens": 4096,
       "temperature": 0.7,
       "top_p": 1.0
   }
   ```

## Contribution

Bring your skills, share your ideas, and contribute your code.

## Installation

The profiler can be installed using our installation script. The script will:

1. Clone the repository to `~/.prof`
2. Set up a Python virtual environment
3. Install all required dependencies
4. Create a wrapper script in `~/bin`

### Prerequisites

- Python 3.12.10+

### Quick Install

Run this command in your terminal:

```bash
curl -sSL https://raw.githubusercontent.com/AlexsanderHamir/prof/main/install.sh | bash
```

### Post-Installation

After installation, you need to add `~/bin` to your PATH. Add this line to your shell configuration file (`.zshrc`, `.bashrc`, etc.):

```bash
export PATH="$HOME/bin:$PATH"
```

Then either:

- Restart your terminal, or
- Run: `source ~/.zshrc` (or your shell's config file)

### Verification

To verify the installation, run:

```bash
prof
```

If you see the `Error: Missing required arguments:`, the installation was successful!

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.
