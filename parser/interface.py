from typing import Dict, List

from parser.helpers import filter_by_ignore_functions, filter_by_ignore_prefixes, filter_by_number, to_ignore_set


def should_keep_line(line: str, profile_values_dict: Dict[int, float], ignore_functions: List[str], ignore_prefixes: List[str]) -> bool:
    if not line:
        return False

    parts = line.split()
    ignore_functions_set = to_ignore_set(ignore_functions)
    ignore_prefixes_set = to_ignore_set(ignore_prefixes)
    return filter_by_number(profile_values_dict, parts) and filter_by_ignore_functions(ignore_functions_set, parts) and filter_by_ignore_prefixes(ignore_prefixes_set, parts)
