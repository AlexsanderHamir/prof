from pathlib import Path
import sys
from parser.interface import should_keep_line
from tests.unit.constants import ONLY_HEADER_INCLUDED


def test_keep_no_lines():
    profile_type = "cpu"
    file_path = Path(__file__).resolve().parent.parent / "BenchmarkGenPool_cpu.txt"
    print("file_path", file_path)

    header_index = 6 if profile_type == "cpu" else 5
    profile_values_dict = {
        0: 1000000000,
        1: 1000000000,
        2: 1000000000,
        3: 1000000000,
        4: 1000000000,
    }
    ignore_functions = []
    ignore_prefixes = []

    try:
        with open(file_path, 'r') as f:
            content = f.readlines()
            header = [line.strip() for line in content[:header_index]]
            body = [line.strip() for line in content[header_index:] if should_keep_line(line.strip(), profile_values_dict, ignore_functions, ignore_prefixes)]
            filtered_content = header + body
            assert len(filtered_content) == ONLY_HEADER_INCLUDED, f"Filtered content length is {len(filtered_content)} but should be {ONLY_HEADER_INCLUDED}"
    except OSError as e:
        print(f"Cannot read profile file {file_path}: {e}", file=sys.stderr)
        assert False
