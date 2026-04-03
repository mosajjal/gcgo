package scc

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the Security Command Center command group.
func NewCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scc",
		Short: "Manage Security Command Center resources",
	}

	cmd.AddCommand(
		newFindingsCommand(creds),
		newSourcesCommand(creds),
		newNotificationsCommand(creds),
	)

	return cmd
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newFindingsCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "findings",
		Short: "Manage SCC findings",
	}

	cmd.AddCommand(
		newFindingsListCommand(creds),
		newFindingsSetStateCommand(creds),
		newFindingsSetMuteCommand(creds),
	)

	return cmd
}

func newFindingsListCommand(creds *auth.Credentials) *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "list SOURCE",
		Short: "List findings under a source resource",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			findings, err := client.ListFindings(ctx, args[0], filter)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), findings)
			}

			headers := []string{"NAME", "CATEGORY", "SEVERITY", "STATE", "MUTE"}
			rows := make([][]string, len(findings))
			for i, finding := range findings {
				rows[i] = []string{finding.Name, finding.Category, finding.Severity, finding.State, finding.Mute}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&filter, "filter", "", "Findings filter")
	return cmd
}

func newFindingsSetStateCommand(creds *auth.Credentials) *cobra.Command {
	var state string

	cmd := &cobra.Command{
		Use:   "set-state FINDING",
		Short: "Set finding state to ACTIVE or INACTIVE",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if state == "" {
				return fmt.Errorf("--state is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.UpdateFindingState(ctx, args[0], state); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated finding state for %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&state, "state", "", "Desired state (ACTIVE or INACTIVE)")
	return cmd
}

func newFindingsSetMuteCommand(creds *auth.Credentials) *cobra.Command {
	var mute string

	cmd := &cobra.Command{
		Use:   "set-mute FINDING",
		Short: "Set finding mute status",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if mute == "" {
				return fmt.Errorf("--mute is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.SetFindingMute(ctx, args[0], mute); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated finding mute state for %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&mute, "mute", "", "Desired mute state (MUTED or UNMUTED)")
	return cmd
}

func newSourcesCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "sources ORG_ID",
		Short: "List SCC sources for an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			sources, err := client.ListSources(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), sources)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "DESCRIPTION"}
			rows := make([][]string, len(sources))
			for i, source := range sources {
				rows[i] = []string{source.Name, source.DisplayName, source.Description}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newNotificationsCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "notifications",
		Short: "Manage SCC notification configs",
	}

	cmd.AddCommand(
		newNotificationsListCommand(creds),
		newNotificationsDescribeCommand(creds),
		newNotificationsCreateCommand(creds),
		newNotificationsDeleteCommand(creds),
	)

	return cmd
}

func newNotificationsListCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list ORG_ID",
		Short: "List notification configs for an organization",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			configs, err := client.ListNotifications(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), configs)
			}

			headers := []string{"NAME", "TOPIC", "FILTER"}
			rows := make([][]string, len(configs))
			for i, cfg := range configs {
				rows[i] = []string{cfg.Name, cfg.PubsubTopic, cfg.Filter}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newNotificationsDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe NAME",
		Short: "Describe a notification config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			cfg, err := client.DescribeNotification(ctx, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), cfg)
		},
	}
}

func newNotificationsCreateCommand(creds *auth.Credentials) *cobra.Command {
	var topic string
	var filter string

	cmd := &cobra.Command{
		Use:   "create ORG_ID CONFIG_ID",
		Short: "Create a notification config",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			if topic == "" {
				return fmt.Errorf("--topic is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			cfg, err := client.CreateNotification(ctx, args[0], args[1], topic, filter)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created notification config %s.\n", cfg.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&topic, "topic", "", "Pub/Sub topic resource")
	cmd.Flags().StringVar(&filter, "filter", "", "Notification filter")
	return cmd
}

func newNotificationsDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete NAME",
		Short: "Delete a notification config",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteNotification(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted notification config %s.\n", args[0])
			return nil
		},
	}
}
