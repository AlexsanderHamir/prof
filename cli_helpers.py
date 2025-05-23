#!/usr/bin/env python3
import sys
import argparse
import os
from typing import Tuple, List, Dict, Optional
from benchmark_helpers import (
    parse_list_argument,
    parse_benchmark_config,
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
setup_parser = subparsers.add_parser("setup", help="Set up the configuration")
setup_group = setup_parser.add_mutually_exclusive_group(required=True)
setup_group.add_argument("--config", help="Path to the configuration JSON file")
setup_group.add_argument("--create-template", action="store_true", help="Create a template configuration file")
setup_parser.add_argument("--output-path", help="Path where to create the template file (only used with --create-template)")

# Clean command
clean_parser = subparsers.add_parser("clean", help="Clean the configuration cache")

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
    '-benchmark-config',
    type=str,
    help='JSON-like string containing benchmark-specific configurations'
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
    """Handle the setup command for configuration management."""
    try:
        if args.create_template:
            ConfigManager.create_template(args.output_path)
            return
            
        if not args.config:
            print("\nError: Please provide a configuration file using --config or create one using --create-template", file=sys.stderr)
            sys.exit(1)
            
        ConfigManager.setup_from_file(args.config)
        print("\nConfiguration completed successfully!")
    except ValueError as e:
        print(f"\nError: {e}", file=sys.stderr)
        sys.exit(1)

def clean_command(args):
    """Handle the clean command to remove configuration cache."""
    try:
        ConfigManager.clean_config()
        print("\nConfiguration cache cleaned successfully!")
    except Exception as e:
        print(f"\nError cleaning configuration cache: {e}", file=sys.stderr)
        sys.exit(1)

def parse_arguments():
    """Parse command line arguments and return the parsed arguments."""
    args = parser.parse_args()
    
    # Map commands to their handler functions
    command_handlers = {
        "setup": setup_command,
        "clean": clean_command
    }
    
    # Execute the appropriate handler if a command was specified
    if args.command in command_handlers:
        command_handlers[args.command](args)
    
    return args

def validate_arguments(args) -> Tuple[List[str], List[str], Optional[Dict]]:
    """Validate parsed arguments and return processed values."""
    benchmarks = parse_list_argument(args.benchmarks)
    profiles = parse_list_argument(args.profiles)
    
    # Parse benchmark configuration if provided
    benchmark_config = None
    if args.benchmark_config:
        try:
            benchmark_config = parse_benchmark_config(args.benchmark_config)
            # Validate that all configured benchmarks exist in the benchmarks list
            invalid_configs = set(benchmark_config.keys()) - set(benchmarks)
            if invalid_configs:
                print(f"Warning: Configurations provided for non-existent benchmarks: {', '.join(invalid_configs)}", file=sys.stderr)
        except ValueError as e:
            print(f"Error: {e}", file=sys.stderr)
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
            print(f"    Prefix: {config['prefix']}")
            if config.get("ignore"):
                print(f"    Ignore: {config['ignore']}")
    else:
        print("\nNo benchmark configuration provided - analyzing all functions")

def run_benchmarks_and_process_profiles(benchmarks: List[str], profiles: List[str], count: int, tag: str, benchmark_config: Optional[Dict]) -> None:
    print("\nRunning benchmarks...")
    for benchmark in benchmarks:
        run_benchmark(benchmark, profiles, count, tag)
    
    print("\nProcessing profiles...")
    for benchmark in benchmarks:
        process_profiles(benchmark, profiles, tag)
        
    print("\nAnalyzing profile functions...")
    analyze_profile_functions(tag, profiles, benchmarks, benchmark_config)
    
    print("\nAll benchmarks and profile processing completed successfully!")

def handle_benchmarks(args):
    """Handle the benchmark command and its associated operations.
    
    Args:
        args: Parsed command line arguments
        
    Returns:
        None
        
    Raises:
        Exception: If any error occurs during benchmark execution
    """
    # Check if configuration exists before running benchmarks
    if (args.general_analyze or args.deep_analyze) and not ConfigManager.is_configured():
        # Try to find config_template.json in current directory
        template_path = os.path.join(os.getcwd(), "config_template.json")
        if os.path.exists(template_path):
            print("\nFound config_template.json in current directory. Attempting automatic setup...")
            try:
                ConfigManager.setup_from_file(template_path)
                print("Automatic configuration completed successfully!")
            except ValueError as e:
                print(f"\nError during automatic configuration: {e}", file=sys.stderr)
                print("\nPlease set up configuration manually:", file=sys.stderr)
                print("1. Create a template config file:", file=sys.stderr)
                print("   prof setup --create-template [--output-path path/to/config.json]", file=sys.stderr)
                print("2. Use an existing config file (after creating template as well):", file=sys.stderr)
                print("   prof setup --config path/to/your/config.json", file=sys.stderr)
                sys.exit(1)
        else:
            print("\nError: Configuration not found. Please run setup first:", file=sys.stderr)
            print("To set up configuration, either:", file=sys.stderr)
            print("1. Create a template config file:", file=sys.stderr)
            print("   prof setup --create-template [--output-path path/to/config.json]", file=sys.stderr)
            print("2. Use an existing config file (after creating template as well):", file=sys.stderr)
            print("   prof setup --config path/to/your/config.json", file=sys.stderr)
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