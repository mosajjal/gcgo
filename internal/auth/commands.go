package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
)

// NewCommand returns the auth command group.
func NewCommand(creds *Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage GCP authentication",
	}

	cmd.AddCommand(
		newLoginCommand(creds),
		newListCommand(creds),
		newPrintAccessTokenCommand(creds),
		newPrintIdentityTokenCommand(creds),
		newRevokeCommand(creds),
		newApplicationDefaultCommand(creds),
		newConfigureDockerCommand(creds),
	)

	return cmd
}

func newLoginCommand(creds *Credentials) *cobra.Command {
	var keyFile string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Authenticate with Google Cloud",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if keyFile != "" {
				if err := creds.StoreServiceAccountKey(keyFile); err != nil {
					return err
				}
				account, err := creds.ActiveAccount()
				if err != nil {
					return err
				}
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Authenticated with service account: %s\n", account)
				return nil
			}

			// Browser-based OAuth flow
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Your browser will open to authenticate with Google Cloud.")
			email, err := creds.BrowserLogin(cmd.Context())
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Authenticated as: %s\n", email)
			return nil
		},
	}

	cmd.Flags().StringVar(&keyFile, "service-account-key", "", "Path to service account JSON key file")

	return cmd
}

func newListCommand(creds *Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "Show active authentication credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			account, err := creds.ActiveAccount()
			if err != nil {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "No active credentials.")
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Run 'gcgo auth login' to authenticate.")
				return nil //nolint:nilerr // intentional: no-creds is informational, not an error
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Active account: %s\n", account)
			if creds.HasStoredCredentials() {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Source: gcgo stored credentials")
			} else {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Source: Application Default Credentials")
			}
			return nil
		},
	}
}

func newPrintAccessTokenCommand(creds *Credentials) *cobra.Command {
	var (
		scopes          []string
		targetPrincipal string
	)

	cmd := &cobra.Command{
		Use:   "print-access-token",
		Short: "Print an access token for the active credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := creds.AccessToken(cmd.Context(), targetPrincipal, scopes)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), token)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&scopes, "scope", nil, "OAuth scope to request (repeatable)")
	cmd.Flags().StringVar(&targetPrincipal, "impersonate-service-account", "", "Target service account email for impersonation")

	return cmd
}

func newPrintIdentityTokenCommand(creds *Credentials) *cobra.Command {
	var (
		audience        string
		targetPrincipal string
		includeEmail    bool
	)

	cmd := &cobra.Command{
		Use:   "print-identity-token",
		Short: "Print an identity token for the active credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if audience == "" {
				return fmt.Errorf("--audience is required")
			}
			token, err := creds.IdentityToken(cmd.Context(), audience, targetPrincipal, includeEmail)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), token)
			return nil
		},
	}

	cmd.Flags().StringVar(&audience, "audience", "", "Audience for the ID token")
	cmd.Flags().StringVar(&targetPrincipal, "impersonate-service-account", "", "Target service account email for impersonation")
	cmd.Flags().BoolVar(&includeEmail, "include-email", false, "Include the service account email claim in the token")

	return cmd
}

func newRevokeCommand(creds *Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "revoke",
		Short: "Remove stored credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if err := creds.Revoke(); err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Stored credentials removed.")
			return nil
		},
	}
}

func newApplicationDefaultCommand(creds *Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "application-default",
		Short: "Manage Application Default Credentials",
	}
	cmd.AddCommand(
		newADCLoginCommand(creds),
		newADCPrintAccessTokenCommand(creds),
	)
	return cmd
}

func newADCLoginCommand(creds *Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "login",
		Short: "Authenticate and store credentials as Application Default Credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Your browser will open to authenticate with Google Cloud.")
			_, path, err := creds.BrowserLoginADC(cmd.Context())
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Credentials saved to: %s\n", path)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "These credentials will be used by any library that requests Application Default Credentials (ADC).")
			return nil
		},
	}
}

func newADCPrintAccessTokenCommand(creds *Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "print-access-token",
		Short: "Print the access token from Application Default Credentials",
		RunE: func(cmd *cobra.Command, _ []string) error {
			token, err := creds.AccessToken(cmd.Context(), "", nil)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), token)
			return nil
		},
	}
}

func newConfigureDockerCommand(_ *Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "configure-docker",
		Short: "Configure Docker to authenticate to GCP registries",
		RunE: func(cmd *cobra.Command, _ []string) error {
			home, err := os.UserHomeDir()
			if err != nil {
				return fmt.Errorf("resolve home dir: %w", err)
			}

			dockerConfigPath := filepath.Join(home, ".docker", "config.json")

			// Read existing config or start fresh
			var config map[string]json.RawMessage
			data, err := os.ReadFile(dockerConfigPath) //nolint:gosec // path is from well-known location
			if err != nil && !os.IsNotExist(err) {
				return fmt.Errorf("read docker config: %w", err)
			}
			if len(data) > 0 {
				if err := json.Unmarshal(data, &config); err != nil {
					return fmt.Errorf("parse docker config: %w", err)
				}
			}
			if config == nil {
				config = make(map[string]json.RawMessage)
			}

			// Merge credHelpers
			var credHelpers map[string]string
			if raw, ok := config["credHelpers"]; ok {
				if err := json.Unmarshal(raw, &credHelpers); err != nil {
					return fmt.Errorf("parse credHelpers: %w", err)
				}
			}
			if credHelpers == nil {
				credHelpers = make(map[string]string)
			}

			registries := []string{
				"gcr.io",
				"us.gcr.io",
				"eu.gcr.io",
				"asia.gcr.io",
				"us-docker.pkg.dev",
				"europe-docker.pkg.dev",
				"asia-docker.pkg.dev",
			}
			for _, r := range registries {
				credHelpers[r] = "gcgo"
			}

			raw, err := json.Marshal(credHelpers)
			if err != nil {
				return fmt.Errorf("marshal credHelpers: %w", err)
			}
			config["credHelpers"] = json.RawMessage(raw)

			out, err := json.MarshalIndent(config, "", "  ")
			if err != nil {
				return fmt.Errorf("marshal docker config: %w", err)
			}

			if err := os.MkdirAll(filepath.Dir(dockerConfigPath), 0o700); err != nil {
				return fmt.Errorf("create docker config dir: %w", err)
			}
			if err := os.WriteFile(dockerConfigPath, out, 0o600); err != nil {
				return fmt.Errorf("write docker config: %w", err)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(),
				"Docker has been configured to use gcgo as credential helper for: %s\n",
				"gcr.io, us.gcr.io, eu.gcr.io, asia.gcr.io, us-docker.pkg.dev, europe-docker.pkg.dev, asia-docker.pkg.dev",
			)
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Note: gcgo must be in your PATH for Docker to use it as a credential helper.")
			return nil
		},
	}
}
