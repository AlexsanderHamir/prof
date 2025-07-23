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
	floatRegexpCompiled    = regexp.MustCompile(floatRegexp)
)

// ProfileFilter collects filters for extracting function names from a profile.
type ProfileFilter struct {
	// Include only lines starting with specified prefix
	FunctionPrefixes []string

	// Ignore all functions after the last dot even if includes the above prefix
	IgnoreFunctions []string
}

type LineObj struct {
	FnName         string
	Flat           float64
	FlatPercentage float64
	SumPercentage  float64
	Cum            float64
	CumPercentage  float64
}

func TurnLinesIntoObjects(profilePath, profileType string) ([]*LineObj, error) {
	var lines []string

	scanner, file, err := shared.GetScanner(profilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	shouldRemove := true
	CollectOrRemoveHeader(scanner, profileType, &lines, shouldRemove)

	GetAllProfileLines(scanner, &lines)

	lineObjs, err := createLineObjects(lines)
	if err != nil {
		return nil, fmt.Errorf("failed creating line objects : %w", err)
	}

	return lineObjs, err
}

func createLineObjects(lines []string) ([]*LineObj, error) {
	var lineObjs []*LineObj

	for _, line := range lines {
		lineParts := strings.Fields(line)
		lineLength := len(lineParts)

		if lineLength < minProfileLinelength {
			return nil, fmt.Errorf("line(%d) is smaller than minimum(%d)", lineLength, minProfileLinelength)
		}

		floats, err := getFloatsFromLineParts(lineParts)
		if err != nil {
			return nil, fmt.Errorf("float extraction failed: %w", err)
		}

		funcName := strings.Join(lineParts[functionNameIndex:], " ")
		lineobj := &LineObj{
			FnName:         funcName,
			Flat:           floats.Flat,
			FlatPercentage: floats.FlatPercentage,
			SumPercentage:  floats.Sum,
			Cum:            floats.Cum,
			CumPercentage:  floats.CumPercentage,
		}

		lineObjs = append(lineObjs, lineobj)
	}

	return lineObjs, nil
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
	if len(lineParts) < minProfileLinelength {
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

func CollectOrRemoveHeader(scanner *bufio.Scanner, profileType string, lines *[]string, shouldRemove bool) {
	lineCount := 0

	headerIndex := 6
	if profileType != "cpu" {
		headerIndex = 5
	}

	for lineCount < headerIndex && scanner.Scan() {
		if !shouldRemove {
			*lines = append(*lines, scanner.Text())
		}
		lineCount++
	}
}
