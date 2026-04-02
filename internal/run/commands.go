package run

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the run command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Manage Cloud Run services",
	}

	cmd.AddCommand(
		newServicesCommand(cfg, creds),
		newDeployCommand(cfg, creds),
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

func runClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newServicesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage Cloud Run services",
	}

	cmd.AddCommand(
		newServicesListCommand(cfg, creds),
		newServicesDescribeCommand(cfg, creds),
		newServicesDeleteCommand(cfg, creds),
	)

	return cmd
}

func newServicesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloud Run services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			services, err := client.ListServices(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), services)
			}

			headers := []string{"NAME", "REGION", "URL"}
			rows := make([][]string, len(services))
			for i, s := range services {
				rows[i] = []string{s.Name, s.Region, s.URI}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")

	return cmd
}

func newServicesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "describe SERVICE",
		Short: "Describe a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			svc, err := client.GetService(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), svc)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")

	return cmd
}

func newServicesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "delete SERVICE",
		Short: "Delete a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteService(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted service %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")

	return cmd
}

func newDeployCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req DeployRequest
	var region string
	var envSlice []string

	cmd := &cobra.Command{
		Use:   "deploy SERVICE --image=IMAGE",
		Short: "Deploy a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			req.Name = args[0]
			req.Env = make(map[string]string)
			for _, e := range envSlice {
				k, v, ok := strings.Cut(e, "=")
				if !ok {
					return fmt.Errorf("invalid env format %q (expected KEY=VALUE)", e)
				}
				req.Env[k] = v
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.Deploy(ctx, project, region, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deployed service %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Image, "image", "", "Container image")
	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&req.Memory, "memory", "", "Memory limit")
	cmd.Flags().StringVar(&req.CPU, "cpu", "", "CPU limit")
	cmd.Flags().Int32Var(&req.Port, "port", 8080, "Container port")
	cmd.Flags().StringSliceVar(&envSlice, "env", nil, "Environment variables (KEY=VALUE)")
	cmd.Flags().BoolVar(&req.AllowUnauthenticated, "allow-unauthenticated", false, "Allow unauthenticated access")
	_ = cmd.MarkFlagRequired("image")

	return cmd
}
