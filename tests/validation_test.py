import json
from pathlib import Path
import subprocess
import os

from exit_codes import EXIT_CODE_MISSING_ARGUMENTS, EXIT_CODE_MISSING_CONFIG_FILE, EXIT_CODE_SUCCESS


def test_no_arguments():
    project_root = Path(__file__).resolve().parent.parent
    prof_path = os.path.join(project_root, 'prof')

    result = subprocess.run([prof_path], capture_output=True, text=True)

    assert result.returncode == EXIT_CODE_MISSING_ARGUMENTS, f"prof failed with error: {result.stderr}"


def test_no_config_file():
    project_root = Path(__file__).resolve().parent.parent
    prof_path = os.path.join(project_root, 'prof')

    result = subprocess.run([prof_path, "-benchmarks", "[BenchmarkGenPool]", "-profiles", "[cpu]", "-tag", "test", "-count", "5"], capture_output=True, text=True)

    assert result.returncode == EXIT_CODE_MISSING_CONFIG_FILE, f"prof failed with error: {result.stderr}"


def test_setup_command():
    project_root = Path(__file__).resolve().parent.parent
    prof_path = os.path.join(project_root, 'prof')

    config_template_path = os.path.join(project_root, "config_template.json")
    result = subprocess.run([prof_path, "setup", "--create-template", "--output-path", config_template_path], capture_output=True, text=True)

    assert result.returncode == EXIT_CODE_SUCCESS, f"prof failed with error: {result.stderr}"

    assert os.path.exists(config_template_path), "config_template.json does not exist"

    with open(config_template_path, "r") as f:
        config = json.load(f)

    expected_keys = {'api_key', 'base_url', 'model_config', 'benchmark_configs'}
    actual_keys = set(config.keys())
    assert actual_keys == expected_keys, f"Expected keys {expected_keys}, but got {actual_keys}"

    os.remove(config_template_path)
