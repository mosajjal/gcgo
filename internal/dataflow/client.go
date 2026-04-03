package dataflow

import (
	"context"
	"fmt"

	"google.golang.org/api/dataflow/v1b3"
	"google.golang.org/api/option"
)

// Job holds Dataflow job fields.
type Job struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Type       string `json:"type"`
	State      string `json:"state"`
	CreateTime string `json:"create_time"`
	Region     string `json:"region"`
}

// Message holds Dataflow job message fields.
type Message struct {
	ID         string `json:"id"`
	Importance string `json:"importance"`
	Text       string `json:"text"`
	Time       string `json:"time"`
}

// Snapshot holds Dataflow snapshot fields.
type Snapshot struct {
	ID           string `json:"id"`
	SourceJobID  string `json:"source_job_id"`
	State        string `json:"state"`
	CreationTime string `json:"creation_time"`
	Description  string `json:"description"`
	Region       string `json:"region"`
}

// CreateSnapshotRequest holds snapshot creation parameters.
type CreateSnapshotRequest struct {
	Description     string
	TTL             string
	SnapshotSources bool
}

// MetricsSummary holds a small summary of Dataflow job metrics.
type MetricsSummary struct {
	MetricTime  string `json:"metric_time"`
	MetricCount int    `json:"metric_count"`
}

// LaunchFlexTemplateRequest holds flex template launch parameters.
type LaunchFlexTemplateRequest struct {
	JobName              string
	ContainerSpecGCSPath string
	Parameters           map[string]string
	ValidateOnly         bool
}

// LaunchTemplateRequest holds classic template launch parameters.
type LaunchTemplateRequest struct {
	JobName      string
	GcsPath      string
	Parameters   map[string]string
	ValidateOnly bool
}

// Client defines Dataflow operations.
type Client interface {
	ListJobs(ctx context.Context, project, region string) ([]*Job, error)
	GetJob(ctx context.Context, project, region, jobID string) (*Job, error)
	GetMetrics(ctx context.Context, project, region, jobID, startTime string) (*MetricsSummary, error)
	LaunchFlexTemplate(ctx context.Context, project, region string, req *LaunchFlexTemplateRequest) (*Job, error)
	GetTemplate(ctx context.Context, project, region, gcsPath string) (*dataflow.GetTemplateResponse, error)
	LaunchTemplate(ctx context.Context, project, region string, req *LaunchTemplateRequest) (*dataflow.LaunchTemplateResponse, error)
	CancelJob(ctx context.Context, project, region, jobID string) error
	DrainJob(ctx context.Context, project, region, jobID string) error
	ListMessages(ctx context.Context, project, region, jobID, minimumImportance, startTime, endTime string) ([]*Message, error)
	ListSnapshots(ctx context.Context, project, region, jobID string) ([]*Snapshot, error)
	GetSnapshot(ctx context.Context, project, region, snapshotID string) (*Snapshot, error)
	CreateSnapshot(ctx context.Context, project, region, jobID string, req *CreateSnapshotRequest) (*Snapshot, error)
	DeleteSnapshot(ctx context.Context, project, region, snapshotID string) error
}

type gcpClient struct {
	svc *dataflow.Service
}

// NewClient creates a Client backed by the real Dataflow API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := dataflow.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create dataflow client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListJobs(ctx context.Context, project, region string) ([]*Job, error) {
	var jobs []*Job
	err := c.svc.Projects.Locations.Jobs.List(project, region).Context(ctx).Pages(ctx, func(resp *dataflow.ListJobsResponse) error {
		for _, j := range resp.Jobs {
			jobs = append(jobs, jobFromAPI(j, region))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list dataflow jobs: %w", err)
	}
	return jobs, nil
}

func (c *gcpClient) GetJob(ctx context.Context, project, region, jobID string) (*Job, error) {
	j, err := c.svc.Projects.Locations.Jobs.Get(project, region, jobID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get dataflow job %s: %w", jobID, err)
	}
	return jobFromAPI(j, region), nil
}

func (c *gcpClient) GetMetrics(ctx context.Context, project, region, jobID, startTime string) (*MetricsSummary, error) {
	call := c.svc.Projects.Locations.Jobs.GetMetrics(project, region, jobID).Context(ctx)
	if startTime != "" {
		call = call.StartTime(startTime)
	}
	metrics, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("get dataflow job metrics for %s: %w", jobID, err)
	}
	return &MetricsSummary{
		MetricTime:  metrics.MetricTime,
		MetricCount: len(metrics.Metrics),
	}, nil
}

func (c *gcpClient) LaunchFlexTemplate(ctx context.Context, project, region string, req *LaunchFlexTemplateRequest) (*Job, error) {
	resp, err := c.svc.Projects.Locations.FlexTemplates.Launch(project, region, &dataflow.LaunchFlexTemplateRequest{
		LaunchParameter: &dataflow.LaunchFlexTemplateParameter{
			JobName:              req.JobName,
			ContainerSpecGcsPath: req.ContainerSpecGCSPath,
			Parameters:           req.Parameters,
		},
		ValidateOnly: req.ValidateOnly,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("launch dataflow flex template %s: %w", req.JobName, err)
	}
	return jobFromAPI(resp.Job, region), nil
}

func (c *gcpClient) GetTemplate(ctx context.Context, project, region, gcsPath string) (*dataflow.GetTemplateResponse, error) {
	resp, err := c.svc.Projects.Locations.Templates.Get(project, region).GcsPath(gcsPath).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get dataflow template %s: %w", gcsPath, err)
	}
	return resp, nil
}

