#!/usr/bin/env python3
import argparse
import sys
from AI_client import analyze_profiles
from config_manager import ConfigManager
from utils_benchmark import BenchmarkConfigError, parse_and_load_benchmark_config, print_configuration, run_benchmarks_and_process_profiles, setup_command, setup_directories, config_setup


def create_parser():
    parser = argparse.ArgumentParser(description="CLI tool for benchmarking Go code with profile analysis")
    subparsers = parser.add_subparsers(dest="command", help="Command to run")

    setup_parser = subparsers.add_parser("setup", help="Create a template configuration file")
    setup_parser.add_argument("--create-template", action="store_true", help="Create a template configuration file")
    setup_parser.add_argument("--output-path", help="Path where to create the template file")

    parser.add_argument('-benchmarks', type=str, help='Comma-separated list of benchmark names (e.g., "[BenchmarkGenPool,BenchmarkSyncPool]")')
    parser.add_argument('-profiles', type=str, help='Comma-separated list of profile types (e.g., "[cpu,memory,mutex]")')
    parser.add_argument('-tag', type=str, help='Tag for the benchmark run (e.g., "test1")')
    parser.add_argument('-count', type=int, help='Number of benchmark iterations (e.g., 5)')
    parser.add_argument('-general_analyze', action='store_true', help='Run general AI analysis on the benchmark results after completion')
    parser.add_argument('-flag_profiles', action='store_true', help='Flag the benchmark results')
    return parser


def parse_arguments():
    parser = create_parser()
    args = parser.parse_args()
    return args


def handle_benchmarks(args):
    config_setup()

    benchmarks, profiles, function_filter_configs = parse_and_load_benchmark_config(args)
    setup_directories(args.tag, benchmarks, profiles)
    print_configuration(benchmarks, profiles, args.tag, args.count, function_filter_configs)

    run_benchmarks_and_process_profiles(benchmarks, profiles, args.count, args.tag, function_filter_configs)

    if args.general_analyze:
        analyze_profiles(args.tag, profiles)
    if args.flag_profiles:
        ConfigManager.is_flagging = True
        analyze_profiles(args.tag, profiles)


def setup(args):
    try:
        setup_command(args)
    except BenchmarkConfigError as e:
        print(f"\nSetup Error: {e}", file=sys.stderr)
        sys.exit(1)
