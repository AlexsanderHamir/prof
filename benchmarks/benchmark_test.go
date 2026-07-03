package benchmarks_test

import (
	"testing"

	"github.com/AlexsanderHamir/prof/benchmarks/utils"
)

func BenchmarkStringProcessor(b *testing.B) {
	processor := utils.NewStringProcessor()
	generator := utils.NewDataGenerator()

	strings := generator.GenerateStrings(1000)
	for _, s := range strings {
		processor.AddString(s)
	}

	b.ResetTimer()
	for range b.N {
		// Keep GenerateStrings on the CPU hot path so filtered profiles still list it
		// (ProcessStrings is ignored in typical prof.json filters; without this, cpu.out
		// can lack any matching benchmarks module symbols).
		_ = generator.GenerateStrings(2000)
		result := processor.ProcessStrings()
		_ = result
	}
}

func BenchmarkFibonacci(b *testing.B) {
	calc := utils.NewCalculator()

	b.ResetTimer()
	for range b.N {
		result := calc.Fibonacci(25)
		_ = result
	}
}

func BenchmarkMatrixMultiplication(b *testing.B) {
	calc := utils.NewCalculator()
	generator := utils.NewDataGenerator()

	matrixA := generator.GenerateMatrix(50, 50)
	matrixB := generator.GenerateMatrix(50, 50)

	b.ResetTimer()
	for range b.N {
		result := calc.MatrixMultiply(matrixA, matrixB)
		_ = result
	}
}

func BenchmarkDataGeneration(b *testing.B) {
	generator := utils.NewDataGenerator()

	b.ResetTimer()
	for range b.N {
		strings := generator.GenerateStrings(500)
		matrix := generator.GenerateMatrix(20, 20)
		_ = strings
		_ = matrix
	}
}
