package tooling

// Adapted from T2A pkgs/agents/runner/adapterkit/exec.go (DefaultStreamExec, ScanStdoutLines).

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
)

const (
	scannerInitialBufferBytes = 64 * 1024
	scannerMaxBufferBytes     = 1024 * 1024
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

// RunWithStdinStreamStdout runs argv[0] with argv[1:], optional stdin, optional environment,
// and optional per-line stdout callback. It returns full captured stdout and stderr, the process
// exit code (0 on success), and a start/wait error when the process did not complete normally.
// A non-zero exit code with nil error means the process exited unsuccessfully.
func RunWithStdinStreamStdout(ctx context.Context, argv []string, opts StreamRunOpts) (stdout []byte, stderr []byte, exitCode int, err error) {
	if ctx == nil {
		ctx = context.Background()
	}
	if len(argv) == 0 {
		return nil, nil, 0, errors.New("tooling: empty argv")
	}
	// #nosec G204 -- argv supplied by callers; sanctioned exec entry point.
	cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
	if opts.Dir != "" {
		cmd.Dir = opts.Dir
	}
	if len(opts.Env) > 0 {
		cmd.Env = opts.Env
	}
	if len(opts.Stdin) > 0 {
		cmd.Stdin = bytes.NewReader(opts.Stdin)
	}
	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, 0, err
	}
	stderrPipe, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, 0, err
	}
	if err := cmd.Start(); err != nil {
		return nil, nil, 0, err
	}

	var stdoutBuf, stderrBuf bytes.Buffer
	stdoutDone := make(chan error, 1)
	stderrDone := make(chan error, 1)
	go func() {
		stdoutDone <- scanStdoutLines(stdoutPipe, &stdoutBuf, opts.OnStdoutLine)
	}()
	go func() {
		_, copyErr := io.Copy(&stderrBuf, stderrPipe)
		stderrDone <- copyErr
	}()

	waitErr := cmd.Wait()
	stdoutErr := <-stdoutDone
	stderrErr := <-stderrDone
	if waitErr == nil {
		stdoutErr = normalizePipeReadError(stdoutErr)
		stderrErr = normalizePipeReadError(stderrErr)
		if stdoutErr != nil {
			return stdoutBuf.Bytes(), stderrBuf.Bytes(), 0, stdoutErr
		}
		if stderrErr != nil {
			return stdoutBuf.Bytes(), stderrBuf.Bytes(), 0, stderrErr
		}
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), 0, nil
	}
	if ctx.Err() != nil {
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), 0, ctx.Err()
	}
	var exitErr *exec.ExitError
	if errors.As(waitErr, &exitErr) {
		return stdoutBuf.Bytes(), stderrBuf.Bytes(), exitErr.ExitCode(), nil
	}
	return stdoutBuf.Bytes(), stderrBuf.Bytes(), 0, waitErr
}

func scanStdoutLines(r io.Reader, dst *bytes.Buffer, onLine func([]byte)) error {
	scanner := bufio.NewScanner(r)
	scanner.Buffer(make([]byte, 0, scannerInitialBufferBytes), scannerMaxBufferBytes)
	for scanner.Scan() {
		line := append([]byte(nil), scanner.Bytes()...)
		dst.Write(line)
		dst.WriteByte('\n')
		if onLine != nil {
			onLine(line)
		}
	}
	return scanner.Err()
}

func normalizePipeReadError(err error) error {
	if isClosedPipeReadError(err) {
		return nil
	}
	return err
}

func isClosedPipeReadError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, os.ErrClosed) {
		return true
	}
	return strings.Contains(strings.ToLower(err.Error()), "file already closed")
}
