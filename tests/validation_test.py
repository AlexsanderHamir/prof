import subprocess
import os

from exit_codes import EXIT_CODE_MISSING_ARGUMENTS, EXIT_CODE_MISSING_CONFIG_FILE


def test_no_arguments():
    project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    prof_path = os.path.join(project_root, 'prof')

    result = subprocess.run([prof_path], capture_output=True, text=True)

    assert result.returncode == EXIT_CODE_MISSING_ARGUMENTS, f"prof failed with error: {result.stderr}"


def test_no_config_file():
    project_root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    prof_path = os.path.join(project_root, 'prof')

    result = subprocess.run([prof_path, "-benchmarks", "[BenchmarkGenPool]", "-profiles", "[cpu]", "-tag", "test", "-count", "5"], capture_output=True, text=True)

    assert result.returncode == EXIT_CODE_MISSING_CONFIG_FILE, f"prof failed with error: {result.stderr}"
