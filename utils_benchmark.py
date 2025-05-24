import json
import os
from pathlib import Path
import re
import shutil
import subprocess
import sys
import time
from typing import Dict, List, Optional, Tuple, Set, Pattern
from config_manager import ConfigManager, ConfigurationError
from dataclasses import dataclass

PROFILE_FLAGS: Dict[str, str] = {
    "cpu": "-cpuprofile=cpu.out",
    "memory": "-memprofile=memory.out",
    "mutex": "-mutexprofile=mutex.out",
    "trace": "-trace=trace.out"
}

PPROF_TEXT_PARAMS = [
    "-nodecount=1000000", "-cum", "-edgefraction=0", "-nodefraction=0", "-top"
]


@dataclass
class BenchmarkConfig:
    benchmark_name: str
    profiles: List[str]
    iteration_count: int
    tag: str


@dataclass
class ProfileAnalysisConfig:
    """Configuration for profile analysis."""
    function_prefixes: List[str]
    ignore_functions: Set[str]


@dataclass
class ProfilePaths:
    """Paths for profile analysis files."""
    profile_text: Path
    profile_binary: Path
    output: Path


class BenchmarkError(Exception):
    """Custom exception for benchmark-related errors."""
    pass


def setup_from_current_directory():
    template_path = os.path.join(os.getcwd(), "config_template.json")
    if os.path.exists(template_path):
        print(
            "\nFound config_template.json in current directory. Attempting automatic setup..."
        )
        try:
            ConfigManager.setup_from_file(template_path)
            print("Automatic configuration completed successfully!")
        except (ValueError, ConfigurationError) as e:
            print_configuration_error(
                f"Error during automatic configuration: {e}")
            raise  # Let the caller handle the error
    else:
        print_configuration_error(
            "Error: Configuration not found. Please run setup first:")
        raise ConfigurationError(
            "Configuration not found. Please run setup first.")


def validate_arguments(args) -> Tuple[List[str], List[str], Optional[Dict]]:
    benchmarks = parse_list_argument(args.benchmarks)
    profiles = parse_list_argument(args.profiles)

    try:
        config = ConfigManager.load()
        benchmark_config = {}
        for benchmark in benchmarks:
            if benchmark in config.benchmark_configs:
                bench_config = config.benchmark_configs[benchmark]
                benchmark_config[benchmark] = {
                    "prefixes": bench_config.prefixes,
                    "ignore": bench_config.ignore
                }
    except ConfigurationError as e:
        print(f"Error loading benchmark configuration: {e}", file=sys.stderr)
        raise  # Let the caller handle the error

    return benchmarks, profiles, benchmark_config


def setup_directories(tag: str, benchmarks: List[str],
                      profiles: List[str]) -> None:
    create_bench_directories(tag, benchmarks)
    create_profile_function_directories(tag, profiles, benchmarks)


def create_bench_directories(tag: str, benchmarks: List[str]):
    bench_dir = "bench"
    tag_dir = os.path.join(bench_dir, tag)
    bin_dir = os.path.join(tag_dir, "bin")
    text_dir = os.path.join(tag_dir, "text")
    description_file = os.path.join(tag_dir, "description.txt")

    try:
        if not os.path.exists(bench_dir):
            os.makedirs(bench_dir)
            print(f"Created directory: {bench_dir}")
        else:
            print(f"Directory '{bench_dir}' already exists")

        if os.path.exists(tag_dir):
            print(f"Directory '{tag_dir}' already exists, cleaning it...")
            clean_directory(tag_dir)

        os.makedirs(bin_dir)
        os.makedirs(text_dir)

        for benchmark in benchmarks:
            os.makedirs(os.path.join(bin_dir, benchmark))
            os.makedirs(os.path.join(text_dir, benchmark))

        with open(description_file, 'w') as f:
            pass

        print(f"Created directory structure: {tag_dir}")
        print(f"  - {bin_dir} (with benchmark subdirectories)")
        print(f"  - {text_dir} (with benchmark subdirectories)")
        print(f"  - {description_file}")

    except Exception as e:
        print(f"Error creating directories: {e}", file=sys.stderr)
        os.exit(1)


def create_profile_function_directories(tag: str, profiles: List[str],
                                        benchmarks: List[str]):
    tag_dir = os.path.join("bench", tag)

    pprof_profiles = [p for p in profiles if p != "trace"]

    for profile in pprof_profiles:
        profile_dir = os.path.join(tag_dir, f"{profile}_functions")
        os.makedirs(profile_dir, exist_ok=True)

        # Create benchmark subdirectories
        for benchmark in benchmarks:
            benchmark_dir = os.path.join(profile_dir, benchmark)
            os.makedirs(benchmark_dir, exist_ok=True)

    print("Created profile function directories")


