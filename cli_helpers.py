#!/usr/bin/env python3
import argparse
from AI_client import analyze_prof_output_deep, analyze_prof_output_general
from utils_benchmark import print_configuration, run_benchmarks_and_process_profiles, setup_command, setup_directories, setup_from_current_directory, validate_arguments

# Create parser at module level
parser = argparse.ArgumentParser(description="CLI tool for benchmarking Go code with profile analysis")
subparsers = parser.add_subparsers(dest="command", help="Command to run")

setup_parser = subparsers.add_parser("setup", help="Create a template configuration file")
setup_parser.add_argument("--create-template", action="store_true", help="Create a template configuration file")
setup_parser.add_argument("--output-path", help="Path where to create the template file")

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

def parse_arguments():
    args = parser.parse_args()
    
    if args.command == "setup":
        setup_command(args)
    
    return args


def handle_benchmarks(args):
    setup_from_current_directory()
    
    benchmarks, profiles, benchmark_config = validate_arguments(args)
    setup_directories(args.tag, benchmarks, profiles)
    print_configuration(benchmarks, profiles, args.tag, args.count, benchmark_config)

    run_benchmarks_and_process_profiles(benchmarks, profiles, args.count, args.tag, benchmark_config)
    
    if args.general_analyze:
        analyze_prof_output_general(args.tag)
    elif args.deep_analyze:
        analyze_prof_output_deep(args.tag)