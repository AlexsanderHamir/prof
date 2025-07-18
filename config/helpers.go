package config

import (
	"errors"
)

func validateAIConfig(config *Config) error {
	if config.AIConfig.APIKey == "" {
		return errors.New("api_key is required")
	}
	if config.AIConfig.BaseURL == "" {
		return errors.New("base_url is required")
	}
	if config.AIConfig.ModelConfig.Model == "" {
		return errors.New("model is required")
	}
	if config.AIConfig.ModelConfig.PromptFileLocation == "" {
		return errors.New("prompt_location is required")
	}

	// Validate AI config logic
	if !config.AIConfig.AllBenchmarks && len(config.AIConfig.SpecificBenchmarks) == 0 {
		return errors.New("when all_benchmarks is false, specific_benchmarks must be provided")
	}
	if !config.AIConfig.AllProfiles && len(config.AIConfig.SpecificProfiles) == 0 {
		return errors.New("when all_profiles is false, specific_profiles must be provided")
	}

	return nil
}
