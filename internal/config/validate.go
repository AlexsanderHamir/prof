package config

import (
	"fmt"
	"strings"
)

// Validate checks cfg after Normalize.
func Validate(cfg *Config) error {
	if cfg == nil {
		return fmt.Errorf("config: nil")
	}
	if cfg.Version <= 0 {
		return fmt.Errorf("config: version must be positive")
	}
	if cfg.Version > CurrentVersion {
		return fmt.Errorf("config: unsupported version %d (max supported %d)", cfg.Version, CurrentVersion)
	}
	if err := validatePercents(cfg.Track.Defaults); err != nil {
		return fmt.Errorf("config track defaults: %w", err)
	}
	for name, p := range cfg.Track.Benchmarks {
		if strings.TrimSpace(name) == "" {
			return fmt.Errorf("config track benchmarks: empty benchmark key")
		}
		if err := validatePercents(p); err != nil {
			return fmt.Errorf("config track benchmarks[%q]: %w", name, err)
		}
	}
	return nil
}

func validatePercents(p TrackPolicy) error {
	if p.MinChangePercent < 0 {
		return fmt.Errorf("min_change_percent must be >= 0")
	}
	if p.MaxRegressionPercent < 0 {
		return fmt.Errorf("max_regression_percent must be >= 0")
	}
	return nil
}
