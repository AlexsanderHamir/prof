package config

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/AlexsanderHamir/prof/internal/workspace"
)

// Load reads and parses prof.json beside go.mod.
func Load() (*Config, error) {
	return LoadFromFile(Filename)
}

// LoadFromFile loads config from filename relative to the module root.
func LoadFromFile(filename string) (*Config, error) {
	path, err := Path(filename)
	if err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var c Config
	if err = json.Unmarshal(stripJSONComments(data), &c); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	Normalize(&c)
	if err = Validate(&c); err != nil {
		return nil, err
	}
	return &c, nil
}

// Path returns the absolute path to filename beside go.mod.
func Path(filename string) (string, error) {
	root, err := workspace.FindModuleRoot()
	if err != nil {
		return "", fmt.Errorf("failed to locate module root for config: %w", err)
	}
	return filepath.Join(root, filename), nil
}

// Save writes cfg to prof.json using an atomic replace.
func Save(cfg *Config) error {
	if cfg == nil {
		return errors.New("config: cannot save nil")
	}
	c := *cfg
	Normalize(&c)
	if err := Validate(&c); err != nil {
		return err
	}

	path, err := Path(Filename)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(configForJSON(c), "", "    ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}
	data = append(data, '\n')

	tmp := path + ".tmp"
	if err = os.WriteFile(tmp, data, workspace.PermFile); err != nil {
		return fmt.Errorf("failed to write config temp file: %w", err)
	}
	if err = os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to replace config file: %w", err)
	}
	return nil
}

// Default returns a minimal starter config; collection and track settings are opt-in.
func Default() *Config {
	cfg := &Config{
		Version: CurrentVersion,
	}
	Normalize(cfg)
	return cfg
}

// configForJSON omits empty collection/track sections so minimal prof.json stays version-only.
func configForJSON(cfg Config) any {
	type fileConfig struct {
		Version    int         `json:"version"`
		Collection *Collection `json:"collection,omitempty"`
		Track      *Track      `json:"track,omitempty"`
	}
	out := fileConfig{Version: cfg.Version}
	if !collectionEmpty(cfg.Collection) {
		col := cfg.Collection
		out.Collection = &col
	}
	if !trackSectionEmpty(cfg.Track) {
		tr := cfg.Track
		out.Track = &tr
	}
	return out
}

func readModulePath(goModPath string) (string, error) {
	data, err := os.ReadFile(goModPath)
	if err != nil {
		return "", err
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "module ") {
			return strings.TrimSpace(line[7:]), nil
		}
	}
	return "", errors.New("module directive not found")
}

// DefaultFromModuleRoot builds Default for the current module root.
func DefaultFromModuleRoot() (*Config, error) {
	return Default(), nil
}

// CreateDefaultFile writes prof.json if it does not exist.
func CreateDefaultFile() error {
	path, err := Path(Filename)
	if err != nil {
		return err
	}
	if _, err = os.Stat(path); err == nil {
		return fmt.Errorf("config file already exists: %s", path)
	}
	if !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("stat config file: %w", err)
	}

	cfg, err := DefaultFromModuleRoot()
	if err != nil {
		return err
	}
	if err = Save(cfg); err != nil {
		return err
	}

	modulePath := ""
	if root, rootErr := workspace.FindModuleRoot(); rootErr == nil {
		if modPath, readErr := readModulePath(filepath.Join(root, "go.mod")); readErr == nil {
			modulePath = modPath
		}
	}
	examplePath, err := Path(ExampleFilename)
	if err != nil {
		return err
	}
	if err = writeFileAtomic(examplePath, []byte(ExampleTemplate(modulePath))); err != nil {
		return fmt.Errorf("failed to write %s: %w", ExampleFilename, err)
	}

	slog.Info("Configuration file created", "path", path, "example", examplePath)
	return nil
}

func writeFileAtomic(path string, content []byte) error {
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, content, workspace.PermFile); err != nil {
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return fmt.Errorf("failed to replace file: %w", err)
	}
	return nil
}

// CreateTemplate is deprecated; use CreateDefaultFile.
func CreateTemplate() error {
	return CreateDefaultFile()
}

// PrintAutoConfiguration logs parsed auto-benchmark arguments and configured filters.
func PrintAutoConfiguration(args *AutoArgs, cfg *Config) {
	slog.Info(
		"Parsed arguments",
		"Benchmarks", args.Benchmarks,
		"Profiles", args.Profiles,
		"Tag", args.Tag,
		"Count", args.Count,
	)

	if cfg == nil {
		slog.Info("No benchmark configuration found in config file - analyzing all functions")
		return
	}

	hasCollection := !functionFilterEmpty(cfg.Collection.Defaults) ||
		len(cfg.Collection.Benchmarks) > 0 ||
		len(cfg.Collection.ManualProfiles) > 0
	if !hasCollection {
		slog.Info("No benchmark configuration found in config file - analyzing all functions")
		return
	}

	slog.Info("Collection filter configuration loaded from prof.json")
	for _, benchmark := range args.Benchmarks {
		filter := ResolveCollectionFilter(cfg, CollectionTargetAuto(benchmark))
		slog.Info("Benchmark filter", "Benchmark", benchmark, "Prefixes", filter.IncludePrefixes, "Ignore", filter.IgnoreFunctions)
	}
}
