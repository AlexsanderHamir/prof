import argparse
import sys
from AI_client import analyze_profiles
from config.config_manager import ConfigManager
from exit_codes import EXIT_CODE_MISSING_ARGUMENTS
from cli.cli_helpers import (parse_and_load_benchmark_config, print_configuration, run_benchmarks_and_process_profiles, setup_directories, config_setup)


def create_parser():
    parser = argparse.ArgumentParser(description="CLI tool for organizing and analyzing Go benchmarks with AI")
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
    return parser


def parse_arguments():
    parser = create_parser()
    try:
        args = parser.parse_args()
    except SystemExit as e:
        if e.code == 2:
            sys.exit(EXIT_CODE_MISSING_ARGUMENTS)
        sys.exit(e.code)
    return args


def handle_benchmarks(args) -> None:
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
