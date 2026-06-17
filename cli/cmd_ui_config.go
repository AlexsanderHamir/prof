package cli

import (
	"errors"
	"fmt"
	"os"

	"github.com/AlecAivazis/survey/v2"
	"github.com/AlexsanderHamir/prof/internal/app"
	"github.com/AlexsanderHamir/prof/internal/intent"
)

func runUIConfigCreate(svc *app.Services) error {
	path, err := svc.Config.Path()
	if err != nil {
		return err
	}

	if _, err = os.Stat(path); err == nil {
		fmt.Fprintf(os.Stdout, "Configuration already exists:\n  %s\n", path)
		fmt.Fprintln(os.Stdout, "Edit this file in your text editor to change filters or regression limits.")
		return nil
	}
	if !os.IsNotExist(err) {
		return err
	}

	create := false
	if err = survey.AskOne(&survey.Confirm{
		Message: "Create prof.json with a commented template next to go.mod?",
		Default: true,
	}, &create); err != nil {
		return err
	}
	if !create {
		return errors.New("configuration cancelled")
	}
	if err = intent.RunValidated(&intent.ConfigCreateIntent{}, svc); err != nil {
		return err
	}

	fmt.Fprintf(os.Stdout, "Created %s\n", path)
	fmt.Fprintln(os.Stdout, "Open it in your text editor. Lines starting with // are comments; uncomment optional sections as needed.")
	return nil
}
