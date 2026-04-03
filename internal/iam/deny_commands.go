package iam

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func denyClient(ctx context.Context, creds *auth.Credentials) (DenyPoliciesClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewDenyPoliciesClient(ctx, opt)
}

func newDenyPoliciesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deny-policies",
		Short: "Manage IAM deny policies",
	}
	cmd.AddCommand(
		newDenyListCommand(cfg, creds),
		newDenyDescribeCommand(cfg, creds),
		newDenyCreateCommand(cfg, creds),
		newDenyDeleteCommand(cfg, creds),
	)
	return cmd
}

func newDenyListCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list ATTACHMENT_POINT",
		Short: "List deny policies for an attachment point",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := denyClient(ctx, creds)
			if err != nil {
				return err
			}
			policies, err := client.ListDenyPolicies(ctx, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), policies)
			}
			headers := []string{"NAME", "DISPLAY_NAME"}
			rows := make([][]string, len(policies))
			for i, p := range policies {
				rows[i] = []string{p.Name, p.DisplayName}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDenyDescribeCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe POLICY_NAME",
		Short: "Describe a deny policy by name",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := denyClient(ctx, creds)
			if err != nil {
				return err
			}
			policy, err := client.DescribeDenyPolicy(ctx, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}
}

func newDenyCreateCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	var attachmentPoint string
	var displayName string
	var rulesFile string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a deny policy",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if attachmentPoint == "" {
				return fmt.Errorf("--attachment-point is required")
			}
			ctx := context.Background()
			client, err := denyClient(ctx, creds)
			if err != nil {
				return err
			}
			policy, err := client.CreateDenyPolicy(ctx, attachmentPoint, displayName, rulesFile)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created deny policy %s.\n", policy.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&attachmentPoint, "attachment-point", "", "Attachment point (e.g. cloudresourcemanager.googleapis.com/projects/PROJECT_NUMBER)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name for the policy")
	cmd.Flags().StringVar(&rulesFile, "rules", "", "Path to JSON file containing deny rules")

	return cmd
}

func newDenyDeleteCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete POLICY_NAME",
		Short: "Delete a deny policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := denyClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteDenyPolicy(ctx, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted deny policy %s.\n", args[0])
			return nil
		},
	}
}
