package tasks

import (
	"context"
	"encoding/base64"
	"fmt"

	cloudtasks "google.golang.org/api/cloudtasks/v2"
	"google.golang.org/api/option"
)

// Queue holds the fields we display.
type Queue struct {
	Name          string  `json:"name"`
	State         string  `json:"state"`
	MaxDispatches float64 `json:"max_dispatches_per_second,omitempty"`
	MaxConcurrent int64   `json:"max_concurrent_dispatches,omitempty"`
	MaxAttempts   int64   `json:"max_attempts,omitempty"`
}

// Task holds the fields we display.
type Task struct {
	Name          string            `json:"name"`
	Method        string            `json:"method,omitempty"`
	URL           string            `json:"url,omitempty"`
	Headers       map[string]string `json:"headers,omitempty"`
	Body          string            `json:"body,omitempty"`
	ScheduleTime  string            `json:"schedule_time,omitempty"`
	CreateTime    string            `json:"create_time,omitempty"`
	DispatchCount int64             `json:"dispatch_count,omitempty"`
}

// CreateQueueRequest holds queue creation parameters.
type CreateQueueRequest struct {
	Name                  string
	MaxDispatchesPerSec   float64
	MaxConcurrentDispatch int64
	MaxAttempts           int64
}

// CreateTaskRequest holds task creation parameters.
type CreateTaskRequest struct {
	Name           string
	URL            string
	Method         string
	Body           string
	Headers        map[string]string
	ScheduleTime   string
	ServiceAccount string
}

// Client defines Cloud Tasks operations.
type Client interface {
	ListQueues(ctx context.Context, project, location string) ([]*Queue, error)
	GetQueue(ctx context.Context, project, location, name string) (*Queue, error)
	CreateQueue(ctx context.Context, project, location string, req *CreateQueueRequest) (*Queue, error)
	DeleteQueue(ctx context.Context, project, location, name string) error

	ListTasks(ctx context.Context, project, location, queue string) ([]*Task, error)
	GetTask(ctx context.Context, project, location, queue, name string) (*Task, error)
	CreateTask(ctx context.Context, project, location, queue string, req *CreateTaskRequest) (*Task, error)
	DeleteTask(ctx context.Context, project, location, queue, name string) error
}

type gcpClient struct {
	svc *cloudtasks.Service
}

// NewClient creates a Client backed by the real Cloud Tasks API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := cloudtasks.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud tasks client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListQueues(ctx context.Context, project, location string) ([]*Queue, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.svc.Projects.Locations.Queues.List(parent).Context(ctx)

	var queues []*Queue
	if err := call.Pages(ctx, func(resp *cloudtasks.ListQueuesResponse) error {
		for _, queue := range resp.Queues {
			queues = append(queues, queueFromAPI(queue))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list queues: %w", err)
	}
	return queues, nil
}

func (c *gcpClient) GetQueue(ctx context.Context, project, location, name string) (*Queue, error) {
	fullName := queueName(project, location, name)
	queue, err := c.svc.Projects.Locations.Queues.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get queue %s: %w", name, err)
	}
	return queueFromAPI(queue), nil
}

func (c *gcpClient) CreateQueue(ctx context.Context, project, location string, req *CreateQueueRequest) (*Queue, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	queue := &cloudtasks.Queue{
		Name: fmt.Sprintf("%s/queues/%s", parent, req.Name),
	}
	if req.MaxDispatchesPerSec > 0 || req.MaxConcurrentDispatch > 0 {
		queue.RateLimits = &cloudtasks.RateLimits{
			MaxDispatchesPerSecond:  req.MaxDispatchesPerSec,
			MaxConcurrentDispatches: req.MaxConcurrentDispatch,
		}
	}
	if req.MaxAttempts > 0 {
		queue.RetryConfig = &cloudtasks.RetryConfig{MaxAttempts: req.MaxAttempts}
	}
	created, err := c.svc.Projects.Locations.Queues.Create(parent, queue).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create queue %s: %w", req.Name, err)
	}
	return queueFromAPI(created), nil
}

