from pathlib import Path
from agents.helpers import extract_all_function_names, group_similar_prefixes, reduce_prefixes


def main():
    project_root = Path(__file__).resolve().parent.parent
    file_path = Path(project_root) / 'BenchmarkGenPool_cpu.txt'

    prefixes = extract_all_function_names(file_path)
    grouped_prefixes = group_similar_prefixes(prefixes, depth=1)

    for prefix, items in grouped_prefixes.items():
        print(f"Group: {prefix}")
        for item in items:
            print(f"  {item}")
        print()


if __name__ == "__main__":
    main()
