package iam

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

type policyTestResult struct {
	Resource             string        `json:"resource"`
	Kind                 string        `json:"kind"`
	RequestedPermissions []string      `json:"requested_permissions,omitempty"`
	GrantedPermissions   []string      `json:"granted_permissions,omitempty"`
	MatchedMember        string        `json:"matched_member,omitempty"`
	MatchedBindings      []*IAMBinding `json:"matched_bindings,omitempty"`
	AllBindings          []*IAMBinding `json:"all_bindings,omitempty"`
}

func newFoldersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "Manage folder IAM policy",
	}
	cmd.AddCommand(newFolderPolicyCommand(cfg, creds))
	return cmd
}

func newOrganizationsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "organizations",
		Short: "Manage organization IAM policy",
	}
	cmd.AddCommand(newOrganizationPolicyCommand(cfg, creds))
	return cmd
}

func newFolderPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return newScopedPolicyCommand("folder", cfg, creds, folderPolicyOps{})
}

func newOrganizationPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return newScopedPolicyCommand("organization", cfg, creds, organizationPolicyOps{})
}

func newScopedPolicyCommand(noun string, cfg *config.Config, creds *auth.Credentials, ops scopedPolicyOps) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policy",
		Short: fmt.Sprintf("Manage %s IAM policy", noun),
	}
	cmd.AddCommand(
		newScopedPolicyGetCommand(noun, cfg, creds, ops),
		newScopedPolicyAddBindingCommand(noun, cfg, creds, ops),
		newScopedPolicyRemoveBindingCommand(noun, cfg, creds, ops),
		newScopedPolicyTroubleshootCommand(noun, cfg, creds, ops),
	)
	return cmd
}

func newProjectPolicyTroubleshootCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var member string
	var permissions []string

	cmd := &cobra.Command{
		Use:   "troubleshoot",
		Short: "Troubleshoot project IAM access",
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
			return printTroubleshootResult(cmd, projectPolicyOps{}, client.(*gcpClient), ctx, "project", project, member, permissions)
		},
	}
	cmd.Flags().StringVar(&member, "member", "", "Member to filter bindings for")
	cmd.Flags().StringSliceVar(&permissions, "permission", nil, "Permission to test (repeatable)")
	return cmd
}

type scopedPolicyOps interface {
	get(ctx context.Context, client *gcpClient, resource string) ([]*IAMBinding, error)
	add(ctx context.Context, client *gcpClient, resource, member, role string) error
	remove(ctx context.Context, client *gcpClient, resource, member, role string) error
	test(ctx context.Context, client *gcpClient, resource string, permissions []string) ([]string, error)
}

type projectPolicyOps struct{}

func (projectPolicyOps) get(ctx context.Context, client *gcpClient, resource string) ([]*IAMBinding, error) {
	return client.GetPolicy(ctx, resource)
}

func (projectPolicyOps) add(ctx context.Context, client *gcpClient, resource, member, role string) error {
	return client.AddBinding(ctx, resource, member, role)
}

func (projectPolicyOps) remove(ctx context.Context, client *gcpClient, resource, member, role string) error {
	return client.RemoveBinding(ctx, resource, member, role)
}

func (projectPolicyOps) test(ctx context.Context, client *gcpClient, resource string, permissions []string) ([]string, error) {
	return client.TestProjectPermissions(ctx, resource, permissions)
}

type folderPolicyOps struct{}

func (folderPolicyOps) get(ctx context.Context, client *gcpClient, resource string) ([]*IAMBinding, error) {
	return client.GetFolderPolicy(ctx, resource)
}

func (folderPolicyOps) add(ctx context.Context, client *gcpClient, resource, member, role string) error {
	return client.AddFolderBinding(ctx, resource, member, role)
}

func (folderPolicyOps) remove(ctx context.Context, client *gcpClient, resource, member, role string) error {
	return client.RemoveFolderBinding(ctx, resource, member, role)
}

func (folderPolicyOps) test(ctx context.Context, client *gcpClient, resource string, permissions []string) ([]string, error) {
	return client.TestFolderPermissions(ctx, resource, permissions)
}

type organizationPolicyOps struct{}

func (organizationPolicyOps) get(ctx context.Context, client *gcpClient, resource string) ([]*IAMBinding, error) {
	return client.GetOrganizationPolicy(ctx, resource)
}

func (organizationPolicyOps) add(ctx context.Context, client *gcpClient, resource, member, role string) error {
	return client.AddOrganizationBinding(ctx, resource, member, role)
}

func (organizationPolicyOps) remove(ctx context.Context, client *gcpClient, resource, member, role string) error {
	return client.RemoveOrganizationBinding(ctx, resource, member, role)
}

func (organizationPolicyOps) test(ctx context.Context, client *gcpClient, resource string, permissions []string) ([]string, error) {
	return client.TestOrganizationPermissions(ctx, resource, permissions)
}

