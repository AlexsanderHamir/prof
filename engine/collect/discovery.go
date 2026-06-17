package collect

import (
	"path/filepath"
	"regexp"
	"strings"
)

func scanForBenchmarks(root string) ([]string, error) {
	pattern := regexp.MustCompile(`(?m)^\s*func\s+(Benchmark[\w\d_]+)\s*\(\s*b\s*\*\s*testing\.B\s*\)\s*{`)
	seen := make(map[string]struct{})
	var names []string

	err := walkTestGoFiles(root, func(_ string, data []byte) error {
		matches := pattern.FindAllSubmatch(data, -1)
		for _, m := range matches {
			if len(m) >= minCaptureGroups {
				name := string(m[1])
				if _, ok := seen[name]; !ok {
					seen[name] = struct{}{}
					names = append(names, name)
				}
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return names, nil
}

func handleDirectory(path string) error {
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") || base == "vendor" {
		return filepath.SkipDir
	}
	return nil
}
