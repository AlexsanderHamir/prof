package collect

import (
	"path/filepath"
	"strings"
)

func stemFromPath(fullPath string) string {
	normalized := strings.ReplaceAll(fullPath, `\`, "/")
	file := filepath.Base(normalized)
	return strings.TrimSuffix(file, filepath.Ext(file))
}
