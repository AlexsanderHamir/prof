package shared

import (
	"bufio"
	"fmt"
	"os"
)

const (
	TRACE                        = "trace"
	MainDirOutput                = "bench"
	Profile_text_files_directory = "text"
	Profile_bin_files_directory  = "bin"
	PermDir                      = 0o755
	PermFile                     = 0o644
)

func GetScanner(filePath string) (*bufio.Scanner, *os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read profile file %s: %w", filePath, err)
	}

	scanner := bufio.NewScanner(file)

	return scanner, file, nil
}
