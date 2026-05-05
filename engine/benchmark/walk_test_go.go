package benchmark

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// walkTestGoFiles walks root like scanForBenchmarks / findBenchmarkPackageDir: skip hidden dirs and vendor,
// then invokes fn for each *_test.go file body.
func walkTestGoFiles(root string, fn func(path string, data []byte) error) error {
	return filepath.WalkDir(root, func(path string, d fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if d.IsDir() {
			return handleDirectory(path)
		}
		if !strings.HasSuffix(path, "_test.go") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return fn(path, data)
	})
}
