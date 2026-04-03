package iam

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func orgPolicyClientHelper(ctx context.Context, creds *auth.Credentials) (OrgPoliciesClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewOrgPoliciesClient(ctx, opt)
}

func newOrgPoliciesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "org-policies",
		Short: "Manage organization policies",
	}
	cmd.AddCommand(
		newOrgPolicyListCommand(cfg, creds),
		newOrgPolicyDescribeCommand(cfg, creds),
		newOrgPolicySetCommand(cfg, creds),
		newOrgPolicyResetCommand(cfg, creds),
	)
	return cmd
}

func newOrgPolicyListCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list RESOURCE",
		Short: "List org policies for a resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := orgPolicyClientHelper(ctx, creds)
			if err != nil {
				return err
			}
			policies, err := client.ListOrgPolicies(ctx, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), policies)
			}
			headers := []string{"NAME", "CONSTRAINT"}
			rows := make([][]string, len(policies))
			for i, p := range policies {
				rows[i] = []string{p.Name, p.Constraint}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newOrgPolicyDescribeCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	var resource string

	cmd := &cobra.Command{
		Use:   "describe CONSTRAINT",
		Short: "Describe an org policy for a constraint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resource == "" {
				return fmt.Errorf("--resource is required")
			}
			ctx := context.Background()
			client, err := orgPolicyClientHelper(ctx, creds)
			if err != nil {
				return err
			}
			policy, err := client.DescribeOrgPolicy(ctx, resource, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}

	cmd.Flags().StringVar(&resource, "resource", "", "Resource (e.g. projects/PROJECT_ID or organizations/ORG_ID)")

	return cmd
}

func newOrgPolicySetCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	var resource string
	var policyFile string

	cmd := &cobra.Command{
		Use:   "set CONSTRAINT",
		Short: "Set an org policy from a JSON file",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resource == "" {
				return fmt.Errorf("--resource is required")
			}
			if policyFile == "" {
				return fmt.Errorf("--policy-file is required")
			}
			ctx := context.Background()
			client, err := orgPolicyClientHelper(ctx, creds)
			if err != nil {
				return err
			}
			policy, err := client.SetOrgPolicy(ctx, resource, args[0], policyFile)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Set org policy %s.\n", policy.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&resource, "resource", "", "Resource (e.g. projects/PROJECT_ID or organizations/ORG_ID)")
	cmd.Flags().StringVar(&policyFile, "policy-file", "", "Path to JSON file containing the policy spec")

	return cmd
}

func newOrgPolicyResetCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	var resource string

	cmd := &cobra.Command{
		Use:   "reset CONSTRAINT",
		Short: "Reset an org policy to its default",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if resource == "" {
				return fmt.Errorf("--resource is required")
			}
			ctx := context.Background()
			client, err := orgPolicyClientHelper(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.ResetOrgPolicy(ctx, resource, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Reset org policy %s on %s.\n", args[0], resource)
			return nil
		},
	}

	cmd.Flags().StringVar(&resource, "resource", "", "Resource (e.g. projects/PROJECT_ID or organizations/ORG_ID)")

	return cmd
}
