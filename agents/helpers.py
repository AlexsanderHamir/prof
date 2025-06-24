import os
from pathlib import Path
import sys


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
