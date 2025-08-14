package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal/shared"
)

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
func CreateTemplate() error {
	outputPath := "./config_template.json"
	runCount := 5
	maxSnapshotCount := 10

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
		Snapshot: SnapshotConfig{
			StorageDirectory: "prof-snapshots",
			Git: SnapshotGitConfig{
				WorkingDirectory: "temp-snapshot-work",
				StandardSavePath: "snapshot-checkout",
				RepositoryURL:    "https://github.com/your-org/your-repo.git",
				GitCommand:       "git",
			},
			DefaultBenchmarks: []string{"BenchmarkGenPool"},
			DefaultProfiles:   []string{"cpu", "memory"},
			DefaultRunCount:   runCount,
			AutoCleanup: SnapshotCleanupConfig{
				MaxAge:           "30d",
				MaxSnapshotCount: maxSnapshotCount,
				KeepTags:         []string{"v1.0", "baseline", "release-*"},
			},
			Metadata: SnapshotMetadataConfig{
				CaptureGitInfo:    true,
				CaptureSystemInfo: true,
				CaptureGoVersion:  true,
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
