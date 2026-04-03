package dns

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the dns command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dns",
		Short: "Manage Cloud DNS resources",
	}

	cmd.AddCommand(
		newManagedZonesCommand(cfg, creds),
		newRecordSetsCommand(cfg, creds),
		newPoliciesCommand(cfg, creds),
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

// managed-zones

func newManagedZonesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "managed-zones",
		Short: "Manage Cloud DNS managed zones",
	}

	cmd.AddCommand(
		newManagedZonesListCommand(cfg, creds),
		newManagedZonesDescribeCommand(cfg, creds),
		newManagedZonesCreateCommand(cfg, creds),
		newManagedZonesDeleteCommand(cfg, creds),
	)

	return cmd
}

func newManagedZonesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List managed zones",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			zones, err := client.ListManagedZones(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), zones)
			}

			headers := []string{"NAME", "DNS_NAME", "VISIBILITY", "DESCRIPTION"}
			rows := make([][]string, len(zones))
			for i, z := range zones {
				rows[i] = []string{z.Name, z.DNSName, z.Visibility, z.Description}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newManagedZonesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ZONE",
		Short: "Describe a managed zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			zone, err := client.GetManagedZone(ctx, project, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), zone)
		},
	}
}

func newManagedZonesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateZoneRequest

	cmd := &cobra.Command{
		Use:   "create ZONE",
		Short: "Create a managed zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			req.Name = args[0]

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			zone, err := client.CreateManagedZone(ctx, project, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created managed zone %q (%s).\n", zone.Name, zone.DNSName)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DNSName, "dns-name", "", "DNS name for the zone (e.g. example.com.)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Zone description")
	cmd.Flags().StringVar(&req.Visibility, "visibility", "public", "Zone visibility: public or private")
	_ = cmd.MarkFlagRequired("dns-name")

	return cmd
}

func newManagedZonesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete ZONE",
		Short: "Delete a managed zone",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteManagedZone(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted managed zone %q.\n", args[0])
			return nil
		},
	}
}

// record-sets

func newRecordSetsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "record-sets",
		Short: "Manage DNS record sets",
	}

	cmd.AddCommand(
		newRecordSetsListCommand(cfg, creds),
		newRecordSetsCreateCommand(cfg, creds),
		newRecordSetsDeleteCommand(cfg, creds),
	)

	return cmd
}

func newRecordSetsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List record sets in a managed zone",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			zone, _ := cmd.Flags().GetString("zone")
			if zone == "" {
				return fmt.Errorf("--zone is required")
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			sets, err := client.ListRecordSets(ctx, project, zone)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), sets)
			}

			headers := []string{"NAME", "TYPE", "TTL", "DATA"}
			rows := make([][]string, len(sets))
			for i, r := range sets {
				rows[i] = []string{r.Name, r.Type, fmt.Sprintf("%d", r.TTL), strings.Join(r.RRDatas, ",")}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().String("zone", "", "Managed zone name")
	return cmd
}

func newRecordSetsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateRecordSetRequest

	cmd := &cobra.Command{
		Use:   "create NAME",
		Short: "Create a record set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			zone, _ := cmd.Flags().GetString("zone")
			if zone == "" {
				return fmt.Errorf("--zone is required")
			}

			req.Name = args[0]

			rrdata, _ := cmd.Flags().GetString("rrdatas")
			if rrdata != "" {
				req.RRDatas = strings.Split(rrdata, ",")
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.CreateRecordSet(ctx, project, zone, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created record set %q in zone %q.\n", req.Name, zone)
			return nil
		},
	}

	cmd.Flags().String("zone", "", "Managed zone name")
	cmd.Flags().StringVar(&req.Type, "type", "A", "Record type (A, AAAA, CNAME, MX, etc.)")
	cmd.Flags().Int64Var(&req.TTL, "ttl", 300, "Time to live in seconds")
	cmd.Flags().String("rrdatas", "", "Comma-separated list of resource record data values")
	return cmd
}

func newRecordSetsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a record set",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			zone, _ := cmd.Flags().GetString("zone")
			if zone == "" {
				return fmt.Errorf("--zone is required")
			}

			rrtype, _ := cmd.Flags().GetString("type")
			if rrtype == "" {
				return fmt.Errorf("--type is required")
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteRecordSet(ctx, project, zone, args[0], rrtype); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted record set %q (%s) from zone %q.\n", args[0], rrtype, zone)
			return nil
		},
	}

	cmd.Flags().String("zone", "", "Managed zone name")
	cmd.Flags().String("type", "", "Record type (A, AAAA, CNAME, MX, etc.)")
	return cmd
}

// policies

func newPoliciesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "Manage Cloud DNS policies",
	}

	cmd.AddCommand(
		newPoliciesListCommand(cfg, creds),
		newPoliciesDescribeCommand(cfg, creds),
		newPoliciesCreateCommand(cfg, creds),
		newPoliciesDeleteCommand(cfg, creds),
	)

	return cmd
}

func newPoliciesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List DNS policies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			policies, err := client.ListPolicies(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), policies)
			}

			headers := []string{"NAME", "INBOUND_FORWARDING", "LOGGING", "DESCRIPTION"}
			rows := make([][]string, len(policies))
			for i, p := range policies {
				rows[i] = []string{
					p.Name,
					fmt.Sprintf("%t", p.EnableInboundForwarding),
					fmt.Sprintf("%t", p.EnableLogging),
					p.Description,
				}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newPoliciesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe POLICY",
		Short: "Describe a DNS policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			policy, err := client.GetPolicy(ctx, project, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}
}

func newPoliciesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreatePolicyRequest

	cmd := &cobra.Command{
		Use:   "create POLICY",
		Short: "Create a DNS policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			req.Name = args[0]

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			policy, err := client.CreatePolicy(ctx, project, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created policy %q.\n", policy.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Description, "description", "", "Policy description")
	cmd.Flags().BoolVar(&req.EnableInboundForwarding, "enable-inbound-forwarding", false, "Enable inbound forwarding")
	cmd.Flags().BoolVar(&req.EnableLogging, "enable-logging", false, "Enable DNS query logging")

	return cmd
}

func newPoliciesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete POLICY",
		Short: "Delete a DNS policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := newClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeletePolicy(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted policy %q.\n", args[0])
			return nil
		},
	}
}
