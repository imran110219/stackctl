package main

import (
	"fmt"
	"os"

	"github.com/example/stackctl/internal/stackctl"
)

func main() {
	if err := stackctl.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}
