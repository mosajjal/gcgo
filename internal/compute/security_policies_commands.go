package compute

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func newSecurityPoliciesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "security-policies",
		Short: "Manage Cloud Armor security policies",
	}
	cmd.AddCommand(
		newSecurityPoliciesListCommand(cfg, creds),
		newSecurityPoliciesDescribeCommand(cfg, creds),
		newSecurityPoliciesCreateCommand(cfg, creds),
		newSecurityPoliciesDeleteCommand(cfg, creds),
		newSecurityPoliciesAddRuleCommand(cfg, creds),
		newSecurityPoliciesRemoveRuleCommand(cfg, creds),
	)
	return cmd
}

func newSecurityPoliciesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List security policies",
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
			policies, err := client.ListSecurityPolicies(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), policies)
			}
			headers := []string{"NAME", "DESCRIPTION", "RULES"}
			rows := make([][]string, len(policies))
			for i, p := range policies {
				rows[i] = []string{p.Name, p.Description, fmt.Sprintf("%d", p.Rules)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newSecurityPoliciesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a security policy",
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
			pol, err := client.GetSecurityPolicy(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), pol)
		},
	}
}

func newSecurityPoliciesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateSecurityPolicyRequest

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a security policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateSecurityPolicy(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created security policy %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Description, "description", "", "Policy description")
	return cmd
}

func newSecurityPoliciesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a security policy",
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
			if err := client.DeleteSecurityPolicy(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted security policy %q.\n", args[0])
			return nil
		},
	}
}

func newSecurityPoliciesAddRuleCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var rule SecurityPolicyRuleRequest

	cmd := &cobra.Command{
		Use:   "add-rule NAME",
		Short: "Add a rule to a security policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if rule.Action == "" {
				return fmt.Errorf("--action is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.AddSecurityPolicyRule(ctx, project, args[0], &rule); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added rule (priority %d) to security policy %q.\n", rule.Priority, args[0])
			return nil
		},
	}
	cmd.Flags().Int32Var(&rule.Priority, "priority", 1000, "Rule priority (lower = higher precedence)")
	cmd.Flags().StringVar(&rule.Action, "action", "", "Rule action: allow or deny(STATUS) e.g. deny(403)")
	cmd.Flags().StringVar(&rule.Description, "description", "", "Rule description")
	cmd.Flags().StringSliceVar(&rule.SrcIPRanges, "src-ip-ranges", nil, "Source IP ranges (CIDR)")
	cmd.Flags().BoolVar(&rule.Preview, "preview", false, "Enable preview mode for the rule")
	return cmd
}

func newSecurityPoliciesRemoveRuleCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var priority int32

	cmd := &cobra.Command{
		Use:   "remove-rule NAME",
		Short: "Remove a rule from a security policy",
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
			if err := client.RemoveSecurityPolicyRule(ctx, project, args[0], priority); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed rule (priority %d) from security policy %q.\n", priority, args[0])
			return nil
		},
	}
	cmd.Flags().Int32Var(&priority, "priority", 0, "Priority of the rule to remove")
	_ = cmd.MarkFlagRequired("priority")
	return cmd
}
