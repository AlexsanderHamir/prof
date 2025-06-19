import argparse
import sys
from AI_client import analyze_profiles
from config_manager import ConfigManager, ConfigurationParsingError, ConfigurationNotFound
from exit_codes import BENCHMARK_DIRECTORY_UNEXPECTED_ERROR, BENCHMARK_FILE_UNEXPECTED_ERROR, CONFIG_PARSING_ERROR, CONFIG_VALIDATION_ERROR, EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR, EXIT_CODE_MISSING_ARGUMENTS, MISSING_CONFIG_FILE, EXIT_CODE_MODULE_ERROR, EXIT_CODE_TEMPLATE_ERROR, MODEL_ANALYSIS_ERROR, PROFILE_READ_ERROR, PROFILE_SAVE_ERROR
from utils_benchmark import (BenchmarkTemplateError, BenchmarkUnexpectedProcessError, BenchmarkDirectoryError, BenchmarkFileError, BenchmarkModuleError, parse_and_load_benchmark_config, print_configuration, run_benchmarks_and_process_profiles, setup_command, setup_directories, config_setup)
from utils_AI_client import ProfileReadError, ProfileSaveError, ModelAnalysisError
from utils_config_manager import ConfigValidationError


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

    except BenchmarkUnexpectedProcessError:
        sys.exit(EXIT_CODE_BENCHMARK_PROCESS_UNEXPECTED_ERROR)
    except BenchmarkModuleError:
        sys.exit(EXIT_CODE_MODULE_ERROR)
    except BenchmarkDirectoryError:
        sys.exit(BENCHMARK_DIRECTORY_UNEXPECTED_ERROR)
    except BenchmarkFileError:
        sys.exit(BENCHMARK_FILE_UNEXPECTED_ERROR)
    except ProfileReadError:
        sys.exit(PROFILE_READ_ERROR)
    except ProfileSaveError:
        sys.exit(PROFILE_SAVE_ERROR)
    except ModelAnalysisError:
        sys.exit(MODEL_ANALYSIS_ERROR)
    except ConfigurationNotFound:
        sys.exit(MISSING_CONFIG_FILE)
    except ConfigValidationError:
        sys.exit(CONFIG_VALIDATION_ERROR)
    except ConfigurationParsingError:
        sys.exit(CONFIG_PARSING_ERROR)


def setup(args) -> None:
    try:
        setup_command(args)
    except BenchmarkTemplateError:
        sys.exit(EXIT_CODE_TEMPLATE_ERROR)
