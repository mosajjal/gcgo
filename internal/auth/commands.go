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
		newPrintAccessTokenCommand(creds),
		newPrintIdentityTokenCommand(creds),
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
