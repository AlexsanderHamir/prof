package config

import (
	"errors"
	"fmt"
	"strings"
)

var (
	errConfigNil             = errors.New("config: nil")
	errConfigVersion         = errors.New("config: version must be positive")
	errTrackEmptyBenchKey    = errors.New("config track benchmarks: empty benchmark key")
	errMinChangeNegative     = errors.New("min_change_percent must be >= 0")
	errMaxRegressionNegative = errors.New("max_regression_percent must be >= 0")
)

// Validate checks cfg after Normalize.
func Validate(cfg *Config) error {
	if cfg == nil {
		return errConfigNil
	}
	if cfg.Version <= 0 {
		return errConfigVersion
	}
	if cfg.Version > CurrentVersion {
		return fmt.Errorf("config: unsupported version %d (max supported %d)", cfg.Version, CurrentVersion)
	}
	if err := validatePercents(cfg.Track.Defaults); err != nil {
		return fmt.Errorf("config track defaults: %w", err)
	}
	for name, p := range cfg.Track.Benchmarks {
		if strings.TrimSpace(name) == "" {
			return errTrackEmptyBenchKey
		}
		if err := validatePercents(p); err != nil {
			return fmt.Errorf("config track benchmarks[%q]: %w", name, err)
		}
	}
	return nil
}

func validatePercents(p TrackPolicy) error {
	if p.MinChangePercent < 0 {
		return errMinChangeNegative
	}
	if p.MaxRegressionPercent < 0 {
		return errMaxRegressionNegative
	}
	return nil
}
