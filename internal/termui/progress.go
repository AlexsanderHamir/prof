package termui

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

// Phase identifies a user-visible collect pipeline step.
type Phase string

const (
	// PhasePrepare covers setup and prelude warnings before benchmarks run.
	PhasePrepare Phase = "prepare"
	// PhaseRunBenchmark covers go test, bench text write, and profile binary move.
	PhaseRunBenchmark Phase = "run_benchmark"
	// PhaseCollectProfiles covers pprof text and PNG generation.
	PhaseCollectProfiles Phase = "collect_profiles"
	// PhaseCollectFunctionProfiles covers parser extraction and per-function pprof lists.
	PhaseCollectFunctionProfiles Phase = "collect_function_profiles"
)

const defaultTermWidth = 120

// Progress describes a long-running step for user-facing labels.
type Progress struct {
	Phase  Phase
	Label  string // benchmark name
	Index  int    // 1-based benchmark index
	Total  int    // benchmark count
	Detail string // optional detail, e.g. "count=5" or "cpu, memory"
}

// WithPhase returns a copy of p with the given phase.
func (p Progress) WithPhase(phase Phase) Progress {
	p.Phase = phase
	return p
}

// WithDetail returns a copy of p with the given detail suffix.
func (p Progress) WithDetail(detail string) Progress {
	p.Detail = detail
	return p
}

// Session drives TTY progress UI for a collect run. Nil is non-interactive.
//
// Interactive layout:
//
//	✓ Preparing
//	    warning: …
//
//	Benchmark 1/2 · BenchmarkStringProcessor
//	  ✓ 0) Run benchmark (count=5)
//	  ✓ 1) Collect profiles (cpu, memory)
//	      warning: …
//	  ✓ 2) Collect per-function text profiles
//
// The step line updates in place while running (spinner), then becomes ✓ when done.
type Session struct {
	w                 io.Writer
	fd                int
	interactive       bool
	termWidthOverride int // tests only; when > 0, used instead of terminal size

	mu                 sync.Mutex
	stageActive        bool
	headerLocked       bool
	detailLines        int
	warnCount          int
	errorDetailEmitted bool
	errorDisplayed     bool
	runningLabel       string
	warnPrefix         string
	lastFrame          string
	spinnerStop        chan struct{}
	spinnerDone        sync.WaitGroup
	benchmarksStarted  int
}

// NewSession reports whether w/fd is an interactive terminal.
func NewSession(w io.Writer, fd int) *Session {
	return &Session{
		w:           w,
		fd:          fd,
		interactive: term.IsTerminal(fd),
	}
}

// Interactive is true when the session can show spinners and styled lines.
func (s *Session) Interactive() bool {
	if s == nil {
		return false
	}
	return s.interactive
}

func formatProgressLabel(p Progress, running bool) string {
	linePrefix, _ := phasePrefixes(p.Phase)
	var b strings.Builder
	b.WriteString(linePrefix)
	switch p.Phase {
	case PhasePrepare:
		b.WriteString("Preparing")
	case PhaseRunBenchmark:
		b.WriteString("0) Run benchmark")
		if p.Detail != "" {
			fmt.Fprintf(&b, " (%s)", p.Detail)
		}
	case PhaseCollectProfiles:
		b.WriteString("1) Collect profiles")
		if p.Detail != "" {
			fmt.Fprintf(&b, " (%s)", p.Detail)
		}
	case PhaseCollectFunctionProfiles:
		b.WriteString("2) Collect per-function text profiles")
	}
	if running {
		b.WriteString("…")
	}
	return b.String()
}

func phasePrefixes(phase Phase) (linePrefix, warnPrefix string) {
	if phase == PhasePrepare {
		return "", "    "
	}
	return "  ", "      "
}

// BeginCollect plays a short transition then prints the collect section header.
func (s *Session) BeginCollect() {
	if s == nil || !s.interactive {
		return
	}
	PrintTransition(s.w, s.fd, CollectSectionTitle)
	PrintSection(s.w, s.fd, CollectSectionTitle)
}

