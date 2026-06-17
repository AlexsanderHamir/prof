package config

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// LoadFromFile loads and validates config from a JSON file beside go.mod.
func LoadFromFile(filename string) (*Config, error) {
	root, err := workspace.FindModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root for config: %w", err)
	}

	configPath := filepath.Join(root, filename)
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	if err = json.Unmarshal(data, &c); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &c, nil
}

// CreateTemplate writes config_template.json beside go.mod with examples.
func CreateTemplate() error {
	root, err := workspace.FindModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to locate module root for template: %w", err)
	}

	outputPath := filepath.Join(root, Filename)

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
		CIConfig: &CIConfig{
			Global: &CITrackingConfig{
				IgnoreFunctions: []string{
					"runtime.gcBgMarkWorker",
					"runtime.systemstack",
					"runtime.mallocgc",
					"reflect.ValueOf",
					"testing.(*B).launch",
				},
				IgnorePrefixes:         []string{"runtime.", "reflect.", "testing."},
				MinChangeThreshold:     5.0,
				MaxRegressionThreshold: 15.0,
				FailOnImprovement:      false,
			},
			Benchmarks: map[string]CITrackingConfig{
				"BenchmarkGenPool": {
					IgnoreFunctions:        []string{"BenchmarkGenPool", "testing.(*B).ResetTimer"},
					MinChangeThreshold:     3.0,
					MaxRegressionThreshold: 10.0,
				},
			},
		},
	}

	if err = os.MkdirAll(filepath.Dir(outputPath), workspace.PermDir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(template, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal template file: %w", err)
	}

	if err = os.WriteFile(outputPath, data, workspace.PermFile); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	slog.Info("Template configuration file created", "path", outputPath)
	slog.Info("Please edit this file with your configuration")
	return nil
}

// PrintAutoConfiguration logs parsed auto-benchmark arguments and optional filters.
func PrintAutoConfiguration(args *AutoArgs, functionFilterPerBench map[string]FunctionFilter) {
	slog.Info(
		"Parsed arguments",
		"Benchmarks", args.Benchmarks,
		"Profiles", args.Profiles,
		"Tag", args.Tag,
		"Count", args.Count,
	)

	if len(functionFilterPerBench) > 0 {
		slog.Info("Benchmark Function Filter Configurations:")
		for benchmark, cfg := range functionFilterPerBench {
			slog.Info("Benchmark Config", "Benchmark", benchmark, "Prefixes", cfg.IncludePrefixes, "Ignore", cfg.IgnoreFunctions)
		}
	} else {
		slog.Info("No benchmark configuration found in config file - analyzing all functions")
	}
}
