package main

import (
	"os"

	"kubectl-multi/pkg/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
} 