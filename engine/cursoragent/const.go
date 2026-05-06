package cursoragent

const (
	// DefaultBinaryName is argv[0] when [Options.BinaryPath] is empty (PATH lookup).
	DefaultBinaryName = "cursor-agent"

	// EnvBinaryOverride is the environment variable name for a full path to cursor-agent.
	EnvBinaryOverride = "PROF_CURSOR_AGENT"

	// MaxStderrTailRunes caps stderr text included in [RunResult.StderrTail] for errors.
	MaxStderrTailRunes = 512
)
