package iam

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func rolesClient(ctx context.Context, creds *auth.Credentials) (RolesClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewRolesClient(ctx, opt)
}

func workloadIdentityClient(ctx context.Context, creds *auth.Credentials) (WorkloadIdentityClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewWorkloadIdentityClient(ctx, opt)
}

func newRolesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "roles",
		Short: "Manage custom IAM roles",
	}
	cmd.AddCommand(
		newRolesListCommand(cfg, creds),
		newRolesDescribeCommand(cfg, creds),
		newRolesCreateCommand(cfg, creds),
		newRolesUpdateCommand(cfg, creds),
		newRolesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newRolesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List custom IAM roles in the current project",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := rolesClient(ctx, creds)
			if err != nil {
				return err
			}
			roles, err := client.ListRoles(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), roles)
			}
			headers := []string{"NAME", "TITLE", "STAGE", "DELETED"}
			rows := make([][]string, len(roles))
			for i, role := range roles {
				rows[i] = []string{role.Name, role.Title, role.Stage, fmt.Sprintf("%t", role.Deleted)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newRolesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ROLE",
		Short: "Describe a custom IAM role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := rolesClient(ctx, creds)
			if err != nil {
				return err
			}
			name := args[0]
			if name == args[0] && len(args[0]) > 0 && args[0][:9] != "projects/" {
				name = fmt.Sprintf("projects/%s/roles/%s", project, args[0])
			}
			role, err := client.DescribeRole(ctx, name)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), role)
		},
	}
}

func newRolesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var title string
	var description string
	var permissions []string

	cmd := &cobra.Command{
		Use:   "create ROLE_ID",
		Short: "Create a custom IAM role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := rolesClient(ctx, creds)
			if err != nil {
				return err
			}
			role, err := client.CreateRole(ctx, project, args[0], title, description, permissions)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), role)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Role title")
	cmd.Flags().StringVar(&description, "description", "", "Role description")
	cmd.Flags().StringSliceVar(&permissions, "permission", nil, "Included permission")
	return cmd
}

func newRolesUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var title string
	var description string
	var permissions []string

	cmd := &cobra.Command{
		Use:   "update ROLE",
		Short: "Update a custom IAM role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := rolesClient(ctx, creds)
			if err != nil {
				return err
			}
			name := args[0]
			if len(args[0]) > 0 && args[0][:9] != "projects/" {
				name = fmt.Sprintf("projects/%s/roles/%s", project, args[0])
			}
			role, err := client.UpdateRole(ctx, name, title, description, permissions)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), role)
		},
	}
	cmd.Flags().StringVar(&title, "title", "", "Role title")
	cmd.Flags().StringVar(&description, "description", "", "Role description")
	cmd.Flags().StringSliceVar(&permissions, "permission", nil, "Included permission")
	return cmd
}

func newRolesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete ROLE",
		Short: "Delete a custom IAM role",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := rolesClient(ctx, creds)
			if err != nil {
				return err
			}
			name := args[0]
			if len(args[0]) > 0 && args[0][:9] != "projects/" {
				name = fmt.Sprintf("projects/%s/roles/%s", project, args[0])
			}
			if err := client.DeleteRole(ctx, name); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted role %s.\n", name)
			return nil
		},
	}
}

func newWorkloadIdentityCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "workload-identity",
		Short: "Manage workload identity federation pools and providers",
	}
	cmd.AddCommand(
		newWorkloadIdentityPoolsCommand(cfg, creds),
		newWorkloadIdentityProvidersCommand(cfg, creds),
	)
	return cmd
}

func newWorkloadIdentityPoolsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pools",
		Short: "Manage workload identity pools",
	}
	cmd.AddCommand(
		newPoolsListCommand(cfg, creds),
		newPoolsDescribeCommand(cfg, creds),
		newPoolsCreateCommand(cfg, creds),
		newPoolsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newPoolsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workload identity pools",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			pools, err := client.ListPools(ctx, project, location)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), pools)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "DISABLED"}
			rows := make([][]string, len(pools))
			for i, pool := range pools {
				rows[i] = []string{pool.Name, pool.DisplayName, pool.State, fmt.Sprintf("%t", pool.Disabled)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	return cmd
}

func newPoolsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "describe POOL",
		Short: "Describe a workload identity pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			name := args[0]
			if len(args[0]) > 0 && args[0][:9] != "projects/" {
				name = fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, args[0])
			}
			pool, err := client.DescribePool(ctx, name)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), pool)
		},
	}
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	return cmd
}

func newPoolsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var displayName string
	var description string

	cmd := &cobra.Command{
		Use:   "create POOL_ID",
		Short: "Create a workload identity pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			pool, err := client.CreatePool(ctx, project, location, args[0], displayName, description)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), pool)
		},
	}
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Pool display name")
	cmd.Flags().StringVar(&description, "description", "", "Pool description")
	return cmd
}

func newPoolsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "delete POOL",
		Short: "Delete a workload identity pool",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			name := args[0]
			if len(args[0]) > 0 && args[0][:9] != "projects/" {
				name = fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, args[0])
			}
			if err := client.DeletePool(ctx, name); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted workload identity pool %s.\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	return cmd
}

func newWorkloadIdentityProvidersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "providers",
		Short: "Manage workload identity providers",
	}
	cmd.AddCommand(
		newProvidersListCommand(cfg, creds),
		newProvidersDescribeCommand(cfg, creds),
		newProvidersCreateCommand(cfg, creds),
		newProvidersDeleteCommand(cfg, creds),
	)
	return cmd
}

func newProvidersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var pool string
	var location string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List workload identity providers in a pool",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if pool == "" {
				return fmt.Errorf("--pool is required")
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			poolName := resolvePoolName(project, location, pool)
			providers, err := client.ListProviders(ctx, poolName)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), providers)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "ISSUER"}
			rows := make([][]string, len(providers))
			for i, provider := range providers {
				rows[i] = []string{provider.Name, provider.DisplayName, provider.State, provider.Issuer}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&pool, "pool", "", "Workload identity pool name or resource")
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	return cmd
}

func newProvidersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var pool string
	var location string

	cmd := &cobra.Command{
		Use:   "describe PROVIDER",
		Short: "Describe a workload identity provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if pool == "" {
				return fmt.Errorf("--pool is required")
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			name := args[0]
			if len(args[0]) > 0 && args[0][:9] != "projects/" {
				name = fmt.Sprintf("%s/providers/%s", resolvePoolName(project, location, pool), args[0])
			}
			provider, err := client.DescribeProvider(ctx, name)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), provider)
		},
	}
	cmd.Flags().StringVar(&pool, "pool", "", "Workload identity pool name or resource")
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	return cmd
}

func newProvidersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var pool string
	var location string
	var displayName string
	var issuerURI string

	cmd := &cobra.Command{
		Use:   "create PROVIDER_ID",
		Short: "Create an OIDC workload identity provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if pool == "" {
				return fmt.Errorf("--pool is required")
			}
			if issuerURI == "" {
				return fmt.Errorf("--issuer-uri is required")
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			provider, err := client.CreateProvider(ctx, resolvePoolName(project, location, pool), args[0], displayName, issuerURI)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), provider)
		},
	}
	cmd.Flags().StringVar(&pool, "pool", "", "Workload identity pool name or resource")
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Provider display name")
	cmd.Flags().StringVar(&issuerURI, "issuer-uri", "", "OIDC issuer URI")
	return cmd
}

func newProvidersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var pool string
	var location string

	cmd := &cobra.Command{
		Use:   "delete PROVIDER",
		Short: "Delete a workload identity provider",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if pool == "" {
				return fmt.Errorf("--pool is required")
			}
			if location == "" {
				location = "global"
			}
			ctx := context.Background()
			client, err := workloadIdentityClient(ctx, creds)
			if err != nil {
				return err
			}
			name := args[0]
			if len(args[0]) > 0 && args[0][:9] != "projects/" {
				name = fmt.Sprintf("%s/providers/%s", resolvePoolName(project, location, pool), args[0])
			}
			if err := client.DeleteProvider(ctx, name); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted workload identity provider %s.\n", name)
			return nil
		},
	}
	cmd.Flags().StringVar(&pool, "pool", "", "Workload identity pool name or resource")
	cmd.Flags().StringVar(&location, "location", "global", "Pool location")
	return cmd
}

func resolvePoolName(project, location, pool string) string {
	if len(pool) > 0 && pool[:9] == "projects/" {
		return pool
	}
	return fmt.Sprintf("projects/%s/locations/%s/workloadIdentityPools/%s", project, location, pool)
}
