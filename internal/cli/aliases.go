package cli

import (
	"encoding/json"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/spf13/cobra"
)

// newWhoamiCommand returns "gcgo whoami" — prints the active account,
// project, region and zone in one shot.
func newWhoamiCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "whoami",
		Short: "Show current identity and active configuration",
		Long:  "Prints the authenticated account, active project, default region, and default zone.",
		RunE: func(cmd *cobra.Command, _ []string) error {
			account, err := creds.ActiveAccount()
			if err != nil {
				account = "(not authenticated — run 'gcgo auth login')"
			}

			project := cfg.Project("")
			if project == "" {
				project = "(not set — run 'gcgo config set project PROJECT_ID')"
			}

			region := cfg.Region()
			if region == "" {
				region = "(not set)"
			}

			zone := cfg.Zone()
			if zone == "" {
				zone = "(not set)"
			}

			format, _ := cmd.Root().PersistentFlags().GetString("format")
			if format == "json" {
				out := map[string]string{
					"account": account,
					"project": project,
					"region":  region,
					"zone":    zone,
				}
				enc := json.NewEncoder(cmd.OutOrStdout())
				enc.SetIndent("", "  ")
				return enc.Encode(out)
			}

			rows := []struct{ key, val string }{
				{"account", account},
				{"project", project},
				{"region", region},
				{"zone", zone},
			}
			for _, r := range rows {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%-10s  %s\n", r.key, r.val)
			}
			return nil
		},
	}
}

// newUseCommand returns "gcgo use PROJECT [--region REGION] [--zone ZONE]"
// as a shorthand for setting config values in one step.
func newUseCommand(cfg *config.Config) *cobra.Command {
	var region, zone string

	cmd := &cobra.Command{
		Use:   "use PROJECT",
		Short: "Set the active project (and optionally region/zone)",
		Long:  "Shorthand for 'gcgo config set project PROJECT'. Optionally set region and zone in the same command.",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := cfg.Set("project", args[0]); err != nil {
				return err
			}
			if region != "" {
				if err := cfg.Set("region", region); err != nil {
					return err
				}
			}
			if zone != "" {
				if err := cfg.Set("zone", zone); err != nil {
					return err
				}
			}
			if err := cfg.Save(); err != nil {
				return fmt.Errorf("save config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "project  %s\n", args[0])
			if region != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "region   %s\n", region)
			}
			if zone != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "zone     %s\n", zone)
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Also set the default region")
	cmd.Flags().StringVar(&zone, "zone", "", "Also set the default zone")
	return cmd
}
