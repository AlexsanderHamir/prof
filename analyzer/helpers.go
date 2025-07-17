package analyzer

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/parser"
	"github.com/AlexsanderHamir/prof/shared"
	"github.com/sashabaranov/go-openai"
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

	analysis, err := requestModelAnalysis(userPrompt, profileData, benchmarkName, profileType, cfg)
	if err != nil {
		return fmt.Errorf("failed to request model analysis: %w", err)
	}

	if err := saveAnalysis(tag, benchmarkName, profileType, analysis, isFlag); err != nil {
		return fmt.Errorf("failed to save analysis: %w", err)
	}

	log.Printf("Successfully analyzed and saved results for %s (%s)\n", benchmarkName, profileType)
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

	for scanner.Scan() {
		line := scanner.Text()
		if parser.ShouldKeepLine(line, profileFilters, ignoreFunctionSet, ignorePrefixSet) {
			*lines = append(*lines, line)
		}
	}
}

func getAllProfileLines(scanner *bufio.Scanner, lines *[]string) {
	for scanner.Scan() {
		*lines = append(*lines, scanner.Text())
	}
}

func readProfileTextFile(filePath, profileType string, cfg *config.Config) (string, error) {
	var lines []string

	scanner, file, err := shared.GetScanner(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	collectHeader(scanner, profileType, &lines)

	isFilterAvailable := cfg.AIConfig.UniversalProfileFilter != nil

	if isFilterAvailable {
		filterProfileBody(cfg, scanner, &lines)
	} else {
		getAllProfileLines(scanner, &lines)
	}

	if err := scanner.Err(); err != nil {
		return "", fmt.Errorf("error reading file: %w", err)
	}

	content := strings.Join(lines, "\n")
	if content == "" {
		return "", fmt.Errorf("profile file %s is empty", filePath)
	}

	return content, nil
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
	if cfg.ModelConfig.PromptLocation == "" {
		return "", fmt.Errorf("prompt_location must be provided in config")
	}

	data, err := os.ReadFile(cfg.ModelConfig.PromptLocation)
	if err != nil {
		return "", fmt.Errorf("failed to read prompt file %s: %w", cfg.ModelConfig.PromptLocation, err)
	}

	return strings.TrimSpace(string(data)), nil
}

func requestModelAnalysis(systemPrompt, profileContent, benchmarkName, profileType string, cfg *config.Config) (string, error) {
	client := openai.NewClient(cfg.APIKey)
	// TODO: ??
	if cfg.BaseURL != "https://api.openai.com/v1" {
		config := openai.DefaultConfig(cfg.APIKey)
		config.BaseURL = cfg.BaseURL
		client = openai.NewClientWithConfig(config)
	}

	// TODO: prompt redundancy
	profileInfo := fmt.Sprintf("BenchmarkName: %s\nProfile Type: %s\n\nProfile Content: %s",
		benchmarkName, profileType, profileContent)

	log.Printf("\nSending request to model: %s\n", cfg.ModelConfig.Model)

	resp, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: cfg.ModelConfig.Model,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleSystem,
					Content: systemPrompt,
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
		return "", fmt.Errorf("no response choices received from model")
	}

	content := resp.Choices[0].Message.Content
	if content == "" {
		return "", fmt.Errorf("no content received from model")
	}

	return content, nil
}

func saveAnalysis(tag, benchmarkName, profileType, analysis string, isFlag bool) error {
	analysisFile := getFilePath(tag, benchmarkName, profileType, isFlag)

	if err := os.MkdirAll(filepath.Dir(analysisFile), permDir); err != nil {
		return fmt.Errorf("failed to create analysis directory: %w", err)
	}

	var content string
	if isFlag {
		content = analysis
	} else {
		content = fmt.Sprintf("Benchmark: %s\nProfile Type: %s\n%s\n\n%s",
			benchmarkName, profileType, strings.Repeat("=", 80), analysis)
	}

	if err := os.WriteFile(analysisFile, []byte(content), permFile); err != nil {
		return fmt.Errorf("cannot save analysis to %s: %w", analysisFile, err)
	}

	log.Printf("Analysis saved to: %s\n", analysisFile)
	return nil
}

func getFilePath(tag, benchmarkName, profileType string, isFlag bool) string {
	if isFlag {
		return filepath.Join("bench", tag, "text", benchmarkName, fmt.Sprintf("%s_%s.txt", benchmarkName, profileType))
	}

	analysisDir := filepath.Join("bench", tag, "AI", "generalistic", benchmarkName)
	return filepath.Join(analysisDir, fmt.Sprintf("generalistic_analysis_%s.txt", profileType))
}
