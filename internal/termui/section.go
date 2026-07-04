package termui

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/term"
)

// SurveySectionTitle is the heading printed before interactive collect prompts.
const SurveySectionTitle = "Configure collection"

// ConfigureWarningPrefix is the left margin for warnings in the configure-collection section.
// Survey questions start at column 0, so warnings align with them.
const ConfigureWarningPrefix = ""

// ConfigureDetailPrefix indents detail lines under a configure-collection stage header.
const ConfigureDetailPrefix = "    "

// FormatWarningLine renders a styled warning line for interactive terminals.
func FormatWarningLine(prefix, msg string) string {
	return WarningPrefixStyle.Render(prefix+"warning: ") + WarningStyle.Render(msg)
}

// PrintWarning writes one styled warning line to w.
func PrintWarning(w io.Writer, prefix, msg string) {
	if w == nil {
		return
	}
	fmt.Fprintln(w, FormatWarningLine(prefix, msg))
}

// PrintSection writes a titled section break: blank line, bold title, faint rule, blank line.
func PrintSection(w io.Writer, fd int, title string) {
	if w == nil || title == "" {
		return
	}

	width := defaultTermWidth
	if fd >= 0 {
		if _, tw, err := term.GetSize(fd); err == nil && tw > 0 {
			width = tw
		}
	}

	fmt.Fprintln(w)
	fmt.Fprintln(w, BenchmarkTitleStyle.Render(title))
	fmt.Fprintln(w, SectionRuleStyle.Render(strings.Repeat("─", width)))
	fmt.Fprintln(w)
}

// StepGap prints a blank line between steps in a multi-step terminal flow.
func StepGap(w io.Writer) {
	if w == nil {
		return
	}
	fmt.Fprintln(w)
}

// EndSection prints a trailing blank line after a section's content.
func EndSection(w io.Writer) {
	StepGap(w)
}

const collectTransitionDuration = 500 * time.Millisecond

var sectionSpinner = spinner.Dot

// PrintTransition shows a brief in-place spinner while handing off to the collect pipeline.
func PrintTransition(w io.Writer, fd int, message string) {
	if w == nil || message == "" {
		return
	}
	if fd >= 0 && !term.IsTerminal(fd) {
		return
	}

	width := defaultTermWidth
	if fd >= 0 {
		if _, tw, err := term.GetSize(fd); err == nil && tw > 0 {
			width = tw
		}
	}

	frames := sectionSpinner.Frames
	ticker := time.NewTicker(sectionSpinner.FPS)
	defer ticker.Stop()

	deadline := time.Now().Add(collectTransitionDuration)
	label := message + "…"
	i := 0
	for time.Now().Before(deadline) {
		frame := frames[i%len(frames)]
		content := spinnerFrameStyle.Render(frame) + " " + LabelStyle.Render(label)
		visible := lipgloss.Width(content)
		padding := width - visible
		if padding < 0 {
			padding = 0
		}
		fmt.Fprint(w, "\r"+content+strings.Repeat(" ", padding))
		i++
		<-ticker.C
	}
	fmt.Fprint(w, ansi.EraseEntireLine+"\r")
}
