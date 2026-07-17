package main

import (
	"fmt"
	"os"

	"github.com/pierinho13/kubectl-peek/cmd"
)

func main() {

	if err := cmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
