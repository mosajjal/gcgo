package projects

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the projects command group.
func NewCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage GCP projects",
	}

	cmd.AddCommand(
		newListCommand(creds),
		newDescribeCommand(creds),
	)

	return cmd
}

func newListCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List accessible projects",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}

			client, err := NewClient(ctx, opt)
			if err != nil {
				return err
			}

			projects, err := client.List(ctx)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), projects)
			}

			headers := []string{"PROJECT_ID", "NAME", "STATE"}
			rows := make([][]string, len(projects))
			for i, p := range projects {
				rows[i] = []string{p.ID, p.Name, p.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe PROJECT_ID",
		Short: "Describe a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}

			client, err := NewClient(ctx, opt)
			if err != nil {
				return err
			}

			project, err := client.Get(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), project)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project ID:     %s\n", project.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:           %s\n", project.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project Number: %s\n", project.Number)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:          %s\n", project.State)
			return nil
		},
	}
}
