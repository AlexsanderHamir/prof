from utils_AI_client import analyze_all_profiles, get_profile_types, validate_benchmark_directories
from pathlib import Path


def analyze_prof_output_general(tag: str):
    try:
        benchmark_names = validate_benchmark_directories(tag)
        base_dir = Path("bench") / tag
        text_dir = base_dir / "text"
        profile_types = get_profile_types(text_dir, benchmark_names[0])
        print(
            f"Found {len(benchmark_names)} benchmarks and {len(profile_types)} profile types"
        )

        analyze_all_profiles(tag, benchmark_names, profile_types)
    except Exception as e:
        print(f"Error: {str(e)}")
        raise
