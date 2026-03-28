package cli

import "github.com/AlexsanderHamir/prof/internal/app"

// Execute runs the CLI application with default (production) services.
func Execute() error {
	return ExecuteWith(nil)
}

// ExecuteWith runs the CLI using the given composition root. Pass nil to use [app.Default].
func ExecuteWith(services *app.Services) error {
	return CreateRootCmd(services).Execute()
}
