import subprocess
import os

from exit_codes import EXIT_CODE_MISSING_ARGUMENTS


def test_no_arguments():
    project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    prof_path = os.path.join(project_root, 'prof')

    result = subprocess.run([prof_path], capture_output=True, text=True)

    assert result.returncode == EXIT_CODE_MISSING_ARGUMENTS, f"prof failed with error: {result.stderr}"