def print_configuration(benchmarks: List[str], profiles: List[str], tag: str,
                        count: int, benchmark_config: Optional[Dict]) -> None:
    print("\nParsed arguments:")
    print(f"Benchmarks: {benchmarks}")
    print(f"Profiles: {profiles}")
    print(f"Tag: {tag}")
    print(f"Count: {count}")
    if benchmark_config:
        print("\nBenchmark configurations:")
        for benchmark, config in benchmark_config.items():
            print(f"  {benchmark}:")
            print(f"    Prefixes: {', '.join(config['prefixes'])}")
            if config.get("ignore"):
                print(f"    Ignore: {config['ignore']}")
    else:
        print(
            "\nNo benchmark configuration found in config file - analyzing all functions"
        )


def run_benchmarks_and_process_profiles(
        benchmarks: List[str], profiles: List[str], count: int, tag: str,
        benchmark_config: Optional[Dict]) -> None:
    print("\nRunning benchmarks sequentially...")
    for benchmark in benchmarks:
        run_benchmark(benchmark, profiles, count, tag)

    print("\nProcessing profiles...")
    for benchmark in benchmarks:
        process_profiles(benchmark, profiles, tag)

    print("\nAnalyzing profile functions...")
    analyze_profile_functions(tag, profiles, benchmarks, benchmark_config)

    print("\nAll benchmarks and profile processing completed successfully!")


def run_benchmark(benchmark: str, profiles: List[str], count: int,
                  tag: str) -> None:
    config = BenchmarkConfig(benchmark, profiles, count, tag)

    # Build command and setup directories
    cmd = build_benchmark_command(config)
    _, text_dir, bin_dir = setup_output_directories(config.benchmark_name,
                                                    config.tag)

    # Run benchmark and capture output
    output_file = text_dir / f"{config.benchmark_name}.txt"
    run_benchmark_command(cmd, output_file)

    # Move generated files
    move_profile_files(config.benchmark_name, config.profiles, bin_dir)
    move_test_files(config.benchmark_name, bin_dir)

    print(f"Completed benchmark: {config.benchmark_name}")


def wait_for_profile_file(profile_file: str, timeout: int = 5) -> bool:
    """Wait for a profile file to be written and become non-empty.
    Returns True if the file exists and is non-empty, False if timeout is reached."""
    start_time = time.time()
    while time.time() - start_time < timeout:
        if os.path.exists(profile_file) and os.path.getsize(profile_file) > 0:
            return True
        time.sleep(0.1)
    return False


def run_pprof_command(
        cmd: List[str],
        output_path: Path,
        binary_mode: bool = False) -> subprocess.CompletedProcess:
    """Run a pprof command and write output to a file.

    Args:
        cmd: List of command arguments
        output_path: Path to write output to
        binary_mode: Whether to open file in binary mode

    Returns:
        CompletedProcess object from subprocess.run

    Raises:
        RuntimeError: If command fails
    """
    mode = 'wb' if binary_mode else 'w'
    try:
        with open(output_path, mode) as f:
            process = subprocess.run(cmd,
                                     stdout=f,
                                     stderr=subprocess.PIPE,
                                     text=not binary_mode,
                                     check=True)
        return process
    except subprocess.CalledProcessError as e:
        error_msg = e.stderr.decode() if isinstance(e.stderr,
                                                    bytes) else e.stderr
        raise RuntimeError(f"pprof command failed: {error_msg}")


def generate_text_profile(profile_file: Path, output_file: Path) -> None:
    """Generate text profile analysis using pprof.

    Args:
        profile_file: Path to the profile file
        output_file: Path to write the text analysis to

    Raises:
        RuntimeError: If profile generation fails
    """
    cmd = ["go", "tool", "pprof", *PPROF_TEXT_PARAMS, str(profile_file)]
    run_pprof_command(cmd, output_file)


def generate_png_visualization(profile_file: Path, output_file: Path) -> None:
    """Generate PNG visualization using pprof.

    Args:
        profile_file: Path to the profile file
        output_file: Path to write the PNG to

    Raises:
        RuntimeError: If PNG generation fails
    """
    cmd = ["go", "tool", "pprof", "-png", str(profile_file)]
    run_pprof_command(cmd, output_file, binary_mode=True)


