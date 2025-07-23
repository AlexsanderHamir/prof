package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/shared"
)

func (cfg *Config) GetIgnoreSets() (map[string]struct{}, map[string]struct{}) {
	ignoreFunctions := cfg.AIConfig.ProfileFilter.IgnoreFunctions
	ignorePrefixes := cfg.AIConfig.ProfileFilter.IgnorePrefixes

	ignoreFunctionSet := make(map[string]struct{})
	for _, f := range ignoreFunctions {
		ignoreFunctionSet[f] = struct{}{}
	}

	ignorePrefixSet := make(map[string]struct{})
	for _, p := range ignorePrefixes {
		ignorePrefixSet[p] = struct{}{}
	}

	return ignoreFunctionSet, ignorePrefixSet
}

func (cfg *Config) GetProfileFilters() map[int]float64 {
	profileFilters := map[int]float64{
		0: cfg.AIConfig.ProfileFilter.Thresholds.Flat,
		1: cfg.AIConfig.ProfileFilter.Thresholds.FlatPercent,
		2: cfg.AIConfig.ProfileFilter.Thresholds.SumPercent,
		3: cfg.AIConfig.ProfileFilter.Thresholds.Cum,
		4: cfg.AIConfig.ProfileFilter.Thresholds.CumPercent,
	}

	return profileFilters
}

// LoadFromFile loads and validates config from a JSON file.
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err = json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// CreateTemplate creates a template configuration file from the actual Config struct
// with pre-built examples.
func CreateTemplate(outputPath string) error {
	if outputPath == "" {
		outputPath = "./config_template.json"
	}

	template := Config{
		FunctionFilter: map[string]FunctionFilter{
			"BenchmarkGenPool": {
				IncludePrefixes: []string{
					"github.com/example/GenPool",
					"github.com/example/GenPool/internal",
					"github.com/example/GenPool/pkg",
				},
				IgnoreFunctions: []string{"init", "TestMain", "BenchmarkMain"},
			},
		},
		AIConfig: AIConfig{
			APIKey:  "your-api-key-here",
			BaseURL: "https://api.openai.com/v1",
			ModelConfig: ModelConfig{
				Model:              "gpt-4-turbo-preview",
				MaxTokens:          0,
				Temperature:        0.0,
				TopP:               0.0,
				PromptFileLocation: "path/to/your/system_prompt.txt",
			},
			AllBenchmarks:      true,
			AllProfiles:        true,
			SpecificBenchmarks: []string{},
			SpecificProfiles:   []string{},
			ProfileFilter: &ProfileFilter{
				Thresholds: FilterValues{
					Flat:        0.0,
					FlatPercent: 0.0,
					SumPercent:  0.0,
					Cum:         0.0,
					CumPercent:  0.0,
				},
				IgnoreFunctions: []string{"init", "TestMain", "BenchmarkMain"},
				IgnorePrefixes: []string{
					"github.com/example/BenchmarkName",
					"github.com/example/BenchmarkName/internal",
					"github.com/example/BenchmarkName/pkg",
				},
			},
		},
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), shared.PermDir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(template, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err = os.WriteFile(outputPath, data, shared.PermFile); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	slog.Info("Template configuration file created", "path", outputPath)
	slog.Info("Please edit this file with your configuration")

	return nil
}

func NewConfigBuilder() *Builder {
	return &Builder{
		config: Config{
			FunctionFilter: make(map[string]FunctionFilter),
			AIConfig: AIConfig{
				APIKey:  "your-api-key-here",
				BaseURL: "https://api.openai.com/v1",
			},
		},
	}
}

func (cb *Builder) WithAPIConfig(apiKey, baseURL string) *Builder {
	cb.config.AIConfig.APIKey = apiKey
	cb.config.AIConfig.BaseURL = baseURL
	return cb
}

func (cb *Builder) WithModelConfig(model string, maxTokens int, temp, topP float32, promptPath string) *Builder {
	cb.config.AIConfig.ModelConfig = ModelConfig{
		Model:              model,
		MaxTokens:          maxTokens,
		Temperature:        temp,
		TopP:               topP,
		PromptFileLocation: promptPath,
	}
	return cb
}

func (cb *Builder) WithFunctionFilter(benchmark string, prefixes, ignoreFunctions []string) *Builder {
	cb.config.FunctionFilter[benchmark] = FunctionFilter{
		IncludePrefixes: prefixes,
		IgnoreFunctions: ignoreFunctions,
	}
	return cb
}

func (cb *Builder) WithAIBenchmarkConfig(allBench, allProf bool, specificBench, specificProf []string) *Builder {
	cb.config.AIConfig.AllBenchmarks = allBench
	cb.config.AIConfig.AllProfiles = allProf
	cb.config.AIConfig.SpecificBenchmarks = specificBench
	cb.config.AIConfig.SpecificProfiles = specificProf
	return cb
}

func (cb *Builder) WithProfileFilter(thresholds FilterValues, ignoreFuncs, ignorePrefixes []string) *Builder {
	cb.config.AIConfig.ProfileFilter = &ProfileFilter{
		Thresholds:      thresholds,
		IgnoreFunctions: ignoreFuncs,
		IgnorePrefixes:  ignorePrefixes,
	}
	return cb
}

func (cb *Builder) Build() Config {
	return cb.config
}

// Preset builders.
func CreateDefaultConfig() Config {
	return NewConfigBuilder().
		WithFunctionFilter("BenchmarkGenPool",
			[]string{
				"github.com/example/GenPool",
				"github.com/example/GenPool/internal",
				"github.com/example/GenPool/pkg",
			},
			[]string{"init", "TestMain", "BenchmarkMain"}).
		WithAPIConfig("your-api-key-here", "https://api.openai.com/v1").
		WithAIBenchmarkConfig(true, true, []string{}, []string{}).
		WithModelConfig("gpt-4-turbo-preview", 0, 0.0, 0.0, "path/to/your/system_prompt.txt").
		WithProfileFilter(
			FilterValues{
				Flat:        0.0,
				FlatPercent: 0.0,
				SumPercent:  0.0,
				Cum:         0.0,
				CumPercent:  0.0,
			},
			[]string{"init", "TestMain", "BenchmarkMain"},
			[]string{
				"github.com/example/BenchmarkName",
				"github.com/example/BenchmarkName/internal",
				"github.com/example/BenchmarkName/pkg",
			}).
		Build()
}
