package compute

import (
	"context"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func newDiskTypesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disk-types",
		Short: "List available persistent disk types",
	}
	cmd.AddCommand(newDiskTypesListCommand(cfg, creds))
	return cmd
}

func newDiskTypesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List disk types available in a zone",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			dts, err := client.ListDiskTypes(ctx, project, zone)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), dts)
			}

			headers := []string{"NAME", "DESCRIPTION"}
			rows := make([][]string, len(dts))
			for i, dt := range dts {
				rows[i] = []string{dt.Name, dt.Description}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().String("project", "", "GCP project ID")
	AddZoneFlag(cmd)
	cmd.Flags().String("format", "table", "Output format: table, json")

	return cmd
}
