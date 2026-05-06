package tooling

import (
	"context"
	"errors"
	"os"
	"os/exec"
)

// LookPath resolves an executable name on PATH. Callers use this instead of importing [os/exec] only for [exec.LookPath].
func LookPath(file string) (string, error) {
	return exec.LookPath(file)
}

// StartDetached runs argv[0] with argv[1:] via [exec.CommandContext.Start] without waiting for completion.
// Use for GUI or long-lived child processes. Stdout and stderr default to [os.Stdout] and [os.Stderr] when nil.
func StartDetached(ctx context.Context, argv []string, opts RunOpts) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(argv) == 0 {
		return errors.New("tooling: empty argv")
	}
	// #nosec G204 -- argv built by callers; this is the sanctioned spawn entry point outside [ExecRunner.Run].
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if opts.Stdout != nil {
		cmd.Stdout = opts.Stdout
	} else {
		cmd.Stdout = os.Stdout
	}
	if opts.Stderr != nil {
		cmd.Stderr = opts.Stderr
	} else {
		cmd.Stderr = os.Stderr
	}
	return cmd.Start()
}
