package cli

// Execute runs the CLI application
func Execute() error {
	return CreateRootCmd().Execute()
}
