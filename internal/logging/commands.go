package logging

import (
	"context"
	"fmt"
	"strconv"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
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
		newMetricsCommand(cfg, creds),
		newExclusionsCommand(cfg, creds),
		newBucketsCommand(cfg, creds),
	)

	return cmd
}

func newLoggingAPIService(ctx context.Context, creds *auth.Credentials) (*loggingapi.Service, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	svc, err := loggingapi.NewService(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("create logging api service: %w", err)
	}
	return svc, nil
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

// — Metrics —

func newMetricsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "metrics",
		Short: "Manage logs-based metrics",
	}
	cmd.AddCommand(
		newMetricsListCommand(cfg, creds),
		newMetricsDescribeCommand(cfg, creds),
		newMetricsCreateCommand(cfg, creds),
		newMetricsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newMetricsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List logs-based metrics",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			parent := "projects/" + project
			var metrics []*loggingapi.LogMetric
			if err := svc.Projects.Metrics.List(parent).Context(ctx).Pages(ctx, func(resp *loggingapi.ListLogMetricsResponse) error {
				metrics = append(metrics, resp.Metrics...)
				return nil
			}); err != nil {
				return fmt.Errorf("list log metrics: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), metrics)
			}

			headers := []string{"NAME", "DESCRIPTION", "FILTER"}
			rows := make([][]string, len(metrics))
			for i, m := range metrics {
				rows[i] = []string{m.Name, m.Description, m.Filter}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	return cmd
}

func newMetricsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe METRIC_NAME",
		Short: "Describe a logs-based metric",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/metrics/" + args[0]
			metric, err := svc.Projects.Metrics.Get(name).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("describe log metric %s: %w", args[0], err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), metric)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", metric.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", metric.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Filter:      %s\n", metric.Filter)
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	return cmd
}

func newMetricsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var description string
	var filter string

	cmd := &cobra.Command{
		Use:   "create METRIC_NAME",
		Short: "Create a logs-based metric",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if filter == "" {
				return fmt.Errorf("--filter is required")
			}

			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			parent := "projects/" + project
			metric := &loggingapi.LogMetric{
				Name:        args[0],
				Description: description,
				Filter:      filter,
			}
			created, err := svc.Projects.Metrics.Create(parent, metric).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("create log metric %s: %w", args[0], err)
			}
			return output.PrintJSON(cmd.OutOrStdout(), created)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().StringVar(&description, "description", "", "Metric description")
	cmd.Flags().StringVar(&filter, "filter", "", "Log filter expression (required)")
	return cmd
}

func newMetricsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete METRIC_NAME",
		Short: "Delete a logs-based metric",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/metrics/" + args[0]
			if _, err := svc.Projects.Metrics.Delete(name).Context(ctx).Do(); err != nil {
				return fmt.Errorf("delete log metric %s: %w", args[0], err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted metric %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	return cmd
}

// — Exclusions —

func newExclusionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "exclusions",
		Short: "Manage log exclusions",
	}
	cmd.AddCommand(
		newExclusionsListCommand(cfg, creds),
		newExclusionsDescribeCommand(cfg, creds),
		newExclusionsCreateCommand(cfg, creds),
		newExclusionsUpdateCommand(cfg, creds),
		newExclusionsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newExclusionsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List log exclusions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			parent := "projects/" + project
			var exclusions []*loggingapi.LogExclusion
			if err := svc.Projects.Exclusions.List(parent).Context(ctx).Pages(ctx, func(resp *loggingapi.ListExclusionsResponse) error {
				exclusions = append(exclusions, resp.Exclusions...)
				return nil
			}); err != nil {
				return fmt.Errorf("list log exclusions: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), exclusions)
			}

			headers := []string{"NAME", "DESCRIPTION", "FILTER", "DISABLED"}
			rows := make([][]string, len(exclusions))
			for i, e := range exclusions {
				rows[i] = []string{e.Name, e.Description, e.Filter, strconv.FormatBool(e.Disabled)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	return cmd
}

func newExclusionsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe EXCLUSION_NAME",
		Short: "Describe a log exclusion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/exclusions/" + args[0]
			excl, err := svc.Projects.Exclusions.Get(name).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("describe log exclusion %s: %w", args[0], err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), excl)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", excl.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description: %s\n", excl.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Filter:      %s\n", excl.Filter)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Disabled:    %t\n", excl.Disabled)
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	return cmd
}

func newExclusionsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var description string
	var filter string
	var disabled bool

	cmd := &cobra.Command{
		Use:   "create EXCLUSION_NAME",
		Short: "Create a log exclusion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if filter == "" {
				return fmt.Errorf("--filter is required")
			}

			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			parent := "projects/" + project
			excl := &loggingapi.LogExclusion{
				Name:        args[0],
				Description: description,
				Filter:      filter,
				Disabled:    disabled,
			}
			created, err := svc.Projects.Exclusions.Create(parent, excl).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("create log exclusion %s: %w", args[0], err)
			}
			return output.PrintJSON(cmd.OutOrStdout(), created)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().StringVar(&description, "description", "", "Exclusion description")
	cmd.Flags().StringVar(&filter, "filter", "", "Log filter expression (required)")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Create exclusion in disabled state")
	return cmd
}

func newExclusionsUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var description string
	var filter string
	var disabled bool

	cmd := &cobra.Command{
		Use:   "update EXCLUSION_NAME",
		Short: "Update a log exclusion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/exclusions/" + args[0]
			excl := &loggingapi.LogExclusion{
				Description: description,
				Filter:      filter,
				Disabled:    disabled,
			}
			updated, err := svc.Projects.Exclusions.Patch(name, excl).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("update log exclusion %s: %w", args[0], err)
			}
			return output.PrintJSON(cmd.OutOrStdout(), updated)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().StringVar(&description, "description", "", "Exclusion description")
	cmd.Flags().StringVar(&filter, "filter", "", "Log filter expression")
	cmd.Flags().BoolVar(&disabled, "disabled", false, "Disable the exclusion")
	return cmd
}

func newExclusionsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete EXCLUSION_NAME",
		Short: "Delete a log exclusion",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/exclusions/" + args[0]
			if _, err := svc.Projects.Exclusions.Delete(name).Context(ctx).Do(); err != nil {
				return fmt.Errorf("delete log exclusion %s: %w", args[0], err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted exclusion %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	return cmd
}

// — Buckets —

func newBucketsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "buckets",
		Short: "Manage log buckets",
	}
	cmd.AddCommand(
		newBucketsListCommand(cfg, creds),
		newBucketsDescribeCommand(cfg, creds),
		newBucketsCreateCommand(cfg, creds),
		newBucketsUpdateCommand(cfg, creds),
		newBucketsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newBucketsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List log buckets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			// Use "-" as location wildcard to list buckets across all locations.
			parent := "projects/" + project + "/locations/-"
			var buckets []*loggingapi.LogBucket
			if err := svc.Projects.Locations.Buckets.List(parent).Context(ctx).Pages(ctx, func(resp *loggingapi.ListBucketsResponse) error {
				buckets = append(buckets, resp.Buckets...)
				return nil
			}); err != nil {
				return fmt.Errorf("list log buckets: %w", err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), buckets)
			}

			headers := []string{"NAME", "LOCATION", "RETENTION_DAYS", "LOCKED"}
			rows := make([][]string, len(buckets))
			for i, b := range buckets {
				rows[i] = []string{b.Name, locationFromBucketName(b.Name), strconv.FormatInt(b.RetentionDays, 10), strconv.FormatBool(b.Locked)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	return cmd
}

func newBucketsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "describe BUCKET_ID",
		Short: "Describe a log bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/locations/" + location + "/buckets/" + args[0]
			bucket, err := svc.Projects.Locations.Buckets.Get(name).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("describe log bucket %s: %w", args[0], err)
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), bucket)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:          %s\n", bucket.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Description:   %s\n", bucket.Description)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "RetentionDays: %d\n", bucket.RetentionDays)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Locked:        %t\n", bucket.Locked)
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().StringVar(&location, "location", "", "Bucket location (required)")
	return cmd
}

func newBucketsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var description string
	var location string
	var retentionDays int64

	cmd := &cobra.Command{
		Use:   "create BUCKET_ID",
		Short: "Create a log bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			parent := "projects/" + project + "/locations/" + location
			bucket := &loggingapi.LogBucket{
				Description:   description,
				RetentionDays: retentionDays,
			}
			created, err := svc.Projects.Locations.Buckets.Create(parent, bucket).BucketId(args[0]).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("create log bucket %s: %w", args[0], err)
			}
			return output.PrintJSON(cmd.OutOrStdout(), created)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().StringVar(&location, "location", "", "Bucket location (required)")
	cmd.Flags().StringVar(&description, "description", "", "Bucket description")
	cmd.Flags().Int64Var(&retentionDays, "retention-days", 30, "Log retention period in days")
	return cmd
}

func newBucketsUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var description string
	var location string
	var retentionDays int64

	cmd := &cobra.Command{
		Use:   "update BUCKET_ID",
		Short: "Update a log bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/locations/" + location + "/buckets/" + args[0]
			bucket := &loggingapi.LogBucket{
				Description:   description,
				RetentionDays: retentionDays,
			}
			updated, err := svc.Projects.Locations.Buckets.Patch(name, bucket).Context(ctx).Do()
			if err != nil {
				return fmt.Errorf("update log bucket %s: %w", args[0], err)
			}
			return output.PrintJSON(cmd.OutOrStdout(), updated)
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().StringVar(&location, "location", "", "Bucket location (required)")
	cmd.Flags().StringVar(&description, "description", "", "Bucket description")
	cmd.Flags().Int64Var(&retentionDays, "retention-days", 0, "Log retention period in days (0 = keep existing)")
	return cmd
}

func newBucketsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "delete BUCKET_ID",
		Short: "Delete a log bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			svc, err := newLoggingAPIService(ctx, creds)
			if err != nil {
				return err
			}

			name := "projects/" + project + "/locations/" + location + "/buckets/" + args[0]
			if _, err := svc.Projects.Locations.Buckets.Delete(name).Context(ctx).Do(); err != nil {
				return fmt.Errorf("delete log bucket %s: %w", args[0], err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted bucket %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().String("project", "", "Project ID")
	cmd.Flags().StringVar(&location, "location", "", "Bucket location (required)")
	return cmd
}

// locationFromBucketName extracts the location segment from a full bucket resource name.
// Format: projects/PROJECT/locations/LOCATION/buckets/BUCKET_ID
func locationFromBucketName(name string) string {
	var parts []string
	start := 0
	for i := 0; i < len(name); i++ {
		if name[i] == '/' {
			parts = append(parts, name[start:i])
			start = i + 1
		}
	}
	parts = append(parts, name[start:])
	for i, p := range parts {
		if p == "locations" && i+1 < len(parts) {
			return parts[i+1]
		}
	}
	return ""
}
