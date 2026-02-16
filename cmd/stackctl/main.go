package main

import (
	"fmt"
	"os"

	"github.com/example/stackctl/internal/stackctl"
	"github.com/example/stackctl/internal/tui"
)

func main() {
	args := os.Args[1:]

	if len(args) > 0 {
		env := extractEnvFlag(args[1:])
		switch args[0] {
		case "setup":
			if err := tui.StartWizard(); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		case "modules":
			if err := tui.StartModuleManager(env); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		case "dash":
			if err := tui.StartDashboard(env); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		case "config":
			if err := tui.StartConfigWizard(env); err != nil {
				fmt.Fprintf(os.Stderr, "error: %v\n", err)
				os.Exit(1)
			}
			return
		}
	}

	if err := stackctl.Run(args); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

func extractEnvFlag(args []string) string {
	for i, arg := range args {
		if arg == "--env" && i+1 < len(args) {
			return args[i+1]
		}
	}
	return ""
}
