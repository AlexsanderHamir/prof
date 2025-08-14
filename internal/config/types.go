package config

// Config holds the main configuration for the prof tool.
type Config struct {
	FunctionFilter map[string]FunctionFilter `json:"function_collection_filter"`
	Snapshot       SnapshotConfig            `json:"snapshot"`
}

// FunctionCollectionFilter defines filters for a specific benchmark,
// the filters are used when deciding which functions to collect
// code line level information for.
type FunctionFilter struct {
	// Prefixes: only collect functions starting with these prefixes
	// Example: []string{"github.com/myorg", "main."}
	IncludePrefixes []string `json:"include_prefixes,omitempty"`

	// IgnoreFunctions ignores the function name after the last dot.
	// Example: "Get,Set" excludes pool.Get() and cache.Set()
	IgnoreFunctions []string `json:"ignore_functions,omitempty"`
}

// SnapshotConfig defines configuration for performance snapshots
type SnapshotConfig struct {
	// StorageDirectory specifies where snapshots are stored relative to project root
	// Default: "prof-snapshots"
	StorageDirectory string `json:"storage_directory,omitempty"`

	// Git configuration for checking out code at specific commits/tags
	Git SnapshotGitConfig `json:"git"`

	// DefaultBenchmarks specifies which benchmarks to run for snapshots if none provided
	DefaultBenchmarks []string `json:"default_benchmarks,omitempty"`

	// DefaultProfiles specifies which profiles to collect for snapshots if none provided
	// Example: ["cpu", "memory", "mutex", "block"]
	DefaultProfiles []string `json:"default_profiles,omitempty"`

	// DefaultRunCount specifies how many benchmark iterations to run for snapshots
	DefaultRunCount int `json:"default_run_count,omitempty"`

	// AutoCleanup configures automatic cleanup of old snapshots
	AutoCleanup SnapshotCleanupConfig `json:"auto_cleanup"`

	// Metadata allows storing additional information with snapshots
	Metadata SnapshotMetadataConfig `json:"metadata"`
}

// SnapshotGitConfig defines Git-related configuration for snapshots
type SnapshotGitConfig struct {
	// WorkingDirectory specifies where to perform git operations
	// If empty, uses a temporary directory. If specified, must be a clean directory.
	WorkingDirectory string `json:"working_directory,omitempty"`

	// StandardSavePath ensures consistent snapshot directory naming across teams
	// for easier .gitignore and .gitattributes configuration
	StandardSavePath string `json:"standard_save_path,omitempty"`

	// RepositoryURL specifies the git repository to clone from
	// If empty, assumes we're already in a git repository
	RepositoryURL string `json:"repository_url,omitempty"`

	// GitCommand allows overriding the git executable path
	// Default: "git"
	GitCommand string `json:"git_command,omitempty"`
}

// SnapshotCleanupConfig defines automatic cleanup behavior for old snapshots
type SnapshotCleanupConfig struct {
	// MaxAge specifies maximum age of snapshots before cleanup (e.g., "30d", "6m").
	// The 0 value will be interpeted as not active.
	MaxAge string `json:"max_age,omitempty"`

	// MaxCount specifies maximum number of snapshots to retain
	// The 0 value will be interpeted as not active.
	MaxSnapshotCount int `json:"max_count,omitempty"`

	// KeepTags specifies tags that should never be automatically cleaned up.
	// Example: ["v1.0", "release-*", "baseline"].
	// If both fields above are nove active, this won't do anything.
	KeepTags []string `json:"keep_tags,omitempty"`
}

// SnapshotMetadataConfig defines what metadata to capture with snapshots
type SnapshotMetadataConfig struct {
	// CaptureGitInfo determines if git commit info should be captured
	CaptureGitInfo bool `json:"capture_git_info"`

	// CaptureSystemInfo determines if system specs should be captured
	CaptureSystemInfo bool `json:"capture_system_info"`

	// CaptureGoVersion determines if Go version should be captured
	CaptureGoVersion bool `json:"capture_go_version"`
}
