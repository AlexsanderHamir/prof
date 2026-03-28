package collector

import (
	"os"

	"github.com/AlexsanderHamir/prof/internal"
)

func ensureDirExists(basePath string) error {
	_, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(basePath, internal.PermDir)
		}
		return err
	}
	return nil
}
