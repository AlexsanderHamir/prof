from utils_AI_client import analyze_all_deep, analyze_all_profiles, get_benchmark_names, get_profile_types, validate_benchmark_directories, validate_benchmark_directory

def analyze_prof_output_general(tag: str) -> None:
    try:
        base_dir = validate_benchmark_directory(tag)
        text_dir = base_dir / "text"
        
        benchmark_names = get_benchmark_names(text_dir)
        profile_types = get_profile_types(text_dir, benchmark_names[0])
        print(f"Found {len(benchmark_names)} benchmarks and {len(profile_types)} profile types")
        
        analyze_all_profiles(tag, benchmark_names, profile_types)
    except ValueError as e:
        print(f"Error: {str(e)}")


def analyze_prof_output_deep(tag: str) -> list[str] | None:
    is_valid, error_msg, benchmark_names = validate_benchmark_directories(tag)
    if not is_valid:
        print(f"Error: {error_msg}")
        return None
        
    print(f"Found {len(benchmark_names)} benchmarks for deep analysis")
    analyze_all_deep(tag, benchmark_names)
    return benchmark_names

