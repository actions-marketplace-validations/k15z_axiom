package main

import (
	"fmt"
	"os"

	"github.com/k15z/axiom/internal/cli"
)

func main() {
	if err := cli.NewRootCmd().Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}
}
