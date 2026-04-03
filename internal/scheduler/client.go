package scheduler

import (
	"context"
	"fmt"
	"strings"

	cloudscheduler "google.golang.org/api/cloudscheduler/v1"
	"google.golang.org/api/option"
)

// Job holds Cloud Scheduler job fields.
type Job struct {
	Name       string `json:"name"`
	Schedule   string `json:"schedule"`
	TimeZone   string `json:"time_zone"`
	State      string `json:"state"`
	HTTPTarget string `json:"http_target,omitempty"`
}

// CreateJobRequest holds parameters for creating a scheduler job.
type CreateJobRequest struct {
	Name       string
	Schedule   string
	TimeZone   string
	URI        string
	HTTPMethod string
	Body       string
}

// Client defines Cloud Scheduler operations.
type Client interface {
	ListJobs(ctx context.Context, project, location string) ([]*Job, error)
	GetJob(ctx context.Context, project, location, jobID string) (*Job, error)
	CreateJob(ctx context.Context, project, location string, req *CreateJobRequest) (*Job, error)
	DeleteJob(ctx context.Context, project, location, jobID string) error
	PauseJob(ctx context.Context, project, location, jobID string) error
	ResumeJob(ctx context.Context, project, location, jobID string) error
	RunJob(ctx context.Context, project, location, jobID string) error
}

type gcpClient struct {
	svc *cloudscheduler.Service
}

// NewClient creates a Client backed by the real Cloud Scheduler API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := cloudscheduler.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create scheduler client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func jobName(project, location, jobID string) string {
	return fmt.Sprintf("projects/%s/locations/%s/jobs/%s", project, location, jobID)
}

func parentName(project, location string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, location)
}

func (c *gcpClient) ListJobs(ctx context.Context, project, location string) ([]*Job, error) {
	call := c.svc.Projects.Locations.Jobs.List(parentName(project, location)).Context(ctx)

	var jobs []*Job
	if err := call.Pages(ctx, func(resp *cloudscheduler.ListJobsResponse) error {
		for _, j := range resp.Jobs {
			jobs = append(jobs, jobFromAPI(j))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return jobs, nil
}

func (c *gcpClient) GetJob(ctx context.Context, project, location, jobID string) (*Job, error) {
	j, err := c.svc.Projects.Locations.Jobs.Get(jobName(project, location, jobID)).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get job %s: %w", jobID, err)
	}
	return jobFromAPI(j), nil
}

func (c *gcpClient) CreateJob(ctx context.Context, project, location string, req *CreateJobRequest) (*Job, error) {
	method := "POST"
	switch strings.ToUpper(req.HTTPMethod) {
	case "GET":
		method = "GET"
	case "PUT":
		method = "PUT"
	case "DELETE":
		method = "DELETE"
	case "PATCH":
		method = "PATCH"
	case "HEAD":
		method = "HEAD"
	case "OPTIONS":
		method = "OPTIONS"
	case "POST", "":
		method = "POST"
	}

	tz := req.TimeZone
	if tz == "" {
		tz = "UTC"
	}

	j := &cloudscheduler.Job{
		Name:     jobName(project, location, req.Name),
		Schedule: req.Schedule,
		TimeZone: tz,
		HttpTarget: &cloudscheduler.HttpTarget{
			Uri:        req.URI,
			HttpMethod: method,
			Body:       req.Body,
		},
	}

	created, err := c.svc.Projects.Locations.Jobs.Create(parentName(project, location), j).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create job %s: %w", req.Name, err)
	}
	return jobFromAPI(created), nil
}

func (c *gcpClient) DeleteJob(ctx context.Context, project, location, jobID string) error {
	if _, err := c.svc.Projects.Locations.Jobs.Delete(jobName(project, location, jobID)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) PauseJob(ctx context.Context, project, location, jobID string) error {
	if _, err := c.svc.Projects.Locations.Jobs.Pause(jobName(project, location, jobID), &cloudscheduler.PauseJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("pause job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) ResumeJob(ctx context.Context, project, location, jobID string) error {
	if _, err := c.svc.Projects.Locations.Jobs.Resume(jobName(project, location, jobID), &cloudscheduler.ResumeJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("resume job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) RunJob(ctx context.Context, project, location, jobID string) error {
	if _, err := c.svc.Projects.Locations.Jobs.Run(jobName(project, location, jobID), &cloudscheduler.RunJobRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("run job %s: %w", jobID, err)
	}
	return nil
}

func jobFromAPI(j *cloudscheduler.Job) *Job {
	job := &Job{
		Name:     j.Name,
		Schedule: j.Schedule,
		TimeZone: j.TimeZone,
		State:    j.State,
	}
	if ht := j.HttpTarget; ht != nil {
		job.HTTPTarget = ht.Uri
	}
	return job
}
