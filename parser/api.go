package parser

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/internal"
)

const (
	funcNameRegexp       = `\.([^.(]+)(?:\([^)]*\))?$`
	floatRegexp          = `\d+(?:\.\d+)?`
	header               = "flat  flat%   sum%        cum   cum%"
	minProfileLinelength = 6
)

// Line Indexes.
const (
	flatIndex           = 0
	flatPercentageIndex = 1
	sumPercentageIndex  = 2
	cumIndex            = 3
	cumPercentageIndex  = 4
	functionNameIndex   = 5
)

var (
	funcNameRegexpCompiled = regexp.MustCompile(funcNameRegexp)
)

// ProfileFilter collects filters for extracting function names from a profile.
type ProfileFilter struct {
	// Include only lines starting with specified prefix
	FunctionPrefixes []string

	// Ignore all functions after the last dot even if includes the above prefix
	IgnoreFunctions []string
}

// GetAllFunctionNames extracts all function names from a profile text file, applying the given filter.
func GetAllFunctionNames(filePath string, filter internal.FunctionFilter) (names []string, err error) {
	scanner, file, err := internal.GetScanner(filePath)
	if err != nil {
		return nil, fmt.Errorf("GetAllFunctionNames Failed: %w", err)
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("file close failed: %w", closeErr)
			}
		}
	}()

	ignoreSet := getFilterSets(filter.IgnoreFunctions)

	var foundHeader bool
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if strings.Contains(line, header) {
			foundHeader = true
			continue
		}

		// Skip lines until we find the header, then process profile data
		if !foundHeader {
			continue
		}

		if funcName := ExtractFunctionName(line, filter.IncludePrefixes, ignoreSet); funcName != "" {
			names = append(names, funcName)
		}
	}

	if err = scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading profile file: %w", err)
	}

	if !foundHeader {
		return nil, errors.New("profile file header not found")
	}

	return names, nil
}
