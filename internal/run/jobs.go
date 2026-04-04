package run

import (
	"context"
	"errors"
	"fmt"
	"strings"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Job holds Cloud Run job fields.
type Job struct {
	Name           string `json:"name"`
	Region         string `json:"region"`
	Image          string `json:"image"`
	ExecutionCount int32  `json:"execution_count"`
	Reconciling    bool   `json:"reconciling"`
}

// Execution holds Cloud Run execution fields.
type Execution struct {
	Name           string `json:"name"`
	Job            string `json:"job"`
	Status         string `json:"status"`
	CreateTime     string `json:"create_time"`
	RunningCount   int32  `json:"running_count"`
	SucceededCount int32  `json:"succeeded_count"`
	FailedCount    int32  `json:"failed_count"`
}

// CreateJobRequest holds parameters for creating a job.
type CreateJobRequest struct {
	Name  string
	Image string
	Env   map[string]string
}

// JobsClient defines Cloud Run jobs operations.
type JobsClient interface {
	ListJobs(ctx context.Context, project, region string) ([]*Job, error)
	GetJob(ctx context.Context, project, region, name string) (*Job, error)
	CreateJob(ctx context.Context, project, region string, req *CreateJobRequest) error
	DeleteJob(ctx context.Context, project, region, name string) error
	ExecuteJob(ctx context.Context, project, region, name string) error
}

// ExecutionsClient defines Cloud Run executions operations.
type ExecutionsClient interface {
	ListExecutions(ctx context.Context, project, region, job string) ([]*Execution, error)
	GetExecution(ctx context.Context, project, region, job, execution string) (*Execution, error)
	CancelExecution(ctx context.Context, project, region, job, execution string) error
}

type gcpJobsClient struct {
	jc *run.JobsClient
}

type gcpExecutionsClient struct {
	ec *run.ExecutionsClient
}

// NewJobsClient creates a JobsClient backed by the real Cloud Run API.
func NewJobsClient(ctx context.Context, opts ...option.ClientOption) (JobsClient, error) {
	jc, err := run.NewJobsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud run jobs client: %w", err)
	}
	return &gcpJobsClient{jc: jc}, nil
}

// NewExecutionsClient creates an ExecutionsClient backed by the real Cloud Run API.
func NewExecutionsClient(ctx context.Context, opts ...option.ClientOption) (ExecutionsClient, error) {
	ec, err := run.NewExecutionsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud run executions client: %w", err)
	}
	return &gcpExecutionsClient{ec: ec}, nil
}

func jobsClient(ctx context.Context, creds *auth.Credentials) (JobsClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewJobsClient(ctx, opt)
}

func executionsClient(ctx context.Context, creds *auth.Credentials) (ExecutionsClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewExecutionsClient(ctx, opt)
}

func (c *gcpJobsClient) ListJobs(ctx context.Context, project, region string) ([]*Job, error) {
	it := c.jc.ListJobs(ctx, &runpb.ListJobsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, region),
	})

	var jobs []*Job
	for {
		j, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list jobs: %w", err)
		}
		jobs = append(jobs, jobFromProto(j, region))
	}
	return jobs, nil
}

func (c *gcpJobsClient) GetJob(ctx context.Context, project, region, name string) (*Job, error) {
	j, err := c.jc.GetJob(ctx, &runpb.GetJobRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, region, name),
	})
	if err != nil {
		return nil, fmt.Errorf("get job %s: %w", name, err)
	}
	return jobFromProto(j, region), nil
}

func (c *gcpJobsClient) CreateJob(ctx context.Context, project, region string, req *CreateJobRequest) error {
	var envVars []*runpb.EnvVar
	for k, v := range req.Env {
		envVars = append(envVars, &runpb.EnvVar{
			Name:   k,
			Values: &runpb.EnvVar_Value{Value: v},
		})
	}

	op, err := c.jc.CreateJob(ctx, &runpb.CreateJobRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, region),
		JobId:  req.Name,
		Job: &runpb.Job{
			Template: &runpb.ExecutionTemplate{
				Template: &runpb.TaskTemplate{
					Containers: []*runpb.Container{
						{
							Image: req.Image,
							Env:   envVars,
						},
					},
				},
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create job %s: %w", req.Name, err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for create job %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpJobsClient) DeleteJob(ctx context.Context, project, region, name string) error {
	op, err := c.jc.DeleteJob(ctx, &runpb.DeleteJobRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, region, name),
	})
	if err != nil {
		return fmt.Errorf("delete job %s: %w", name, err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for delete job %s: %w", name, err)
	}
	return nil
}

func (c *gcpJobsClient) ExecuteJob(ctx context.Context, project, region, name string) error {
	op, err := c.jc.RunJob(ctx, &runpb.RunJobRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, region, name),
	})
	if err != nil {
		return fmt.Errorf("execute job %s: %w", name, err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for execute job %s: %w", name, err)
	}
	return nil
}

func (c *gcpExecutionsClient) ListExecutions(ctx context.Context, project, region, job string) ([]*Execution, error) {
	it := c.ec.ListExecutions(ctx, &runpb.ListExecutionsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, region, job),
	})

	var execs []*Execution
	for {
		e, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list executions: %w", err)
		}
		execs = append(execs, execFromProto(e))
	}
	return execs, nil
}

