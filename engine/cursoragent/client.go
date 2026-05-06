package cursoragent

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/AlexsanderHamir/prof/engine/tooling"
)

const defaultProbeTimeout = 5 * time.Second

// StreamExec matches the T2A StreamExecFunc shape: spawn name with args, stdin, env, dir, optional per-line stdout tap.
type StreamExec func(ctx context.Context, dir string, env []string, stdin []byte, name string, onStdoutLine func([]byte), args ...string) (stdout, stderr []byte, exitCode int, err error)

// ProbeFunc is the small exec surface used by [Client.Probe] in tests.
type ProbeFunc func(ctx context.Context, name string, args ...string) (stdout, stderr []byte, exitCode int, err error)

// Options configures [Client].
type Options struct {
	// BinaryPath is the cursor-agent executable or bare name (empty means [DefaultBinaryName]).
	BinaryPath string
	// DefaultModel is passed as --model when [RunRequest.Model] is empty.
	DefaultModel string
	// DefaultTimeout wraps [Client.Run] when [RunRequest.Timeout] is zero; zero means no extra deadline beyond ctx.
	DefaultTimeout time.Duration
	// ExtraEnvKeys are merged into the allowlist for parent env passthrough (tests, rare host needs).
	ExtraEnvKeys []string
	// StreamExec runs cursor-agent; nil uses [tooling.RunWithStdinStreamStdout].
	StreamExec StreamExec
	// ProbeFn overrides the version probe; nil uses StreamExec with --version.
	ProbeFn ProbeFunc
	// HomePaths override redaction roots; nil uses [liveHomePaths].
	HomePaths []string
}

// Client runs cursor-agent non-interactively.
type Client struct {
	opts       Options
	streamExec StreamExec
	homePaths  []string
}

// NewClient returns a client with defaults applied (stream exec, home paths for redaction).
func NewClient(opts Options) *Client {
	hp := opts.HomePaths
	if len(hp) == 0 {
		hp = liveHomePaths()
	}
	se := opts.StreamExec
	if se == nil {
		se = defaultStreamExec
	}
	return &Client{opts: opts, streamExec: se, homePaths: hp}
}

func defaultStreamExec(ctx context.Context, dir string, env []string, stdin []byte, name string, onStdoutLine func([]byte), args ...string) ([]byte, []byte, int, error) {
	argv := append([]string{name}, args...)
	return tooling.RunWithStdinStreamStdout(ctx, argv, tooling.StreamRunOpts{
		Dir:          dir,
		Env:          env,
		Stdin:        stdin,
		OnStdoutLine: onStdoutLine,
	})
}

// RunRequest is one cursor-agent invocation.
type RunRequest struct {
	Prompt     []byte
	WorkingDir string
	Model      string
	Timeout    time.Duration
	// OnStdoutLine receives each raw stdout line (stream-json) while the process runs.
	OnStdoutLine func([]byte)
	// ExtraEnv is merged into the child allowlisted environment (values must be safe to forward).
	ExtraEnv map[string]string
}

// RunResult is the parsed outcome of [Client.Run].
type RunResult struct {
	Text            string
	ExitCode        int
	StderrTail      string
	ResolvedModel   string
	MissingTerminal bool
}

func (c *Client) effectiveModel(req RunRequest) string {
	if m := strings.TrimSpace(req.Model); m != "" {
		return m
	}
	return strings.TrimSpace(c.opts.DefaultModel)
}

