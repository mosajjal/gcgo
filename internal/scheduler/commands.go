package scheduler

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the scheduler command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "scheduler",
		Short: "Manage Cloud Scheduler",
	}

	cmd.AddCommand(newJobsCommand(cfg, creds))
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

func schedulerClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newJobsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage Cloud Scheduler jobs",
	}

	cmd.AddCommand(
		newJobsListCommand(cfg, creds),
		newJobsDescribeCommand(cfg, creds),
		newJobsCreateCommand(cfg, creds),
		newJobsDeleteCommand(cfg, creds),
		newJobsPauseCommand(cfg, creds),
		newJobsResumeCommand(cfg, creds),
		newJobsRunCommand(cfg, creds),
	)

	return cmd
}

func newJobsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List scheduler jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := schedulerClient(ctx, creds)
			if err != nil {
				return err
			}

			jobs, err := client.ListJobs(ctx, project, location)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), jobs)
			}

			headers := []string{"NAME", "SCHEDULE", "TIME_ZONE", "STATE"}
			rows := make([][]string, len(jobs))
			for i, j := range jobs {
				rows[i] = []string{j.Name, j.Schedule, j.TimeZone, j.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location")
	return cmd
}

func newJobsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "describe JOB",
		Short: "Describe a scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			ctx := context.Background()
			client, err := schedulerClient(ctx, creds)
			if err != nil {
				return err
			}

			job, err := client.GetJob(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), job)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", job.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Schedule:    %s\n", job.Schedule)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Time Zone:   %s\n", job.TimeZone)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", job.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "HTTP Target: %s\n", job.HTTPTarget)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location")
	return cmd
}

func newJobsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location string
		req      CreateJobRequest
	)

	cmd := &cobra.Command{
		Use:   "create JOB --schedule=CRON --uri=URI",
		Short: "Create a scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			req.Name = args[0]

			ctx := context.Background()
			client, err := schedulerClient(ctx, creds)
			if err != nil {
				return err
			}

			job, err := client.CreateJob(ctx, project, location, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created job %s.\n", job.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location")
	cmd.Flags().StringVar(&req.Schedule, "schedule", "", "Cron schedule (e.g. '*/5 * * * *')")
	cmd.Flags().StringVar(&req.URI, "uri", "", "HTTP target URI")
	cmd.Flags().StringVar(&req.HTTPMethod, "http-method", "POST", "HTTP method (GET, POST, PUT, DELETE, PATCH)")
	cmd.Flags().StringVar(&req.Body, "body", "", "HTTP request body")
	cmd.Flags().StringVar(&req.TimeZone, "time-zone", "UTC", "Time zone (e.g. 'America/New_York')")
	_ = cmd.MarkFlagRequired("schedule")
	_ = cmd.MarkFlagRequired("uri")

	return cmd
}

func newJobsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "delete JOB",
		Short: "Delete a scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			ctx := context.Background()
			client, err := schedulerClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteJob(ctx, project, location, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted job %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location")
	return cmd
}

func newJobsPauseCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "pause JOB",
		Short: "Pause a scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			ctx := context.Background()
			client, err := schedulerClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.PauseJob(ctx, project, location, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Paused job %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location")
	return cmd
}

func newJobsResumeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "resume JOB",
		Short: "Resume a scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			ctx := context.Background()
			client, err := schedulerClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.ResumeJob(ctx, project, location, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Resumed job %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location")
	return cmd
}

func newJobsRunCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "run JOB",
		Short: "Run a scheduler job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}

			ctx := context.Background()
			client, err := schedulerClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.RunJob(ctx, project, location, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Ran job %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Location")
	return cmd
}
