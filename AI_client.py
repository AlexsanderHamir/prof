from utils_AI_client import analyze_all_profiles, validate_benchmark_directories, ProfileReadError, ProfileSaveError, ModelAnalysisError
from typing import List


def analyze_profiles(tag: str, profile_types: List[str]):
    """
    Analyze benchmark profiles for a given tag and list of profile types.

    Args:
        tag (str): The tag identifying the benchmark run.
        profile_types (List[str]): List of profile types to analyze (e.g., cpu, memory).
    Raises:
        Exception: If analysis or validation fails.
    """
    try:
        benchmark_names = validate_benchmark_directories(tag)
        print(f"Found {len(benchmark_names)} benchmarks and {len(profile_types)} profile types")

        analyze_all_profiles(tag, benchmark_names, profile_types)
    except ProfileReadError as e:
        print(f"Profile read error: {str(e)}")
        raise
    except ProfileSaveError as e:
        print(f"Profile save error: {str(e)}")
        raise
    except ModelAnalysisError as e:
        print(f"Model analysis error: {str(e)}")
        raise
    except Exception as e:
        print(f"Unexpected error: {str(e)}")
        raise
