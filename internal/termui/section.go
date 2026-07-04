package termui

import (
	"fmt"
	"io"
	"strings"

	"golang.org/x/term"
)

// SurveySectionTitle is the heading printed before interactive collect prompts.
const SurveySectionTitle = "Configure collection"

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

// EndSection prints a trailing blank line after a section's content.
func EndSection(w io.Writer) {
	if w == nil {
		return
	}
	fmt.Fprintln(w)
}
