package analyzer

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/config"
	"github.com/AlexsanderHamir/prof/parser"
	"github.com/sashabaranov/go-openai"
)

const (
	permDir  = 0o755
	permFile = 0o644
)

// ValidateBenchmarkDirectories checks if the benchmark directories exist for a given tag and returns the benchmark names.
func ValidateBenchmarkDirectories(tag string) ([]string, error) {
	baseDir := filepath.Join("bench", tag)

	if _, err := os.Stat(baseDir); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no benchmark data found for tag '%s'", tag)
	}

	textDir := filepath.Join(baseDir, "text")
	if _, err := os.Stat(textDir); errors.Is(err, os.ErrNotExist) {
		return nil, fmt.Errorf("no text profiles found in %s", textDir)
	}

	entries, err := os.ReadDir(textDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read text directory: %w", err)
	}

	var benchmarkNames []string
	for _, entry := range entries {
		if entry.IsDir() {
			benchmarkNames = append(benchmarkNames, entry.Name())
		}
	}

	if len(benchmarkNames) == 0 {
		return nil, fmt.Errorf("no benchmark directories found in %s", textDir)
	}

	return benchmarkNames, nil
}

// AnalyzeAllProfiles runs analysis for all benchmarks and profile types for a given tag.
func AnalyzeAllProfiles(tag string, benchmarkNames, profileTypes []string, cfg *config.Config, isFlag bool) error {
	log.Printf("\nStarting comprehensive analysis for tag: %s\n", tag)
	log.Printf("Benchmarks: %v\n", benchmarkNames)
	log.Printf("Profile types: %v\n", profileTypes)
	log.Printf("================================================================================\n")

	for _, benchmarkName := range benchmarkNames {
		for _, profileType := range profileTypes {
			if profileType == "trace" {
				continue
			}

			log.Printf("\nAnalyzing %s (%s)...\n", benchmarkName, profileType)
			if err := sendToModel(tag, benchmarkName, profileType, cfg, isFlag); err != nil {
				return fmt.Errorf("failed to analyze %s (%s): %w", benchmarkName, profileType, err)
			}
		}
	}

	return nil
}

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

func readProfileTextFile(filePath, profileType string, cfg *config.Config) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("cannot read profile file %s: %w", filePath, err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	lineCount := 0

	// Read header
	headerIndex := 6
	if profileType != "cpu" {
		headerIndex = 5
	}

	for scanner.Scan() && lineCount < headerIndex {
		lines = append(lines, scanner.Text())
		lineCount++
	}

	// Read and filter body
	if cfg.AIConfig.UniversalProfileFilter != nil {
		profileValues := map[int]float64{
			0: cfg.AIConfig.UniversalProfileFilter.ProfileValues.Flat,
			1: cfg.AIConfig.UniversalProfileFilter.ProfileValues.FlatPercent,
			2: cfg.AIConfig.UniversalProfileFilter.ProfileValues.SumPercent,
			3: cfg.AIConfig.UniversalProfileFilter.ProfileValues.Cum,
			4: cfg.AIConfig.UniversalProfileFilter.ProfileValues.CumPercent,
		}

		for scanner.Scan() {
			line := scanner.Text()
			if parser.ShouldKeepLine(line, profileValues,
				cfg.AIConfig.UniversalProfileFilter.IgnoreFunctions,
				cfg.AIConfig.UniversalProfileFilter.IgnorePrefixes) {
				lines = append(lines, line)
			}
		}
	} else {
		// No filtering, include all lines
		for scanner.Scan() {
			lines = append(lines, scanner.Text())
		}
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
	if cfg.BaseURL != "https://api.openai.com/v1" {
		config := openai.DefaultConfig(cfg.APIKey)
		config.BaseURL = cfg.BaseURL
		client = openai.NewClientWithConfig(config)
	}

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

	// Ensure directory exists
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
