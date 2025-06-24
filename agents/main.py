from pathlib import Path
from agents.helpers import compress_profile_content


def main():
    project_root = Path(__file__).resolve().parent.parent
    file_path = Path(project_root) / 'BenchmarkGenPool_cpu.txt'

    compress_profile_content(file_path)


if __name__ == "__main__":
    main()
