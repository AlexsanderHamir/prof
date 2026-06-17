package collect

import (
	"os"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

func ensureDirExists(basePath string) error {
	_, err := os.Stat(basePath)
	if err != nil {
		if os.IsNotExist(err) {
			return os.MkdirAll(basePath, workspace.PermDir)
		}
		return err
	}
	return nil
}
