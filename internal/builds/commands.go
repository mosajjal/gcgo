package builds

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the builds command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "builds",
		Short: "Manage Cloud Build resources",
	}

	cmd.AddCommand(
		newListCommand(cfg, creds),
		newDescribeCommand(cfg, creds),
		newCancelCommand(cfg, creds),
		newTriggersCommand(cfg, creds),
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

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List builds",
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

			builds, err := client.ListBuilds(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), builds)
			}

			headers := []string{"ID", "STATUS", "SOURCE", "CREATE_TIME"}
			rows := make([][]string, len(builds))
			for i, b := range builds {
				rows[i] = []string{b.ID, b.Status, b.Source, b.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe BUILD_ID",
		Short: "Describe a build",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			build, err := client.GetBuild(ctx, project, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), build)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", build.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:      %s\n", build.Status)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Source:      %s\n", build.Source)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create Time: %s\n", build.CreateTime)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Log URL:     %s\n", build.LogURL)
			return nil
		},
	}
}

func newCancelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel BUILD_ID",
		Short: "Cancel a running build",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.CancelBuild(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cancelled build %s.\n", args[0])
			return nil
		},
	}
}

// Triggers subcommands

func newTriggersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "triggers",
		Short: "Manage build triggers",
	}

	cmd.AddCommand(
		newTriggersListCommand(cfg, creds),
		newTriggersDescribeCommand(cfg, creds),
		newTriggersCreateCommand(cfg, creds),
		newTriggersDeleteCommand(cfg, creds),
		newTriggersRunCommand(cfg, creds),
	)
	return cmd
}

func newTriggersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List build triggers",
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

			triggers, err := client.ListTriggers(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), triggers)
			}

			headers := []string{"ID", "NAME", "DISABLED", "CREATE_TIME"}
			rows := make([][]string, len(triggers))
			for i, t := range triggers {
				rows[i] = []string{t.ID, t.Name, fmt.Sprintf("%v", t.Disabled), t.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newTriggersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe TRIGGER_ID",
		Short: "Describe a build trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			trigger, err := client.GetTrigger(ctx, project, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), trigger)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:          %s\n", trigger.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", trigger.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", trigger.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Disabled:    %v\n", trigger.Disabled)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create Time: %s\n", trigger.CreateTime)
			return nil
		},
	}
}

func newTriggersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateTriggerRequest

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a build trigger",
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

			trigger, err := client.CreateTrigger(ctx, project, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created trigger %s (%s).\n", trigger.Name, trigger.ID)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Name, "name", "", "Trigger name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Trigger description")
	cmd.Flags().StringVar(&req.RepoName, "repo-name", "", "Repository name")
	cmd.Flags().StringVar(&req.BranchName, "branch-name", "", "Branch pattern")
	cmd.Flags().StringVar(&req.Filename, "filename", "cloudbuild.yaml", "Build config filename")
	_ = cmd.MarkFlagRequired("name")
	return cmd
}

func newTriggersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete TRIGGER_ID",
		Short: "Delete a build trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteTrigger(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted trigger %s.\n", args[0])
			return nil
		},
	}
}

func newTriggersRunCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "run TRIGGER_ID",
		Short: "Run a build trigger",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.RunTrigger(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Triggered build for trigger %s.\n", args[0])
			return nil
		},
	}
}
