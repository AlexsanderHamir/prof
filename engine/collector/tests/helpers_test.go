package tests

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/AlexsanderHamir/prof/internal"
)

// Test helper function to create a mock binary file for testing
func createMockBinaryFile(t *testing.T, dir, filename string) string {
	content := []byte("mock binary content for testing")
	filepath := filepath.Join(dir, filename)
	err := os.WriteFile(filepath, content, internal.PermFile)
	if err != nil {
		t.Fatalf("Failed to create mock binary file: %v", err)
	}
	return filepath
}
