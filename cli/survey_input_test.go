package cli

import (
	"testing"

	"github.com/AlecAivazis/survey/v2"
)

func TestCleanInput_CleanupIsNoop(t *testing.T) {
	t.Parallel()

	var p survey.Prompt = &cleanInput{Input: survey.Input{Message: "test"}}
	c, ok := p.(interface {
		Cleanup(*survey.PromptConfig, interface{}) error
	})
	if !ok {
		t.Fatal("cleanInput should implement Cleanup")
	}
	if err := c.Cleanup(&survey.PromptConfig{}, "answer"); err != nil {
		t.Fatalf("Cleanup() = %v", err)
	}
}
