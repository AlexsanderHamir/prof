from utils_AI_client import analyze_all_profiles, validate_benchmark_directories
from typing import List


def analyze_prof_output_general(tag: str, profile_types: List[str]):
    try:
        benchmark_names = validate_benchmark_directories(tag)
        print(
            f"Found {len(benchmark_names)} benchmarks and {len(profile_types)} profile types"
        )

        analyze_all_profiles(tag, benchmark_names, profile_types)
    except Exception as e:
        print(f"Error: {str(e)}")
        raise
