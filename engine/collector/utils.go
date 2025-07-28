package collector

import (
	"path"
	"path/filepath"
	"strings"
)

func getPprofTextParams() []string {
	return []string{
		"-cum",
		"-edgefraction=0",
		"-nodefraction=0",
		"-top",
	}
}

func getFileName(fullPath string) string {
	file := filepath.Base(fullPath)
	fileName := strings.TrimSuffix(file, filepath.Ext(file))

	return fileName
}

func createProfileDirectory(tagDir, fileName string) (string, error) {
	profileDirPath := path.Join(tagDir, fileName)
	if err := ensureDirExists(profileDirPath); err != nil {
		return "", err
	}

	return profileDirPath, nil
}
