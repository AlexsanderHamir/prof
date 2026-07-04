package config

const (
	// Filename is the default config file beside go.mod.
	Filename = "prof.json"
	// ExampleFilename is a commented reference copy written beside prof.json on init.
	ExampleFilename = "prof.json.example"
	// CurrentVersion is the supported prof.json schema version.
	CurrentVersion = 1
	// MissingConfigUserWarning is shown when prof.json is absent during collect.
	MissingConfigUserWarning = "No prof.json found; proceeding without function filters (run prof config init to add one)."
)
