package tooling

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

// ExecRunner implements [Runner] using [exec.CommandContext].
type ExecRunner struct{}

// NewExecRunner returns a [Runner] that invokes the system shell path resolution for argv[0].
func NewExecRunner() Runner {
	return ExecRunner{}
}

// Run executes the command. When opts.Stdout is nil and opts.Combined is false, stdout only is captured (like [exec.Cmd.Output]).
// When opts.Combined is true, stdout and stderr are merged (like [exec.Cmd.CombinedOutput]).
func (ExecRunner) Run(ctx context.Context, argv []string, opts RunOpts) ([]byte, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(argv) == 0 {
		return nil, errors.New("tooling: empty argv")
	}
	// #nosec G204 -- argv supplied by callers; ExecRunner is the single execution gate.
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
		if opts.Stderr != nil {
			cmd.Stderr = opts.Stderr
		} else {
			cmd.Stderr = os.Stderr
		}
		err := cmd.Run()
		return nil, err
	}
	if opts.Combined {
		return cmd.CombinedOutput()
	}
	return cmd.Output()
}
