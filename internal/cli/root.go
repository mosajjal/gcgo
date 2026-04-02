package cli

import (
	"fmt"
	"os"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/compute"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/container"
	"github.com/mosajjal/gcgo/internal/iam"
	"github.com/mosajjal/gcgo/internal/logging"
	"github.com/mosajjal/gcgo/internal/projects"
	"github.com/mosajjal/gcgo/internal/run"
	"github.com/mosajjal/gcgo/internal/storage"
	"github.com/mosajjal/gcgo/internal/version"
	"github.com/spf13/cobra"
)

// NewRootCommand builds the root cobra command with all subcommands wired in.
func NewRootCommand() *cobra.Command {
	root := &cobra.Command{
		Use:           "gcgo",
		Short:         "A fast Google Cloud CLI",
		SilenceUsage:  true,
		SilenceErrors: true,
	}

	root.PersistentFlags().String("project", "", "GCP project ID (overrides config)")
	root.PersistentFlags().String("format", "table", "Output format: table, json")
	root.PersistentFlags().Bool("quiet", false, "Suppress non-essential output")

	root.AddCommand(newVersionCommand())

	cfg, err := config.Load()
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "warning: could not load config: %v\n", err)
		cfg = &config.Config{}
	}

	credDir, err := auth.DefaultCredDir()
	if err != nil {
		credDir = ""
	}
	creds := auth.New(credDir)

	root.AddCommand(config.NewCommand(cfg))
	root.AddCommand(auth.NewCommand(creds))
	root.AddCommand(projects.NewCommand(creds))
	root.AddCommand(compute.NewCommand(cfg, creds))
	root.AddCommand(iam.NewCommand(cfg, creds))
	root.AddCommand(storage.NewCommand(cfg, creds))
	root.AddCommand(container.NewCommand(cfg, creds))
	root.AddCommand(run.NewCommand(cfg, creds))
	root.AddCommand(logging.NewCommand(cfg, creds))

	return root
}

func newVersionCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(_ *cobra.Command, _ []string) {
			info := version.Info()
			fmt.Printf("gcgo %s (commit: %s, built: %s, %s/%s)\n",
				info["version"], info["git_commit"], info["build_time"],
				info["os"], info["arch"])
		},
	}
}
