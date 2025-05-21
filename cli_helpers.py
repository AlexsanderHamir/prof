#!/usr/bin/env python3
import sys
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
from AI_client import analyze_prof_output

def parse_arguments():
    """Parse and validate command line arguments."""
    import argparse
    parser = argparse.ArgumentParser(description="CLI tool for benchmarking")
    
    parser.add_argument(
        '-benchmarks',
        type=str,
        required=True,
        help='Comma-separated list of benchmark types (e.g., -benchmarks "[BenchmarkGenPool,BenchmarkSyncPool]")'
    )
    
    parser.add_argument(
        '-profiles',
        type=str,
        required=True,
        help='Comma-separated list of profile types (e.g., -profiles "[cpu,memory,mutex]")'
    )
    
    parser.add_argument(
        '-tag',
        type=str,
        required=True,
        help='Tag for the benchmark run (e.g., -tag "test1")'
    )
    
    parser.add_argument(
        '-count',
        type=int,
        required=True,
        help='Number of benchmark iterations (e.g., -count 5)'
    )

    parser.add_argument(
        '-benchmark-config',
        type=str,
        help='JSON-like string containing benchmark-specific configurations. Example: \'{"BenchmarkGenPool":{"prefix":"github.com/AlexsanderHamir/GenPool","ignore":"func1,performWorkload"},"BenchmarkSyncPool":{"prefix":"sync"}}\''
    )

    parser.add_argument(
        '-analyze',
        action='store_true',
        help='Run AI analysis on the benchmark results after completion'
    )

    return parser.parse_args()

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

def run_benchmarks_and_process_profiles(benchmarks: List[str], profiles: List[str], 
                                      count: int, tag: str, 
                                      benchmark_config: Optional[Dict]) -> None:
    """Execute benchmarks and process their profiles."""
    print("\nRunning benchmarks...")
    for benchmark in benchmarks:
        run_benchmark(benchmark, profiles, count, tag)
    
    print("\nProcessing profiles...")
    for benchmark in benchmarks:
        process_profiles(benchmark, profiles, tag)
        
    print("\nAnalyzing profile functions...")
    analyze_profile_functions(tag, profiles, benchmarks, benchmark_config)
    
    print("\nAll benchmarks and profile processing completed successfully!")

def run_ai_analysis(tag: str) -> None:
    """Run AI analysis on the benchmark results if requested."""
    print("\nStarting AI analysis of benchmark results...")
    try:
        analyze_prof_output(tag)
        print("\nAI analysis completed successfully!")
    except ValueError as e:
        print(f"\nError: {e}", file=sys.stderr)
        print("Please set the DEEPSEEK_API_KEY environment variable to use AI analysis", file=sys.stderr)
    except Exception as e:
        print(f"\nError during AI analysis: {e}", file=sys.stderr)
        # Don't exit with error, as the benchmarks were successful 