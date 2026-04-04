package projects

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

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
		newCreateCommand(creds),
		newDeleteCommand(creds),
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

func newCreateCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create PROJECT_ID",
		Short: "Create a new project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			name, _ := cmd.Flags().GetString("name")
			if name == "" {
				name = projectID
			}
			labels, _ := cmd.Flags().GetStringToString("labels")

			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}

			client, err := NewClient(ctx, opt)
			if err != nil {
				return err
			}

			if err := client.CreateProject(ctx, projectID, name, labels); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created project %s\n", projectID)
			return nil
		},
	}
	cmd.Flags().String("name", "", "Display name for the project (defaults to project ID)")
	cmd.Flags().StringToString("labels", nil, "Labels to attach (key=value,...)")
	return cmd
}

func newDeleteCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete PROJECT_ID",
		Short: "Delete a project",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			projectID := args[0]
			quiet, _ := cmd.Flags().GetBool("quiet")

			if !quiet {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Delete project %s? (y/N): ", projectID)
				scanner := bufio.NewScanner(os.Stdin)
				scanner.Scan()
				if strings.TrimSpace(scanner.Text()) != "y" {
					return fmt.Errorf("aborted")
				}
			}

			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}

			client, err := NewClient(ctx, opt)
			if err != nil {
				return err
			}

			if err := client.DeleteProject(ctx, projectID); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted project %s\n", projectID)
			return nil
		},
	}
	cmd.Flags().Bool("quiet", false, "Skip confirmation prompt")
	return cmd
}
