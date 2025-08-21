package internal

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
)

func GetScanner(filePath string) (*bufio.Scanner, *os.File, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read profile file %s: %w", filePath, err)
	}

	scanner := bufio.NewScanner(file)

	return scanner, file, nil
}

// CleanOrCreateDir cleans a directory if it exists, or creates one if it.
func CleanOrCreateDir(dir string) error {
	info, err := os.Stat(dir)
	if err != nil {
		if os.IsNotExist(err) {
			if err = os.MkdirAll(dir, PermDir); err != nil {
				return fmt.Errorf("failed to create %s directory: %w", dir, err)
			}
			return nil
		}
		return err
	}

	if !info.IsDir() {
		return fmt.Errorf("path is not a directory: %s", dir)
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read directory: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dir, entry.Name())
		if err = os.RemoveAll(path); err != nil {
			return fmt.Errorf("failed to remove %s: %w", path, err)
		}
	}

	return nil
}

// FindGoModuleRoot searches upwards from the current working directory for a directory
// containing a go.mod file and returns its absolute path. If none is found, an error is returned.
func FindGoModuleRoot() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	for {
		if _, err = os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", errors.New("go.mod not found from current directory upwards")
		}
		dir = parent
	}
}

func PrintConfiguration(benchArgs *BenchArgs, functionFilterPerBench map[string]FunctionFilter) {
	slog.Info(
		"Parsed arguments",
		"Benchmarks", benchArgs.Benchmarks,
		"Profiles", benchArgs.Profiles,
		"Tag", benchArgs.Tag,
		"Count", benchArgs.Count,
	)

	hasBenchFunctionFilters := len(functionFilterPerBench) > 0
	if hasBenchFunctionFilters {
		slog.Info("Benchmark Function Filter Configurations:")
		for benchmark, cfg := range functionFilterPerBench {
			slog.Info("Benchmark Config", "Benchmark", benchmark, "Prefixes", cfg.IncludePrefixes, "Ignore", cfg.IgnoreFunctions)
		}
	} else {
		slog.Info("No benchmark configuration found in config file - analyzing all functions")
	}
}

// LoadFromFile loads and validates config from a JSON file.
func LoadFromFile(filename string) (*Config, error) {
	root, err := FindGoModuleRoot()
	if err != nil {
		return nil, fmt.Errorf("failed to locate module root for config: %w", err)
	}

	configPath := filepath.Join(root, filename)
	data, err := os.ReadFile(configPath)
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
	root, err := FindGoModuleRoot()
	if err != nil {
		return fmt.Errorf("failed to locate module root for template: %w", err)
	}

	outputPath := filepath.Join(root, "config_template.json")

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
	}

	if err = os.MkdirAll(filepath.Dir(outputPath), PermDir); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	data, err := json.MarshalIndent(template, "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal template: %w", err)
	}

	if err = os.WriteFile(outputPath, data, PermFile); err != nil {
		return fmt.Errorf("failed to write template file: %w", err)
	}

	slog.Info("Template configuration file created", "path", outputPath)
	slog.Info("Please edit this file with your configuration")

	return nil
}
