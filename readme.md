# Go Benchmark Profiler

![Build](https://github.com/AlexsanderHamir/prof/actions/workflows/test.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Issues](https://img.shields.io/github/issues/AlexsanderHamir/Prof)
![Last Commit](https://img.shields.io/github/last-commit/AlexsanderHamir/Prof)
![Code Size](https://img.shields.io/github/languages/code-size/AlexsanderHamir/Prof)
![Version](https://img.shields.io/github/v/tag/AlexsanderHamir/Prof?sort=semver)


This tool simplifies complex performance analysis by consolidating multiple pprof commands into a single step. It automatically collects all relevant profiling data, organizes it, makes it searchable within your workspace, and enhances the process with AI-powered insights.

[Example Profile Analysis Video](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/prof.mp4)

## Why Use This Tool?

### The Manual Way vs. Our Way

**Traditional Manual Profiling:**

```bash
# Run benchmarks with profiling
go test -run=^$ -bench=^BenchmarkGenPool$ -count 5 -benchmem -cpuprofile=cpu.out -memprofile=mem.out -trace=trace.out

# Analyze CPU profile
go tool pprof -nodecount=1000 -cum -edgefraction=0 -nodefraction=0 cpu.out

# For each function you want to inspect
list .*pool.*Get
list .*pool.*Put
# ... repeat for every function

# Analyze memory profile
go tool pprof -nodecount=1000 -cum -edgefraction=0 -nodefraction=0 mem.out
# ... repeat the same process

# Organize results manually
# Create directories, move files, document changes
```

**With Our Tool:**

```bash
# One command does everything
prof -benchmarks "[BenchmarkGenPool]" -profiles "[cpu,memory]" -tag "initialBench" -count 5
```

### What You Get

1. **Complete Data Collection**: Automatically collects all profiling data you'd ever need, including code-line level mapping for every function
2. **Organized Workspace**: All files are automatically organized into tagged directories with clear structure
3. **Searchable Results**: Instead of running multiple pprof commands, just search your workspace for function names
4. **Documentation**: Built-in description files help you document changes and their performance impact
5. **AI-Powered Insights**: Optional AI analysis provides intelligent performance recommendations

### Real-World Benefits

- **Save Hours**: What takes 30+ minutes manually of annoying back and forth work becomes a single command of a couple seconds
- **Never Miss Data**: Automatic collection ensures you have all the profiling information you need
- **Track Progress**: Tagged directories and description files help you document performance improvements
- **Team Collaboration**: Organized, searchable results make it easy to share findings with your team
- **Reproducible Analysis**: Consistent data collection and organization across all profiling sessions

## Table of Contents

[Why Automate Profiling?](#why-automate-profiling)

[Usage](#usage)

[Output Examples](#output-examples)

[Configuration](#configuration)

[Installation](#installation)

[AI Analysis](#ai-analysis)

[Contribution](#contribution)


## 🔁 Why Automate Profiling?

Manually running `pprof`, filtering functions, and inspecting profiles is time-consuming, and during a long profiling section it becomes error prone.

### 🛠️ The Traditional Workflow

```bash
# Run benchmarks and generate profiles
go test -bench=^BenchmarkGenPool$ -count 5 -benchmem \
  -cpuprofile=cpu.out -memprofile=mem.out -trace=trace.out

# Inspect profiles manually
go tool pprof cpu.out
list .*pool.*Get
list .*pool.*Put
# Repeat for every function of interest...
```

### ✅ The Automated Way

```bash
prof -benchmarks "[BenchmarkGenPool]" -profiles "[cpu,memory]" -tag "initialBench" -count 5
```

## ⚙️ What You Get with One Command

* **Comprehensive Profiling** – Automatically captures CPU, memory, and mutex profiles text files, along with code-level performance data for every function based on your configuration.
* **Structured Output** – Results saved under clean, tagged directories.
* **Quick Search** – Use `Cmd+P` in VSCode to jump to any function.
* **Documentation (Optional)** – Creates documentation text files so you can add context for each tag.
* **AI Insights (Optional)** – Get summaries and recommendations using your own prompts.

## 🌟 Key Benefits

* ⏱️ **Faster Iteration** – From hours to seconds
* 📤 **Team-Friendly** – Share clean, consistent results
* 🧠 **Codebase Snapshots** – Capture performance state with minimal config

## 🧩 Bonus Features

* 🔍 **Scoped Analysis** – Target specific packages/functions or exclude noise
* 🤖 **AI-Driven Reports** – Automated interpretations tailored to your needs

## Usage

> ⚠️ Always run commands from the directory where your benchmark file is located.

### Step 1: Create a Template Configuration

Generate a starter config file with:

```bash
prof setup --create-template
```

### Stage 2: Customize the Configuration File

The configuration dictates what will be collected from `pprof`, and what the AI should analyze.

See [Configuration](#configuration), and [AI Analysis](#ai-analysis).


### Step 3: Run Benchmarks and Collect Profiles

Use the following command to run benchmarks, collect profiles, and store results:

```bash
prof -benchmarks "[BenchmarkGenPool, BenchmarkSyncPool]" -profiles "[cpu,memory,mutex]" -tag "test1" -count 1
```

This command:

1. Runs each specified benchmark (`-benchmarks`) the given number of times (`-count`).
2. Collects the selected profiles (e.g., CPU, memory, mutex).
3. Saves results in a directory named after the tag (`test1`).
4. Extracts and stores line-level code mappings for all functions in each profile.

### Check Version Information

To check your current version and see if updates are available:

```bash
prof -version
```

Example output:

```
Current version: 1.0.25
Latest version: v1.0.25 (up to date)
```

## Output Examples

Want to see what the output looks like before running the tool? Check out the [`output_example/bench/`](output_example/bench/) directory in this repository, which contains real examples of the output.

## Configuration

The configuration file (`config_template.json`) controls the data to be collected, and AI behavior. Here's a detailed breakdown of each section:

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
    "prompt_location": "path/to/your/system_prompt.txt"
}
```

### Benchmark Configurations
The `benchmark_configs` section lets you control which functions are included for collecting code-line performance data and which ones to exclude.

1. If you provide one or more **prefixes**, only functions whose path contain those prefixes will be included into your workspace.

2. Even when using prefixes, you can explicitly **ignore** specific functions by name (matching the part after the last dot). For example, in `github.com/AlexsanderHamir/GenPool/pool.BenchmarkGenPool.func1`, specifying `func1` in the ignore list will exclude that function—even though it matches the prefix `github.com/example/GenPool`.

The `benchmark_configs` section lets you customize analysis for each benchmark:

```json
"benchmark_configs": {
    "BenchmarkGenPool": {
        "prefixes": [
          // only functions with this prefixes will be collected
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
   - Example: `"github.com/your-project/core"`

2. **Ignore**:
   - Comma-separated list of function names to exclude
   - Example: `"init,TestMain,BenchmarkMain,setup,teardown"`

### Example Use Cases:

```json
{
  "api_key": "your-api-key-here",
  "base_url": "https://api.openai.com/v1",
  "model_config": {
    "model": "gpt-4-turbo-preview",
    "max_tokens": 4096,
    "temperature": 0.7,
    "prompt_location": "./prompts/custom_analysis.txt"
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

Enable AI analysis by adding the `-general_analyze` or `-flag_profiles` flag:

```bash
prof -benchmarks "[BenchmarkGenPool]" -profiles "[cpu,memory]" -tag "test1" -general_analyze
```

**Analysis Modes:**

1. **`-general_analyze`**: Creates a separate analysis file per profile containing AI insights based on your custom prompt
2. **`-flag_profiles`**: Rewrites the original profile files (`text/Benchmark_profile.txt`) with AI-enhanced content according to your prompt

### Configuration

The `ai_config` section in your configuration file controls which benchmarks and profiles are analyzed, as well as how the data is filtered before being sent to the AI.

#### Basic Configuration

```json
"ai_config": {
    "all_benchmarks": true,
    "all_profiles": true,
    "specific_benchmarks": [],
    "specific_profiles": [],
    "universal_profile_filter": {
        "profile_values": {
            "flat": 0.0,
            "flat%": 0.0,
            "sum%": 0.0,
            "cum": 0.0,
            "cum%": 0.0
        },
        "ignore_functions": ["init", "TestMain", "BenchmarkMain"],
        "ignore_prefixes": ["github.com/example/BenchmarkName"]
    }
}
```

#### Configuration Options

**Benchmark and Profile Selection:**

- **`all_benchmarks`** (boolean): When `true`, analyzes all benchmarks found in the tag directory. When `false`, only analyzes benchmarks listed in `specific_benchmarks`
- **`all_profiles`** (boolean): When `true`, analyzes all profile types (cpu, memory, mutex). When `false`, only analyzes profiles listed in `specific_profiles`
- **`specific_benchmarks`** (array): List of benchmark names to analyze when `all_benchmarks` is `false`
- **`specific_profiles`** (array): List of profile types to analyze when `all_profiles` is `false`

**Important Rules:**

- If `all_benchmarks` is `true`, `specific_benchmarks` must be empty
- If `all_profiles` is `true`, `specific_profiles` must be empty
- If `all_benchmarks` is `false`, you must provide `specific_benchmarks`
- If `all_profiles` is `false`, you must provide `specific_profiles`

#### Data Filtering

The `universal_profile_filter` controls which profile data is sent to the AI, helping to focus the analysis on the most relevant information.

**Profile Value Filtering:**

The `profile_values` section filters out profile entries based on their performance metrics. Any line with values less than or equal to the specified thresholds will be excluded from AI analysis.

```json
"profile_values": {
    "flat": 0.0,      // Flat time (s) - excludes functions with flat time ≤ this value
    "flat%": 0.0,     // Flat percentage - excludes functions with flat% ≤ this value
    "sum%": 0.0,      // Sum percentage - excludes functions with sum% ≤ this value
    "cum": 0.0,       // Cumulative time (s) - excludes functions with cum time ≤ this value
    "cum%": 0.0       // Cumulative percentage - excludes functions with cum% ≤ this value
}
```

**Examples:**

- `"flat%": 1.0` - Only include functions that consume more than 1% of flat time
- `"cum": 3` - Only include functions with cumulative time greater than 3s
- `"flat": 0.0` - Include all functions regardless of flat time (default behavior)

**Function Filtering:**

- **`ignore_functions`** (array): List of function names to exclude from analysis. The tool matches the function name after the last dot. For example:
  - `"math/rand.Intn"` → specify `"Intn"` to ignore
  - `"github.com/example/pkg.Pool.Get"` → specify `"Get"` to ignore
- **`ignore_prefixes`** (array): List of package prefixes to exclude. Functions from these packages will be filtered out:
  - `"github.com/example/BenchmarkName"` - excludes all functions from this package
  - `"github.com/example/BenchmarkName/internal"` - excludes internal package functions

#### Example Configurations

**Analyze Only Specific Benchmarks:**

```json
"ai_config": {
    "all_benchmarks": false,
    "all_profiles": true,
    "specific_benchmarks": ["BenchmarkGenPool", "BenchmarkSyncPool"],
    "specific_profiles": [],
    "universal_profile_filter": {
        "profile_values": {
            "flat": 0.0,
            "flat%": 0.5,
            "sum%": 0.0,
            "cum": 0.0,
            "cum%": 0.0
        },
        "ignore_functions": ["init", "TestMain"],
        "ignore_prefixes": ["runtime", "testing"]
    }
}
```

**Focus on High-Impact Functions:**

```json
"ai_config": {
    "all_benchmarks": true,
    "all_profiles": false,
    "specific_benchmarks": [],
    "specific_profiles": ["cpu", "memory"],
    "universal_profile_filter": {
        "profile_values": {
            "flat": 0.0,
            "flat%": 2.0,
            "sum%": 0.0,
            "cum": 0.0,
            "cum%": 5.0
        },
        "ignore_functions": ["init", "TestMain", "BenchmarkMain", "setup", "teardown"],
        "ignore_prefixes": ["runtime", "testing", "reflect"]
    }
}
```

**Minimal Filtering for Comprehensive Analysis:**

```json
"ai_config": {
    "all_benchmarks": true,
    "all_profiles": true,
    "specific_benchmarks": [],
    "specific_profiles": [],
    "universal_profile_filter": {
        "profile_values": {
            "flat": 0.0,
            "flat%": 0.0,
            "sum%": 0.0,
            "cum": 0.0,
            "cum%": 0.0
        },
        "ignore_functions": ["init", "TestMain"],
        "ignore_prefixes": []
    }
}
```

## Contribution

We welcome contributions of all kinds! Whether you have ideas for new features, improvements to existing functionality, bug reports, or just want to help expand this software - we'd love to hear from you. This project is actively being developed and expanded, and your input is invaluable in making it even better.

### What We're Looking For

- **Feature Ideas**: Have an idea for a new capability? We're excited to hear about it!
- **Performance Improvements**: Suggestions for making the tool faster or more efficient
- **UI/UX Enhancements**: Ways to make the tool more user-friendly
- **Documentation**: Help improve guides, examples, or code comments
- **Bug Reports**: Found an issue? Let us know so we can fix it
- **Code Contributions**: Pull requests for new features or fixes
- **Testing**: Help improve test coverage or add new test cases
- **Community**: Share how you're using the tool, provide feedback, or help others

### Getting Started

This section will help you set up your local development environment to contribute to the project.

### Prerequisites

Before setting up your development environment, ensure you have:

- **Python 3.12+** - The project requires Python 3.12.10 or higher
- **Go 1.21+** - Required for running benchmarks and testing
- **Git** - For version control

### Local Development Setup

#### 1. Clone the Repository

```bash
git clone https://github.com/AlexsanderHamir/prof.git
cd prof
```

#### 2. Set Up Python Virtual Environment

Create and activate a virtual environment:

```bash
# Create virtual environment
python3 -m venv venv

# Activate virtual environment
# On macOS/Linux:
source venv/bin/activate
# On Windows:
# venv\Scripts\activate
```

#### 3. Install Dependencies

Install the required Python packages:

```bash
pip install -r requirements.txt
```

#### 4. Set Up Local Testing

Make the `prof` script executable and create a local alias for testing:

```bash
# Make the script executable
chmod +x prof

# Create a local alias (optional, for easier testing), example:
alias profDev="/Users/alexsandergomes/Documents/prof_AI/prof"
```

#### 5. Verify Installation

Test that everything is working:

```bash
# Test the command
profDev

# You should see: "Error: Missing required arguments:" - this means it's working!
```

### Development Workflow

#### Running Tests

The project includes end-to-end and unit tests to ensure functionality:

```bash
# Run all tests
pytest

# Run tests with verbose output
pytest -v

# Run specific test file
pytest tests/e2e/benchmark_test.py
```

#### Testing Manually

To test your local changes manually, run the `profDev` command in any golang project where the benchmarks are located:

```bash
profDev -benchmarks "[BenchmarkSimple]" -profiles "[cpu,memory]" -tag "test" -count 1
```

#### Style Configuration:

- **Indent width**: 4 spaces (no tabs)
- **Column limit**: 300 characters
- **Style**: PEP 8 compliant
- **Formatter**: YAPF

For Cursor/VS Code users, you can configure your editor with:

```json
{
  "[python]": {
    "editor.formatOnSaveMode": "file",
    "editor.formatOnSave": true,
    "editor.defaultFormatter": "eeyore.yapf"
  },
  "yapf.args": [
    "--style",
    "{based_on_style: pep8, indent_width: 4, column_limit: 300}"
  ]
}
```

#### Making Changes

1. Create a new branch for your feature/fix:

   ```bash
   git checkout -b feature/your-feature-name
   ```

2. Make your changes and test them thoroughly

3. Run the test suite to ensure nothing is broken:

   ```bash
   pytest
   ```

4. Commit your changes with clear, descriptive commit messages

5. Push your branch and create a pull request

### Project Structure

Understanding the project structure will help you contribute effectively:

```
prof_AI/
├── prof                    # Main executable script
├── cli/                    # Command-line interface modules
│   ├── interface.py        # Argument parsing and main CLI logic
│   └── helpers.py          # CLI helper functions
├── analyzer/               # Profile analysis modules
│   ├── interface.py        # Analysis interface
│   └── helpers.py          # Analysis helper functions
├── config/                 # Configuration management
│   ├── config_manager.py   # Configuration handling
│   └── helpers.py          # Config helper functions
├── tests/                  # Test suite
│   └── e2e/               # End-to-end tests
├── requirements.txt        # Python dependencies
└── install.sh             # Installation script
```

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
