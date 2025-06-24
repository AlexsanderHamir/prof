import argparse
import json
from pathlib import Path
import re
import shutil
import subprocess
import sys
import time
from typing import Dict, List, Optional, Tuple, Set, Any
from config.config_manager import ConfigManager
from dataclasses import dataclass

from exit_codes import BENCHMARK_DIRECTORY_UNEXPECTED_ERROR, BENCHMARK_FILE_UNEXPECTED_ERROR, CONFIG_PARSING_ERROR, EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR, EXIT_CODE_MISSING_BRACKETS, EXIT_CODE_MISSING_EMPTY_LIST, EXIT_CODE_MODULE_ERROR, EXIT_CODE_TEMPLATE_ERROR, PROFILE_FILE_INVALID_HEADER, PROFILE_FILE_MISSING, PROFILE_FILE_UNEXPECTED_ERROR

PROFILE_FLAGS: Dict[str, str] = {"cpu": "-cpuprofile=cpu.out", "memory": "-memprofile=memory.out", "mutex": "-mutexprofile=mutex.out", "trace": "-trace=trace.out"}

PPROF_TEXT_PARAMS = ["-nodecount=100000000", "-cum", "-edgefraction=0", "-nodefraction=0", "-top"]


@dataclass
class BenchmarkParamsWrapper:
    benchmark_name: str
    profiles: List[str]
    iteration_count: int
    tag: str


@dataclass
class ProfileFilter:
    function_prefixes: List[str]
    ignore_functions: Set[str]


@dataclass
class ProfilePaths:
    profile_text: Path
    profile_binary: Path
    output: Path


def config_setup():
    paths = [Path.cwd() / "config_template.json"]
    if ConfigManager._config_path:
        paths.append(ConfigManager._config_path)

    for path in paths:
        if path.exists():
            print(f"\nFound config_template.json at {path}. Attempting automatic setup...")

    ConfigManager.setup_from_file(path)
    print("Automatic configuration completed successfully!")


def validate_list_arguments(benchmarks_arg: str, profiles_arg: str) -> None:
    if benchmarks_arg.strip() == "[]":
        print("Benchmarks argument cannot be an empty list (i.e., '[]'). Please provide at least one benchmark.", file=sys.stderr)
        sys.exit(EXIT_CODE_MISSING_EMPTY_LIST)
    if profiles_arg.strip() == "[]":
        print("Profiles argument cannot be an empty list (i.e., '[]'). Please provide at least one profile.", file=sys.stderr)
        sys.exit(EXIT_CODE_MISSING_EMPTY_LIST)

    if not benchmarks_arg.strip().startswith("[") or not benchmarks_arg.strip().endswith("]"):
        print("Benchmarks argument must be wrapped in brackets (e.g., '[BenchmarkGenPool]'). Please provide a properly formatted list.", file=sys.stderr)
        sys.exit(EXIT_CODE_MISSING_BRACKETS)
    if not profiles_arg.strip().startswith("[") or not profiles_arg.strip().endswith("]"):
        print("Profiles argument must be wrapped in brackets (e.g., '[cpu,memory]'). Please provide a properly formatted list.", file=sys.stderr)
        sys.exit(EXIT_CODE_MISSING_BRACKETS)

    def check_commas(arg: str, arg_name: str):
        stripped = arg.strip()[1:-1].strip()  # remove brackets and whitespace
        if stripped:
            # Detect two or more words separated by whitespace but not by a comma
            if re.search(r'\b\w+\b\s+\b\w+\b', stripped) and ',' not in stripped:
                print(f"{arg_name} argument items must be separated by commas, not spaces. Please provide a properly comma-separated list.", file=sys.stderr)
                sys.exit(CONFIG_PARSING_ERROR)
            # Split by comma, check for empty items or consecutive/missing commas
            items = [item.strip() for item in stripped.split(",")]
            if any(not item for item in items):
                print(f"{arg_name} argument contains empty items or consecutive/missing commas. Please provide a properly comma-separated list.", file=sys.stderr)
                sys.exit(CONFIG_PARSING_ERROR)
            # Check for any item that still contains multiple words (e.g., 'foo bar')
            for item in items:
                if ' ' in item:
                    print(f"{arg_name} argument items must be single words and separated by commas. Found: '{item}'", file=sys.stderr)
                    sys.exit(CONFIG_PARSING_ERROR)

    check_commas(benchmarks_arg, "Benchmarks")
    check_commas(profiles_arg, "Profiles")