// BeginBenchmark prints a section header before the three steps for one benchmark.
func (s *Session) BeginBenchmark(index, total int, name string) {
	if s == nil || !s.interactive {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	fmt.Fprintln(s.w)
	s.benchmarksStarted++

	var title string
	if total > 1 {
		title = fmt.Sprintf("Benchmark %d/%d · %s", index, total, name)
	} else {
		title = name
	}
	fmt.Fprintln(s.w, BenchmarkTitleStyle.Render(title))
}

var dotSpinner = spinner.Dot

// ErrStagedDisplay marks an error whose message was already printed under a stage header.
var ErrStagedDisplay = errors.New("termui: error rendered under stage")

// StagedDisplay wraps err when the interactive session already showed it under a stage.
func StagedDisplay(err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%w: %w", ErrStagedDisplay, err)
}

// ErrorWasStaged reports whether err was wrapped with ErrStagedDisplay.
func ErrorWasStaged(err error) bool {
	return errors.Is(err, ErrStagedDisplay)
}

// RunWhile runs fn while showing a persistent spinner when the session is interactive.
func (s *Session) RunWhile(p Progress, fn func() error) error {
	if fn == nil {
		return errors.New("termui: nil task")
	}
	if s == nil || !s.interactive {
		return fn()
	}

	s.mu.Lock()
	s.stageActive = true
	s.headerLocked = false
	s.detailLines = 0
	s.warnCount = 0
	s.errorDetailEmitted = false
	_, s.warnPrefix = phasePrefixes(p.Phase)
	s.runningLabel = formatProgressLabel(p, true)
	s.startSpinnerLocked()
	s.mu.Unlock()

	fnErr := fn()

	s.mu.Lock()
	s.signalSpinnerStopLocked()
	doneLabel := formatProgressLabel(p, false)
	failed := fnErr != nil
	s.mu.Unlock()

	s.spinnerDone.Wait()

	s.mu.Lock()
	if fnErr != nil {
		s.captureStageErrorLocked(fnErr)
	}
	s.finishStageLocked(doneLabel, failed)
	s.stageActive = false
	s.mu.Unlock()

	return fnErr
}

func (s *Session) startSpinnerLocked() {
	stop := make(chan struct{})
	s.spinnerStop = stop
	s.spinnerDone.Add(1)

	frames := dotSpinner.Frames
	s.lastFrame = frames[0]
	s.paintSpinnerHeaderLocked(s.lastFrame)

	go func() {
		defer s.spinnerDone.Done()
		ticker := time.NewTicker(dotSpinner.FPS)
		defer ticker.Stop()

		i := 1
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				s.mu.Lock()
				if !s.headerLocked {
					frame := frames[i%len(frames)]
					s.lastFrame = frame
					s.paintSpinnerHeaderLocked(frame)
				}
				s.mu.Unlock()
				i++
			}
		}
	}()
}

func (s *Session) paintSpinnerHeaderLocked(frame string) {
	content := spinnerFrameStyle.Render(frame) + " " + LabelStyle.Render(s.runningLabel)
	s.overwriteLineLocked(content, false)
}

func (s *Session) lockHeaderLocked() {
	if s.headerLocked {
		return
	}
	s.headerLocked = true
	// Commit the spinner line (erase + newline) so warnings append below a stable header.
	content := spinnerFrameStyle.Render(s.lastFrame) + " " + LabelStyle.Render(s.runningLabel)
	s.overwriteLineLocked(content, true)
}

func (s *Session) signalSpinnerStopLocked() {
	if s.spinnerStop == nil {
		return
	}
	close(s.spinnerStop)
	s.spinnerStop = nil
}

func (s *Session) finishStageLocked(doneLabel string, failed bool) {
	mark := DoneStyle.Render("✓")
	if failed {
		mark = FailStyle.Render("✗")
	}
	line := mark + " " + doneLabel
	if !failed && s.warnCount > 0 {
		line += warnCountSuffix(s.warnCount)
	}

	if s.detailLines > 0 {
		s.seekStageHeaderLocked()
		s.overwriteLineLocked(line, false)
		s.seekAfterStageBlockLocked()
		return
	}

	s.overwriteLineLocked(line, true)
}

func warnCountSuffix(count int) string {
	noun := "warnings"
	if count == 1 {
		noun = "warning"
	}
	return FaintStyle.Render(fmt.Sprintf(" (%d %s)", count, noun))
}

