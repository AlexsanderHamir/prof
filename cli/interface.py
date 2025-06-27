import sys
from analyzer.interface import analyze_profiles
from config.config_manager import ConfigManager
from exit_codes import EXIT_CODE_MISSING_ARGUMENTS
from cli.helpers import (create_parser, parse_and_load_benchmark_config, print_configuration, run_benchmarks_and_process_profiles, setup_directories, config_setup, validate_required_args)
from version import format_version_output, check_version


def handle_version():
    current_version, latest_version = check_version()
    print(format_version_output(current_version, latest_version))


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
    if args.command or not validate_required_args(args):
        sys.exit(EXIT_CODE_MISSING_ARGUMENTS)

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
