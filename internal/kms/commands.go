package kms

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
	cloudkms "google.golang.org/api/cloudkms/v1"
)

// NewCommand returns the kms command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "kms",
		Short: "Manage Cloud KMS resources",
	}

	cmd.AddCommand(
		newKeyRingsCommand(cfg, creds),
		newKeysCommand(cfg, creds),
		newEncryptCommand(cfg, creds),
		newDecryptCommand(cfg, creds),
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

func newKeyRingsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keyrings",
		Short: "Manage KMS key rings",
	}

	cmd.AddCommand(
		newKeyRingsListCommand(cfg, creds),
		newKeyRingsCreateCommand(cfg, creds),
		newKeyRingsDescribeCommand(cfg, creds),
		newKeyRingsIAMCommand(cfg, creds),
	)

	return cmd
}

func newKeyRingsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List key rings",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			rings, err := client.ListKeyRings(ctx, project, location)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), rings)
			}

			headers := []string{"NAME", "CREATE_TIME"}
			rows := make([][]string, len(rings))
			for i, r := range rings {
				rows[i] = []string{r.Name, r.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location (e.g. global, us-east1)")

	return cmd
}

func newKeyRingsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "create KEY_RING_ID",
		Short: "Create a key ring",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			kr, err := client.CreateKeyRing(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created key ring %s.\n", kr.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location (e.g. global, us-east1)")

	return cmd
}

func newKeyRingsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "describe KEY_RING_ID",
		Short: "Describe a key ring",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, args[0])
			kr, err := client.DescribeKeyRing(ctx, name)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), kr)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", kr.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create Time: %s\n", kr.CreateTime)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location (e.g. global, us-east1)")

	return cmd
}

func newKeyRingsIAMCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage KMS key ring IAM policies",
	}

	cmd.AddCommand(
		newKeyRingGetPolicyCommand(cfg, creds),
		newKeyRingSetPolicyCommand(cfg, creds),
		newKeyRingTestPermissionsCommand(cfg, creds),
	)

	return cmd
}

func newKeyRingGetPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "get-policy KEY_RING_ID",
		Short: "Get a key ring IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, args[0])
			policy, err := client.GetKeyRingPolicy(ctx, name)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location (e.g. global, us-east1)")
	return cmd
}

func newKeyRingSetPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, member, role string
	var remove bool

	cmd := &cobra.Command{
		Use:   "set-policy KEY_RING_ID",
		Short: "Update a key ring IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			if member == "" || role == "" {
				return fmt.Errorf("--member and --role are required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, args[0])
			policy, err := client.GetKeyRingPolicy(ctx, name)
			if err != nil {
				return err
			}
			updated := applyKMSBinding(policy, member, role, remove)
			policy, err = client.SetKeyRingPolicy(ctx, name, updated)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location (e.g. global, us-east1)")
	cmd.Flags().StringVar(&member, "member", "", "Member (e.g. user:foo@example.com)")
	cmd.Flags().StringVar(&role, "role", "", "Role (e.g. roles/viewer)")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove the member from the role instead of adding it")
	return cmd
}

func newKeyRingTestPermissionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var permissions []string

	cmd := &cobra.Command{
		Use:   "test-permissions KEY_RING_ID",
		Short: "Test key ring IAM permissions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			if len(permissions) == 0 {
				return fmt.Errorf("at least one --permission is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, args[0])
			granted, err := client.TestKeyRingPermissions(ctx, name, permissions)
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

	cmd.Flags().StringVar(&location, "location", "", "KMS location (e.g. global, us-east1)")
	cmd.Flags().StringArrayVar(&permissions, "permission", nil, "Permission to test")
	return cmd
}

func newKeysCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "keys",
		Short: "Manage KMS crypto keys",
	}

	cmd.AddCommand(
		newKeysListCommand(cfg, creds),
		newKeysCreateCommand(cfg, creds),
		newKeysDescribeCommand(cfg, creds),
		newKeysSetPrimaryVersionCommand(cfg, creds),
		newKeysIAMCommand(cfg, creds),
		newKeyVersionsCommand(cfg, creds),
	)

	return cmd
}

func newKeysListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List crypto keys in a key ring",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			if keyring == "" {
				return fmt.Errorf("--keyring is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keyRingName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, keyring)
			keys, err := client.ListKeys(ctx, keyRingName)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), keys)
			}

			headers := []string{"NAME", "PURPOSE", "PRIMARY_VERSION"}
			rows := make([][]string, len(keys))
			for i, k := range keys {
				rows[i] = []string{k.Name, k.Purpose, k.PrimaryVersion}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")

	return cmd
}

func newKeysCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring, purpose string

	cmd := &cobra.Command{
		Use:   "create KEY_ID",
		Short: "Create a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			if keyring == "" {
				return fmt.Errorf("--keyring is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keyRingName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s", project, location, keyring)
			k, err := client.CreateKey(ctx, keyRingName, args[0], purpose)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created key %s.\n", k.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringVar(&purpose, "purpose", "encryption", "Key purpose: encryption, asymmetric-signing, asymmetric-encryption")

	return cmd
}

func newKeysDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring string

	cmd := &cobra.Command{
		Use:   "describe KEY_ID",
		Short: "Describe a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			if keyring == "" {
				return fmt.Errorf("--keyring is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, args[0])
			k, err := client.DescribeKey(ctx, name)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), k)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:             %s\n", k.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Purpose:          %s\n", k.Purpose)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Primary Version:  %s\n", k.PrimaryVersion)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Protection Level: %s\n", k.ProtectionLevel)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create Time:      %s\n", k.CreateTime)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")

	return cmd
}

func newKeysIAMCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage KMS crypto key IAM policies",
	}

	cmd.AddCommand(
		newKeyGetPolicyCommand(cfg, creds),
		newKeySetPolicyCommand(cfg, creds),
		newKeyTestPermissionsCommand(cfg, creds),
	)

	return cmd
}

func newKeyGetPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring string

	cmd := &cobra.Command{
		Use:   "get-policy KEY_ID",
		Short: "Get a crypto key IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" {
				return fmt.Errorf("--location and --keyring are required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, args[0])
			policy, err := client.GetKeyPolicy(ctx, name)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	return cmd
}

func newKeySetPolicyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring, member, role string
	var remove bool

	cmd := &cobra.Command{
		Use:   "set-policy KEY_ID",
		Short: "Update a crypto key IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" {
				return fmt.Errorf("--location and --keyring are required")
			}
			if member == "" || role == "" {
				return fmt.Errorf("--member and --role are required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, args[0])
			policy, err := client.GetKeyPolicy(ctx, name)
			if err != nil {
				return err
			}
			updated := applyKMSBinding(policy, member, role, remove)
			policy, err = client.SetKeyPolicy(ctx, name, updated)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringVar(&member, "member", "", "Member (e.g. user:foo@example.com)")
	cmd.Flags().StringVar(&role, "role", "", "Role (e.g. roles/viewer)")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove the member from the role instead of adding it")
	return cmd
}

func newKeyTestPermissionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring string
	var permissions []string

	cmd := &cobra.Command{
		Use:   "test-permissions KEY_ID",
		Short: "Test crypto key IAM permissions",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" {
				return fmt.Errorf("--location and --keyring are required")
			}
			if len(permissions) == 0 {
				return fmt.Errorf("at least one --permission is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			name := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, args[0])
			granted, err := client.TestKeyPermissions(ctx, name, permissions)
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

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringArrayVar(&permissions, "permission", nil, "Permission to test")
	return cmd
}

func newKeysSetPrimaryVersionCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring, version string

	cmd := &cobra.Command{
		Use:   "set-primary-version KEY_ID",
		Short: "Set the primary version for a crypto key",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" || version == "" {
				return fmt.Errorf("--location, --keyring, and --version are all required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, args[0])
			key, err := client.UpdatePrimaryVersion(ctx, keyName, version)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Set primary version to %s for %s.\n", key.PrimaryVersion, key.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringVar(&version, "version", "", "Crypto key version ID or full resource name")
	return cmd
}

func newKeyVersionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "versions",
		Short: "Manage crypto key versions",
	}

	cmd.AddCommand(
		newKeyVersionsListCommand(cfg, creds),
		newKeyVersionsCreateCommand(cfg, creds),
		newKeyVersionsDescribeCommand(cfg, creds),
		newKeyVersionsDestroyCommand(cfg, creds),
		newKeyVersionsEnableCommand(cfg, creds),
		newKeyVersionsDisableCommand(cfg, creds),
		newKeyVersionsAsymmetricSignCommand(cfg, creds),
	)

	return cmd
}

func newKeyVersionsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring, key string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a new key version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" || key == "" {
				return fmt.Errorf("--location, --keyring, and --key are all required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, key)
			version, err := client.CreateKeyVersion(ctx, keyName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created key version %s.\n", version.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringVar(&key, "key", "", "Crypto key ID")
	return cmd
}

func newKeyVersionsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring, key string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List key versions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" || key == "" {
				return fmt.Errorf("--location, --keyring, and --key are all required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, key)
			versions, err := client.ListKeyVersions(ctx, keyName)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), versions)
			}

			headers := []string{"NAME", "STATE", "ALGORITHM", "PROTECTION_LEVEL"}
			rows := make([][]string, len(versions))
			for i, v := range versions {
				rows[i] = []string{v.Name, v.State, v.Algorithm, v.ProtectionLevel}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringVar(&key, "key", "", "Crypto key ID")

	return cmd
}

func newKeyVersionsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe VERSION_NAME",
		Short: "Describe a key version",
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

			v, err := client.DescribeKeyVersion(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), v)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:             %s\n", v.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:            %s\n", v.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Algorithm:        %s\n", v.Algorithm)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Protection Level: %s\n", v.ProtectionLevel)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Create Time:      %s\n", v.CreateTime)
			return nil
		},
	}
}

func newKeyVersionsDestroyCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "destroy VERSION_NAME",
		Short: "Destroy a key version",
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

			if err := client.DestroyKeyVersion(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Scheduled destruction of version %s.\n", args[0])
			return nil
		},
	}
}

func newKeyVersionsEnableCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "enable VERSION_NAME",
		Short: "Enable a key version",
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

			if err := client.EnableKeyVersion(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Enabled version %s.\n", args[0])
			return nil
		},
	}
}

func newKeyVersionsDisableCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "disable VERSION_NAME",
		Short: "Disable a key version",
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

			if err := client.DisableKeyVersion(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Disabled version %s.\n", args[0])
			return nil
		},
	}
}

func newEncryptCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring, key, plaintextFile, ciphertextFile string

	cmd := &cobra.Command{
		Use:   "encrypt",
		Short: "Encrypt data with a KMS key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" || key == "" {
				return fmt.Errorf("--location, --keyring, and --key are all required")
			}
			if plaintextFile == "" {
				return fmt.Errorf("--plaintext-file is required")
			}
			if ciphertextFile == "" {
				return fmt.Errorf("--ciphertext-file is required")
			}

			plaintext, err := os.ReadFile(plaintextFile)
			if err != nil {
				return fmt.Errorf("read plaintext file: %w", err)
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, key)
			ciphertext, err := client.Encrypt(ctx, keyName, plaintext)
			if err != nil {
				return err
			}

			if err := os.WriteFile(ciphertextFile, ciphertext, 0o600); err != nil {
				return fmt.Errorf("write ciphertext file: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Encrypted %s -> %s.\n", plaintextFile, ciphertextFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringVar(&key, "key", "", "Crypto key ID")
	cmd.Flags().StringVar(&plaintextFile, "plaintext-file", "", "Path to plaintext input file")
	cmd.Flags().StringVar(&ciphertextFile, "ciphertext-file", "", "Path to ciphertext output file")

	return cmd
}

func newKeyVersionsAsymmetricSignCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var dataFile string

	cmd := &cobra.Command{
		Use:   "asymmetric-sign VERSION_NAME",
		Short: "Sign data with an asymmetric key version",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if dataFile == "" {
				return fmt.Errorf("--data-file is required")
			}
			payload, err := os.ReadFile(dataFile)
			if err != nil {
				return fmt.Errorf("read data file: %w", err)
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			result, err := client.AsymmetricSign(ctx, args[0], payload)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), result)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:             %s\n", result.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Protection Level: %s\n", result.ProtectionLevel)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Signature:        %s\n", base64.StdEncoding.EncodeToString(result.Signature))
			return nil
		},
	}

	cmd.Flags().StringVar(&dataFile, "data-file", "", "Path to file containing data to sign")
	return cmd
}

func newDecryptCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location, keyring, key, plaintextFile, ciphertextFile string

	cmd := &cobra.Command{
		Use:   "decrypt",
		Short: "Decrypt data with a KMS key",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" || keyring == "" || key == "" {
				return fmt.Errorf("--location, --keyring, and --key are all required")
			}
			if ciphertextFile == "" {
				return fmt.Errorf("--ciphertext-file is required")
			}
			if plaintextFile == "" {
				return fmt.Errorf("--plaintext-file is required")
			}

			ciphertext, err := os.ReadFile(ciphertextFile)
			if err != nil {
				return fmt.Errorf("read ciphertext file: %w", err)
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			keyName := fmt.Sprintf("projects/%s/locations/%s/keyRings/%s/cryptoKeys/%s", project, location, keyring, key)
			plaintext, err := client.Decrypt(ctx, keyName, ciphertext)
			if err != nil {
				return err
			}

			if err := os.WriteFile(plaintextFile, plaintext, 0o600); err != nil {
				return fmt.Errorf("write plaintext file: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Decrypted %s -> %s.\n", ciphertextFile, plaintextFile)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "KMS location")
	cmd.Flags().StringVar(&keyring, "keyring", "", "Key ring ID")
	cmd.Flags().StringVar(&key, "key", "", "Crypto key ID")
	cmd.Flags().StringVar(&ciphertextFile, "ciphertext-file", "", "Path to ciphertext input file")
	cmd.Flags().StringVar(&plaintextFile, "plaintext-file", "", "Path to plaintext output file")

	return cmd
}

func applyKMSBinding(policy *cloudkms.Policy, member, role string, remove bool) *cloudkms.Policy {
	if policy == nil {
		policy = &cloudkms.Policy{}
	}

	bindings := make([]*cloudkms.Binding, 0, len(policy.Bindings))
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
		if !containsString(binding.Members, member) {
			binding.Members = append(binding.Members, member)
		}
		bindings = append(bindings, binding)
	}

	if !found && !remove {
		bindings = append(bindings, &cloudkms.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	policy.Bindings = bindings
	return policy
}

func containsString(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}
	return false
}