def parse_and_load_benchmark_config(args) -> Tuple[List[str], List[str], Dict[str, Dict[str, Any]]]:
    validate_list_arguments(args.benchmarks, args.profiles)
    benchmarks = parse_list_argument(args.benchmarks)
    profiles = parse_list_argument(args.profiles)
    benchmark_filters = filter_configs(benchmarks)

    return benchmarks, profiles, benchmark_filters


def filter_configs(benchmarks: List[str]) -> Dict[str, Dict[str, Any]]:
    config = ConfigManager.get_config()
    function_filter_configs: Dict[str, Dict[str, Any]] = {}
    for benchmark in benchmarks:
        if benchmark in config.benchmark_filters:
            bench_config = config.benchmark_filters[benchmark]
            function_filter_configs[benchmark] = {"prefixes": bench_config.prefixes, "ignore": bench_config.ignore}

    return function_filter_configs


def setup_directories(tag: str, benchmarks: List[str], profiles: List[str]) -> None:
    create_bench_directories(tag, benchmarks)
    create_profile_function_directories(tag, profiles, benchmarks)


def create_bench_directories(tag: str, benchmarks: List[str]):
    bench_dir = Path("bench")
    tag_dir = bench_dir / tag
    bin_dir = tag_dir / "bin"
    text_dir = tag_dir / "text"
    description_file = tag_dir / "description.txt"

    try:
        if not bench_dir.exists():
            bench_dir.mkdir()
            print(f"Created directory: {bench_dir}")
        else:
            print(f"Directory '{bench_dir}' already exists")

        if tag_dir.exists():
            print(f"Directory '{tag_dir}' already exists, cleaning it...")
            clean_directory(tag_dir)

        bin_dir.mkdir(parents=True)
        text_dir.mkdir(parents=True)

        for benchmark in benchmarks:
            (bin_dir / benchmark).mkdir(parents=True)
            (text_dir / benchmark).mkdir(parents=True)

        description_file.touch()

        print(f"Created directory structure: {tag_dir}")
        print(f"  - {bin_dir} (with benchmark subdirectories)")
        print(f"  - {text_dir} (with benchmark subdirectories)")
        print(f"  - {description_file}")

    except Exception as e:
        print(f"Error creating directories: {e}", file=sys.stderr)
        sys.exit(BENCHMARK_DIRECTORY_UNEXPECTED_ERROR)


def create_profile_function_directories(tag: str, profiles: List[str], benchmarks: List[str]):
    tag_dir = Path("bench") / tag

    pprof_profiles = [p for p in profiles if p != "trace"]

    for profile in pprof_profiles:
        profile_dir = tag_dir / f"{profile}_functions"
        profile_dir.mkdir(exist_ok=True)

        for benchmark in benchmarks:
            benchmark_dir = profile_dir / benchmark
            benchmark_dir.mkdir(exist_ok=True)

    print("Created profile function directories")


def print_configuration(benchmarks: List[str], profiles: List[str], tag: str, count: int, function_filter_configs: Dict[str, Dict[str, Any]]) -> None:
    print("\nParsed arguments:")
    print(f"Benchmarks: {benchmarks}")
    print(f"Profiles: {profiles}")
    print(f"Tag: {tag}")
    print(f"Count: {count}")
    if function_filter_configs:
        print("\nBenchmark Function Filter Configurations:")
        for benchmark, config in function_filter_configs.items():
            print(f"  {benchmark}:")
            print(f"    Prefixes: {', '.join(config['prefixes'])}")
            if config.get("ignore"):
                print(f"    Ignore: {config['ignore']}")
    else:
        print("\nNo benchmark configuration found in config file - analyzing all functions")


def run_benchmarks_and_process_profiles(benchmarks: List[str], profiles: List[str], count: int, tag: str, function_filter_configs: Dict[str, Dict[str, Any]]) -> None:
    print("\nStarting benchmark pipeline...")

    for benchmark in benchmarks:
        print(f"\nRunning benchmark: {benchmark}")
        run_benchmark(benchmark, profiles, count, tag)

        print(f"\nProcessing profiles for {benchmark}...")
        process_profiles(benchmark, profiles, tag)

        print(f"\nAnalyzing profile functions for {benchmark}...")
        analysis_config = get_profile_analysis_config(benchmark, function_filter_configs)
        analyze_benchmark_profile_functions(tag, profiles, benchmark, analysis_config)

        print(f"Completed pipeline for benchmark: {benchmark}")

    print("\nAll benchmarks and profile processing completed successfully!")


