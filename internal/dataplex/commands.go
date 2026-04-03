package dataplex

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the dataplex command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataplex",
		Short: "Manage Dataplex resources",
	}

	cmd.AddCommand(newLakesCommand(cfg, creds), newZonesCommand(cfg, creds))
	return cmd
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func requireRegion(cmd *cobra.Command, cfg *config.Config) (string, error) {
	region, _ := cmd.Flags().GetString("region")
	if region == "" {
		region = cfg.Region()
	}
	if region == "" {
		return "", fmt.Errorf("no region set (use --region or 'gcgo config set region REGION')")
	}
	return region, nil
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newLakesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lakes",
		Short: "Manage Dataplex lakes",
	}

	cmd.AddCommand(
		newLakesListCommand(cfg, creds),
		newLakesDescribeCommand(cfg, creds),
		newLakesCreateCommand(cfg, creds),
		newLakesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newLakesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List lakes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			lakes, err := client.ListLakes(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), lakes)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "REGION"}
			rows := make([][]string, len(lakes))
			for i, l := range lakes {
				rows[i] = []string{l.Name, l.DisplayName, l.State, l.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newLakesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe LAKE",
		Short: "Describe a lake",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			lake, err := client.GetLake(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), lake)
		},
	}
}

func newLakesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateLakeRequest
	cmd := &cobra.Command{
		Use:   "create LAKE",
		Short: "Create a lake",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			req.LakeID = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateLake(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created lake %q.\n", req.LakeID)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Display name")
	return cmd
}

func newLakesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete LAKE",
		Short: "Delete a lake",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteLake(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted lake %q.\n", args[0])
			return nil
		},
	}
}

func newZonesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zones",
		Short: "Manage Dataplex zones",
	}

	cmd.AddCommand(
		newZonesListCommand(cfg, creds),
		newZonesDescribeCommand(cfg, creds),
		newZonesCreateCommand(cfg, creds),
		newZonesDeleteCommand(cfg, creds),
	)
	return cmd
}

func requireLake(cmd *cobra.Command) (string, error) {
	lake, _ := cmd.Flags().GetString("lake")
	if lake == "" {
		return "", fmt.Errorf("--lake is required")
	}
	return lake, nil
}

func newZonesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var lake string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List zones in a lake",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if lake == "" {
				return fmt.Errorf("--lake is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			zones, err := client.ListZones(ctx, project, region, lake)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), zones)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "TYPE"}
			rows := make([][]string, len(zones))
			for i, z := range zones {
				rows[i] = []string{z.Name, z.DisplayName, z.State, z.Type}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&lake, "lake", "", "Lake ID")
	return cmd
}

func newZonesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var lake string
	cmd := &cobra.Command{
		Use:   "describe ZONE",
		Short: "Describe a zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if lake == "" {
				return fmt.Errorf("--lake is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			zone, err := client.GetZone(ctx, project, region, lake, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), zone)
		},
	}
	cmd.Flags().StringVar(&lake, "lake", "", "Lake ID")
	return cmd
}

func newZonesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		lake string
		req  CreateZoneRequest
	)
	cmd := &cobra.Command{
		Use:   "create ZONE",
		Short: "Create a zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if lake == "" {
				return fmt.Errorf("--lake is required")
			}
			req.ZoneID = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateZone(ctx, project, region, lake, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created zone %q.\n", req.ZoneID)
			return nil
		},
	}
	cmd.Flags().StringVar(&lake, "lake", "", "Lake ID")
	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&req.Type, "type", "RAW_DATA", "Zone type")
	return cmd
}

func newZonesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var lake string
	cmd := &cobra.Command{
		Use:   "delete ZONE",
		Short: "Delete a zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if lake == "" {
				return fmt.Errorf("--lake is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteZone(ctx, project, region, lake, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted zone %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&lake, "lake", "", "Lake ID")
	return cmd
}
