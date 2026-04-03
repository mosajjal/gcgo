package monitoring

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the monitoring command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "monitoring",
		Short: "Cloud Monitoring",
	}

	cmd.AddCommand(
		newDashboardsCommand(cfg, creds),
		newPoliciesCommand(cfg, creds),
		newChannelsCommand(cfg, creds),
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

// Dashboards.

func newDashboardsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dashboards",
		Short: "Manage monitoring dashboards",
	}

	cmd.AddCommand(
		newDashboardListCommand(cfg, creds),
		newDashboardDescribeCommand(creds),
		newDashboardCreateCommand(cfg, creds),
		newDashboardDeleteCommand(creds),
	)

	return cmd
}

func newDashboardListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List dashboards",
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

			dashes, err := client.ListDashboards(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), dashes)
			}

			headers := []string{"NAME", "DISPLAY_NAME"}
			rows := make([][]string, len(dashes))
			for i, d := range dashes {
				rows[i] = []string{d.Name, d.DisplayName}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDashboardDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			d, err := client.GetDashboard(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), d)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", d.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", d.DisplayName)
			return nil
		},
	}
}

func newDashboardCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var displayName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a dashboard",
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

			d, err := client.CreateDashboard(ctx, project, displayName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created dashboard %s.\n", d.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "Dashboard display name")
	_ = cmd.MarkFlagRequired("display-name")

	return cmd
}

func newDashboardDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a dashboard",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteDashboard(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted dashboard %s.\n", args[0])
			return nil
		},
	}
}

// Alert Policies.

func newPoliciesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "policies",
		Short: "Manage alerting policies",
	}

	cmd.AddCommand(
		newPolicyListCommand(cfg, creds),
		newPolicyDescribeCommand(creds),
		newPolicyCreateCommand(cfg, creds),
		newPolicyDeleteCommand(creds),
	)

	return cmd
}

func newPolicyListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List alerting policies",
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

			policies, err := client.ListAlertPolicies(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), policies)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "ENABLED"}
			rows := make([][]string, len(policies))
			for i, p := range policies {
				rows[i] = []string{p.Name, p.DisplayName, fmt.Sprintf("%v", p.Enabled)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newPolicyDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe an alerting policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			p, err := client.GetAlertPolicy(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), p)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", p.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", p.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Enabled:      %v\n", p.Enabled)
			return nil
		},
	}
}

func newPolicyCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var displayName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an alerting policy",
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

			p, err := client.CreateAlertPolicy(ctx, project, displayName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created alert policy %s.\n", p.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "Policy display name")
	_ = cmd.MarkFlagRequired("display-name")

	return cmd
}

func newPolicyDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete an alerting policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteAlertPolicy(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted alert policy %s.\n", args[0])
			return nil
		},
	}
}

// Notification Channels.

func newChannelsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "channels",
		Short: "Manage notification channels",
	}

	cmd.AddCommand(
		newChannelListCommand(cfg, creds),
		newChannelDescribeCommand(creds),
		newChannelCreateCommand(cfg, creds),
		newChannelDeleteCommand(creds),
	)

	return cmd
}

func newChannelListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List notification channels",
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

			channels, err := client.ListNotificationChannels(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), channels)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "TYPE", "ENABLED"}
			rows := make([][]string, len(channels))
			for i, ch := range channels {
				rows[i] = []string{ch.Name, ch.DisplayName, ch.Type, fmt.Sprintf("%v", ch.Enabled)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newChannelDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a notification channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			ch, err := client.GetNotificationChannel(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), ch)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", ch.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", ch.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:         %s\n", ch.Type)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Enabled:      %v\n", ch.Enabled)
			return nil
		},
	}
}

func newChannelCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var displayName string
	var channelType string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a notification channel",
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

			ch, err := client.CreateNotificationChannel(ctx, project, displayName, channelType)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created notification channel %s.\n", ch.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&displayName, "display-name", "", "Channel display name")
	cmd.Flags().StringVar(&channelType, "type", "", "Channel type")
	_ = cmd.MarkFlagRequired("display-name")
	_ = cmd.MarkFlagRequired("type")

	return cmd
}

func newChannelDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a notification channel",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteNotificationChannel(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted notification channel %s.\n", args[0])
			return nil
		},
	}
}
