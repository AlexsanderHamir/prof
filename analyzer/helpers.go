package analyzer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/args"
	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
	"github.com/sashabaranov/go-openai"
)

const (
	formatLineLength = 80
)

func sendToModel(tag, benchmarkName, profileType string, cfg *config.Config, isFlag bool) error {
	profileData, err := getBenchmarkFile(tag, benchmarkName, profileType, cfg)
	if err != nil {
		return fmt.Errorf("failed to get benchmark file: %w", err)
	}

	if profileData == "" {
		return fmt.Errorf("no content found for %s (%s)", benchmarkName, profileType)
	}

	userPrompt, err := getUserPrompt(cfg)
	if err != nil {
		return fmt.Errorf("failed to get user prompt: %w", err)
	}

	args := &args.ModelCallRequiredArgs{
		SystemPrompt:   userPrompt,
		ProfileContent: profileData,
		BenchmarkName:  benchmarkName,
		ProfileType:    profileType,
	}

	analysis, err := requestModelAnalysis(args, cfg)
	if err != nil {
		return fmt.Errorf("failed to request model analysis: %w", err)
	}

	if err = saveAnalysis(tag, benchmarkName, profileType, analysis, isFlag); err != nil {
		return fmt.Errorf("failed to save analysis: %w", err)
	}

	slog.Info("Successfully analyzed and saved results", "benchmarkName", benchmarkName, "profileType", profileType)
	return nil
}

func getBenchmarkFile(tag, benchmarkName, profileType string, cfg *config.Config) (string, error) {
	baseDir := filepath.Join("bench", tag)
	textDir := filepath.Join(baseDir, "text", benchmarkName)
	profileFile := filepath.Join(textDir, fmt.Sprintf("%s_%s.txt", benchmarkName, profileType))

	return readProfileTextFile(profileFile, profileType, cfg)
}

func filterProfileBody(cfg *config.Config, scanner *bufio.Scanner, lines *[]string) {
	profileFilters := cfg.GetProfileFilters()
	ignoreFunctionSet, ignorePrefixSet := cfg.GetIgnoreSets()

	options := &args.LineFilterArgs{
		ProfileFilters:    profileFilters,
		IgnoreFunctionSet: ignoreFunctionSet,
		IgnorePrefixSet:   ignorePrefixSet,
	}

	for scanner.Scan() {
		line := scanner.Text()

		if parser.ShouldKeepLine(line, options) {
			*lines = append(*lines, line)
		}
	}
}

func getAllProfileLines(scanner *bufio.Scanner, lines *[]string) {
	for scanner.Scan() {
		*lines = append(*lines, scanner.Text())
	}
}

func readProfileTextFile(filePath, profileType string, cfg *config.Config) (fileContent string, err error) {
	var lines []string

	scanner, file, err := shared.GetScanner(filePath)
	if err != nil {
		return "", err
	}

	defer func() {
		if closeErr := file.Close(); closeErr != nil {
			if err == nil {
				err = fmt.Errorf("file close failed: %w", closeErr)
			}
		}
	}()

	collectHeader(scanner, profileType, &lines)

	isFilterAvailable := cfg.AIConfig.ProfileFilter != nil

	if isFilterAvailable {
		filterProfileBody(cfg, scanner, &lines)
	} else {
		getAllProfileLines(scanner, &lines)
	}

	if err = scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	fileContent = strings.Join(lines, "\n")
	if fileContent == "" {
		return "", fmt.Errorf("profile file %s is empty", filePath)
	}

	return fileContent, nil
}

// POTENTIAL IMPROVEMENT: shouldn't this be part of the parser ?
func collectHeader(scanner *bufio.Scanner, profileType string, lines *[]string) {
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

func getUserPrompt(cfg *config.Config) (string, error) {
	if cfg.ModelConfig.PromptFileLocation == "" {
		return "", errors.New("prompt_location must be provided in config")
	}

	data, err := os.ReadFile(cfg.ModelConfig.PromptFileLocation)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", cfg.ModelConfig.PromptFileLocation, err)
	}

	return strings.TrimSpace(string(data)), nil
}

func requestModelAnalysis(args *args.ModelCallRequiredArgs, cfg *config.Config) (string, error) {
	client := openai.NewClient(cfg.APIKey)
	// TODO: ??
	if cfg.BaseURL != "https://api.openai.com/v1" {
		config := openai.DefaultConfig(cfg.APIKey)
		config.BaseURL = cfg.BaseURL
		client = openai.NewClientWithConfig(config)
	}

	// TODO: prompt redundancy
	profileInfo := fmt.Sprintf("BenchmarkName: %s\nProfile Type: %s\n\nProfile Content: %s",
		args.BenchmarkName, args.ProfileType, args.ProfileContent)

	slog.Info("Sending request to model", "model", cfg.ModelConfig.Model)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: cfg.ModelConfig.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: args.SystemPrompt,
				},
				{
					Role:    openai.ChatMessageRoleUser,
					Content: profileInfo,
				},
			},
			MaxTokens:   cfg.ModelConfig.MaxTokens,
			Temperature: cfg.ModelConfig.Temperature,
			TopP:        cfg.ModelConfig.TopP,
		},
	)

	if err != nil {
		return "", fmt.Errorf("error during model analysis request: %w", err)
	}

	if len(resp.Choices) == 0 {
		return "", errors.New("no response choices received from model")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return "", errors.New("no content received from model")
	}

	return content, nil
}

func saveAnalysis(tag, benchmarkName, profileType, analysis string, isFlag bool) error {
	analysisFile := getFilePath(tag, benchmarkName, profileType, isFlag)

	if err := os.MkdirAll(filepath.Dir(analysisFile), shared.PermDir); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	var content string
	if isFlag {
		content = analysis
	} else {
		content = fmt.Sprintf("Benchmark: %s\nProfile Type: %s\n%s\n\n%s",
			benchmarkName, profileType, strings.Repeat("=", formatLineLength), analysis)
	}

	if err := os.WriteFile(analysisFile, []byte(content), shared.PermFile); err != nil {
		return fmt.Errorf("cannot save analysis to %s: %w", analysisFile, err)
	}

	slog.Info("Analysis saved", "file", analysisFile)
	return nil
}

func getFilePath(tag, benchmarkName, profileType string, isFlag bool) string {
	if isFlag {
		fileName := fmt.Sprintf("%s_%s.txt", benchmarkName, profileType)
		return filepath.Join(shared.MainDirOutput, tag, shared.ProfileTextDir, benchmarkName, fileName)
	}

	analysisDir := filepath.Join(shared.MainDirOutput, tag, "AI", "generalistic", benchmarkName)
	return filepath.Join(analysisDir, fmt.Sprintf("generalistic_analysis_%s.txt", profileType))
}
