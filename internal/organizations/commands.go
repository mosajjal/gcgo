package organizations

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the organizations command group.
func NewCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Manage GCP organizations",
	}

	cmd.AddCommand(
		newListCommand(creds),
		newDescribeCommand(creds),
		newGetIAMPolicyCommand(creds),
	)

	return cmd
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newListCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List accessible organizations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			orgs, err := client.List(ctx)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), orgs)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "STATE"}
			rows := make([][]string, len(orgs))
			for i, o := range orgs {
				rows[i] = []string{o.Name, o.DisplayName, o.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ORG_ID",
		Short: "Describe an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			org, err := client.Get(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), org)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:           %s\n", org.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name:   %s\n", org.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:          %s\n", org.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Directory ID:   %s\n", org.DirectoryID)
			return nil
		},
	}
}

func newGetIAMPolicyCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "get-iam-policy ORG_ID",
		Short: "Get IAM policy for an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			bindings, err := client.GetIAMPolicy(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), bindings)
			}

			headers := []string{"ROLE", "MEMBERS"}
			rows := make([][]string, len(bindings))
			for i, b := range bindings {
				rows[i] = []string{b.Role, strings.Join(b.Members, ", ")}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}