def run_benchmark(benchmark: str, profiles: List[str], count: int, tag: str) -> None:
    config = BenchmarkParamsWrapper(benchmark, profiles, count, tag)

    cmd = build_benchmark_command(config)
    text_dir, bin_dir = setup_output_directories(config.benchmark_name, config.tag)

    output_file = text_dir / f"{config.benchmark_name}.txt"
    run_benchmark_command(cmd, output_file)

    move_profile_files(config.benchmark_name, config.profiles, bin_dir)
    move_test_files(config.benchmark_name, bin_dir)

    print(f"Completed benchmark: {config.benchmark_name}")


# TODO: space for improvement.
def wait_for_profile_file(profile_file: Path, timeout: int = 5) -> bool:
    start_time = time.time()
    while time.time() - start_time < timeout:
        if profile_file.exists() and profile_file.stat().st_size > 0:
            return True
        time.sleep(0.1)
    return False


def run_pprof_command(cmd: List[str], output_path: Path, binary_mode: bool = False) -> subprocess.CompletedProcess:
    mode = 'wb' if binary_mode else 'w'

    try:
        with open(output_path, mode) as f:
            return subprocess.run(cmd, stdout=f, stderr=subprocess.PIPE, text=not binary_mode, check=True)

    except Exception as e:
        if isinstance(e, subprocess.CalledProcessError):
            stderr = e.stderr.decode() if isinstance(e.stderr, bytes) else e.stderr
            print(f"pprof command failed:\n{stderr}", file=sys.stderr)
        else:
            print(f"Error running pprof command: {e}", file=sys.stderr)

        sys.exit(EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR)


def generate_text_profile(profile_file: Path, output_file: Path) -> None:
    cmd = ["go", "tool", "pprof", *PPROF_TEXT_PARAMS, str(profile_file)]
    run_pprof_command(cmd, output_file)


def generate_png_visualization(profile_file: Path, output_file: Path) -> None:
    cmd = ["go", "tool", "pprof", "-png", str(profile_file)]
    run_pprof_command(cmd, output_file, binary_mode=True)


def process_profile(profile: str, benchmark: str, profile_file: Path, text_dir: Path, profile_functions_dir: Path) -> None:
    if not profile_file.exists():
        print(f"Warning: Profile file not found: {profile_file}", file=sys.stderr)
        sys.exit(PROFILE_FILE_MISSING)

    output_file = text_dir / f"{benchmark}_{profile}.txt"
    png_file = profile_functions_dir / f"{benchmark}_{profile}.png"

    try:
        generate_text_profile(profile_file, output_file)
        print(f"Processed {profile} profile for {benchmark}")

        generate_png_visualization(profile_file, png_file)
        print(f"Generated PNG visualization for {profile} profile of {benchmark} in {profile_functions_dir}")

    except Exception as e:
        print(f"Error processing {profile} profile for {benchmark}: {e}", file=sys.stderr)
        sys.exit(EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR)


def process_profiles(benchmark: str, profiles: List[str], tag: str) -> None:
    tag_dir = Path("bench") / tag
    bin_dir = tag_dir / "bin" / benchmark
    text_dir = tag_dir / "text" / benchmark

    pprof_profiles = [p for p in profiles if p != "trace"]

    for profile in pprof_profiles:
        profile_file = bin_dir / f"{benchmark}_{profile}.out"
        profile_functions_dir = tag_dir / f"{profile}_functions" / benchmark

        try:
            process_profile(profile, benchmark, profile_file, text_dir, profile_functions_dir)
        except Exception as e:
            print(f"Error processing profiles: {e}", file=sys.stderr)
            sys.exit(EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR)


def get_profile_analysis_config(benchmark: str, function_filter_configs: Dict[str, Dict[str, Any]]) -> ProfileFilter:

    config = function_filter_configs.get(benchmark, {})
    prefixes = config.get("prefixes", [])
    if not isinstance(prefixes, list):
        prefixes = []
    ignore_str = config.get("ignore", "")
    ignore_functions = set(parse_list_argument(ignore_str)) if ignore_str else set()
    return ProfileFilter(function_prefixes=prefixes, ignore_functions=ignore_functions)


def get_profile_paths(tag: str, benchmark: str, profile: str) -> ProfilePaths:

    tag_dir = Path("bench") / tag
    return ProfilePaths(profile_text=tag_dir / "text" / benchmark / f"{benchmark}_{profile}.txt", profile_binary=tag_dir / "bin" / benchmark / f"{benchmark}_{profile}.out", output=tag_dir / f"{profile}_functions" / benchmark)