def process_profile(profile: str, benchmark: str, profile_file: Path,
                    text_dir: Path, profile_functions_dir: Path) -> None:
    if not profile_file.exists():
        print(f"Warning: Profile file not found: {profile_file}",
              file=sys.stderr)
        return

    output_file = text_dir / f"{benchmark}_{profile}.txt"
    png_file = profile_functions_dir / f"{benchmark}_{profile}.png"

    try:
        # Generate text profile
        generate_text_profile(profile_file, output_file)
        print(f"Processed {profile} profile for {benchmark}")

        # Generate PNG visualization
        generate_png_visualization(profile_file, png_file)
        print(
            f"Generated PNG visualization for {profile} profile of {benchmark} in {profile_functions_dir}"
        )

    except RuntimeError as e:
        print(f"Error processing {profile} profile for {benchmark}: {e}",
              file=sys.stderr)
        raise


def process_profiles(benchmark: str, profiles: List[str], tag: str) -> None:
    tag_dir = Path("bench") / tag
    bin_dir = tag_dir / "bin" / benchmark
    text_dir = tag_dir / "text" / benchmark

    # Skip trace profile as it's not processed with pprof
    pprof_profiles = [p for p in profiles if p != "trace"]

    for profile in pprof_profiles:
        profile_file = bin_dir / f"{benchmark}_{profile}.out"
        profile_functions_dir = tag_dir / f"{profile}_functions" / benchmark

        try:
            process_profile(profile, benchmark, profile_file, text_dir,
                            profile_functions_dir)
        except RuntimeError as e:
            print(f"Error processing profiles: {e}", file=sys.stderr)
            raise  # Let the caller handle the error


def get_profile_analysis_config(
    benchmark: str, benchmark_config: Optional[Dict[str, Dict[str, str]]]
) -> ProfileAnalysisConfig:
    """Get configuration for profile analysis of a specific benchmark.
    
    Args:
        benchmark: Name of the benchmark
        benchmark_config: Optional benchmark configuration dictionary
        
    Returns:
        ProfileAnalysisConfig with prefixes and ignore functions
    """
    config = benchmark_config.get(benchmark, {}) if benchmark_config else {}
    return ProfileAnalysisConfig(
        function_prefixes=config.get("prefixes", []),
        ignore_functions=set(parse_list_argument(config.get("ignore", "")))
        if config.get("ignore") else set())


def get_profile_paths(tag: str, benchmark: str, profile: str) -> ProfilePaths:
    """Get all relevant paths for profile analysis.
    
    Args:
        tag: Analysis tag
        benchmark: Benchmark name
        profile: Profile type
        
    Returns:
        ProfilePaths containing all necessary file paths
    """
    tag_dir = Path("bench") / tag
    return ProfilePaths(profile_text=tag_dir / "text" / benchmark /
                        f"{benchmark}_{profile}.txt",
                        profile_binary=tag_dir / "bin" / benchmark /
                        f"{benchmark}_{profile}.out",
                        output=tag_dir / f"{profile}_functions" / benchmark)


def extract_function_name(line: str, function_prefixes: List[str],
                          ignore_functions: Set[str]) -> Optional[str]:
    """Extract and validate a function name from a profile line.
    
    Args:
        line: Profile line to parse
        function_prefixes: List of allowed function prefixes
        ignore_functions: Set of functions to ignore
        
    Returns:
        Extracted function name if valid, None otherwise
    """
    parts = line.split()
    if len(
            parts
    ) < 6:  # Need at least 6 columns (flat, flat%, sum%, cum, cum%, function)
        return None

    func_name = " ".join(parts[5:])

    # Check prefixes if specified
    if function_prefixes and not any(prefix in func_name
                                     for prefix in function_prefixes):
        return None

    # Extract function name after last dot
    match = re.search(r'\.([^.(]+)(?:\([^)]*\))?$', func_name)
    if not match:
        return None

    func_name = match.group(1).strip().replace(" ", "")
    return func_name if func_name and func_name not in ignore_functions else None