func (c *Client) resolvedBinaryForExec() (string, error) {
	name := strings.TrimSpace(c.opts.BinaryPath)
	if name == "" {
		name = DefaultBinaryName
	}
	if strings.ContainsAny(name, `/\`) {
		if _, statErr := os.Stat(name); statErr != nil {
			return "", fmt.Errorf("%w: %w\n\n%s", ErrBinaryNotFound, statErr, FixBinaryHelpBlock())
		}
		return name, nil
	}
	if _, lookErr := exec.LookPath(name); lookErr != nil {
		return "", fmt.Errorf("%w: %w\n\n%s", ErrBinaryNotFound, lookErr, FixBinaryHelpBlock())
	}
	return ResolveBinaryPath(name), nil
}

func (c *Client) buildEnv(extra map[string]string) []string {
	return buildChildEnv(extra, c.opts.ExtraEnvKeys)
}

func (c *Client) argvFor(model string) []string {
	out := []string{"--print", "--output-format", "stream-json"}
	m := strings.TrimSpace(model)
	if m == "" {
		m = strings.TrimSpace(c.opts.DefaultModel)
	}
	if m != "" {
		out = append(out, "--model", m)
	}
	out = append(out, "--force")
	return out
}

// Probe runs "<binary> --version" with a short timeout and returns the resolved path and version line.
func (c *Client) Probe(ctx context.Context) (resolvedPath, version string, err error) {
	bin, err := c.resolvedBinaryForExec()
	if err != nil {
		return "", "", err
	}
	resolved := ResolveBinaryPath(bin)
	if c.opts.ProbeFn != nil {
		probeCtx, cancel := context.WithTimeout(ctx, defaultProbeTimeout)
		defer cancel()
		stdout, stderr, code, probeErr := c.opts.ProbeFn(probeCtx, resolved, "--version")
		if probeErr != nil {
			return resolved, "", fmt.Errorf("cursoragent: probe exec: %w\n\n%s", probeErr, FixBinaryHelpBlock())
		}
		if code != 0 {
			return resolved, "", fmt.Errorf("cursoragent: probe %q: exit %d stderr=%q\n\n%s", resolved, code, trimForLog(stderr, 256), FixBinaryHelpBlock())
		}
		v := firstNonEmptyLine(stdout)
		if v == "" {
			v = firstNonEmptyLine(stderr)
		}
		if v == "" {
			return resolved, "", fmt.Errorf("cursoragent: probe %q: empty --version output\n\n%s", resolved, FixBinaryHelpBlock())
		}
		return resolved, v, nil
	}

	probeCtx, cancel := context.WithTimeout(ctx, defaultProbeTimeout)
	defer cancel()
	env := c.buildEnv(nil)
	stdout, stderr, code, probeErr := c.streamExec(probeCtx, "", env, nil, resolved, nil, "--version")
	if probeErr != nil {
		if errors.Is(probeErr, context.DeadlineExceeded) || errors.Is(probeErr, context.Canceled) {
			return resolved, "", fmt.Errorf("%w: %w\n\n%s", ErrTimeout, probeErr, FixBinaryHelpBlock())
		}
		return resolved, "", fmt.Errorf("cursoragent: probe exec: %w\n\n%s", probeErr, FixBinaryHelpBlock())
	}
	if code != 0 {
		return resolved, "", fmt.Errorf("cursoragent: probe %q: exit %d stderr=%q\n\n%s", resolved, code, Redact(trimForLog(stderr, 256), c.homePaths), FixBinaryHelpBlock())
	}
	v := firstNonEmptyLine(stdout)
	if v == "" {
		v = firstNonEmptyLine(stderr)
	}
	if v == "" {
		return resolved, "", fmt.Errorf("cursoragent: probe %q: empty --version output\n\n%s", resolved, FixBinaryHelpBlock())
	}
	return resolved, v, nil
}

// Run executes cursor-agent with the prompt on stdin and parses stream-json stdout into [RunResult.Text].
func (c *Client) Run(ctx context.Context, req RunRequest) (RunResult, error) {
	if len(strings.TrimSpace(string(req.Prompt))) == 0 {
		return RunResult{}, errors.New("cursoragent: empty Prompt")
	}
	if strings.TrimSpace(req.WorkingDir) == "" {
		return RunResult{}, errors.New("cursoragent: empty WorkingDir")
	}
	fi, err := os.Stat(req.WorkingDir)
	if err != nil || !fi.IsDir() {
		return RunResult{}, fmt.Errorf("cursoragent: WorkingDir not a directory: %w", err)
	}

	bin, err := c.resolvedBinaryForExec()
	if err != nil {
		return RunResult{}, err
	}
	resolved := ResolveBinaryPath(bin)

	runCtx, cancel := withOptionalTimeout(ctx, effectiveRunTimeout(c.opts.DefaultTimeout, req.Timeout))
	defer cancel()

	env := c.buildEnv(req.ExtraEnv)
	args := c.argvFor(c.effectiveModel(req))

	onLine := req.OnStdoutLine
	stdout, stderr, code, execErr := c.streamExec(runCtx, req.WorkingDir, env, req.Prompt, resolved, onLine, args...)
	stderrStr := Redact(string(stderr), c.homePaths)

	if execErr != nil {
		if errors.Is(runCtx.Err(), context.DeadlineExceeded) || errors.Is(runCtx.Err(), context.Canceled) {
			return RunResult{ExitCode: code, StderrTail: clipTail(stderrStr, MaxStderrTailRunes)}, fmt.Errorf("%w: %w", ErrTimeout, execErr)
		}
		return RunResult{ExitCode: code, StderrTail: clipTail(stderrStr, MaxStderrTailRunes)}, fmt.Errorf("%w: %w\n\n%s", ErrInvalidOutput, execErr, FixBinaryHelpBlock())
	}
	if code != 0 {
		return RunResult{
			ExitCode:   code,
			StderrTail: clipTail(stderrStr, MaxStderrTailRunes),
		}, fmt.Errorf("%w: exit %d stderr=%q", ErrNonZeroExit, code, clipTail(stderrStr, 400))
	}

	parsed, parseErr := parseStdout(stdout)
	if parseErr != nil {
		hint := Redact(string(stdout), c.homePaths)
		return RunResult{
			ExitCode:   code,
			StderrTail: clipTail(stderrStr, MaxStderrTailRunes),
		}, fmt.Errorf("%w: %w\n%s", ErrInvalidOutput, parseErr, clipTail(hint, MaxStderrTailRunes))
	}

	text := Redact(strings.TrimSpace(parsed.Result), c.homePaths)
	if parsed.IsError {
		return RunResult{
			Text:            text,
			ExitCode:        1,
			StderrTail:      clipTail(stderrStr, MaxStderrTailRunes),
			ResolvedModel:   parsed.ResolvedModel,
			MissingTerminal: parsed.MissingTerminalResult,
		}, fmt.Errorf("%w: agent reported is_error=true", ErrNonZeroExit)
	}

	return RunResult{
		Text:            text,
		ExitCode:        0,
		StderrTail:      clipTail(stderrStr, MaxStderrTailRunes),
		ResolvedModel:   parsed.ResolvedModel,
		MissingTerminal: parsed.MissingTerminalResult,
	}, nil
}

func effectiveRunTimeout(defaultTimeout, reqTimeout time.Duration) time.Duration {
	if reqTimeout > 0 {
		return reqTimeout
	}
	return defaultTimeout
}

func withOptionalTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if d <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}

func clipTail(s string, maxRunes int) string {
	if maxRunes <= 0 {
		return s
	}
	r := []rune(s)
	if len(r) <= maxRunes {
		return s
	}
	return string(r[:maxRunes]) + "…"
}