func newScopedPolicyGetCommand(noun string, cfg *config.Config, creds *auth.Credentials, ops scopedPolicyOps) *cobra.Command {
	return &cobra.Command{
		Use:   fmt.Sprintf("get %s", strings.ToUpper(noun)),
		Short: fmt.Sprintf("Get %s IAM policy", noun),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			resource := args[0]
			if noun == "project" {
				resource, err = requireProject(cmd, cfg)
				if err != nil {
					return err
				}
			}
			return printPolicyBindings(cmd, ops, client.(*gcpClient), ctx, resource)
		},
	}
}

func newScopedPolicyAddBindingCommand(noun string, cfg *config.Config, creds *auth.Credentials, ops scopedPolicyOps) *cobra.Command {
	var member, role string

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("add-binding %s", strings.ToUpper(noun)),
		Short: fmt.Sprintf("Add a %s IAM policy binding", noun),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			resource := args[0]
			if noun == "project" {
				resource, err = requireProject(cmd, cfg)
				if err != nil {
					return err
				}
			}
			if err := ops.add(ctx, client.(*gcpClient), resource, member, role); err != nil {
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

func newScopedPolicyRemoveBindingCommand(noun string, cfg *config.Config, creds *auth.Credentials, ops scopedPolicyOps) *cobra.Command {
	var member, role string

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("remove-binding %s", strings.ToUpper(noun)),
		Short: fmt.Sprintf("Remove a %s IAM policy binding", noun),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			resource := args[0]
			if noun == "project" {
				resource, err = requireProject(cmd, cfg)
				if err != nil {
					return err
				}
			}
			if err := ops.remove(ctx, client.(*gcpClient), resource, member, role); err != nil {
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

func newScopedPolicyTroubleshootCommand(noun string, cfg *config.Config, creds *auth.Credentials, ops scopedPolicyOps) *cobra.Command {
	var member string
	var permissions []string

	cmd := &cobra.Command{
		Use:   fmt.Sprintf("troubleshoot %s", strings.ToUpper(noun)),
		Short: fmt.Sprintf("Troubleshoot %s IAM access", noun),
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			resource := args[0]
			if noun == "project" {
				resource, err = requireProject(cmd, cfg)
				if err != nil {
					return err
				}
			}
			return printTroubleshootResult(cmd, ops, client.(*gcpClient), ctx, noun, resource, member, permissions)
		},
	}
	cmd.Flags().StringVar(&member, "member", "", "Member to filter bindings for")
	cmd.Flags().StringSliceVar(&permissions, "permission", nil, "Permission to test (repeatable)")
	return cmd
}

func printPolicyBindings(cmd *cobra.Command, ops scopedPolicyOps, client *gcpClient, ctx context.Context, resource string) error {
	bindings, err := ops.get(ctx, client, resource)
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
}

func printTroubleshootResult(cmd *cobra.Command, ops scopedPolicyOps, client *gcpClient, ctx context.Context, kind, resource, member string, permissions []string) error {
	bindings, err := ops.get(ctx, client, resource)
	if err != nil {
		return err
	}
	granted := permissions
	if len(permissions) > 0 {
		granted, err = ops.test(ctx, client, resource, permissions)
		if err != nil {
			return err
		}
	}
	result := policyTestResult{
		Resource:             resource,
		Kind:                 kind,
		RequestedPermissions: permissions,
		GrantedPermissions:   granted,
		AllBindings:          bindings,
		MatchedMember:        member,
	}
	if member != "" {
		result.MatchedBindings = filterBindingsByMember(bindings, member)
	}
	format, _ := cmd.Flags().GetString("format")
	if format == "json" {
		return output.PrintJSON(cmd.OutOrStdout(), result)
	}

	if len(result.RequestedPermissions) > 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Granted permissions for %s:\n", resource)
		for _, permission := range result.GrantedPermissions {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "- %s\n", permission)
		}
	}
	if member != "" {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "\nBindings for %s:\n", member)
		for _, binding := range result.MatchedBindings {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "- %s: %s\n", binding.Role, strings.Join(binding.Members, ", "))
		}
	}
	if len(bindings) > 0 && member == "" && len(permissions) == 0 {
		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Bindings for %s:\n", resource)
		for _, binding := range bindings {
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "- %s: %s\n", binding.Role, strings.Join(binding.Members, ", "))
		}
	}
	return nil
}

func filterBindingsByMember(bindings []*IAMBinding, member string) []*IAMBinding {
	var filtered []*IAMBinding
	for _, binding := range bindings {
		var members []string
		for _, m := range binding.Members {
			if m == member {
				members = append(members, m)
			}
		}
		if len(members) > 0 {
			filtered = append(filtered, &IAMBinding{Role: binding.Role, Members: members})
		}
	}
	return filtered
}
