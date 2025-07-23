package utils

import (
	"crypto/sha256"
	"fmt"
	"sort"
	"strings"
)

// StringProcessor provides string manipulation utilities
type StringProcessor struct {
	data []string
}

// NewStringProcessor creates a new string processor
func NewStringProcessor() *StringProcessor {
	return &StringProcessor{
		data: make([]string, 0),
	}
}

// AddString adds a string to the processor
func (sp *StringProcessor) AddString(s string) {
	sp.data = append(sp.data, s)
}

// ProcessStrings performs various operations on stored strings
func (sp *StringProcessor) ProcessStrings() map[string]interface{} {
	result := make(map[string]interface{})
	
	// Sort strings
	sorted := make([]string, len(sp.data))
	copy(sorted, sp.data)
	sort.Strings(sorted)
	result["sorted"] = sorted
	
	// Calculate total length
	totalLen := 0
	for _, s := range sp.data {
		totalLen += len(s)
	}
	result["total_length"] = totalLen
	
	// Generate hashes
	hashes := make([]string, len(sp.data))
	for i, s := range sp.data {
		hash := sha256.Sum256([]byte(s))
		hashes[i] = fmt.Sprintf("%x", hash)
	}
	result["hashes"] = hashes
	
	return result
}

// Calculator provides mathematical operations
type Calculator struct{}

// NewCalculator creates a new calculator
func NewCalculator() *Calculator {
	return &Calculator{}
}

// Fibonacci calculates fibonacci number (CPU intensive)
func (c *Calculator) Fibonacci(n int) int {
	if n <= 1 {
		return n
	}
	return c.Fibonacci(n-1) + c.Fibonacci(n-2)
}

// MatrixMultiply performs matrix multiplication
func (c *Calculator) MatrixMultiply(a, b [][]int) [][]int {
	if len(a) == 0 || len(b) == 0 || len(a[0]) != len(b) {
		return nil
	}
	
	rows, cols := len(a), len(b[0])
	result := make([][]int, rows)
	for i := range result {
		result[i] = make([]int, cols)
	}
	
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			for k := 0; k < len(b); k++ {
				result[i][j] += a[i][k] * b[k][j]
			}
		}
	}
	
	return result
}

// DataGenerator generates test data
type DataGenerator struct{}

// NewDataGenerator creates a new data generator
func NewDataGenerator() *DataGenerator {
	return &DataGenerator{}
}

// GenerateStrings creates a slice of strings
func (dg *DataGenerator) GenerateStrings(count int) []string {
	result := make([]string, count)
	for i := 0; i < count; i++ {
		result[i] = fmt.Sprintf("generated_string_%d_%s", i, strings.Repeat("x", i%100))
	}
	return result
}

// GenerateMatrix creates a matrix of given size
func (dg *DataGenerator) GenerateMatrix(rows, cols int) [][]int {
	matrix := make([][]int, rows)
	for i := range matrix {
		matrix[i] = make([]int, cols)
		for j := range matrix[i] {
			matrix[i][j] = (i + j) % 100
		}
	}
	return matrix
}