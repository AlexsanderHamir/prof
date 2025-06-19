import os
from pathlib import Path
from tests.constants import BENCHMARK_FILE_NAME, BENCHMARK_MAIN_DIR_NAME, BENCHMARK_PRIME_COUNT, BENCHMARK_MATRIX_MULTIPLY, BENCHMARK_TAG_NAME


def create_benchmark_file(benchmark_path: str) -> None:
    with open(os.path.join(benchmark_path, BENCHMARK_FILE_NAME), "w") as f:
        f.write(f"""package benchmark

    import (
        "testing"
        "math"
    )

    func {BENCHMARK_PRIME_COUNT}(b *testing.B) {{
        for b.Loop() {{
            countPrimes(10000)
        }}
    }}

    func {BENCHMARK_MATRIX_MULTIPLY}(b *testing.B) {{
        a := makeMatrix(100, 100)
        bMat := makeMatrix(100, 100)
        for b.Loop() {{
            multiplyMatrix(a, bMat)
        }}
    }}

    func countPrimes(limit int) int {{
        count := 0
        for i := 2; i < limit; i++ {{
            if isPrime(i) {{
                count++
            }}
        }}
        return count
    }}

    func isPrime(n int) bool {{
        if n < 2 {{
            return false
        }}
        sqrtN := int(math.Sqrt(float64(n)))
        for i := 2; i <= sqrtN; i++ {{
            if n%i == 0 {{
                return false
            }}
        }}
        return true
    }}

    func makeMatrix(rows, cols int) [][]float64 {{
        m := make([][]float64, rows)
        for i := range m {{
            m[i] = make([]float64, cols)
            for j := range m[i] {{
                m[i][j] = float64(i + j)
            }}
        }}
        return m
    }}

    func multiplyMatrix(a, b [][]float64) [][]float64 {{
        rows, cols := len(a), len(b[0])
        result := make([][]float64, rows)
        for i := range result {{
            result[i] = make([]float64, cols)
            for j := 0; j < cols; j++ {{
                sum := 0.0
                for k := 0; k < len(b); k++ {{
                    sum += a[i][k] * b[k][j]
                }}
                result[i][j] = sum
            }}
        }}
        return result
    }}
    """)


def verify_required_structure_exists(base_path: Path, directories_to_check: list[Path]) -> None:
    for directory in directories_to_check:
        assert directory.exists(), f"Directory {directory} does not exist"

        for file_path in directory.iterdir():
            assert file_path.is_file() and file_path.stat().st_size > 100, f"File {file_path} is empty"

    desc_path = base_path / "description.txt"
    assert desc_path.exists(), f"Description file {desc_path} does not exist"


def verify_benchmark_output_structure(benchmark_path: str, benchmark_name: str) -> None:
    base_path = Path(benchmark_path) / BENCHMARK_MAIN_DIR_NAME / BENCHMARK_TAG_NAME

    directories_to_check = [
        base_path / "bin" / benchmark_name,
        base_path / "cpu_functions" / benchmark_name,
        base_path / "memory_functions" / benchmark_name,
        base_path / "mutex_functions" / benchmark_name,
        base_path / "text" / benchmark_name,
    ]

    verify_required_structure_exists(base_path, directories_to_check)
