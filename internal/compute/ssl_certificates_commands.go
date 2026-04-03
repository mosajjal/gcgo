package compute

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func newSSLCertificatesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ssl-certificates",
		Short: "Manage SSL certificates",
	}
	cmd.AddCommand(
		newSSLCertificatesListCommand(cfg, creds),
		newSSLCertificatesDescribeCommand(cfg, creds),
		newSSLCertificatesCreateCommand(cfg, creds),
		newSSLCertificatesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newSSLCertificatesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List SSL certificates",
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
			certs, err := client.ListSSLCertificates(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), certs)
			}
			headers := []string{"NAME", "TYPE", "STATUS", "DOMAINS"}
			rows := make([][]string, len(certs))
			for i, c := range certs {
				rows[i] = []string{c.Name, c.Type, c.Status, strings.Join(c.Domains, ",")}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newSSLCertificatesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe an SSL certificate",
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
			cert, err := client.GetSSLCertificate(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), cert)
		},
	}
}

func newSSLCertificatesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateSSLCertificateRequest

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create an SSL certificate",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			if len(req.Domains) == 0 && (req.CertFile == "" || req.KeyFile == "") {
				return fmt.Errorf("either --domains (managed) or both --cert-file and --key-file (self-managed) are required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateSSLCertificate(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created SSL certificate %q.\n", req.Name)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&req.Domains, "domains", nil, "Domains for a managed certificate (e.g. example.com,www.example.com)")
	cmd.Flags().StringVar(&req.CertFile, "cert-file", "", "Path to PEM certificate file (self-managed)")
	cmd.Flags().StringVar(&req.KeyFile, "key-file", "", "Path to PEM private key file (self-managed)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Certificate description")
	return cmd
}

func newSSLCertificatesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete an SSL certificate",
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
			if err := client.DeleteSSLCertificate(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted SSL certificate %q.\n", args[0])
			return nil
		},
	}
}
