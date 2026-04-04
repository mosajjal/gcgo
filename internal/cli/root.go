package cli

import (
	"fmt"
	"os"

	"github.com/mosajjal/gcgo/internal/ai"
	"github.com/mosajjal/gcgo/internal/alloydb"
	"github.com/mosajjal/gcgo/internal/artifacts"
	"github.com/mosajjal/gcgo/internal/asset"
	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/bigtable"
	"github.com/mosajjal/gcgo/internal/billing"
	"github.com/mosajjal/gcgo/internal/builds"
	"github.com/mosajjal/gcgo/internal/composer"
	"github.com/mosajjal/gcgo/internal/compute"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/container"
	"github.com/mosajjal/gcgo/internal/datacatalog"
	"github.com/mosajjal/gcgo/internal/dataflow"
	"github.com/mosajjal/gcgo/internal/dns"
	"github.com/mosajjal/gcgo/internal/dataplex"
	"github.com/mosajjal/gcgo/internal/dataproc"
	"github.com/mosajjal/gcgo/internal/deploy"
	"github.com/mosajjal/gcgo/internal/eventarc"
	"github.com/mosajjal/gcgo/internal/firestore"
	"github.com/mosajjal/gcgo/internal/folders"
	"github.com/mosajjal/gcgo/internal/functions"
	"github.com/mosajjal/gcgo/internal/iam"
	"github.com/mosajjal/gcgo/internal/kms"
	"github.com/mosajjal/gcgo/internal/logging"
	"github.com/mosajjal/gcgo/internal/monitoring"
	"github.com/mosajjal/gcgo/internal/organizations"
	"github.com/mosajjal/gcgo/internal/projects"
	"github.com/mosajjal/gcgo/internal/pubsub"
	"github.com/mosajjal/gcgo/internal/redis"
	"github.com/mosajjal/gcgo/internal/run"
	"github.com/mosajjal/gcgo/internal/scc"
	"github.com/mosajjal/gcgo/internal/scheduler"
	"github.com/mosajjal/gcgo/internal/secrets"
	"github.com/mosajjal/gcgo/internal/services"
	"github.com/mosajjal/gcgo/internal/spanner"
	"github.com/mosajjal/gcgo/internal/sql"
	"github.com/mosajjal/gcgo/internal/storage"
	"github.com/mosajjal/gcgo/internal/tasks"
	"github.com/mosajjal/gcgo/internal/version"
	"github.com/mosajjal/gcgo/internal/workflows"
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
	root.PersistentFlags().String("impersonate-service-account", "", "Service account email to impersonate for API calls")
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
	root.PersistentPreRunE = func(cmd *cobra.Command, _ []string) error {
		target, err := cmd.Root().PersistentFlags().GetString("impersonate-service-account")
		if err != nil {
			return fmt.Errorf("read impersonation flag: %w", err)
		}
		creds.SetImpersonateTarget(target)
		return nil
	}

	root.AddCommand(newWhoamiCommand(cfg, creds))
	root.AddCommand(newUseCommand(cfg))
	root.AddCommand(config.NewCommand(cfg))
	root.AddCommand(auth.NewCommand(creds))
	root.AddCommand(projects.NewCommand(creds))
	root.AddCommand(ai.NewCommand(cfg, creds))
	root.AddCommand(alloydb.NewCommand(cfg, creds))
	root.AddCommand(bigtable.NewCommand(cfg, creds))
	root.AddCommand(compute.NewCommand(cfg, creds))
	root.AddCommand(iam.NewCommand(cfg, creds))
	root.AddCommand(storage.NewCommand(cfg, creds))
	root.AddCommand(container.NewCommand(cfg, creds))
	root.AddCommand(deploy.NewCommand(cfg, creds))
	root.AddCommand(eventarc.NewCommand(cfg, creds))
	root.AddCommand(firestore.NewCommand(cfg, creds))
	root.AddCommand(run.NewCommand(cfg, creds))
	root.AddCommand(logging.NewCommand(cfg, creds))
	root.AddCommand(pubsub.NewCommand(cfg, creds))
	root.AddCommand(redis.NewCommand(cfg, creds))
	root.AddCommand(scheduler.NewCommand(cfg, creds))
	root.AddCommand(artifacts.NewCommand(cfg, creds))
	root.AddCommand(builds.NewCommand(cfg, creds))
	root.AddCommand(functions.NewCommand(cfg, creds))
	root.AddCommand(asset.NewCommand(creds))
	root.AddCommand(billing.NewCommand(creds))
	root.AddCommand(folders.NewCommand(creds))
	root.AddCommand(organizations.NewCommand(creds))
	root.AddCommand(composer.NewCommand(cfg, creds))
	root.AddCommand(datacatalog.NewCommand(cfg, creds))
	root.AddCommand(dataflow.NewCommand(cfg, creds))
	root.AddCommand(dataplex.NewCommand(cfg, creds))
	root.AddCommand(dataproc.NewCommand(cfg, creds))
	root.AddCommand(spanner.NewCommand(cfg, creds))
	root.AddCommand(sql.NewCommand(cfg, creds))
	root.AddCommand(services.NewCommand(cfg, creds))
	root.AddCommand(tasks.NewCommand(cfg, creds))
	root.AddCommand(kms.NewCommand(cfg, creds))
	root.AddCommand(secrets.NewCommand(cfg, creds))
	root.AddCommand(monitoring.NewCommand(cfg, creds))
	root.AddCommand(scc.NewCommand(creds))
	root.AddCommand(workflows.NewCommand(cfg, creds))
	root.AddCommand(dns.NewCommand(cfg, creds))

	configureHelp(root)

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
