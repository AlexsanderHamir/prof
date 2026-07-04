package cli

import "github.com/AlecAivazis/survey/v2"

// cleanInput wraps survey.Input and skips Cleanup.
//
// survey.AskOne always calls Cleanup after Input.Prompt, which re-renders
// "? question: answer" on a new line. When the terminal echoes typed input
// (ShowCursor) or readline emits an extra newline (Windows), that cleanup line
// duplicates the answer the user already sees — it is not a second prompt.
type cleanInput struct {
	survey.Input
}

func (cleanInput) Cleanup(*survey.PromptConfig, interface{}) error {
	return nil
}
