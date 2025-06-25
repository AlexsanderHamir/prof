from pathlib import Path
import sys
from typing import Dict, List, Set

from shared import ProfileFilter
from exit_codes import PROFILE_FILE_INVALID_HEADER, PROFILE_FILE_MISSING, PROFILE_FILE_UNEXPECTED_ERROR
from parser.helpers import extract_function_name, filter_by_ignore_functions, filter_by_ignore_prefixes, filter_by_number, to_ignore_set


def should_keep_line(line: str, profile_values_dict: Dict[int, float], ignore_functions: List[str], ignore_prefixes: List[str]) -> bool:
    if not line:
        return False

    parts = line.split()
    ignore_functions_set = to_ignore_set(ignore_functions)
    ignore_prefixes_set = to_ignore_set(ignore_prefixes)
    return filter_by_number(profile_values_dict, parts) and filter_by_ignore_functions(ignore_functions_set, parts) and filter_by_ignore_prefixes(ignore_prefixes_set, parts)


def extract_all_function_names(profile_text_file: Path, config: ProfileFilter) -> Set[str]:
    if not profile_text_file.exists():
        sys.exit(PROFILE_FILE_MISSING)

    functions = set()
    found_header = False

    try:
        with open(profile_text_file, 'r') as f:
            for line in f:
                line = line.strip()
                if not line:
                    continue

                if "flat  flat%   sum%        cum   cum%" in line:
                    found_header = True
                    continue

                if not found_header:
                    continue

                if func_name := extract_function_name(line, config.function_prefixes, config.ignore_functions):
                    functions.add(func_name)

        if not found_header:
            sys.exit(PROFILE_FILE_INVALID_HEADER)
        return functions

    except Exception:
        sys.exit(PROFILE_FILE_UNEXPECTED_ERROR)
