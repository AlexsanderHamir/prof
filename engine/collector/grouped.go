package collector

import (
	"fmt"
	"os"

	"github.com/AlexsanderHamir/prof/internal"
	"github.com/AlexsanderHamir/prof/parser"
)

// WriteGroupedPackageProfile writes a package/module grouped markdown report for a pprof binary.
func WriteGroupedPackageProfile(binaryPath, outputPath string, filter internal.FunctionFilter) error {
	text, err := parser.OrganizeProfileByPackageV2(binaryPath, filter)
	if err != nil {
		return fmt.Errorf("organize profile by package: %w", err)
	}
	return os.WriteFile(outputPath, []byte(text), internal.PermFile)
}
