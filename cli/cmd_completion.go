package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCompletionCmd(root *cobra.Command) *cobra.Command {
	return &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Write a shell completion script for prof to stdout.",
		Long: `Generate tab-completion scripts for prof.

Typical usage:

  Bash (write to a file sourced from ~/.bashrc):

    prof completion bash > prof-completion.bash

  Zsh:

    prof completion zsh > _prof

  Fish:

    prof completion fish > prof.fish

  PowerShell:

    prof completion powershell > prof.ps1

Then install the script using your shell's completion conventions.`,
		Args:                  cobra.ExactArgs(1),
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			w := cmd.OutOrStdout()
			switch args[0] {
			case "bash":
				return root.GenBashCompletionV2(w, true)
			case "zsh":
				return root.GenZshCompletion(w)
			case "fish":
				return root.GenFishCompletion(w, true)
			case "powershell":
				return root.GenPowerShellCompletion(w)
			default:
				return fmt.Errorf("unknown shell: %s", args[0])
			}
		},
	}
}
