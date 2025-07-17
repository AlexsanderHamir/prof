package config

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
)

const (
	permDir  = 0o755
	permFile = 0o644
)

// Config holds the main configuration for the prof tool.
type Config struct {
	APIKey           string                     `json:"api_key"`
	BaseURL          string                     `json:"base_url"`
	ModelConfig      ModelConfig                `json:"model_config"`
	BenchmarkConfigs map[string]BenchmarkFilter `json:"benchmark_configs"`
	AIConfig         AIConfig                   `json:"ai_config"`
}

func (cfg *Config) GetProfileFilters() map[int]float64 {
	profileValues := map[int]float64{
		0: cfg.AIConfig.UniversalProfileFilter.ProfileValues.Flat,
		1: cfg.AIConfig.UniversalProfileFilter.ProfileValues.FlatPercent,
		2: cfg.AIConfig.UniversalProfileFilter.ProfileValues.SumPercent,
		3: cfg.AIConfig.UniversalProfileFilter.ProfileValues.Cum,
		4: cfg.AIConfig.UniversalProfileFilter.ProfileValues.CumPercent,
	}

	return profileValues
}

// ModelConfig holds the configuration for the AI model.
type ModelConfig struct {
	Model          string  `json:"model"`
	MaxTokens      int     `json:"max_tokens"`
	Temperature    float32 `json:"temperature"`
	TopP           float32 `json:"top_p"`
	PromptLocation string  `json:"prompt_location"`
}

// BenchmarkFilter defines filters for a specific benchmark.
type BenchmarkFilter struct {
	Prefixes []string `json:"prefixes"`
	Ignore   string   `json:"ignore,omitempty"`
}

// AIConfig holds configuration for AI-driven analysis.
type AIConfig struct {
	AllBenchmarks          bool                    `json:"all_benchmarks"`
	AllProfiles            bool                    `json:"all_profiles"`
	SpecificBenchmarks     []string                `json:"specific_benchmarks"`
	SpecificProfiles       []string                `json:"specific_profiles"`
	UniversalProfileFilter *UniversalProfileFilter `json:"universal_profile_filter,omitempty"`
}

// UniversalProfileFilter defines universal filters for profile analysis.
type UniversalProfileFilter struct {
	ProfileValues   ProfileValues `json:"profile_values"`
	IgnoreFunctions []string      `json:"ignore_functions,omitempty"`
	IgnorePrefixes  []string      `json:"ignore_prefixes,omitempty"`
}

// ProfileValues holds threshold values for filtering profile data.
type ProfileValues struct {
	Flat        float64 `json:"flat"`
	FlatPercent float64 `json:"flat%"`
	SumPercent  float64 `json:"sum%"`
	Cum         float64 `json:"cum"`
	CumPercent  float64 `json:"cum%"`
}

// LoadFromFile loads and validates a Config from a JSON file.
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := validateConfig(&config); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &config, nil
}

func validateConfig(config *Config) error {
	if config.APIKey == "" {
		return fmt.Errorf("api_key is required")
	}
	if config.BaseURL == "" {
		return fmt.Errorf("base_url is required")
	}
	if config.ModelConfig.Model == "" {
		return fmt.Errorf("model is required")
	}
	if config.ModelConfig.PromptLocation == "" {
		return fmt.Errorf("prompt_location is required")
	}

	// Validate AI config logic
	if !config.AIConfig.AllBenchmarks && len(config.AIConfig.SpecificBenchmarks) == 0 {
		return fmt.Errorf("when all_benchmarks is false, specific_benchmarks must be provided")
	}
	if !config.AIConfig.AllProfiles && len(config.AIConfig.SpecificProfiles) == 0 {
		return fmt.Errorf("when all_profiles is false, specific_profiles must be provided")
	}

	return nil
}

// CreateTemplate creates a template configuration file at the specified output path.
func CreateTemplate(outputPath string) error {
	if outputPath == "" {
		outputPath = "./config_template.json"
	}

	template := map[string]interface{}{
		"api_key":  "your-api-key-here",
		"base_url": "https://api.openai.com/v1",
		"model_config": map[string]interface{}{
			"model":           "gpt-4-turbo-preview",
			"max_tokens":      4096,
			"temperature":     0.7,
			"top_p":           1.0,
			"prompt_location": "path/to/your/system_prompt.txt",
		},
		"benchmark_configs": map[string]interface{}{
			"BenchmarkGenPool": map[string]interface{}{
				"prefixes": []string{
					"github.com/example/GenPool",
					"github.com/example/GenPool/internal",
					"github.com/example/GenPool/pkg",
				},
				"ignore": "init,TestMain,BenchmarkMain",
			},
			"BenchmarkSyncPool": map[string]interface{}{
				"prefixes": []string{"github.com/example/SyncPool"},
				"ignore":   "setup,teardown",
			},
			"BenchmarkCustomPool": map[string]interface{}{
				"prefixes": []string{
					"github.com/example/CustomPool",
					"github.com/example/CustomPool/optimizations",
				},
			},
		},
		"ai_config": map[string]interface{}{
			"all_benchmarks":      true,
			"all_profiles":        true,
			"specific_benchmarks": []string{},
			"specific_profiles":   []string{},
			"universal_profile_filter": map[string]interface{}{
				"profile_values": map[string]interface{}{
					"flat":  0.0,
					"flat%": 0.0,
					"sum%":  0.0,
					"cum":   0.0,
					"cum%":  0.0,
				},
				"ignore_functions": []string{"init", "TestMain", "BenchmarkMain"},
				"ignore_prefixes": []string{
					"github.com/example/BenchmarkName",
					"github.com/example/BenchmarkName/internal",
					"github.com/example/BenchmarkName/pkg",
				},
			},
		},
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputPath), permDir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(template, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err := os.WriteFile(outputPath, data, permFile); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	log.Printf("Template configuration file created at: %s\n", outputPath)
	log.Printf("\nThe template includes example benchmark configurations with multiple prefixes.\n")
	log.Printf("For each benchmark, you can specify:\n")
	log.Printf("  - prefixes: A list of package prefixes to analyze\n")
	log.Printf("  - ignore: Optional comma-separated list of functions to exclude\n")
	log.Printf("\nPlease edit this file with your configuration.\n")

	return nil
}
