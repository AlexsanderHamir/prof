package benchmark

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/internal"
)

const (
	// Minimum number of regex capture groups expected for benchmark function
	minCaptureGroups = 2
)

// DiscoverBenchmarks scans the Go module for benchmark functions and returns their names.
// A benchmark is identified by functions matching:
//
//	func BenchmarkXxx(b *testing.B) { ... }
func DiscoverBenchmarks() ([]string, error) {
	root, err := internal.FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root: %w", err)
	}

	return scanForBenchmarks(root)
}

// scanForBenchmarks walks the directory tree looking for benchmark functions
func scanForBenchmarks(root string) ([]string, error) {
	pattern := regexp.MustCompile(`(?m)^\s*func\s+(Benchmark[\w\d_]+)\s*\(\s*b\s*\*\s*testing\.B\s*\)\s*{`)
	seen := make(map[string]struct{})
	var names []string

	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return handleDirectory(path)
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		return processTestFile(path, pattern, seen, &names)
	})
	if err != nil {
		return nil, err
	}

	return names, nil
}

// handleDirectory determines if a directory should be traversed or skipped
func handleDirectory(path string) error {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") || base == "vendor" {
		return filepath.SkipDir
	}
	return nil
}

// processTestFile extracts benchmark function names from a test file
func processTestFile(path string, pattern *regexp.Regexp, seen map[string]struct{}, names *[]string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}

	matches := pattern.FindAllSubmatch(data, -1)
	for _, m := range matches {
		if len(m) >= minCaptureGroups {
			name := string(m[1])
			if _, ok := seen[name]; !ok {
				seen[name] = struct{}{}
				*names = append(*names, name)
			}
		}
	}
	return nil
}
