package composer

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the composer command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "composer",
		Short: "Manage Cloud Composer environments",
	}

	cmd.AddCommand(
		newEnvironmentsCommand(cfg, creds),
	)

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

func newEnvironmentsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "environments",
		Short: "Manage Composer environments",
	}

	cmd.AddCommand(
		newEnvironmentsListCommand(cfg, creds),
		newEnvironmentsDescribeCommand(cfg, creds),
		newEnvironmentsCreateCommand(cfg, creds),
		newEnvironmentsDeleteCommand(cfg, creds),
		newEnvironmentsUpdateCommand(cfg, creds),
	)

	return cmd
}

func newEnvironmentsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Composer environments",
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
			envs, err := client.ListEnvironments(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), envs)
			}
			headers := []string{"NAME", "STATE", "CONFIG", "REGION"}
			rows := make([][]string, len(envs))
			for i, e := range envs {
				rows[i] = []string{e.Name, e.State, e.Config, e.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newEnvironmentsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ENVIRONMENT",
		Short: "Describe a Composer environment",
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
			env, err := client.GetEnvironment(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), env)
		},
	}
}

func newEnvironmentsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateEnvironmentRequest

	cmd := &cobra.Command{
		Use:   "create ENVIRONMENT",
		Short: "Create a Composer environment",
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
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateEnvironment(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created environment %q.\n", req.Name)
			return nil
		},
	}

	cmd.Flags().Int64Var(&req.NodeCount, "node-count", 1, "Worker node count")
	cmd.Flags().StringVar(&req.MachineType, "machine-type", "", "Worker machine type")
	cmd.Flags().StringVar(&req.ImageVersion, "image-version", "", "Composer image version")

	return cmd
}

func newEnvironmentsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENVIRONMENT",
		Short: "Delete a Composer environment",
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
			if err := client.DeleteEnvironment(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted environment %q.\n", args[0])
			return nil
		},
	}
}

func newEnvironmentsUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req UpdateEnvironmentRequest

	cmd := &cobra.Command{
		Use:   "update ENVIRONMENT",
		Short: "Update a Composer environment",
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
			if err := client.UpdateEnvironment(ctx, project, region, args[0], &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated environment %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().Int64Var(&req.NodeCount, "node-count", 0, "Worker node count")
	cmd.Flags().StringToStringVar(&req.Labels, "label", nil, "Labels in key=value form")

	return cmd
}
