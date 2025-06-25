import json
from pathlib import Path
import shutil
import subprocess
import os

import pytest

from exit_codes import EXIT_CODE_MISSING_ARGUMENTS, EXIT_CODE_MISSING_BRACKETS, MISSING_CONFIG_FILE, EXIT_CODE_MODULE_ERROR, EXIT_CODE_SUCCESS
from tests.e2e.constants import BENCHMARK_GEN_POOL


def test_no_arguments():
    project_root = Path(__file__).resolve().parent.parent.parent
    prof_path = os.path.join(project_root, 'prof')

    result = subprocess.run([prof_path], capture_output=True, text=True)

    assert result.returncode == EXIT_CODE_MISSING_ARGUMENTS, f"prof failed with error: {result.stderr}"


def test_no_config_file():
    project_root = Path(__file__).resolve().parent.parent.parent
    prof_path = os.path.join(project_root, 'prof')

    result = subprocess.run([prof_path, "-benchmarks", f"[{BENCHMARK_GEN_POOL}]", "-profiles", "[cpu]", "-tag", "test", "-count", "5"], capture_output=True, text=True)

    assert result.returncode == MISSING_CONFIG_FILE, f"prof failed with error: {result.stderr}"


def test_setup_command():
    project_root = Path(__file__).resolve().parent.parent.parent
    prof_path = os.path.join(project_root, 'prof')

    config_template_path = os.path.join(project_root, "config_template.json")

    result = subprocess.run([prof_path, "setup", "--create-template", "--output-path", config_template_path], capture_output=True, text=True)
    assert result.returncode == EXIT_CODE_SUCCESS, f"prof failed with error: {result.stderr}"
    assert os.path.exists(config_template_path), "config_template.json does not exist"

    try:
        with open(config_template_path, "r") as f:
            config = json.load(f)

        expected_keys = {'api_key', 'base_url', 'model_config', 'benchmark_configs', 'ai_config'}
        actual_keys = set(config.keys())
        assert actual_keys == expected_keys, f"Expected keys {expected_keys}, but got {actual_keys}"

    finally:
        if os.path.exists(config_template_path):
            os.remove(config_template_path)


@pytest.mark.parametrize("args,expected_code", [
    (["-benchmarks", f"[{BENCHMARK_GEN_POOL}]", "-profiles", "cpu", "-tag", "test", "-count", "5"], EXIT_CODE_MISSING_BRACKETS),
    (["-benchmarks", BENCHMARK_GEN_POOL, "-profiles", "[cpu]", "-tag", "test", "-count", "5"], EXIT_CODE_MISSING_BRACKETS),
    (["-benchmarks", f"{BENCHMARK_GEN_POOL}]", "-profiles", "[cpu]", "-tag", "test", "-count", "5"], EXIT_CODE_MISSING_BRACKETS),
    (["-benchmarks", f"[{BENCHMARK_GEN_POOL}", "-profiles", "[cpu]", "-tag", "test", "-count", "5"], EXIT_CODE_MISSING_BRACKETS),
    (["-benchmarks", f"[{BENCHMARK_GEN_POOL}]", "-profiles", "cpu]", "-tag", "test", "-count", "5"], EXIT_CODE_MISSING_BRACKETS),
    (["-benchmarks", f"[{BENCHMARK_GEN_POOL}]", "-profiles", "[cpu", "-tag", "test", "-count", "5"], EXIT_CODE_MISSING_BRACKETS),
    (["-benchmarks", f"[{BENCHMARK_GEN_POOL}]", "-profiles", "[cpu]", "-tag", "test", "-count", "5"], EXIT_CODE_MODULE_ERROR),
    (["-benchmarks", f"[{BENCHMARK_GEN_POOL}, BenchmarkGenPool2, BenchmarkGenPool3]", "-profiles", "[cpu, memory, mutex]", "-tag", "test", "-count", "5"], EXIT_CODE_MODULE_ERROR),
])
def test_brackets_variants(args, expected_code):
    project_root = Path(__file__).resolve().parent.parent.parent
    prof_path = os.path.join(project_root, 'prof')
    benchmark_path = os.path.join(project_root, "bench")

    config_template_path = os.path.join(project_root, "config_template.json")
    subprocess.run([prof_path, "setup", "--create-template", "--output-path", config_template_path], capture_output=True, text=True, check=True)

    try:
        result = subprocess.run([prof_path] + args, capture_output=True, text=True)
        assert result.returncode == expected_code, f"Args {args} failed with: {result.stderr}"
    finally:
        if os.path.exists(config_template_path):
            os.remove(config_template_path)
        if os.path.exists(benchmark_path):
            shutil.rmtree(benchmark_path)