func (c *gcpClient) LaunchTemplate(ctx context.Context, project, region string, req *LaunchTemplateRequest) (*dataflow.LaunchTemplateResponse, error) {
	resp, err := c.svc.Projects.Locations.Templates.Launch(project, region, &dataflow.LaunchTemplateParameters{
		JobName:    req.JobName,
		Parameters: req.Parameters,
	}).GcsPath(req.GcsPath).ValidateOnly(req.ValidateOnly).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("launch dataflow template %s: %w", req.JobName, err)
	}
	return resp, nil
}

func (c *gcpClient) CancelJob(ctx context.Context, project, region, jobID string) error {
	_, err := c.svc.Projects.Locations.Jobs.Update(project, region, jobID, &dataflow.Job{
		RequestedState: "JOB_STATE_CANCELLED",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancel dataflow job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) DrainJob(ctx context.Context, project, region, jobID string) error {
	_, err := c.svc.Projects.Locations.Jobs.Update(project, region, jobID, &dataflow.Job{
		RequestedState: "JOB_STATE_DRAINED",
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("drain dataflow job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) ListMessages(ctx context.Context, project, region, jobID, minimumImportance, startTime, endTime string) ([]*Message, error) {
	call := c.svc.Projects.Locations.Jobs.Messages.List(project, region, jobID).Context(ctx)
	if minimumImportance != "" {
		call = call.MinimumImportance(minimumImportance)
	}
	if startTime != "" {
		call = call.StartTime(startTime)
	}
	if endTime != "" {
		call = call.EndTime(endTime)
	}

	var messages []*Message
	if err := call.Pages(ctx, func(resp *dataflow.ListJobMessagesResponse) error {
		for _, message := range resp.JobMessages {
			messages = append(messages, messageFromAPI(message))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list dataflow job messages for %s: %w", jobID, err)
	}
	return messages, nil
}

func (c *gcpClient) ListSnapshots(ctx context.Context, project, region, jobID string) ([]*Snapshot, error) {
	call := c.svc.Projects.Locations.Snapshots.List(project, region).Context(ctx)
	if jobID != "" {
		call = call.JobId(jobID)
	}

	resp, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("list dataflow snapshots: %w", err)
	}
	var snapshots []*Snapshot
	for _, snapshot := range resp.Snapshots {
		snapshots = append(snapshots, snapshotFromAPI(snapshot))
	}
	return snapshots, nil
}

func (c *gcpClient) GetSnapshot(ctx context.Context, project, region, snapshotID string) (*Snapshot, error) {
	snapshot, err := c.svc.Projects.Locations.Snapshots.Get(project, region, snapshotID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get dataflow snapshot %s: %w", snapshotID, err)
	}
	return snapshotFromAPI(snapshot), nil
}

func (c *gcpClient) CreateSnapshot(ctx context.Context, project, region, jobID string, req *CreateSnapshotRequest) (*Snapshot, error) {
	snapshot, err := c.svc.Projects.Locations.Jobs.Snapshot(project, region, jobID, &dataflow.SnapshotJobRequest{
		Description:     req.Description,
		Location:        region,
		SnapshotSources: req.SnapshotSources,
		Ttl:             req.TTL,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create dataflow snapshot from job %s: %w", jobID, err)
	}
	return snapshotFromAPI(snapshot), nil
}

func (c *gcpClient) DeleteSnapshot(ctx context.Context, project, region, snapshotID string) error {
	if _, err := c.svc.Projects.Locations.Snapshots.Delete(project, region, snapshotID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete dataflow snapshot %s: %w", snapshotID, err)
	}
	return nil
}

func jobFromAPI(j *dataflow.Job, region string) *Job {
	return &Job{
		ID:         j.Id,
		Name:       j.Name,
		Type:       j.Type,
		State:      j.CurrentState,
		CreateTime: j.CreateTime,
		Region:     region,
	}
}

func messageFromAPI(message *dataflow.JobMessage) *Message {
	if message == nil {
		return nil
	}
	return &Message{
		ID:         message.Id,
		Importance: message.MessageImportance,
		Text:       message.MessageText,
		Time:       message.Time,
	}
}

func snapshotFromAPI(snapshot *dataflow.Snapshot) *Snapshot {
	if snapshot == nil {
		return nil
	}
	return &Snapshot{
		ID:           snapshot.Id,
		SourceJobID:  snapshot.SourceJobId,
		State:        snapshot.State,
		CreationTime: snapshot.CreationTime,
		Description:  snapshot.Description,
		Region:       snapshot.Region,
	}
}
