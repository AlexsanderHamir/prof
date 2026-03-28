package collector

import (
	"path/filepath"
	"strings"
)

func pprofTextListArgs() []string {
	return []string{
		"-cum",
		"-edgefraction=0",
		"-nodefraction=0",
		"-top",
	}
}

func stemFromPath(fullPath string) string {
	file := filepath.Base(fullPath)
	return strings.TrimSuffix(file, filepath.Ext(file))
}

func profileSubdir(tagDir, fileName string) (string, error) {
	profileDirPath := filepath.Join(tagDir, fileName)
	if err := ensureDirExists(profileDirPath); err != nil {
		return "", err
	}
	return profileDirPath, nil
}
