"""
Constants for benchmark names used in tests and helpers.
"""

BENCHMARK_FILE_NAME = "benchmark_test.go"
BENCHMARK_TEST_DIR_NAME = "bench_test"
BENCHMARK_TAG_NAME = "testing"
BENCHMARK_MAIN_DIR_NAME = "bench"

BENCHMARK_PRIME_COUNT = "BenchmarkPrimeCount"
BENCHMARK_MATRIX_MULTIPLY = "BenchmarkMatrixMultiply"
BENCHMARK_GEN_POOL = "BenchmarkGenPool"

AVAILABLE_BENCHMARKS = [
    BENCHMARK_PRIME_COUNT,
    BENCHMARK_MATRIX_MULTIPLY,
    BENCHMARK_GEN_POOL,
]