func (c *gcpClient) DeleteQueue(ctx context.Context, project, location, name string) error {
	fullName := queueName(project, location, name)
	if _, err := c.svc.Projects.Locations.Queues.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete queue %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListTasks(ctx context.Context, project, location, queue string) ([]*Task, error) {
	parent := queueName(project, location, queue)
	call := c.svc.Projects.Locations.Queues.Tasks.List(parent).ResponseView("FULL").Context(ctx)

	var tasks []*Task
	if err := call.Pages(ctx, func(resp *cloudtasks.ListTasksResponse) error {
		for _, task := range resp.Tasks {
			tasks = append(tasks, taskFromAPI(task))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list tasks: %w", err)
	}
	return tasks, nil
}

func (c *gcpClient) GetTask(ctx context.Context, project, location, queue, name string) (*Task, error) {
	fullName := taskName(project, location, queue, name)
	task, err := c.svc.Projects.Locations.Queues.Tasks.Get(fullName).ResponseView("FULL").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get task %s: %w", name, err)
	}
	return taskFromAPI(task), nil
}

func (c *gcpClient) CreateTask(ctx context.Context, project, location, queue string, req *CreateTaskRequest) (*Task, error) {
	parent := queueName(project, location, queue)
	task := &cloudtasks.Task{
		HttpRequest: &cloudtasks.HttpRequest{
			HttpMethod: req.Method,
			Url:        req.URL,
			Headers:    req.Headers,
		},
		ScheduleTime: req.ScheduleTime,
	}
	if req.Name != "" {
		task.Name = taskName(project, location, queue, req.Name)
	}
	if req.Body != "" {
		task.HttpRequest.Body = base64.StdEncoding.EncodeToString([]byte(req.Body))
	}
	if req.ServiceAccount != "" {
		task.HttpRequest.OidcToken = &cloudtasks.OidcToken{ServiceAccountEmail: req.ServiceAccount}
	}

	created, err := c.svc.Projects.Locations.Queues.Tasks.Create(parent, &cloudtasks.CreateTaskRequest{
		ResponseView: "FULL",
		Task:         task,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create task %s: %w", req.Name, err)
	}
	return taskFromAPI(created), nil
}

func (c *gcpClient) DeleteTask(ctx context.Context, project, location, queue, name string) error {
	fullName := taskName(project, location, queue, name)
	if _, err := c.svc.Projects.Locations.Queues.Tasks.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete task %s: %w", name, err)
	}
	return nil
}

func queueFromAPI(queue *cloudtasks.Queue) *Queue {
	if queue == nil {
		return nil
	}
	item := &Queue{Name: queue.Name, State: queue.State}
	if queue.RateLimits != nil {
		item.MaxDispatches = queue.RateLimits.MaxDispatchesPerSecond
		item.MaxConcurrent = queue.RateLimits.MaxConcurrentDispatches
	}
	if queue.RetryConfig != nil {
		item.MaxAttempts = queue.RetryConfig.MaxAttempts
	}
	return item
}

func taskFromAPI(task *cloudtasks.Task) *Task {
	if task == nil {
		return nil
	}
	item := &Task{
		Name:          task.Name,
		ScheduleTime:  task.ScheduleTime,
		CreateTime:    task.CreateTime,
		DispatchCount: task.DispatchCount,
	}
	if task.HttpRequest != nil {
		item.Method = task.HttpRequest.HttpMethod
		item.URL = task.HttpRequest.Url
		item.Headers = task.HttpRequest.Headers
		item.Body = decodeBody(task.HttpRequest.Body)
	}
	return item
}

func decodeBody(raw string) string {
	if raw == "" {
		return ""
	}
	decoded, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return raw
	}
	return string(decoded)
}

func queueName(project, location, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/queues/%s", project, location, name)
}

func taskName(project, location, queue, name string) string {
	return fmt.Sprintf("%s/tasks/%s", queueName(project, location, queue), name)
}
