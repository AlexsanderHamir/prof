import sys
from analyzer.helpers import analyze_all_profiles, validate_benchmark_directories
from typing import List

from config.config_manager import ConfigManager


def analyze_profiles(tag: str, profile_types: List[str]):
    config = ConfigManager.get_config()

    benchmark_names = (validate_benchmark_directories(tag) if config.ai_config.all_benchmarks else config.ai_config.specific_benchmarks) or []
    selected_profile_types = (profile_types if config.ai_config.all_profiles else config.ai_config.specific_profiles) or []

    print(f"Found {benchmark_names} benchmarks and {selected_profile_types} profile types \n")

    analyze_all_profiles(tag, benchmark_names, selected_profile_types)
