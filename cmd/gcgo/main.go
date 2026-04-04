package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/cli"
)

func main() {
	// When symlinked as docker-credential-gcgo, act as a Docker credential helper
	// instead of the normal CLI. This is set up by "gcgo auth configure-docker".
	if filepath.Base(os.Args[0]) == "docker-credential-gcgo" {
		credDir, _ := auth.DefaultCredDir()
		creds := auth.New(credDir)
		if err := auth.RunDockerCredentialHelper(creds, os.Args[1:]); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		return
	}

	root := cli.NewRootCommand()
	if err := root.Execute(); err != nil {
		cli.FormatError(os.Stderr, err)
		fmt.Fprintln(os.Stderr)
		os.Exit(1)
	}
}
