package main

import (
	"fmt"
	"os"

	"github.com/AlexsanderHamir/prof/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "prof: %v\n", err)
		os.Exit(1)
	}
}
