package intent

import (
	"errors"

	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/config"
)

// ConfigCreateIntent writes the default prof.json beside go.mod.
type ConfigCreateIntent struct{}

// Kind implements [Executable].
func (i *ConfigCreateIntent) Kind() Kind { return KindConfigCreate }

// Validate implements [Executable].
func (i *ConfigCreateIntent) Validate() error { return nil }

// Run implements [Executable].
func (i *ConfigCreateIntent) Run(svc *app.Services) error {
	return svc.Config.CreateDefaultFile()
}

// ConfigSaveIntent validates and saves prof.json.
type ConfigSaveIntent struct {
	Config *config.Config
}

// Kind implements [Executable].
func (i *ConfigSaveIntent) Kind() Kind { return KindConfigSave }

// Validate implements [Executable].
func (i *ConfigSaveIntent) Validate() error {
	if i.Config == nil {
		return errors.New("config save intent: config is required")
	}
	return config.Validate(i.Config)
}

// Run implements [Executable].
func (i *ConfigSaveIntent) Run(svc *app.Services) error {
	return svc.Config.Save(i.Config)
}
