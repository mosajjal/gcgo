package logging

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/mosajjal/gcgo/internal/placeholder"
	"github.com/spf13/cobra"
	loggingapi "google.golang.org/api/logging/v2"
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
		newSinksCommand(cfg, creds),
		newMetricsCommand(),
		newExclusionsCommand(),
		newBucketsCommand(),
	)

	return cmd
}

func newMetricsCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/logging/docs/logs-based-metrics"
	return placeholder.NewGroup(
		"metrics",
		"Manage logs-based metrics",
		docsURL,
		placeholder.NewCommand("list", "List logs-based metrics", docsURL),
		placeholder.NewCommand("describe", "Describe a logs-based metric", docsURL),
		placeholder.NewCommand("create", "Create a logs-based metric", docsURL),
		placeholder.NewCommand("delete", "Delete a logs-based metric", docsURL),
	)
}

func newExclusionsCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/logging/docs/exclusions"
	return placeholder.NewGroup(
		"exclusions",
		"Manage log exclusions",
		docsURL,
		placeholder.NewCommand("list", "List log exclusions", docsURL),
		placeholder.NewCommand("describe", "Describe a log exclusion", docsURL),
		placeholder.NewCommand("create", "Create a log exclusion", docsURL),
		placeholder.NewCommand("update", "Update a log exclusion", docsURL),
		placeholder.NewCommand("delete", "Delete a log exclusion", docsURL),
	)
}

func newBucketsCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/logging/docs/store-log-entries"
	return placeholder.NewGroup(
		"buckets",
		"Manage log buckets",
		docsURL,
		placeholder.NewCommand("list", "List log buckets", docsURL),
		placeholder.NewCommand("describe", "Describe a log bucket", docsURL),
		placeholder.NewCommand("create", "Create a log bucket", docsURL),
		placeholder.NewCommand("update", "Update a log bucket", docsURL),
		placeholder.NewCommand("delete", "Delete a log bucket", docsURL),
	)
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func requireParent(cmd *cobra.Command, cfg *config.Config) (string, error) {
	parent, _ := cmd.Flags().GetString("parent")
	if parent != "" {
		return parent, nil
	}
	project, err := requireProject(cmd, cfg)
	if err != nil {
		return "", err
	}
	return "projects/" + project, nil
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

func newSinksCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sinks",
		Short: "Manage Cloud Logging sinks",
	}
	cmd.AddCommand(
		newSinksListCommand(cfg, creds),
		newSinksDescribeCommand(creds),
		newSinksCreateCommand(cfg, creds),
		newSinksDeleteCommand(creds),
	)
	return cmd
}

func newSinksListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var parent string
	var filter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List logging sinks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			parent, err := requireParent(cmd, cfg)
			if err != nil {
				return err
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

			sinks, err := client.ListSinks(ctx, parent, filter)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), sinks)
			}

			headers := []string{"NAME", "DESTINATION", "FILTER", "WRITER_IDENTITY"}
			rows := make([][]string, len(sinks))
			for i, sink := range sinks {
				rows[i] = []string{sink.Name, sink.Destination, sink.Filter, sink.WriterIdentity}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Sink filter expression")
	cmd.Flags().StringVar(&parent, "parent", "", "Parent resource (projects/..., folders/..., organizations/..., billingAccounts/...)")
	return cmd
}

func newSinksDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe SINK_NAME",
		Short: "Describe a logging sink",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}
			client, err := NewClient(ctx, opt)
			if err != nil {
				return err
			}

			sink, err := client.GetSink(ctx, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), sink)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:              %s\n", sink.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Resource Name:     %s\n", sink.ResourceName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description:       %s\n", sink.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Destination:       %s\n", sink.Destination)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Filter:            %s\n", sink.Filter)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Disabled:          %t\n", sink.Disabled)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Include Children:   %t\n", sink.IncludeChildren)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Intercept Children: %t\n", sink.InterceptChildren)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Writer Identity:    %s\n", sink.WriterIdentity)
			return nil
		},
	}
}

func newSinksCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var parent string
	var destination string
	var filter string
	var description string
	var includeChildren bool
	var interceptChildren bool
	var uniqueWriterIdentity bool

	cmd := &cobra.Command{
		Use:   "create SINK_ID",
		Short: "Create a logging sink",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if parent == "" {
				parent, _ = requireParent(cmd, cfg)
			}
			if destination == "" {
				return fmt.Errorf("--destination is required")
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

			sink := &loggingapi.LogSink{
				Name:              args[0],
				Description:       description,
				Destination:       destination,
				Filter:            filter,
				IncludeChildren:   includeChildren,
				InterceptChildren: interceptChildren,
			}
			if uniqueWriterIdentity {
				sink.ForceSendFields = append(sink.ForceSendFields, "WriterIdentity")
			}

			created, err := client.CreateSink(ctx, parent, sink)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), created)
		},
	}
	cmd.Flags().StringVar(&parent, "parent", "", "Parent resource (projects/..., folders/..., organizations/..., billingAccounts/...)")
	cmd.Flags().StringVar(&destination, "destination", "", "Sink destination")
	cmd.Flags().StringVar(&filter, "filter", "", "Log filter")
	cmd.Flags().StringVar(&description, "description", "", "Sink description")
	cmd.Flags().BoolVar(&includeChildren, "include-children", false, "Include child resources")
	cmd.Flags().BoolVar(&interceptChildren, "intercept-children", false, "Intercept logs from child resources")
	cmd.Flags().BoolVar(&uniqueWriterIdentity, "unique-writer-identity", false, "Use a unique writer identity")
	return cmd
}

func newSinksDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete SINK_NAME",
		Short: "Delete a logging sink",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			opt, err := creds.ClientOption(ctx)
			if err != nil {
				return err
			}
			client, err := NewClient(ctx, opt)
			if err != nil {
				return err
			}
			if err := client.DeleteSink(ctx, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted sink %s.\n", args[0])
			return nil
		},
	}
}