def extract_function_name(line: str, function_prefixes: List[str], ignore_functions: Set[str]) -> Optional[str]:
    parts = line.split()
    if len(parts) < 6:
        return None

    func_name = " ".join(parts[5:])

    if function_prefixes and not any(prefix in func_name for prefix in function_prefixes):
        return None

    match = re.search(r'\.([^.(]+)(?:\([^)]*\))?$', func_name)
    if not match:
        return None

    func_name = match.group(1).strip().replace(" ", "")
    return func_name if func_name and func_name not in ignore_functions else None


def extract_all_function_names(profile_text_file: Path, config: ProfileFilter) -> Set[str]:
    if not profile_text_file.exists():
        sys.exit(PROFILE_FILE_MISSING)

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

                if func_name := extract_function_name(line, config.function_prefixes, config.ignore_functions):
                    functions.add(func_name)

        if not found_header:
            sys.exit(PROFILE_FILE_INVALID_HEADER)
        return functions

    except Exception:
        sys.exit(PROFILE_FILE_UNEXPECTED_ERROR)


def extract_single_function_content(func: str, paths: ProfilePaths) -> None:
    output_file = paths.output / f"{func}.txt"
    cmd = ["go", "tool", "pprof", f"-list={func}", str(paths.profile_binary)]

    try:
        with open(output_file, 'w') as f:
            subprocess.run(cmd, stdout=f, stderr=subprocess.PIPE, text=True, check=True)
        print(f"Collected function {func}")
    except subprocess.CalledProcessError as e:
        print(f"Error collecting function {func}: {e.stderr}", file=sys.stderr)
        sys.exit(EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR)


def analyze_benchmark_profile_functions(tag: str, profiles: List[str], benchmark: str, analysis_config: ProfileFilter) -> None:
    pprof_profiles = [p for p in profiles if p != "trace"]

    for profile in pprof_profiles:
        try:
            paths = get_profile_paths(tag, benchmark, profile)

            paths.output.mkdir(parents=True, exist_ok=True)

            functions = extract_all_function_names(paths.profile_text, analysis_config)
            for func in functions:
                extract_single_function_content(func, paths)
        except Exception as e:
            print(f"Error analyzing profile functions: {e}", file=sys.stderr)
            sys.exit(PROFILE_FILE_UNEXPECTED_ERROR)


def print_configuration_error(error_msg: Optional[str] = None) -> None:
    if error_msg:
        print(f"\n{error_msg}", file=sys.stderr)

    print("\nPlease set up configuration manually:", file=sys.stderr)
    print("1. Create a template config file:", file=sys.stderr)
    print("   prof setup --create-template [--output-path path/to/config.json]", file=sys.stderr)
    print("2. Use an existing config file (after creating template as well):", file=sys.stderr)
    print("   prof setup --config path/to/your/config.json", file=sys.stderr)


def setup_command(args):
    if args.create_template:
        ConfigManager.create_template(args.output_path)
        print("\nTemplate configuration file created successfully!")
    else:
        sys.exit(EXIT_CODE_TEMPLATE_ERROR)


def validate_required_args(args) -> bool:
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
    tag_dir = Path("bench") / tag
    if tag_dir.exists():
        try:
            shutil.rmtree(tag_dir)
            print(f"Cleaned up tag directory: {tag_dir}")
        except Exception as e:
            print(f"Error cleaning up tag directory {tag_dir}: {e}", file=sys.stderr)


def clean_directory(directory: Path):
    if directory.exists():
        try:
            for item in directory.iterdir():
                if item.is_file():
                    item.unlink()
                elif item.is_dir():
                    shutil.rmtree(item)
            print(f"Cleaned directory: {directory}")
        except Exception as e:
            print(f"Error cleaning directory {directory}: {e}", file=sys.stderr)
            sys.exit(BENCHMARK_DIRECTORY_UNEXPECTED_ERROR)


def parse_list_argument(arg: str) -> List[str]:
    arg = arg.strip('[]')
    return [item.strip() for item in arg.split(',')]


def parse_benchmark_config(config_str: str) -> Dict[str, Dict[str, str]]:
    try:
        config_str = config_str.replace("'", '"')
        config = json.loads(config_str)

        for benchmark, settings in config.items():
            if not isinstance(settings, dict):
                raise ValueError(f"Invalid settings format for benchmark {benchmark}")
            if "prefix" not in settings:
                raise ValueError(f"Missing 'prefix' for benchmark {benchmark}")
            if "ignore" in settings and not isinstance(settings["ignore"], str):
                raise ValueError(f"'ignore' must be a string for benchmark {benchmark}")

        return config
    except json.JSONDecodeError as e:
        raise ValueError(f"Invalid JSON format: {e}")
    except Exception as e:
        raise ValueError(f"Error parsing benchmark config: {e}")


