package main

import (
	"os"

	"kubectl-multi/pkg/cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}