from typing import List

from analyzer.helpers import get_benchmark_file, validate_benchmark_directories


def multi_agent_analysis(tag: str, profile_types: List[str]):
    benchmark_names = validate_benchmark_directories(tag)
    print(f"Found {len(benchmark_names)} benchmarks and {len(profile_types)} profile types \n")

    profile_types = [profile_type for profile_type in profile_types if "trace" not in profile_type]
    for benchmark in benchmark_names:
        for profile_type in profile_types:
            print(f"\nAnalyzing {benchmark} ({profile_type})...")

            contentDict = get_benchmark_file(tag, benchmark, profile_type)
            content = contentDict["text_content"]
            print(content)
