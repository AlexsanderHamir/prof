package config

import (
	"errors"
	"fmt"
)

var (
	errConfigNil     = errors.New("config: nil")
	errConfigVersion = errors.New("config: version must be positive")
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
	return nil
}
