package tooling

import (
	"context"
	"io"
)

// StreamRunOpts configures [RunWithStdinStreamStdout].
type StreamRunOpts struct {
	// Dir is the working directory for the child process. Empty means the current process directory.
	Dir string
	// Env is the environment for the child. When nil, the child inherits the current process environment ([os.Environ]).
	Env []string
	// Stdin is optional input written to the child's standard input.
	Stdin []byte
	// OnStdoutLine, when non-nil, is called with each complete stdout line (without the trailing newline) as the process runs.
	OnStdoutLine func([]byte)
}

// RunOpts configures a single [Runner.Run] invocation.
type RunOpts struct {
	// Dir is the working directory for the child process. Empty means the current process directory.
	Dir string
	// Combined, when Stdout is nil, selects CombinedOutput instead of stdout-only Output.
	Combined bool
	// Stdout, when non-nil, receives the child stdout; the returned byte slice is nil on success.
	Stdout io.Writer
	// Stderr is used when Stdout is set. When nil, os.Stderr is used.
	Stderr io.Writer
}

// Runner runs an external command given argv[0] as the executable and argv[1:] as arguments.
type Runner interface {
	Run(ctx context.Context, argv []string, opts RunOpts) ([]byte, error)
}