def extract_functions_from_profile(profile_text_file: Path,
                                   config: ProfileAnalysisConfig) -> Set[str]:
    """Extract function names from a profile text file.
    
    Args:
        profile_text_file: Path to the profile text file
        config: Profile analysis configuration
        
    Returns:
        Set of extracted function names
        
    Raises:
        FileNotFoundError: If profile file doesn't exist
        RuntimeError: If profile file is invalid
    """
    if not profile_text_file.exists():
        raise FileNotFoundError(
            f"Profile text file not found: {profile_text_file}")

    functions = set()
    found_header = False

    try:
        with open(profile_text_file, 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue

                if "flat  flat%   sum%        cum   cum%" in line:
                    found_header = True
                    continue

                if not found_header:
                    continue

                if func_name := extract_function_name(line,
                                                      config.function_prefixes,
                                                      config.ignore_functions):
                    functions.add(func_name)

        if not found_header:
            raise RuntimeError("Profile file is invalid: header not found")

        return functions

    except Exception as e:
        raise RuntimeError(
            f"Error reading profile file {profile_text_file}: {e}")


def analyze_single_function(func: str, paths: ProfilePaths) -> None:
    """Analyze a single function using pprof.
    
    Args:
        func: Function name to analyze
        paths: ProfilePaths containing all necessary file paths
        
    Raises:
        RuntimeError: If analysis fails
    """
    output_file = paths.output / f"{func}.txt"
    cmd = ["go", "tool", "pprof", f"-list={func}", str(paths.profile_binary)]

    try:
        with open(output_file, 'w') as f:
            subprocess.run(cmd,
                           stdout=f,
                           stderr=subprocess.PIPE,
                           text=True,
                           check=True)
        print(f"Analyzed function {func}")
    except subprocess.CalledProcessError as e:
        raise RuntimeError(f"Error analyzing function {func}: {e.stderr}")


def analyze_profile_functions(
        tag: str,
        profiles: List[str],
        benchmarks: List[str],
        benchmark_config: Optional[Dict[str, Dict[str, str]]] = None) -> None:
    pprof_profiles = [p for p in profiles if p != "trace"]

    for profile in pprof_profiles:
        for benchmark in benchmarks:
            try:
                config = get_profile_analysis_config(benchmark,
                                                     benchmark_config)
                paths = get_profile_paths(tag, benchmark, profile)

                paths.output.mkdir(parents=True, exist_ok=True)

                functions = extract_functions_from_profile(
                    paths.profile_text, config)
                for func in functions:
                    try:
                        analyze_single_function(func, paths)
                    except RuntimeError as e:
                        print(
                            f"Error analyzing function {func} for {benchmark} ({profile}): {e}"
                        )
                        continue

            except FileNotFoundError as e:
                print(f"Error processing {benchmark} ({profile}): {e}")
                continue
            except Exception as e:
                print(f"Error processing {benchmark} ({profile}): {e}")
                sys.exit(1)


def print_configuration_error(error_msg: Optional[str] = None) -> None:
    if error_msg:
        print(f"\n{error_msg}", file=sys.stderr)

    print("\nPlease set up configuration manually:", file=sys.stderr)
    print("1. Create a template config file:", file=sys.stderr)
    print(
        "   prof setup --create-template [--output-path path/to/config.json]",
        file=sys.stderr)
    print("2. Use an existing config file (after creating template as well):",
          file=sys.stderr)
    print("   prof setup --config path/to/your/config.json", file=sys.stderr)


def setup_command(args):
    if args.create_template:
        ConfigManager.create_template(args.output_path)
        print("\nTemplate configuration file created successfully!")
    else:
        print(
            "\nError: Please use --create-template to create a configuration template",
            file=sys.stderr)
        sys.exit(1)


def check_required_args(args) -> bool:
    """Check if all required arguments are provided and print any missing ones.

    Args:
        args: Parsed command line arguments

    Returns:
        bool: True if all required arguments are present, False otherwise
    """
    missing_args = []
    if not args.benchmarks:
        missing_args.append("benchmarks")
    if not args.profiles:
        missing_args.append("profiles")
    if not args.tag:
        missing_args.append("tag")
    if not args.count:
        missing_args.append("count")

    if missing_args:
        print("\nError: Missing required arguments:", file=sys.stderr)
        for arg in missing_args:
            print(f"  - {arg}", file=sys.stderr)
        print("\nPlease provide all required arguments.\n", file=sys.stderr)
        return False
    return True


def cleanup_tag_directory(tag: str):
    """Clean up the tag directory if it exists."""
    tag_dir = os.path.join("bench", tag)
    if os.path.exists(tag_dir):
        try:
            shutil.rmtree(tag_dir)
            print(f"Cleaned up tag directory: {tag_dir}")
        except Exception as e:
            print(f"Error cleaning up tag directory {tag_dir}: {e}",
                  file=sys.stderr)


def clean_directory(directory: str):
    """Delete all contents of a directory if it exists."""
    if os.path.exists(directory):
        try:
            # Remove all contents of the directory
            for item in os.listdir(directory):
                item_path = os.path.join(directory, item)
                if os.path.isfile(item_path):
                    os.remove(item_path)
                elif os.path.isdir(item_path):
                    shutil.rmtree(item_path)
            print(f"Cleaned directory: {directory}")
        except Exception as e:
            print(f"Error cleaning directory {directory}: {e}",
                  file=sys.stderr)
            raise RuntimeError(f"Failed to clean directory {directory}: {e}")


def parse_list_argument(arg: str) -> List[str]:
    # Remove brackets if present
    arg = arg.strip('[]')
    # Split by comma and strip whitespace
    return [item.strip() for item in arg.split(',')]


def parse_benchmark_config(config_str: str) -> Dict[str, Dict[str, str]]:
    """Parse a JSON-like string containing benchmark-specific configurations.
    Example format: '{"BenchmarkGenPool":{"prefix":"github.com/AlexsanderHamir/GenPool","ignore":"func1,performWorkload"}}'"""
    try:
        # Replace single quotes with double quotes for valid JSON
        config_str = config_str.replace("'", '"')
        config = json.loads(config_str)

        # Validate the structure
        for benchmark, settings in config.items():
            if not isinstance(settings, dict):
                raise ValueError(
                    f"Invalid settings format for benchmark {benchmark}")
            if "prefix" not in settings:
                raise ValueError(f"Missing 'prefix' for benchmark {benchmark}")
            if "ignore" in settings and not isinstance(settings["ignore"],
                                                       str):
                raise ValueError(
                    f"'ignore' must be a string for benchmark {benchmark}")

        return config
    except json.JSONDecodeError as e:
        raise ValueError(f"Invalid JSON format: {e}")
    except Exception as e:
        raise ValueError(f"Error parsing benchmark config: {e}")


def build_benchmark_command(config: BenchmarkConfig) -> List[str]:
    cmd = [
        "go", "test", "-run=^$", f"-bench=^{config.benchmark_name}$",
        "-benchmem", f"-count={config.iteration_count}"
    ]

    # Add requested profile flags
    for profile in config.profiles:
        if profile in PROFILE_FLAGS:
            cmd.append(PROFILE_FLAGS[profile])

    return cmd


def setup_output_directories(benchmark_name: str,
                             tag: str) -> tuple[Path, Path, Path]:
    tag_dir = Path("bench") / tag
    text_dir = tag_dir / "text" / benchmark_name
    bin_dir = tag_dir / "bin" / benchmark_name

    # Create directories if they don't exist
    text_dir.mkdir(parents=True, exist_ok=True)
    bin_dir.mkdir(parents=True, exist_ok=True)

    return tag_dir, text_dir, bin_dir


def run_benchmark_command(cmd: List[str], output_file: Path) -> None:
    """Executes the benchmark command and writes output to file.

    Args:
        cmd: Command to execute
        output_file: Path to write output

    Raises:
        BenchmarkError: If benchmark execution fails
    """
    try:
        with open(output_file, 'w') as f:
            process = subprocess.run(cmd,
                                     stdout=f,
                                     stderr=subprocess.STDOUT,
                                     text=True,
                                     check=True)
    except subprocess.CalledProcessError as e:
        with open(output_file, 'r') as f:
            error_output = f.read()
        raise BenchmarkError(
            f"Benchmark failed with exit code {e.returncode}:\n{error_output}")


def move_profile_files(benchmark_name: str, profiles: List[str],
                       bin_dir: Path) -> None:
    """Moves generated profile files to the benchmark directory.

    Args:
        benchmark_name: Name of the benchmark
        profiles: List of requested profiles
        bin_dir: Directory to move files to
    """
    for profile in profiles:
        if profile not in PROFILE_FLAGS:
            continue

        profile_file = Path(PROFILE_FLAGS[profile].split('=')[1])
        if not profile_file.exists():
            continue

        if not wait_for_profile_file(str(profile_file)):
            print(
                f"Warning: Profile file {profile_file} was not fully written within timeout",
                file=sys.stderr)
            continue

        new_profile_file = bin_dir / f"{benchmark_name}_{profile}.out"
        shutil.move(str(profile_file), str(new_profile_file))


def move_test_files(benchmark_name: str, bin_dir: Path) -> None:
    """Moves generated .test files to the benchmark directory.

    Args:
        benchmark_name: Name of the benchmark
        bin_dir: Directory to move files to
    """
    for item in Path('.').glob('*.test'):
        if not wait_for_profile_file(str(item)):
            print(
                f"Warning: Test file {item} was not fully written within timeout",
                file=sys.stderr)
            continue

        new_test_file = bin_dir / f"{benchmark_name}_{item.name}"
        shutil.move(str(item), str(new_test_file))
        print(f"Moved test file: {item} -> {new_test_file}")
