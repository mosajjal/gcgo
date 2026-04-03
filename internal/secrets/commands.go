package secrets

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
	secretmanager "google.golang.org/api/secretmanager/v1"
)

// NewCommand returns the secrets command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "secrets",
		Short: "Manage Secret Manager secrets",
	}

	cmd.AddCommand(
		newListCommand(cfg, creds),
		newCreateCommand(cfg, creds),
		newDeleteCommand(cfg, creds),
		newDescribeCommand(cfg, creds),
		newUpdateCommand(cfg, creds),
		newIAMCommand(cfg, creds),
		newVersionsCommand(cfg, creds),
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
		Short: "List secrets",
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

			secrets, err := client.List(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), secrets)
			}

			headers := []string{"NAME", "CREATE_TIME"}
			rows := make([][]string, len(secrets))
			for i, s := range secrets {
				rows[i] = []string{s.Name, s.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "create SECRET_ID",
		Short: "Create a secret",
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

			s, err := client.Create(ctx, project, args[0])
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created secret %s.\n", s.Name)
			return nil
		},
	}
}

func newDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete SECRET_NAME",
		Short: "Delete a secret",
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

			name := resolveSecretName(project, args[0])
			if err := client.Delete(ctx, name); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted secret %s.\n", args[0])
			return nil
		},
	}
}

func newDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe SECRET_NAME",
		Short: "Describe a secret",
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

			name := resolveSecretName(project, args[0])
			s, err := client.Describe(ctx, name)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), s)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", s.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create Time: %s\n", s.CreateTime)
			return nil
		},
	}
}

func newUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var labelPairs []string

	cmd := &cobra.Command{
		Use:   "update SECRET_NAME",
		Short: "Update secret labels",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if len(labelPairs) == 0 {
				return fmt.Errorf("at least one --label is required")
			}
			labels := make(map[string]string, len(labelPairs))
			for _, pair := range labelPairs {
				key, value, ok := strings.Cut(pair, "=")
				if !ok || key == "" {
					return fmt.Errorf("invalid --label %q, expected key=value", pair)
				}
				labels[key] = value
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			secret, err := client.UpdateLabels(ctx, resolveSecretName(project, args[0]), labels)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), secret)
		},
	}
	cmd.Flags().StringArrayVar(&labelPairs, "label", nil, "Secret label in key=value form")
	return cmd
}

func newIAMCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage Secret Manager IAM policies",
	}

	cmd.AddCommand(
		newGetPolicyCommand(cfg, creds),
		newSetPolicyCommand(cfg, creds),
		newTestPermissionsCommand(cfg, creds),
	)

	return cmd
}

func newGetPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "get-policy SECRET_NAME",
		Short: "Get a secret IAM policy",
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

			policy, err := client.GetPolicy(ctx, resolveSecretName(project, args[0]))
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}
}

func newSetPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var member, role string
	var remove bool

	cmd := &cobra.Command{
		Use:   "set-policy SECRET_NAME",
		Short: "Update a secret IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if member == "" || role == "" {
				return fmt.Errorf("--member and --role are required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := resolveSecretName(project, args[0])
			policy, err := client.GetPolicy(ctx, name)
			if err != nil {
				return err
			}
			updated := applySecretBinding(policy, member, role, remove)
			policy, err = client.SetPolicy(ctx, name, updated)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}

	cmd.Flags().StringVar(&member, "member", "", "Member (e.g. user:foo@example.com)")
	cmd.Flags().StringVar(&role, "role", "", "Role (e.g. roles/secretmanager.viewer)")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove the member from the role instead of adding it")
	return cmd
}

func newTestPermissionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var permissions []string

	cmd := &cobra.Command{
		Use:   "test-permissions SECRET_NAME",
		Short: "Test secret IAM permissions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if len(permissions) == 0 {
				return fmt.Errorf("at least one --permission is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			granted, err := client.TestPermissions(ctx, resolveSecretName(project, args[0]), permissions)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), granted)
			}

			rows := make([][]string, len(granted))
			for i, permission := range granted {
				rows[i] = []string{permission}
			}
			return output.PrintTable(cmd.OutOrStdout(), []string{"PERMISSION"}, rows)
		},
	}

	cmd.Flags().StringArrayVar(&permissions, "permission", nil, "Permission to test")
	return cmd
}

func newVersionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Manage secret versions",
	}

	cmd.AddCommand(
		newVersionsListCommand(cfg, creds),
		newVersionsDescribeCommand(cfg, creds),
		newVersionsAddCommand(cfg, creds),
		newVersionsAccessCommand(cfg, creds),
		newVersionsDestroyCommand(cfg, creds),
		newVersionsDisableCommand(cfg, creds),
		newVersionsEnableCommand(cfg, creds),
	)

	return cmd
}

func newVersionsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe VERSION_NAME",
		Short: "Describe a secret version",
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

			version, err := client.DescribeVersion(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), version)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", version.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", version.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create Time: %s\n", version.CreateTime)
			return nil
		},
	}
}

func newVersionsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list SECRET_NAME",
		Short: "List secret versions",
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

			name := resolveSecretName(project, args[0])
			versions, err := client.ListVersions(ctx, name)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), versions)
			}

			headers := []string{"NAME", "STATE", "CREATE_TIME"}
			rows := make([][]string, len(versions))
			for i, v := range versions {
				rows[i] = []string{v.Name, v.State, v.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newVersionsAddCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var dataFile string

	cmd := &cobra.Command{
		Use:   "add SECRET_NAME",
		Short: "Add a new secret version",
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

			var payload []byte
			if dataFile != "" {
				payload, err = os.ReadFile(dataFile)
				if err != nil {
					return fmt.Errorf("read data file: %w", err)
				}
			} else {
				return fmt.Errorf("--data-file is required")
			}

			name := resolveSecretName(project, args[0])
			v, err := client.AddVersion(ctx, name, payload)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created version %s.\n", v.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to file containing secret data")

	return cmd
}

func newVersionsAccessCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "access VERSION_NAME",
		Short: "Access a secret version's data",
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

			data, err := client.AccessVersion(ctx, args[0])
			if err != nil {
				return err
			}

			_, _ = fmt.Fprint(cmd.OutOrStdout(), string(data))
			return nil
		},
	}
}

func newVersionsDestroyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "destroy VERSION_NAME",
		Short: "Destroy a secret version",
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

			if err := client.DestroyVersion(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Destroyed version %s.\n", args[0])
			return nil
		},
	}
}

func newVersionsDisableCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "disable VERSION_NAME",
		Short: "Disable a secret version",
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

			if err := client.DisableVersion(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Disabled version %s.\n", args[0])
			return nil
		},
	}
}

func newVersionsEnableCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "enable VERSION_NAME",
		Short: "Enable a secret version",
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

			if err := client.EnableVersion(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Enabled version %s.\n", args[0])
			return nil
		},
	}
}

func resolveSecretName(project, name string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/secrets/%s", project, name)
}

func applySecretBinding(policy *secretmanager.Policy, member, role string, remove bool) *secretmanager.Policy {
	if policy == nil {
		policy = &secretmanager.Policy{}
	}

	bindings := make([]*secretmanager.Binding, 0, len(policy.Bindings))
	found := false
	for _, binding := range policy.Bindings {
		if binding.Role != role {
			bindings = append(bindings, binding)
			continue
		}
		found = true
		if remove {
			members := make([]string, 0, len(binding.Members))
			for _, existing := range binding.Members {
				if existing != member {
					members = append(members, existing)
				}
			}
			if len(members) > 0 {
				binding.Members = members
				bindings = append(bindings, binding)
			}
			continue
		}
		if !containsSecretString(binding.Members, member) {
			binding.Members = append(binding.Members, member)
		}
		bindings = append(bindings, binding)
	}

	if !found && !remove {
		bindings = append(bindings, &secretmanager.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	policy.Bindings = bindings
	return policy
}

func containsSecretString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
