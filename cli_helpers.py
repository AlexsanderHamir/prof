#!/usr/bin/env python3
import sys
import argparse
import os
from typing import Tuple, List, Dict, Optional
from benchmark_helpers import (
    parse_list_argument,
    create_bench_directories,
    create_profile_function_directories,
    run_benchmark,
    process_profiles,
    analyze_profile_functions,
    cleanup_tag_directory
)
from AI_client import analyze_prof_output, analyze_prof_output_deep
from config_manager import ConfigManager

# Create parser at module level
parser = argparse.ArgumentParser(description="CLI tool for benchmarking Go code with profile analysis")
subparsers = parser.add_subparsers(dest="command", help="Command to run")

# Setup command
setup_parser = subparsers.add_parser("setup", help="Create a template configuration file")
setup_parser.add_argument("--create-template", action="store_true", help="Create a template configuration file")
setup_parser.add_argument("--output-path", help="Path where to create the template file")

# Make benchmarks the default command by adding it to the main parser
parser.add_argument(
    '-benchmarks',
    type=str,
    help='Comma-separated list of benchmark names (e.g., "[BenchmarkGenPool,BenchmarkSyncPool]")'
)
parser.add_argument(
    '-profiles',
    type=str,
    help='Comma-separated list of profile types (e.g., "[cpu,memory,mutex]")'
)
parser.add_argument(
    '-tag',
    type=str,
    help='Tag for the benchmark run (e.g., "test1")'
)
parser.add_argument(
    '-count',
    type=int,
    help='Number of benchmark iterations (e.g., 5)'
)
parser.add_argument(
    '-general_analyze',
    action='store_true',
    help='Run general AI analysis on the benchmark results after completion'
)
parser.add_argument(
    '-deep_analyze',
    action='store_true',
    help='Run deep AI analysis on the benchmark results after completion'
)

def setup_command(args):
    if args.create_template:
        ConfigManager.create_template(args.output_path)
        print("\nTemplate configuration file created successfully!")
    else:
        print("\nError: Please use --create-template to create a configuration template", file=sys.stderr)
        sys.exit(1)

def parse_arguments():
    args = parser.parse_args()
    
    if args.command == "setup":
        setup_command(args)
    
    return args

def validate_arguments(args) -> Tuple[List[str], List[str], Optional[Dict]]:
    """Validate parsed arguments and return processed values."""
    benchmarks = parse_list_argument(args.benchmarks)
    profiles = parse_list_argument(args.profiles)
    
    # Get benchmark configuration from config file
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
    except ValueError as e:
        print(f"Error loading benchmark configuration: {e}", file=sys.stderr)
        sys.exit(1)
    
    # Validate count
    if args.count <= 0:
        print("Error: count must be greater than 0", file=sys.stderr)
        cleanup_tag_directory(args.tag)
        sys.exit(1)
        
    return benchmarks, profiles, benchmark_config

def setup_directories(tag: str, benchmarks: List[str], profiles: List[str]) -> None:
    """Create necessary directories for the benchmark run."""
    create_bench_directories(tag, benchmarks)
    create_profile_function_directories(tag, profiles, benchmarks)

def print_configuration(benchmarks: List[str], profiles: List[str], tag: str, 
                       count: int, benchmark_config: Optional[Dict]) -> None:
    """Print the configuration details for verification."""
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
        print("\nNo benchmark configuration found in config file - analyzing all functions")

def run_benchmarks_and_process_profiles(benchmarks: List[str], profiles: List[str], count: int, tag: str, benchmark_config: Optional[Dict]) -> None:
    print("\nRunning benchmarks sequentially...")
    for benchmark in benchmarks:
        run_benchmark(benchmark, profiles, count, tag)
    
    print("\nProcessing profiles...")
    for benchmark in benchmarks:
        process_profiles(benchmark, profiles, tag)
        
    print("\nAnalyzing profile functions...")
    analyze_profile_functions(tag, profiles, benchmarks, benchmark_config)
    
    print("\nAll benchmarks and profile processing completed successfully!")

def print_configuration_error(error_msg: Optional[str] = None) -> None:
    """Print configuration error messages and setup instructions.
    
    Args:
        error_msg: Optional error message to display before the setup instructions
    """
    if error_msg:
        print(f"\n{error_msg}", file=sys.stderr)
    
    print("\nPlease set up configuration manually:", file=sys.stderr)
    print("1. Create a template config file:", file=sys.stderr)
    print("   prof setup --create-template [--output-path path/to/config.json]", file=sys.stderr)
    print("2. Use an existing config file (after creating template as well):", file=sys.stderr)
    print("   prof setup --config path/to/your/config.json", file=sys.stderr)

def handle_benchmarks(args):
    # Always check for configuration
    template_path = os.path.join(os.getcwd(), "config_template.json")
    if os.path.exists(template_path):
        print("\nFound config_template.json in current directory. Attempting automatic setup...")
        try:
            ConfigManager.setup_from_file(template_path)
            print("Automatic configuration completed successfully!")
        except ValueError as e:
            print_configuration_error(f"Error during automatic configuration: {e}")
            sys.exit(1)
    else:
        print_configuration_error("Error: Configuration not found. Please run setup first:")
        sys.exit(1)
    
    if not all([args.benchmarks, args.profiles, args.tag, args.count]):
        print("\nError: All of -benchmarks, -profiles, -tag, and -count are required for benchmarking", file=sys.stderr)
        sys.exit(1)
    
    benchmarks, profiles, benchmark_config = validate_arguments(args)
    setup_directories(args.tag, benchmarks, profiles)
    print_configuration(benchmarks, profiles, args.tag, args.count, benchmark_config)

    run_benchmarks_and_process_profiles(benchmarks, profiles, args.count, args.tag, benchmark_config)
    
    if args.general_analyze:
        analyze_prof_output(args.tag)
    elif args.deep_analyze:
        analyze_prof_output_deep(args.tag)

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