package collector

import (
	"os"

	"github.com/AlexsanderHamir/prof/shared"
)

func ensureDirExists(basePath string) error {
	_, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(basePath, shared.PermDir)
		}
		return err
	}

	return nil
}