func (s *Session) seekStageHeaderLocked() {
	n := s.detailLines + 1
	if n > 0 {
		fmt.Fprint(s.w, ansi.CursorUp(n))
	}
}

func (s *Session) seekAfterStageBlockLocked() {
	n := s.detailLines + 1
	if n > 0 {
		fmt.Fprint(s.w, ansi.CursorDown(n))
		// Width-padded overwrite leaves the cursor at the terminal edge; reset column
		// so the next write starts at the left margin.
		fmt.Fprint(s.w, "\r")
	}
}

func (s *Session) termWidth() int {
	if s.termWidthOverride > 0 {
		return s.termWidthOverride
	}
	_, width, err := term.GetSize(s.fd)
	if err != nil || width <= 0 {
		return defaultTermWidth
	}
	return width
}

func (s *Session) overwriteLineLocked(content string, newline bool) {
	width := s.termWidth()
	visible := lipgloss.Width(content)
	padding := width - visible
	if padding < 0 {
		padding = 0
	}
	fmt.Fprint(s.w, ansi.EraseEntireLine+"\r"+content+strings.Repeat(" ", padding))
	if newline {
		fmt.Fprint(s.w, "\n")
	}
}

func formatWarningLine(prefix, msg string) string {
	return FormatWarningLine(prefix, msg)
}

// StageDetailKind identifies a stage-scoped diagnostic line.
type StageDetailKind int

const (
	// StageWarn is a recoverable issue; the stage may still succeed.
	StageWarn StageDetailKind = iota
	// StageError is a failure tied to the active stage.
	StageError
)

func formatStageDetailLine(kind StageDetailKind, prefix, msg string) string {
	switch kind {
	case StageError:
		return ErrorPrefixStyle.Render(prefix+"error: ") + ErrorStyle.Render(msg)
	case StageWarn:
		return formatWarningLine(prefix, msg)
	default:
		return formatWarningLine(prefix, msg)
	}
}

func shortUserMessage(err error) string {
	if err == nil {
		return ""
	}
	msg := err.Error()
	if idx := strings.IndexByte(msg, '\n'); idx >= 0 {
		first := strings.TrimSpace(msg[:idx])
		if first != "" {
			return first + " (truncated)"
		}
	}
	return msg
}

func (s *Session) writeStageDetailLinesLocked(kind StageDetailKind, msg string) {
	for _, line := range splitDetailMessage(msg) {
		for _, formatted := range wrapDetailLines(kind, s.warnPrefix, line, s.termWidth()) {
			fmt.Fprintln(s.w, formatted)
			s.detailLines++
		}
	}
}

func splitDetailMessage(msg string) []string {
	parts := strings.Split(msg, "\n")
	var lines []string
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			lines = append(lines, p)
		}
	}
	if len(lines) == 0 {
		return []string{msg}
	}
	return lines
}

func stageDetailHead(kind StageDetailKind, prefix string) string {
	switch kind {
	case StageError:
		return ErrorPrefixStyle.Render(prefix + "error: ")
	default:
		return WarningPrefixStyle.Render(prefix + "warning: ")
	}
}

func stageDetailBodyStyle(kind StageDetailKind) lipgloss.Style {
	switch kind {
	case StageError:
		return ErrorStyle
	default:
		return WarningStyle
	}
}

// wrapDetailLines splits a long diagnostic into terminal-width lines. The first
// line includes the styled prefix; continuations align under the message text.
func wrapDetailLines(kind StageDetailKind, prefix, msg string, maxWidth int) []string {
	full := formatStageDetailLine(kind, prefix, msg)
	if maxWidth <= 0 || lipgloss.Width(full) <= maxWidth {
		return []string{full}
	}

	head := stageDetailHead(kind, prefix)
	headWidth := lipgloss.Width(head)
	bodyStyle := stageDetailBodyStyle(kind)
	contPad := strings.Repeat(" ", headWidth)

	firstMax := maxWidth - headWidth
	if firstMax < 1 {
		firstMax = 1
	}
	contMax := maxWidth - headWidth
	if contMax < 1 {
		contMax = 1
	}

	var lines []string
	remaining := strings.TrimSpace(msg)
	first := true
	for remaining != "" {
		limit := contMax
		if first {
			limit = firstMax
		}
		chunk, rest := takeWrappedChunk(remaining, limit)
		if chunk == "" {
			break
		}
		body := bodyStyle.Render(chunk)
		if first {
			lines = append(lines, head+body)
			first = false
		} else {
			lines = append(lines, contPad+body)
		}
		remaining = strings.TrimSpace(rest)
	}
	if len(lines) == 0 {
		return []string{full}
	}
	return lines
}

