package compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func newZonesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zones",
		Short: "List available zones",
	}
	cmd.AddCommand(newZonesListCommand(cfg, creds))
	return cmd
}

func newZonesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available zones",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, _ := cmd.Flags().GetString("region")

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			zones, err := client.ListZones(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), zones)
			}
			headers := []string{"NAME", "REGION", "STATUS"}
			rows := make([][]string, len(zones))
			for i, z := range zones {
				rows[i] = []string{z.Name, z.Region, z.Status}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("project", "", "GCP project ID")
	cmd.Flags().String("region", "", "Filter by region")
	cmd.Flags().String("format", "table", "Output format: table, json")
	return cmd
}

func newRegionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "regions",
		Short: "List available regions",
	}
	cmd.AddCommand(newRegionsListCommand(cfg, creds))
	return cmd
}

func newRegionsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List available regions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			regions, err := client.ListRegions(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), regions)
			}
			headers := []string{"NAME", "STATUS", "ZONES"}
			rows := make([][]string, len(regions))
			for i, r := range regions {
				rows[i] = []string{r.Name, r.Status, strings.Join(r.Zones, ", ")}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("project", "", "GCP project ID")
	cmd.Flags().String("format", "table", "Output format: table, json")
	return cmd
}

func newMachineTypesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "machine-types",
		Short: "List available machine types",
	}
	cmd.AddCommand(newMachineTypesListCommand(cfg, creds))
	return cmd
}

func newMachineTypesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List machine types available in a zone",
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

			mts, err := client.ListMachineTypes(ctx, project, zone)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), mts)
			}
			headers := []string{"NAME", "VCPUS", "MEMORY_MB", "DESCRIPTION"}
			rows := make([][]string, len(mts))
			for i, m := range mts {
				rows[i] = []string{m.Name, fmt.Sprintf("%d", m.VCPUs), fmt.Sprintf("%d", m.MemoryMb), m.Description}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("project", "", "GCP project ID")
	cmd.Flags().String("zone", "", "Zone (falls back to config)")
	cmd.Flags().String("format", "table", "Output format: table, json")
	return cmd
}
