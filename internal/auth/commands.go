package auth

import (
	"fmt"

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
		newRevokeCommand(creds),
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
