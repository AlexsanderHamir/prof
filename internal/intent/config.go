package intent

import (
	"github.com/AlexsanderHamir/prof/internal/app"
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
