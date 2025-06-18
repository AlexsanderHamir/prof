import argparse
import sys
from AI_client import analyze_profiles
from config_manager import ConfigManager
from utils_benchmark import (BenchmarkConfigError, BenchmarkError, BenchmarkProfileError, BenchmarkDirectoryError, BenchmarkFileError, parse_and_load_benchmark_config, print_configuration, run_benchmarks_and_process_profiles, setup_command, setup_directories, config_setup)
from utils_AI_client import ProfileReadError, ProfileSaveError, ModelAnalysisError


def create_parser():
    parser = argparse.ArgumentParser(description="CLI tool for organizing and analyzing Go benchmarks with AI")
    subparsers = parser.add_subparsers(dest="command", help="Command to run")

    setup_parser = subparsers.add_parser("setup", help="Set up configuration for the benchmarking tool")
    setup_parser.add_argument("--create-template", action="store_true", help="Generate a new template configuration file for benchmarks")
    setup_parser.add_argument("--output-path", help="Destination path for the generated template configuration file (default: ./config_template.json)")

    parser.add_argument('-benchmarks', type=str, help="List of benchmark names to run, formatted as a Python list (e.g., '[BenchmarkGenPool,BenchmarkSyncPool]')")
    parser.add_argument('-profiles', type=str, help="List of profile types to collect, formatted as a Python list (e.g., '[cpu,memory,mutex]')")
    parser.add_argument('-tag', type=str, help="A unique tag or label for this benchmark run (e.g., 'test1')")
    parser.add_argument('-count', type=int, help="Number of times to repeat each benchmark (e.g., 5)")
    parser.add_argument('-general_analyze', action='store_true', help="After benchmarks complete, run general AI analysis on the results")
    parser.add_argument('-flag_profiles', action='store_true', help="Flag the benchmark results for further review or processing")
    return parser


def parse_arguments():
    parser = create_parser()
    args = parser.parse_args()
    return args


def handle_benchmarks(args):
    try:
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

    except BenchmarkConfigError as e:
        print(f"\nSetup Error: {e}", file=sys.stderr)
        sys.exit(1)
    except BenchmarkError as e:
        print(f"\nBenchmark error: {e}", file=sys.stderr)
        sys.exit(2)
    except BenchmarkProfileError as e:
        print(f"\nBenchmark profile error: {e}", file=sys.stderr)
        sys.exit(3)
    except BenchmarkDirectoryError as e:
        print(f"\nBenchmark directory error: {e}", file=sys.stderr)
        sys.exit(4)
    except BenchmarkFileError as e:
        print(f"\nBenchmark file error: {e}", file=sys.stderr)
        sys.exit(5)
    except ProfileReadError as e:
        print(f"\nProfile read error: {e}", file=sys.stderr)
        sys.exit(6)
    except ProfileSaveError as e:
        print(f"\nProfile save error: {e}", file=sys.stderr)
        sys.exit(7)
    except ModelAnalysisError as e:
        print(f"\nModel analysis error: {e}", file=sys.stderr)
        sys.exit(8)
    except Exception as e:
        print(f"\nUnexpected error: {e}", file=sys.stderr)
        sys.exit(99)


def setup(args):
    try:
        setup_command(args)
    except BenchmarkConfigError as e:
        print(f"\nSetup Error: {e}", file=sys.stderr)
        sys.exit(1)
