package services

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the services command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage Service Usage",
	}

	cmd.AddCommand(
		newListCommand(cfg, creds),
		newDescribeCommand(cfg, creds),
		newEnableCommand(cfg, creds),
		newDisableCommand(cfg, creds),
	)
	return cmd
}

func servicesClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func newListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := servicesClient(ctx, creds)
			if err != nil {
				return err
			}
			services, err := client.ListServices(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), services)
			}
			headers := []string{"NAME", "TITLE", "STATE"}
			rows := make([][]string, len(services))
			for i, svc := range services {
				rows[i] = []string{svc.Name, svc.Title, svc.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe SERVICE",
		Short: "Describe a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := servicesClient(ctx, creds)
			if err != nil {
				return err
			}
			svc, err := client.GetService(ctx, project, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), svc)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:  %s\n", svc.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Title: %s\n", svc.Title)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State: %s\n", svc.State)
			return nil
		},
	}
}

func newEnableCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "enable SERVICE",
		Short: "Enable a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := servicesClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.EnableService(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Enabled %s.\n", args[0])
			return nil
		},
	}
}

func newDisableCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "disable SERVICE",
		Short: "Disable a service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := servicesClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DisableService(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Disabled %s.\n", args[0])
			return nil
		},
	}
}
