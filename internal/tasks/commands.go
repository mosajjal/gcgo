package tasks

import (
	"context"
	"fmt"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the Cloud Tasks command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage Cloud Tasks",
	}

	cmd.AddCommand(
		newQueuesCommand(cfg, creds),
		newTasksCommand(cfg, creds),
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

func requireLocation(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("location")
	if flagVal != "" {
		return flagVal, nil
	}
	location := cfg.Region()
	if location == "" {
		return "", fmt.Errorf("--location is required (or set region in config)")
	}
	return location, nil
}

func tasksClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newQueuesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "queues",
		Short: "Manage queues",
	}
	cmd.AddCommand(
		newQueuesListCommand(cfg, creds),
		newQueuesDescribeCommand(cfg, creds),
		newQueuesCreateCommand(cfg, creds),
		newQueuesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newQueuesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List queues",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			queues, err := client.ListQueues(ctx, project, location)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), queues)
			}
			headers := []string{"NAME", "STATE", "MAX_QPS", "MAX_CONCURRENCY"}
			rows := make([][]string, len(queues))
			for i, queue := range queues {
				rows[i] = []string{
					queue.Name,
					queue.State,
					fmt.Sprintf("%.2f", queue.MaxDispatches),
					fmt.Sprintf("%d", queue.MaxConcurrent),
				}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	return cmd
}

func newQueuesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "describe QUEUE",
		Short: "Describe a queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			queue, err := client.GetQueue(ctx, project, location, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), queue)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:           %s\n", queue.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:          %s\n", queue.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Max QPS:        %.2f\n", queue.MaxDispatches)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Max Concurrent: %d\n", queue.MaxConcurrent)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Max Attempts:   %d\n", queue.MaxAttempts)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	return cmd
}

func newQueuesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var maxDispatches float64
	var maxConcurrent int64
	var maxAttempts int64

	cmd := &cobra.Command{
		Use:   "create QUEUE",
		Short: "Create a queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			queue, err := client.CreateQueue(ctx, project, location, &CreateQueueRequest{
				Name:                  args[0],
				MaxDispatchesPerSec:   maxDispatches,
				MaxConcurrentDispatch: maxConcurrent,
				MaxAttempts:           maxAttempts,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created queue %s.\n", queue.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	cmd.Flags().Float64Var(&maxDispatches, "max-dispatches-per-second", 0, "Maximum dispatches per second")
	cmd.Flags().Int64Var(&maxConcurrent, "max-concurrent-dispatches", 0, "Maximum concurrent dispatches")
	cmd.Flags().Int64Var(&maxAttempts, "max-attempts", 0, "Maximum retry attempts")
	return cmd
}

func newQueuesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "delete QUEUE",
		Short: "Delete a queue",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteQueue(ctx, project, location, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted queue %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	return cmd
}

func newTasksCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "tasks",
		Short: "Manage tasks in a queue",
	}
	cmd.AddCommand(
		newTasksListCommand(cfg, creds),
		newTasksDescribeCommand(cfg, creds),
		newTasksCreateCommand(cfg, creds),
		newTasksDeleteCommand(cfg, creds),
	)
	return cmd
}

func newTasksListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var queue string
	cmd := &cobra.Command{
		Use:   "list --queue=QUEUE",
		Short: "List tasks in a queue",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if queue == "" {
				return fmt.Errorf("--queue is required")
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			tasks, err := client.ListTasks(ctx, project, location, queue)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), tasks)
			}
			headers := []string{"NAME", "METHOD", "URL", "SCHEDULE", "DISPATCHES"}
			rows := make([][]string, len(tasks))
			for i, task := range tasks {
				rows[i] = []string{task.Name, task.Method, task.URL, task.ScheduleTime, fmt.Sprintf("%d", task.DispatchCount)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	cmd.Flags().StringVar(&queue, "queue", "", "Queue name")
	return cmd
}

func newTasksDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var queue string
	cmd := &cobra.Command{
		Use:   "describe TASK --queue=QUEUE",
		Short: "Describe a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if queue == "" {
				return fmt.Errorf("--queue is required")
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			task, err := client.GetTask(ctx, project, location, queue, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), task)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", task.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Method:       %s\n", task.Method)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "URL:          %s\n", task.URL)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ScheduleTime: %s\n", task.ScheduleTime)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "CreateTime:   %s\n", task.CreateTime)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Dispatches:   %d\n", task.DispatchCount)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Body:         %s\n", task.Body)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	cmd.Flags().StringVar(&queue, "queue", "", "Queue name")
	return cmd
}

func newTasksCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var queue string
	var url string
	var method string
	var body string
	var scheduleTime string
	var headers []string
	var serviceAccount string

	cmd := &cobra.Command{
		Use:   "create TASK --queue=QUEUE --url=URL",
		Short: "Create a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if queue == "" {
				return fmt.Errorf("--queue is required")
			}
			if url == "" {
				return fmt.Errorf("--url is required")
			}
			parsedHeaders, err := parseHeaders(headers)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			task, err := client.CreateTask(ctx, project, location, queue, &CreateTaskRequest{
				Name:           args[0],
				URL:            url,
				Method:         strings.ToUpper(method),
				Body:           body,
				Headers:        parsedHeaders,
				ScheduleTime:   scheduleTime,
				ServiceAccount: serviceAccount,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created task %s.\n", task.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	cmd.Flags().StringVar(&queue, "queue", "", "Queue name")
	cmd.Flags().StringVar(&url, "url", "", "HTTP target URL")
	cmd.Flags().StringVar(&method, "method", "POST", "HTTP method")
	cmd.Flags().StringVar(&body, "body", "", "HTTP request body")
	cmd.Flags().StringVar(&scheduleTime, "schedule-time", "", "RFC3339 schedule time")
	cmd.Flags().StringArrayVar(&headers, "header", nil, "HTTP header as KEY=VALUE")
	cmd.Flags().StringVar(&serviceAccount, "service-account", "", "OIDC service account email")
	return cmd
}

func newTasksDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var queue string
	cmd := &cobra.Command{
		Use:   "delete TASK --queue=QUEUE",
		Short: "Delete a task",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if queue == "" {
				return fmt.Errorf("--queue is required")
			}
			ctx := context.Background()
			client, err := tasksClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteTask(ctx, project, location, queue, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted task %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Cloud Tasks location")
	cmd.Flags().StringVar(&queue, "queue", "", "Queue name")
	return cmd
}

func parseHeaders(items []string) (map[string]string, error) {
	headers := make(map[string]string, len(items))
	for _, item := range items {
		key, value, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			return nil, fmt.Errorf("invalid header %q, expected KEY=VALUE", item)
		}
		headers[key] = value
	}
	return headers, nil
}