func (c *gcpExecutionsClient) GetExecution(ctx context.Context, project, region, job, execution string) (*Execution, error) {
	e, err := c.ec.GetExecution(ctx, &runpb.GetExecutionRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s/executions/%s", project, region, job, execution),
	})
	if err != nil {
		return nil, fmt.Errorf("get execution %s: %w", execution, err)
	}
	return execFromProto(e), nil
}

func (c *gcpExecutionsClient) CancelExecution(ctx context.Context, project, region, job, execution string) error {
	op, err := c.ec.CancelExecution(ctx, &runpb.CancelExecutionRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/jobs/%s/executions/%s", project, region, job, execution),
	})
	if err != nil {
		return fmt.Errorf("cancel execution %s: %w", execution, err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for cancel execution %s: %w", execution, err)
	}
	return nil
}

func jobFromProto(j *runpb.Job, region string) *Job {
	image := ""
	if j.GetTemplate() != nil && j.GetTemplate().GetTemplate() != nil {
		containers := j.GetTemplate().GetTemplate().GetContainers()
		if len(containers) > 0 {
			image = containers[0].GetImage()
		}
	}
	return &Job{
		Name:           j.GetName(),
		Region:         region,
		Image:          image,
		ExecutionCount: j.GetExecutionCount(),
		Reconciling:    j.GetReconciling(),
	}
}

func execFromProto(e *runpb.Execution) *Execution {
	var createTime string
	if e.GetCreateTime() != nil {
		createTime = e.GetCreateTime().AsTime().String()
	}
	var status string
	if e.GetCompletionTime() != nil {
		status = "COMPLETED"
	} else if e.GetDeleteTime() != nil {
		status = "CANCELLED"
	} else {
		status = "RUNNING"
	}
	return &Execution{
		Name:           e.GetName(),
		Job:            e.GetJob(),
		Status:         status,
		CreateTime:     createTime,
		RunningCount:   e.GetRunningCount(),
		SucceededCount: e.GetSucceededCount(),
		FailedCount:    e.GetFailedCount(),
	}
}

func newJobsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage Cloud Run jobs",
	}

	cmd.AddCommand(
		newJobsListCommand(cfg, creds),
		newJobsDescribeCommand(cfg, creds),
		newJobsCreateCommand(cfg, creds),
		newJobsDeleteCommand(cfg, creds),
		newJobsRunCommand(cfg, creds),
		newExecutionsCommand(cfg, creds),
	)

	return cmd
}

func newExecutionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "executions",
		Short: "Manage Cloud Run job executions",
	}

	cmd.AddCommand(
		newExecutionsListCommand(cfg, creds),
		newExecutionsDescribeCommand(cfg, creds),
		newExecutionsCancelCommand(cfg, creds),
	)

	return cmd
}

func newJobsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloud Run jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := jobsClient(ctx, creds)
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

			headers := []string{"NAME", "REGION", "IMAGE", "EXECUTIONS", "RECONCILING"}
			rows := make([][]string, len(jobs))
			for i, job := range jobs {
				rows[i] = []string{
					job.Name,
					job.Region,
					job.Image,
					fmt.Sprintf("%d", job.ExecutionCount),
					fmt.Sprintf("%v", job.Reconciling),
				}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newJobsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "describe JOB",
		Short: "Describe a Cloud Run job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := jobsClient(ctx, creds)
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

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newJobsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var envSlice []string
	var req CreateJobRequest

	cmd := &cobra.Command{
		Use:   "create JOB",
		Short: "Create a Cloud Run job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			req.Name = args[0]
			req.Env = make(map[string]string)
			for _, e := range envSlice {
				k, v, ok := strings.Cut(e, "=")
				if !ok {
					return fmt.Errorf("invalid env format %q (expected KEY=VALUE)", e)
				}
				req.Env[k] = v
			}

			ctx := context.Background()
			client, err := jobsClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.CreateJob(ctx, project, region, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created job %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&req.Image, "image", "", "Container image")
	cmd.Flags().StringSliceVar(&envSlice, "env", nil, "Environment variables (KEY=VALUE)")
	_ = cmd.MarkFlagRequired("image")

	return cmd
}

func newJobsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "delete JOB",
		Short: "Delete a Cloud Run job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := jobsClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteJob(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted job %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newJobsRunCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "run JOB",
		Short: "Execute a Cloud Run job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := jobsClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.ExecuteJob(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started job %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newExecutionsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "list JOB",
		Short: "List executions for a Cloud Run job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := executionsClient(ctx, creds)
			if err != nil {
				return err
			}

			execs, err := client.ListExecutions(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), execs)
			}

			headers := []string{"NAME", "JOB", "STATUS", "CREATE_TIME"}
			rows := make([][]string, len(execs))
			for i, exec := range execs {
				rows[i] = []string{exec.Name, exec.Job, exec.Status, exec.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newExecutionsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "describe JOB EXECUTION",
		Short: "Describe a Cloud Run job execution",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := executionsClient(ctx, creds)
			if err != nil {
				return err
			}

			exec, err := client.GetExecution(ctx, project, region, args[0], args[1])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), exec)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newExecutionsCancelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "cancel JOB EXECUTION",
		Short: "Cancel a Cloud Run job execution",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := executionsClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.CancelExecution(ctx, project, region, args[0], args[1]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cancelled execution %s for job %s.\n", args[1], args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}
