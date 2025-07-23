package main

import (
	"fmt"
	"os"

	"github.com/AlexsanderHamir/prof/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "\n🚨🚨🚨 FATAL ERROR 🚨🚨🚨\n")
		fmt.Fprintf(os.Stderr, "%v\n", err)
		fmt.Fprintf(os.Stderr, "🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨🚨\n\n")
		os.Exit(1)
	}
}
