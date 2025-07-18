package shared

import (
	"bufio"
	"fmt"
	"os"
)

const (
	InfoCollectionSuccess = "All benchmarks and profile processing completed successfully!"
)

const (
	TRACE              = "trace"
	MainDirOutput      = "bench"
	ProfileTextDir     = "text"
	ProfileBinDir      = "bin"
	PermDir            = 0o755
	PermFile           = 0o644
	FunctionsDirSuffix = "_functions"
)

func GetScanner(filePath string) (*bufio.Scanner, *os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read profile file %s: %w", filePath, err)
	}

	scanner := bufio.NewScanner(file)

	return scanner, file, nil
}
