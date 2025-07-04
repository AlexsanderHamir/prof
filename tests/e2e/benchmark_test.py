import os
from pathlib import Path
import subprocess

from tests.e2e.helpers import create_benchmark_file, verify_benchmark_output_structure
from tests.e2e.constants import BENCHMARK_MATRIX_MULTIPLY, BENCHMARK_PRIME_COUNT, BENCHMARK_TAG_NAME, BENCHMARK_TEST_DIR_NAME

import shutil


def test_single_benchmark():
    project_root = Path(__file__).resolve().parent.parent.parent
    prof_path = os.path.join(project_root, 'prof')

    benchmark_path = os.path.join(project_root, BENCHMARK_TEST_DIR_NAME)
    os.makedirs(benchmark_path, exist_ok=True)

    try:
        create_benchmark_file(benchmark_path)
        subprocess.run(["go", "mod", "init", "benchmark"], cwd=benchmark_path, capture_output=True, text=True)
        subprocess.run([prof_path, "setup", "--create-template"], cwd=benchmark_path, capture_output=True, text=True)

        result = subprocess.run([prof_path, "-benchmarks", f"[{BENCHMARK_PRIME_COUNT}]", "-profiles", "[cpu, memory, mutex]", "-tag", BENCHMARK_TAG_NAME, "-count", "1"], cwd=benchmark_path, capture_output=True, text=True)
        assert result.returncode == 0, f"prof failed with error: {result.stderr}"

        verify_benchmark_output_structure(benchmark_path, BENCHMARK_PRIME_COUNT)

    finally:
        shutil.rmtree(benchmark_path, ignore_errors=True)


def test_multiple_benchmarks():
    project_root = Path(__file__).resolve().parent.parent.parent
    prof_path = os.path.join(project_root, 'prof')

    benchmark_path = os.path.join(project_root, BENCHMARK_TEST_DIR_NAME)
    os.makedirs(benchmark_path, exist_ok=True)

    try:
        create_benchmark_file(benchmark_path)
        subprocess.run(["go", "mod", "init", "benchmark"], cwd=benchmark_path, capture_output=True, text=True)
        subprocess.run([prof_path, "setup", "--create-template"], cwd=benchmark_path, capture_output=True, text=True)

        result = subprocess.run([prof_path, "-benchmarks", f"[{BENCHMARK_PRIME_COUNT}, {BENCHMARK_MATRIX_MULTIPLY}]", "-profiles", "[cpu, memory, mutex]", "-tag", BENCHMARK_TAG_NAME, "-count", "1"], cwd=benchmark_path, capture_output=True, text=True)
        assert result.returncode == 0, f"prof failed with error: {result.stderr}"

        verify_benchmark_output_structure(benchmark_path, BENCHMARK_PRIME_COUNT)
        verify_benchmark_output_structure(benchmark_path, BENCHMARK_MATRIX_MULTIPLY)

    finally:
        shutil.rmtree(benchmark_path, ignore_errors=True)
