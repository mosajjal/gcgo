package main

import (
	"fmt"
	"os"

	"github.com/mosajjal/gcgo/internal/cli"
)

func main() {
	root := cli.NewRootCommand()
	if err := root.Execute(); err != nil {
		cli.FormatError(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
}
