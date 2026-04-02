package logging

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the logging command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "logging",
		Short: "Manage Cloud Logging",
	}

	cmd.AddCommand(
		newReadCommand(cfg, creds),
		newTailCommand(cfg, creds),
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

func newReadCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var limit int

	cmd := &cobra.Command{
		Use:   "read [FILTER]",
		Short: "Read log entries",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}

			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}
			client, err := NewClient(ctx, opt)
			if err != nil {
				return err
			}

			entries, err := client.ReadLogs(ctx, project, filter, limit)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), entries)
			}

			headers := []string{"TIMESTAMP", "SEVERITY", "LOG_NAME", "PAYLOAD"}
			rows := make([][]string, len(entries))
			for i, e := range entries {
				payload := e.Payload
				if len(payload) > 120 {
					payload = payload[:120] + "..."
				}
				rows[i] = []string{e.Timestamp, e.Severity, e.LogName, payload}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().IntVar(&limit, "limit", 50, "Maximum number of entries")

	return cmd
}

func newTailCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "tail [FILTER]",
		Short: "Stream log entries in real-time",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			filter := ""
			if len(args) > 0 {
				filter = args[0]
			}

			ctx := cmd.Context()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}

			return TailLogs(ctx, cmd.OutOrStdout(), project, filter, opt)
		},
	}
}
