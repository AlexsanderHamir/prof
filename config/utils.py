import sys
from typing import Any
from exit_codes import CONFIG_VALIDATION_ERROR


def fail(msg: str) -> None:
    print(f"[config error] {msg}", file=sys.stderr)
    sys.exit(CONFIG_VALIDATION_ERROR)


def check_benchmark_profile_logic(all_benchmarks, all_profiles, specific_benchmarks, specific_profiles) -> None:
    if all_benchmarks and all_profiles:
        if specific_benchmarks or specific_profiles:
            fail("When all_benchmarks and all_profiles are both True, specific_benchmarks and specific_profiles must be empty")

    if all_benchmarks and specific_benchmarks:
        fail("When all_benchmarks is True, specific_benchmarks must be empty")

    if all_profiles and specific_profiles:
        fail("When all_profiles is True, specific_profiles must be empty")

    if not all_benchmarks and not specific_benchmarks:
        fail("When all_benchmarks is False, provide specific_benchmarks")

    if not all_profiles and not specific_profiles:
        fail("When all_profiles is False, provide specific_profiles")


def validate_string_list(lst, name: str) -> None:
    if not isinstance(lst, list):
        fail(f"{name} must be a list")
    for i, item in enumerate(lst):
        if not isinstance(item, str):
            fail(f"{name}[{i}] must be a string, got {type(item).__name__}")


def validate_universal_profile_filter(universal_profile_filter: Any) -> None:
    if not isinstance(universal_profile_filter, dict):
        fail("universal_profile_filter must be a dictionary")

    if "profile_values" not in universal_profile_filter:
        fail("universal_profile_filter must contain 'profile_values'")

    profile_values = universal_profile_filter["profile_values"]
    if not isinstance(profile_values, dict):
        fail("profile_values must be a dictionary")

    required_fields = ["flat", "flat%", "sum%", "cum", "cum%"]
    for field in required_fields:
        if field not in profile_values:
            fail(f"profile_values must contain '{field}'")
        if not isinstance(profile_values[field], (int, float)):
            fail(f"profile_values '{field}' must be a number")

    if "ignore_functions" in universal_profile_filter:
        validate_string_list(universal_profile_filter["ignore_functions"], "ignore_functions")

    if "ignore_prefixes" in universal_profile_filter:
        validate_string_list(universal_profile_filter["ignore_prefixes"], "ignore_prefixes")
