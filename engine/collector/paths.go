package collector

import (
	"path/filepath"
	"strings"
)

// stemFromPath returns the filename without extension. fullPath may come from
// another OS (e.g. Windows-style paths in tests or copied paths); normalize
// '\' to '/' before filepath.Base so Linux CI agrees with Windows.
func stemFromPath(fullPath string) string {
	normalized := strings.ReplaceAll(fullPath, `\`, "/")
	file := filepath.Base(normalized)
	return strings.TrimSuffix(file, filepath.Ext(file))
}

func profileSubdir(tagDir, fileName string) (string, error) {
	profileDirPath := filepath.Join(tagDir, fileName)
	if err := ensureDirExists(profileDirPath); err != nil {
		return "", err
	}
	return profileDirPath, nil
}
