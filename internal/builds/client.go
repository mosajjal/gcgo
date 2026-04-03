package builds

import (
	"context"
	"fmt"

	cloudbuild "google.golang.org/api/cloudbuild/v1"
	"google.golang.org/api/option"
)

// Build holds the fields we care about.
type Build struct {
	ID         string `json:"id"`
	ProjectID  string `json:"project_id"`
	Status     string `json:"status"`
	Source     string `json:"source"`
	CreateTime string `json:"create_time"`
	Duration   string `json:"duration"`
	LogURL     string `json:"log_url"`
}

// Trigger holds build trigger fields.
type Trigger struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	CreateTime  string `json:"create_time"`
	Disabled    bool   `json:"disabled"`
}

// CreateTriggerRequest holds parameters for trigger creation.
type CreateTriggerRequest struct {
	Name        string
	Description string
	RepoName    string
	BranchName  string
	Filename    string
}

// Client defines the operations we use for Cloud Build.
type Client interface {
	ListBuilds(ctx context.Context, project string) ([]*Build, error)
	GetBuild(ctx context.Context, project, buildID string) (*Build, error)
	CancelBuild(ctx context.Context, project, buildID string) error
	ListTriggers(ctx context.Context, project string) ([]*Trigger, error)
	GetTrigger(ctx context.Context, project, triggerID string) (*Trigger, error)
	CreateTrigger(ctx context.Context, project string, req *CreateTriggerRequest) (*Trigger, error)
	DeleteTrigger(ctx context.Context, project, triggerID string) error
	RunTrigger(ctx context.Context, project, triggerID string) error
}

type gcpClient struct {
	svc *cloudbuild.Service
}

// NewClient creates a Client backed by the real Cloud Build API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := cloudbuild.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud build client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListBuilds(ctx context.Context, project string) ([]*Build, error) {
	call := c.svc.Projects.Builds.List(project).Context(ctx)

	var builds []*Build
	if err := call.Pages(ctx, func(resp *cloudbuild.ListBuildsResponse) error {
		for _, b := range resp.Builds {
			builds = append(builds, buildFromAPI(b))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list builds: %w", err)
	}
	return builds, nil
}

func (c *gcpClient) GetBuild(ctx context.Context, project, buildID string) (*Build, error) {
	b, err := c.svc.Projects.Builds.Get(project, buildID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get build %s: %w", buildID, err)
	}
	return buildFromAPI(b), nil
}

func (c *gcpClient) CancelBuild(ctx context.Context, project, buildID string) error {
	if _, err := c.svc.Projects.Builds.Cancel(project, buildID, &cloudbuild.CancelBuildRequest{
		ProjectId: project,
		Id:        buildID,
	}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("cancel build %s: %w", buildID, err)
	}
	return nil
}

func (c *gcpClient) ListTriggers(ctx context.Context, project string) ([]*Trigger, error) {
	call := c.svc.Projects.Triggers.List(project).Context(ctx)

	var triggers []*Trigger
	if err := call.Pages(ctx, func(resp *cloudbuild.ListBuildTriggersResponse) error {
		for _, t := range resp.Triggers {
			triggers = append(triggers, triggerFromAPI(t))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list triggers: %w", err)
	}
	return triggers, nil
}

func (c *gcpClient) GetTrigger(ctx context.Context, project, triggerID string) (*Trigger, error) {
	t, err := c.svc.Projects.Triggers.Get(project, triggerID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get trigger %s: %w", triggerID, err)
	}
	return triggerFromAPI(t), nil
}

func (c *gcpClient) CreateTrigger(ctx context.Context, project string, req *CreateTriggerRequest) (*Trigger, error) {
	t := &cloudbuild.BuildTrigger{
		Name:        req.Name,
		Description: req.Description,
		Filename:    req.Filename,
	}
	if req.RepoName != "" || req.BranchName != "" {
		t.TriggerTemplate = &cloudbuild.RepoSource{
			RepoName:   req.RepoName,
			BranchName: req.BranchName,
		}
	}

	created, err := c.svc.Projects.Triggers.Create(project, t).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create trigger %s: %w", req.Name, err)
	}
	return triggerFromAPI(created), nil
}

func (c *gcpClient) DeleteTrigger(ctx context.Context, project, triggerID string) error {
	if _, err := c.svc.Projects.Triggers.Delete(project, triggerID).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete trigger %s: %w", triggerID, err)
	}
	return nil
}

func (c *gcpClient) RunTrigger(ctx context.Context, project, triggerID string) error {
	if _, err := c.svc.Projects.Triggers.Run(project, triggerID, &cloudbuild.RepoSource{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("run trigger %s: %w", triggerID, err)
	}
	return nil
}

func buildFromAPI(b *cloudbuild.Build) *Build {
	build := &Build{
		ID:         b.Id,
		ProjectID:  b.ProjectId,
		Status:     b.Status,
		LogURL:     b.LogUrl,
		CreateTime: b.CreateTime,
		Duration:   b.Timeout,
	}
	if src := b.Source; src != nil {
		if repo := src.RepoSource; repo != nil {
			build.Source = repo.RepoName
		} else if storage := src.StorageSource; storage != nil {
			build.Source = fmt.Sprintf("gs://%s/%s", storage.Bucket, storage.Object)
		}
	}
	return build
}

func triggerFromAPI(t *cloudbuild.BuildTrigger) *Trigger {
	return &Trigger{
		ID:          t.Id,
		Name:        t.Name,
		Description: t.Description,
		Disabled:    t.Disabled,
		CreateTime:  t.CreateTime,
	}
}
