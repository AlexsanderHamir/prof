package main

import (
	"testing"
	"test-environment/utils"
)

func BenchmarkStringProcessor(b *testing.B) {
	processor := utils.NewStringProcessor()
	generator := utils.NewDataGenerator()

	// Generate test data
	strings := generator.GenerateStrings(1000)
	for _, s := range strings {
		processor.AddString(s)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := processor.ProcessStrings()
		_ = result
	}
}

func BenchmarkFibonacci(b *testing.B) {
	calc := utils.NewCalculator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := calc.Fibonacci(25)
		_ = result
	}
}

func BenchmarkMatrixMultiplication(b *testing.B) {
	calc := utils.NewCalculator()
	generator := utils.NewDataGenerator()

	// Generate test matrices
	matrixA := generator.GenerateMatrix(50, 50)
	matrixB := generator.GenerateMatrix(50, 50)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		result := calc.MatrixMultiply(matrixA, matrixB)
		_ = result
	}
}

func BenchmarkDataGeneration(b *testing.B) {
	generator := utils.NewDataGenerator()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		strings := generator.GenerateStrings(500)
		matrix := generator.GenerateMatrix(20, 20)
		_ = strings
		_ = matrix
	}
}
