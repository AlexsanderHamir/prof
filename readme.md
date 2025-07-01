# Go Benchmark Profiler

![Build](https://github.com/AlexsanderHamir/prof/actions/workflows/test.yml/badge.svg)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
![Issues](https://img.shields.io/github/issues/AlexsanderHamir/Prof)
![Last Commit](https://img.shields.io/github/last-commit/AlexsanderHamir/Prof)
![Code Size](https://img.shields.io/github/languages/code-size/AlexsanderHamir/Prof)
![Version](https://img.shields.io/github/v/tag/AlexsanderHamir/Prof?sort=semver)

**Go Benchmark Profiler** is a high-efficiency tool for automated performance profiling in Go. It simplifies and accelerates benchmarking workflows by:

* 🔁 Combining multiple `pprof` steps into a single command
* 📁 Organizing outputs by benchmark, tag, and profile type
* 🧠 Generating AI-powered insights based on custom prompts
* 🔍 Making performance data searchable and shareable

[▶ Example Profile Analysis Video](https://cdn.jsdelivr.net/gh/AlexsanderHamir/assets@main/prof.mp4)

## Table of Contents

[Why Use This Tool ?](#why-use-this-tool)

[Usage](#usage)

[Output Examples](#output-examples)

[Configuration](#configuration)

[Installation](#installation)

[AI Analysis](#ai-analysis)

[Contribution](#contribution)


## Why Use This Tool?

Skip the manual grind. Instead of running multiple profiling commands, filtering through function lists, and manually inspecting each one...

### The Traditional Way:

```bash
# Run benchmarks and generate profiles
go test -bench=^BenchmarkGenPool$ -count 5 -benchmem -cpuprofile=cpu.out -memprofile=mem.out -trace=trace.out

# Manually inspect each profile
go tool pprof cpu.out
list .*pool.*Get
list .*pool.*Put
# Repeat for every function...
````

### With This Tool:

```bash
# One command does everything
prof -benchmarks "[BenchmarkGenPool]" -profiles "[cpu,memory]" -tag "initialBench" -count 5
```

### What You Get in One Command

* **Complete Profiling** – Automatically collects CPU, memory, and code-line-level data for every function.
* **Organized Output** – Tagged directories with clear structure for easy navigation.
* **Searchable Workspace** – Just use Command+P (in VSCode or similar) to find any function by name.
* **Built-in Docs** – Description files help you log what changed and how it affected performance.
* **Optional AI Insights** – Get intelligent analysis tailored to your prompt.

### Real-World Benefits

* ⏱️ **Save Time** – Replace hours of manual inspection with seconds.
* 🤝 **Collaborate Easily** – Share structured results with your team.
* 🧠 **Snapshot Your Codebase** – Not specifying the functions or prefixes to ignore will result in a codebase performance snapshot.

### Bonus Features

* **Customizable Function Scope** – Collect all functions by default, or filter by prefix and exclusion rules.
* **AI-Powered Analysis** – Plug in your own prompts to generate focused performance insights.

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


```json
"benchmark_configs": {
    "BenchmarkGenPool": {
        "prefixes": [
            "github.com/example/GenPool",
            "github.com/example/GenPool/internal",
            "github.com/example/GenPool/pkg"
        ],
        "ignore": "init,TestMain,BenchmarkMain,func1"
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

**Important Rules:**

- If `all_benchmarks` is `true`, `specific_benchmarks` must be empty (vice-versa)
- If `all_profiles` is `true`, `specific_profiles` must be empty (vice-versa)

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

## Contribution

Share your ideas !!

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

To test the your local changes manually, run the `profDev` command in any golang project where the benchmarks are located:

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
├── cli/                    # Command-line module
│   ├── interface.py        # Main logic
│   └── helpers.py          # helper functions
├── analyzer/               # AI analysis module
│   ├── interface.py        
│   └── helpers.py          
├── config/                 # Configuration management module
│   ├── config_manager.py   
│   └── helpers.py
├── parser/                 # Text parser
│   ├── interface.py      
│   └── helpers.py          
├── tests/                  # Test module
│   └── e2e/                
|   └── unit/               
├── requirements.txt        
└── install.sh              # Installation script
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
