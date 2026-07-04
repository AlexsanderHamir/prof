package cli

import (
	"github.com/AlecAivazis/survey/v2"
	"github.com/AlecAivazis/survey/v2/terminal"
)

// cleanInput wraps survey.Input and replaces Cleanup.
//
// survey.AskOne normally calls Cleanup to re-print "? question: answer", which
// duplicates the line the user already typed. We skip that re-print and only
// erase the blank line readline leaves after Enter so the next prompt can follow
// immediately when there is no warning between steps.
type cleanInput struct {
	survey.Input
}

func (c *cleanInput) Cleanup(_ *survey.PromptConfig, _ interface{}) error {
	stdio := c.Stdio()
	if stdio.Out == nil {
		return nil
	}
	return terminal.EraseLine(stdio.Out, terminal.ERASE_LINE_ALL)
}
