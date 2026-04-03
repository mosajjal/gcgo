package iam

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the iam command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage IAM resources",
	}

	cmd.AddCommand(
		newServiceAccountsCommand(cfg, creds),
		newPolicyCommand(cfg, creds),
		newDenyPoliciesCommand(cfg, creds),
		newOrgPoliciesCommand(cfg, creds),
		newRolesCommand(cfg, creds),
		newFoldersCommand(cfg, creds),
		newOrganizationsCommand(cfg, creds),
		newWorkloadIdentityCommand(cfg, creds),
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

// Service accounts

func newServiceAccountsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "service-accounts",
		Short: "Manage service accounts",
	}

	cmd.AddCommand(
		newSAListCommand(cfg, creds),
		newSACreateCommand(cfg, creds),
		newSADeleteCommand(cfg, creds),
		newKeysCommand(cfg, creds),
	)

	return cmd
}

func newSAListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List service accounts",
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

			accounts, err := client.ListServiceAccounts(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), accounts)
			}

			headers := []string{"EMAIL", "DISPLAY_NAME", "DISABLED"}
			rows := make([][]string, len(accounts))
			for i, sa := range accounts {
				rows[i] = []string{sa.Email, sa.DisplayName, fmt.Sprintf("%v", sa.Disabled)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newSACreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var displayName string

	cmd := &cobra.Command{
		Use:   "create ACCOUNT_ID",
		Short: "Create a service account",
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

			sa, err := client.CreateServiceAccount(ctx, project, args[0], displayName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created service account %s.\n", sa.Email)
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name")

	return cmd
}

func newSADeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete EMAIL",
		Short: "Delete a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteServiceAccount(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted service account %s.\n", args[0])
			return nil
		},
	}
}

// Keys

func newKeysCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage service account keys",
	}

	cmd.AddCommand(
		newKeysListCommand(cfg, creds),
		newKeysCreateCommand(cfg, creds),
		newKeysDeleteCommand(cfg, creds),
	)

	return cmd
}

func newKeysListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list EMAIL",
		Short: "List keys for a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keys, err := client.ListKeys(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), keys)
			}

			headers := []string{"NAME", "KEY_ALGORITHM", "VALID_AFTER", "VALID_BEFORE"}
			rows := make([][]string, len(keys))
			for i, k := range keys {
				rows[i] = []string{k.Name, k.KeyAlgorithm, k.ValidAfter, k.ValidBefore}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newKeysCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var outFile string

	cmd := &cobra.Command{
		Use:   "create EMAIL",
		Short: "Create a key for a service account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			data, err := client.CreateKey(ctx, args[0])
			if err != nil {
				return err
			}

			if outFile != "" {
				if err := os.WriteFile(outFile, data, 0o600); err != nil {
					return fmt.Errorf("write key file: %w", err)
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Key written to %s.\n", outFile)
			} else {
				_, _ = fmt.Fprint(cmd.OutOrStdout(), string(data))
			}
			return nil
		},
	}

	cmd.Flags().StringVar(&outFile, "output-file", "", "Write key to file")

	return cmd
}

func newKeysDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var iamAccount string

	cmd := &cobra.Command{
		Use:   "delete KEY_ID",
		Short: "Delete a service account key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keyName := args[0]
			if iamAccount != "" && !strings.Contains(keyName, "/") {
				keyName = fmt.Sprintf("projects/-/serviceAccounts/%s/keys/%s", iamAccount, keyName)
			}

			if err := client.DeleteKey(ctx, keyName); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted key.\n")
			return nil
		},
	}

	cmd.Flags().StringVar(&iamAccount, "iam-account", "", "Service account email")

	return cmd
}

// Policy

func newPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: "Manage project IAM policy",
	}

	cmd.AddCommand(
		newPolicyGetCommand(cfg, creds),
		newPolicyAddBindingCommand(cfg, creds),
		newPolicyRemoveBindingCommand(cfg, creds),
		newProjectPolicyTroubleshootCommand(cfg, creds),
	)

	return cmd
}

func newPolicyGetCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "get",
		Short: "Get project IAM policy",
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

			bindings, err := client.GetPolicy(ctx, project)
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

func newPolicyAddBindingCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var member, role string

	cmd := &cobra.Command{
		Use:   "add-binding",
		Short: "Add an IAM policy binding",
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

			if err := client.AddBinding(ctx, project, member, role); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Added %s with role %s.\n", member, role)
			return nil
		},
	}

	cmd.Flags().StringVar(&member, "member", "", "Member (e.g. user:foo@bar.com)")
	cmd.Flags().StringVar(&role, "role", "", "Role (e.g. roles/viewer)")
	_ = cmd.MarkFlagRequired("member")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}

func newPolicyRemoveBindingCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var member, role string

	cmd := &cobra.Command{
		Use:   "remove-binding",
		Short: "Remove an IAM policy binding",
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

			if err := client.RemoveBinding(ctx, project, member, role); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed %s from role %s.\n", member, role)
			return nil
		},
	}

	cmd.Flags().StringVar(&member, "member", "", "Member (e.g. user:foo@bar.com)")
	cmd.Flags().StringVar(&role, "role", "", "Role (e.g. roles/viewer)")
	_ = cmd.MarkFlagRequired("member")
	_ = cmd.MarkFlagRequired("role")

	return cmd
}
