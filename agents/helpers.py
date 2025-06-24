from collections import defaultdict
import os
from pathlib import Path
import re
import sys
from typing import Dict, List, Set
from exit_codes import PROFILE_FILE_INVALID_HEADER, PROFILE_FILE_MISSING, PROFILE_FILE_UNEXPECTED_ERROR


def get_system_prompt() -> str:
    project_root = Path(__file__).resolve().parent.parent
    prompt_path = os.path.join(project_root, 'templates/profile_prompt.txt')
    try:
        with open(prompt_path, "r", encoding="utf-8") as f:
            return f.read()
    except FileNotFoundError:
        print(f"Error: The prompt file was not found at '{prompt_path}'.")
        sys.exit(1)
    except PermissionError:
        print(f"Error: Permission denied when trying to read '{prompt_path}'.")
        sys.exit(1)
    except OSError as e:
        print(f"Error: An unexpected OS error occurred while reading '{prompt_path}': {e}")
        sys.exit(1)


def extract_prefix(line: str) -> str | None:
    parts = line.strip().split()
    if len(parts) < 6:
        return None

    func_name = " ".join(parts[5:])
    func_name = func_name.replace(" (inline)", "")

    return func_name.rsplit(".", 1)[0] if "." in func_name else None


def extract_all_function_names(profile_text_file: Path) -> Set[str]:
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

                if func_name := extract_prefix(line):
                    functions.add(func_name)

        if not found_header:
            sys.exit(PROFILE_FILE_INVALID_HEADER)
        return functions

    except Exception:
        sys.exit(PROFILE_FILE_UNEXPECTED_ERROR)


def group_similar_prefixes(prefixes: Set[str], depth: int = 1) -> Dict[str, List[str]]:
    grouped = defaultdict(list)
    for s in prefixes:
        parts = re.split(r'[./]', s)
        prefix = ".".join(parts[:depth]) if len(parts) >= depth else s
        grouped[prefix].append(s)
    return grouped


def longest_common_prefix_groups(groups: Dict[str, List[str]]) -> Set[str]:

    def common_prefix(paths: List[str]) -> str:
        if not paths:
            return ""
        split_paths = [re.split(r'[./]', p) for p in paths]
        min_len = min(len(p) for p in split_paths)
        result = []
        for i in range(min_len):
            tokens = {p[i] for p in split_paths}
            if len(tokens) == 1:
                result.append(tokens.pop())
            else:
                break
        return ".".join(result)

    return {common_prefix(group) for group in groups.values()}


def compress_profile_content(file_path: Path) -> None:
    prefixes = extract_all_function_names(file_path)
    print(prefixes)
    grouped_prefixes = group_similar_prefixes(prefixes, depth=1)
    longest_common_prefixes = longest_common_prefix_groups(grouped_prefixes)
    print(longest_common_prefixes)
