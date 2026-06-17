package intent

import (
	"github.com/AlexsanderHamir/prof/internal/app"
)

// SetupIntent writes prof.json (alias for ConfigCreateIntent).
type SetupIntent struct{}

// Kind implements [Executable].
func (i *SetupIntent) Kind() Kind { return KindSetup }

// Validate implements [Executable]; setup has no input constraints.
func (i *SetupIntent) Validate() error { return nil }

// Run implements [Executable].
func (i *SetupIntent) Run(svc *app.Services) error {
	return (&ConfigCreateIntent{}).Run(svc)
}
