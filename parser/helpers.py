import re
from typing import Dict, List, Set


def filter_by_ignore_functions(ignore_functions_set: Set[str], parts: List[str]) -> bool:
    if not ignore_functions_set:
        return True

    full_function_name = clean_function_name(" ".join(parts[5:]))

    if full_function_name in ignore_functions_set:
        return False

    return True


def clean_function_name(s: str) -> str:
    s = s.replace(" (inline)", "").strip()
    return s.rsplit(".", 1)[-1]


def filter_by_number(profile_values_dict: Dict[int, float], parts: List[str]) -> bool:
    for i in range(5):
        config_value = profile_values_dict[i]
        line_value = extract_float(parts[i])

        if config_value == 0.0:
            continue

        if line_value <= config_value:
            return False

    return True


def extract_float(s: str) -> float:
    match = re.search(r"\d+(?:\.\d+)?", s)
    if not match:
        raise ValueError(f"No float found in '{s}'")
    return float(match.group())


def to_ignore_set(ignore_functions: List[str]) -> Set[str]:
    return set(ignore_functions)


def filter_by_ignore_prefixes(ignore_prefixes_set: Set[str], parts: List[str]) -> bool:
    if not ignore_prefixes_set:
        return True

    full_function_name = " ".join(parts[5:])
    full_function_name = full_function_name.replace(" (inline)", "").strip()
    for ignore_prefix in ignore_prefixes_set:
        if full_function_name.startswith(ignore_prefix):
            return False

    return True
