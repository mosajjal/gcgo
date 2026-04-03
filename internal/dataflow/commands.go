package dataflow

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the dataflow command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataflow",
		Short: "Manage Dataflow jobs",
	}

	cmd.AddCommand(newJobsCommand(cfg, creds))
	cmd.AddCommand(newSnapshotsCommand(cfg, creds))
	cmd.AddCommand(newFlexTemplatesCommand(cfg, creds))
	cmd.AddCommand(newTemplatesCommand(cfg, creds))

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

func requireRegion(cmd *cobra.Command, cfg *config.Config) (string, error) {
	region, _ := cmd.Flags().GetString("region")
	if region == "" {
		region = cfg.Region()
	}
	if region == "" {
		return "", fmt.Errorf("no region set (use --region or 'gcgo config set region REGION')")
	}
	return region, nil
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newJobsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage Dataflow jobs",
	}

	cmd.AddCommand(
		newJobsListCommand(cfg, creds),
		newJobsDescribeCommand(cfg, creds),
		newJobsMessagesCommand(cfg, creds),
		newJobsMetricsCommand(cfg, creds),
		newJobsCancelCommand(cfg, creds),
		newJobsDrainCommand(cfg, creds),
	)

	return cmd
}

func newJobsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Dataflow jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			jobs, err := client.ListJobs(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), jobs)
			}
			headers := []string{"ID", "NAME", "TYPE", "STATE", "CREATE_TIME", "REGION"}
			rows := make([][]string, len(jobs))
			for i, job := range jobs {
				rows[i] = []string{job.ID, job.Name, job.Type, job.State, job.CreateTime, job.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newJobsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe JOB",
		Short: "Describe a Dataflow job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			job, err := client.GetJob(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), job)
		},
	}
}

func newJobsMessagesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var minimumImportance, startTime, endTime string

	cmd := &cobra.Command{
		Use:   "messages JOB",
		Short: "List messages for a Dataflow job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			messages, err := client.ListMessages(ctx, project, region, args[0], minimumImportance, startTime, endTime)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), messages)
			}
			headers := []string{"TIME", "IMPORTANCE", "TEXT"}
			rows := make([][]string, len(messages))
			for i, message := range messages {
				rows[i] = []string{message.Time, message.Importance, message.Text}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&minimumImportance, "minimum-importance", "", "Minimum message importance")
	cmd.Flags().StringVar(&startTime, "start-time", "", "RFC3339 start time")
	cmd.Flags().StringVar(&endTime, "end-time", "", "RFC3339 end time")
	return cmd
}

func newJobsMetricsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var startTime string

	cmd := &cobra.Command{
		Use:   "metrics JOB",
		Short: "Summarize metrics for a Dataflow job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			metrics, err := client.GetMetrics(ctx, project, region, args[0], startTime)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), metrics)
		},
	}
	cmd.Flags().StringVar(&startTime, "start-time", "", "RFC3339 metrics start time")
	return cmd
}

func newJobsCancelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel JOB",
		Short: "Cancel a Dataflow job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CancelJob(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cancelled job %q.\n", args[0])
			return nil
		},
	}
}

func newJobsDrainCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "drain JOB",
		Short: "Drain a Dataflow job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DrainJob(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Drained job %q.\n", args[0])
			return nil
		},
	}
}

func newSnapshotsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Manage Dataflow snapshots",
	}
	cmd.AddCommand(
		newSnapshotsListCommand(cfg, creds),
		newSnapshotsDescribeCommand(cfg, creds),
		newSnapshotsCreateCommand(cfg, creds),
		newSnapshotsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newSnapshotsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var jobID string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Dataflow snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			snapshots, err := client.ListSnapshots(ctx, project, region, jobID)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), snapshots)
			}
			headers := []string{"ID", "JOB", "STATE", "CREATED", "REGION"}
			rows := make([][]string, len(snapshots))
			for i, snapshot := range snapshots {
				rows[i] = []string{snapshot.ID, snapshot.SourceJobID, snapshot.State, snapshot.CreationTime, snapshot.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&jobID, "job", "", "Filter snapshots by job ID")
	return cmd
}

func newSnapshotsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe SNAPSHOT",
		Short: "Describe a Dataflow snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			snapshot, err := client.GetSnapshot(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), snapshot)
		},
	}
}

func newSnapshotsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateSnapshotRequest

	cmd := &cobra.Command{
		Use:   "create JOB",
		Short: "Create a Dataflow snapshot from a job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			snapshot, err := client.CreateSnapshot(ctx, project, region, args[0], &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created snapshot %s.\n", snapshot.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Description, "description", "", "Snapshot description")
	cmd.Flags().StringVar(&req.TTL, "ttl", "", "Snapshot TTL duration")
	cmd.Flags().BoolVar(&req.SnapshotSources, "snapshot-sources", false, "Snapshot supported sources")
	return cmd
}

func newSnapshotsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete SNAPSHOT",
		Short: "Delete a Dataflow snapshot",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteSnapshot(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted snapshot %q.\n", args[0])
			return nil
		},
	}
}

func newFlexTemplatesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "flex-templates",
		Short: "Manage Dataflow flex templates",
	}
	cmd.AddCommand(newFlexTemplatesLaunchCommand(cfg, creds))
	return cmd
}

func newFlexTemplatesLaunchCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req LaunchFlexTemplateRequest
	var parameters []string

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch a Dataflow flex template",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.JobName == "" || req.ContainerSpecGCSPath == "" {
				return fmt.Errorf("--job-name and --container-spec-gcs-path are required")
			}
			req.Parameters = make(map[string]string, len(parameters))
			for _, param := range parameters {
				key, value, ok := strings.Cut(param, "=")
				if !ok || key == "" {
					return fmt.Errorf("invalid --param %q, expected key=value", param)
				}
				req.Parameters[key] = value
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			job, err := client.LaunchFlexTemplate(ctx, project, region, &req)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), job)
		},
	}
	cmd.Flags().StringVar(&req.JobName, "job-name", "", "Dataflow job name")
	cmd.Flags().StringVar(&req.ContainerSpecGCSPath, "container-spec-gcs-path", "", "GCS path to the flex template spec")
	cmd.Flags().StringArrayVar(&parameters, "param", nil, "Template parameter in key=value form")
	cmd.Flags().BoolVar(&req.ValidateOnly, "validate-only", false, "Validate launch request without starting the job")
	return cmd
}

func newTemplatesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "templates",
		Short: "Manage Dataflow classic templates",
	}
	cmd.AddCommand(
		newTemplatesGetCommand(cfg, creds),
		newTemplatesLaunchCommand(cfg, creds),
	)
	return cmd
}

func newTemplatesGetCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var gcsPath string

	cmd := &cobra.Command{
		Use:   "get",
		Short: "Get a classic Dataflow template",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if gcsPath == "" {
				return fmt.Errorf("--gcs-path is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			resp, err := client.GetTemplate(ctx, project, region, gcsPath)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), resp)
		},
	}
	cmd.Flags().StringVar(&gcsPath, "gcs-path", "", "GCS path to the template")
	return cmd
}

func newTemplatesLaunchCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var gcsPath string
	var jobName string
	var parameters []string
	var validateOnly bool

	cmd := &cobra.Command{
		Use:   "launch",
		Short: "Launch a classic Dataflow template",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if gcsPath == "" {
				return fmt.Errorf("--gcs-path is required")
			}
			if jobName == "" {
				return fmt.Errorf("--job-name is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			params := map[string]string{}
			for _, kv := range parameters {
				key, value, ok := strings.Cut(kv, "=")
				if !ok || key == "" {
					return fmt.Errorf("invalid --parameter %q, expected key=value", kv)
				}
				params[key] = value
			}
			resp, err := client.LaunchTemplate(ctx, project, region, &LaunchTemplateRequest{
				JobName:      jobName,
				GcsPath:      gcsPath,
				Parameters:   params,
				ValidateOnly: validateOnly,
			})
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), resp)
		},
	}
	cmd.Flags().StringVar(&gcsPath, "gcs-path", "", "GCS path to the template")
	cmd.Flags().StringVar(&jobName, "job-name", "", "Dataflow job name")
	cmd.Flags().StringArrayVar(&parameters, "parameter", nil, "Runtime parameter key=value")
	cmd.Flags().BoolVar(&validateOnly, "validate-only", false, "Validate only")
	return cmd
}
