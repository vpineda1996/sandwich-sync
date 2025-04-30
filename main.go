package main

import (
	"fmt"
	"os"

	"github.com/vpineda1996/sandwhich-sync/cmd/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
