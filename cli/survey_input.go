package cli

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Survey question styling matches github.com/AlecAivazis/survey/v2 defaults (green+hb ?, bold message).
var (
	configureQuestionIcon    = lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Bold(true)
	configureQuestionMessage = lipgloss.NewStyle().Bold(true)
	configureQuestionDefault = lipgloss.NewStyle().Foreground(lipgloss.Color("15"))
)

func formatConfigurePrompt(message, defaultVal string) string {
	icon := configureQuestionIcon.Render("?")
	msg := configureQuestionMessage.Render(message)
	if defaultVal != "" {
		def := configureQuestionDefault.Render("(" + defaultVal + ")")
		return icon + " " + msg + " " + def + " "
	}
	return icon + " " + msg + " "
}

// askConfigureLine prints a survey-style "? message" prompt and reads one line.
// It does not add extra blank lines between consecutive prompts.
func askConfigureLine(r *bufio.Reader, w io.Writer, message, defaultVal string) (string, error) {
	if r == nil {
		return "", errors.New("reader is nil")
	}
	fmt.Fprint(w, formatConfigurePrompt(message, defaultVal))
	line, err := r.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" && defaultVal != "" {
		return defaultVal, nil
	}
	if line == "" {
		return "", errors.New("required")
	}
	return line, nil
}
