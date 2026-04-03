package workflows

import (
	"context"
	"fmt"
	"os"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the workflows command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workflows",
		Short: "Manage Workflows",
	}

	cmd.AddCommand(
		newListCommand(cfg, creds),
		newDescribeCommand(cfg, creds),
		newDeployCommand(cfg, creds),
		newDeleteCommand(cfg, creds),
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

func requireLocation(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("location")
	if flagVal != "" {
		return flagVal, nil
	}
	location := cfg.Region()
	if location == "" {
		return "", fmt.Errorf("--location is required (or set region in config)")
	}
	return location, nil
}

func workflowsClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workflows",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := workflowsClient(ctx, creds)
			if err != nil {
				return err
			}
			workflows, err := client.List(ctx, project, location)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), workflows)
			}
			headers := []string{"NAME", "STATE", "REVISION", "UPDATED"}
			rows := make([][]string, len(workflows))
			for i, workflow := range workflows {
				rows[i] = []string{workflow.Name, workflow.State, workflow.RevisionID, workflow.UpdateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Workflow location")
	return cmd
}

func newDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "describe WORKFLOW",
		Short: "Describe a workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := workflowsClient(ctx, creds)
			if err != nil {
				return err
			}
			workflow, err := client.Get(ctx, project, location, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), workflow)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:           %s\n", workflow.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description:    %s\n", workflow.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:          %s\n", workflow.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Revision:       %s\n", workflow.RevisionID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ServiceAccount: %s\n", workflow.ServiceAccount)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Workflow location")
	return cmd
}

func newDeployCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var description string
	var sourceFile string
	var serviceAccount string

	cmd := &cobra.Command{
		Use:   "deploy WORKFLOW --source-file=FILE",
		Short: "Create or update a workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if sourceFile == "" {
				return fmt.Errorf("--source-file is required")
			}
			source, err := os.ReadFile(sourceFile)
			if err != nil {
				return fmt.Errorf("read source file %s: %w", sourceFile, err)
			}
			ctx := context.Background()
			client, err := workflowsClient(ctx, creds)
			if err != nil {
				return err
			}
			workflow, err := client.Deploy(ctx, project, location, &DeployRequest{
				Name:           args[0],
				Description:    description,
				SourceContents: string(source),
				ServiceAccount: serviceAccount,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deployed workflow %s.\n", workflow.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Workflow location")
	cmd.Flags().StringVar(&description, "description", "", "Workflow description")
	cmd.Flags().StringVar(&sourceFile, "source-file", "", "Workflow source YAML file")
	cmd.Flags().StringVar(&serviceAccount, "service-account", "", "Workflow service account")
	return cmd
}

func newDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "delete WORKFLOW",
		Short: "Delete a workflow",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := workflowsClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.Delete(ctx, project, location, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted workflow %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Workflow location")
	return cmd
}