// takeWrappedChunk returns the largest prefix of text that fits in limit visible
// columns, preferring word boundaries; a single overlong word is hard-broken.
func takeWrappedChunk(text string, limit int) (chunk, rest string) {
	if limit <= 0 {
		return "", text
	}
	text = strings.TrimLeft(text, " ")
	if text == "" {
		return "", ""
	}
	if runeWidth(text) <= limit {
		return text, ""
	}

	words := strings.Fields(text)
	if len(words) == 0 {
		return hardBreakRunes(text, limit)
	}

	var b strings.Builder
	for i, word := range words {
		sep := ""
		if b.Len() > 0 {
			sep = " "
		}
		candidate := sep + word
		if runeWidth(candidate) > limit && b.Len() == 0 {
			return hardBreakRunes(word, limit)
		}
		if runeWidth(b.String()+candidate) > limit {
			restWords := words[i:]
			return strings.TrimSpace(b.String()), strings.Join(restWords, " ")
		}
		b.WriteString(candidate)
	}
	return b.String(), ""
}

func runeWidth(s string) int {
	return lipgloss.Width(s)
}

func hardBreakRunes(text string, limit int) (chunk, rest string) {
	runes := []rune(text)
	if len(runes) == 0 {
		return "", ""
	}
	for n := len(runes); n > 0; n-- {
		part := string(runes[:n])
		if runeWidth(part) <= limit {
			return part, string(runes[n:])
		}
	}
	return string(runes[:1]), string(runes[1:])
}

func (s *Session) captureStageErrorLocked(err error) {
	if err == nil || s.errorDetailEmitted {
		return
	}
	s.lockHeaderLocked()
	s.writeStageDetailLinesLocked(StageError, shortUserMessage(err))
	s.errorDetailEmitted = true
	s.errorDisplayed = true
}

func (s *Session) stageDetail(kind StageDetailKind, msg string) {
	if s == nil || !s.interactive {
		if kind == StageWarn {
			slog.Warn(msg)
		} else {
			slog.Error(msg)
		}
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.stageActive {
		if kind == StageWarn {
			slog.Warn(msg)
		} else {
			slog.Error(msg)
		}
		return
	}

	s.lockHeaderLocked()
	s.writeStageDetailLinesLocked(kind, msg)
	if kind == StageWarn {
		s.warnCount++
	}
	if kind == StageError {
		s.errorDetailEmitted = true
		s.errorDisplayed = true
	}
}

// Error writes a styled error on its own line below the active stage header.
func (s *Session) Error(msg string) {
	s.stageDetail(StageError, msg)
}

// ErrorDisplayed reports whether an error was rendered under a stage header.
func (s *Session) ErrorDisplayed() bool {
	if s == nil {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.errorDisplayed
}

// Warn writes a styled warning on its own line below the active stage header.
func (s *Session) Warn(msg string) {
	s.stageDetail(StageWarn, msg)
}

// Success writes a completion message (styled on TTY, slog otherwise).
func (s *Session) Success(msg string) {
	if s == nil || !s.interactive {
		slog.Info(msg)
		return
	}
	fmt.Fprintln(s.w)
	fmt.Fprintln(s.w, SuccessStyle.Render(msg))
	fmt.Fprintln(s.w)
}

// newSessionForTest builds a session with a forced interactive flag (tests only).
func newSessionForTest(w io.Writer) *Session {
	return &Session{w: w, interactive: true, termWidthOverride: defaultTermWidth}
}

// newNarrowSessionForTest builds an interactive session with a narrow terminal width.
func newNarrowSessionForTest(w io.Writer, width int) *Session {
	return &Session{w: w, interactive: true, termWidthOverride: width}
}

var spinnerFrameStyle = lipgloss.NewStyle().Foreground(lipgloss.Color(AccentColor))
