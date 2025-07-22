package parser

import (
	"bufio"
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/shared"
)

const (
	funcNameRegexp    = `\.([^.(]+)(?:\([^)]*\))?$`
	floatRegexp       = `\d+(?:\.\d+)?`
	header            = "flat  flat%   sum%        cum   cum%"
	profileLinelength = 6
	functionNameIndex = 5
)

var (
	funcNameRegexpCompiled = regexp.MustCompile(funcNameRegexp)
	floatRegexpCompiled    = regexp.MustCompile(floatRegexp)
)

// ProfileFilter collects filters for extracting function names from a profile.
type ProfileFilter struct {
	// Include only lines starting with specified prefix
	FunctionPrefixes []string

	// Ignore all functions after the last dot even if includes the above prefix
	IgnoreFunctions []string
}

// GetAllFunctionNames extracts all function names from a profile text file, applying the given filter.
func GetAllFunctionNames(filePath string, filter ProfileFilter) (names []string, err error) {
	scanner, file, err := shared.GetScanner(filePath)
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

		if funcName := extractFunctionName(line, filter.FunctionPrefixes, ignoreSet); funcName != "" {
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

// ShouldKeepLine determines if a line from a profile should be kept based on profile values and ignore filters.
func ShouldKeepLine(line string, agrs *args.LineFilterArgs) bool {
	if line == "" {
		return false
	}

	lineParts := strings.Fields(line)
	if len(lineParts) < profileLinelength {
		return false
	}

	if !filterByNumber(agrs.ProfileFilters, lineParts) {
		return false
	}

	if !filterByIgnoreFunctions(agrs.IgnoreFunctionSet, lineParts) {
		return false
	}

	return filterByIgnorePrefixes(agrs.IgnorePrefixSet, lineParts)
}

func GetAllProfileLines(scanner *bufio.Scanner, lines *[]string) {
	for scanner.Scan() {
		*lines = append(*lines, scanner.Text())
	}
}

func FilterProfileBody(cfg *config.Config, scanner *bufio.Scanner, lines *[]string) {
	profileFilters := cfg.GetProfileFilters()
	ignoreFunctionSet, ignorePrefixSet := cfg.GetIgnoreSets()

	options := &args.LineFilterArgs{
		ProfileFilters:    profileFilters,
		IgnoreFunctionSet: ignoreFunctionSet,
		IgnorePrefixSet:   ignorePrefixSet,
	}

	for scanner.Scan() {
		line := scanner.Text()

		if ShouldKeepLine(line, options) {
			*lines = append(*lines, line)
		}
	}
}

func CollectHeader(scanner *bufio.Scanner, profileType string, lines *[]string) {
	lineCount := 0

	headerIndex := 6
	if profileType != "cpu" {
		headerIndex = 5
	}

	for scanner.Scan() && lineCount < headerIndex {
		*lines = append(*lines, scanner.Text())
		lineCount++
	}
}
