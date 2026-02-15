package main

import (
	"fmt"
	"os"

	"github.com/example/stackctl/internal/stackctl"
	"github.com/example/stackctl/internal/tui"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 && args[0] == "setup" {
		if err := tui.StartWizard(); err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if err := stackctl.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