def build_benchmark_command(config: BenchmarkParamsWrapper) -> List[str]:
    cmd = ["go", "test", "-run=^$", f"-bench=^{config.benchmark_name}$", "-benchmem", f"-count={config.iteration_count}"]

    for profile in config.profiles:
        if profile in PROFILE_FLAGS:
            cmd.append(PROFILE_FLAGS[profile])

    return cmd


def setup_output_directories(benchmark_name: str, tag: str) -> Tuple[Path, Path]:
    tag_dir = Path("bench") / tag
    text_dir = tag_dir / "text" / benchmark_name
    bin_dir = tag_dir / "bin" / benchmark_name

    text_dir.mkdir(parents=True, exist_ok=True)
    bin_dir.mkdir(parents=True, exist_ok=True)

    return text_dir, bin_dir


def run_benchmark_command(cmd: List[str], output_file: Path) -> None:
    try:
        with open(output_file, 'w') as f:
            subprocess.run(cmd, stdout=f, stderr=subprocess.STDOUT, text=True, check=True)
    except subprocess.CalledProcessError as e:
        with open(output_file, 'r') as f:
            error_output = f.read()

        if "go: cannot find main module" in error_output:
            print(f"{error_output}", file=sys.stderr)
            sys.exit(EXIT_CODE_MODULE_ERROR)

        print(f"Benchmark process failed with exit code {e.returncode}:\n{error_output}", file=sys.stderr)
        sys.exit(EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR)
    except Exception as e:
        print(f"Error running benchmark command or writing to {output_file}: {e}", file=sys.stderr)
        sys.exit(EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR)


def move_profile_files(benchmark_name: str, profiles: List[str], bin_dir: Path) -> None:
    for profile in profiles:
        if profile not in PROFILE_FLAGS:
            continue

        profile_file = Path(PROFILE_FLAGS[profile].split('=')[1])
        if not profile_file.exists():
            continue

        if not wait_for_profile_file(profile_file):
            print(f"Warning: Profile file {profile_file} was not fully written within timeout", file=sys.stderr)
            continue

        new_profile_file = bin_dir / f"{benchmark_name}_{profile}.out"
        try:
            shutil.move(str(profile_file), str(new_profile_file))
        except Exception as e:
            print(f"Error moving profile file {profile_file} to {new_profile_file}: {e}", file=sys.stderr)
            sys.exit(BENCHMARK_FILE_UNEXPECTED_ERROR)


def move_test_files(benchmark_name: str, bin_dir: Path) -> None:
    for item in Path('.').glob('*.test'):
        if not wait_for_profile_file(item):
            print(f"Warning: Test file {item} was not fully written within timeout", file=sys.stderr)
            continue

        new_test_file = bin_dir / f"{benchmark_name}_{item.name}"
        try:
            shutil.move(str(item), str(new_test_file))
            print(f"Moved test file: {item} -> {new_test_file}")
        except Exception as e:
            print(f"Error moving test file {item} to {new_test_file}: {e}", file=sys.stderr)
            sys.exit(BENCHMARK_FILE_UNEXPECTED_ERROR)


def create_parser():
    parser = argparse.ArgumentParser(description="CLI tool for organizing and analyzing Go benchmarks with AI")

    parser.add_argument('-version', '--version', action='store_true', help='Show version information and check for updates')

    subparsers = parser.add_subparsers(dest="command", help="Command to run")

    setup_parser = subparsers.add_parser("setup", help="Set up configuration for the benchmarking tool")
    setup_parser.add_argument("--create-template", action="store_true", help="Generate a new template configuration file for benchmarks")
    setup_parser.add_argument("--output-path", help="Destination path for the generated template configuration file (default: ./config_template.json)")

    parser.add_argument('-benchmarks', help='Benchmarks to run')
    parser.add_argument('-profiles', help='Profiles to use')
    parser.add_argument('-tag', help='Tag for the run')
    parser.add_argument('-count', type=int, help='Number of runs')
    parser.add_argument('-general_analyze', action='store_true', help="After benchmarks complete, run general AI analysis on the results")
    parser.add_argument('-flag_profiles', action='store_true', help="Flag the benchmark results for further review or processing")
    parser.add_argument('-multi_agent_analysis', action='store_true', help="Run multi-agent analysis on the benchmark results")
    return parser
