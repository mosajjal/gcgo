package deploy

import (
	"context"
	"fmt"

	clouddeploy "google.golang.org/api/clouddeploy/v1"
	"google.golang.org/api/option"
)

// DeliveryPipeline holds the fields we display.
type DeliveryPipeline struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Suspended   bool   `json:"suspended"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
	Uid         string `json:"uid"`
}

// Release holds the fields we display.
type Release struct {
	Name               string `json:"name"`
	Description        string `json:"description"`
	RenderState        string `json:"render_state"`
	SkaffoldConfigUri  string `json:"skaffold_config_uri"`
	SkaffoldConfigPath string `json:"skaffold_config_path"`
	SkaffoldVersion    string `json:"skaffold_version"`
	CreateTime         string `json:"create_time"`
	Uid                string `json:"uid"`
}

// CreateDeliveryPipelineRequest holds parameters for pipeline creation.
type CreateDeliveryPipelineRequest struct {
	Name        string
	Description string
	Suspended   bool
	Targets     []string
}

// CreateReleaseRequest holds parameters for release creation.
type CreateReleaseRequest struct {
	Name               string
	Description        string
	SkaffoldConfigUri  string
	SkaffoldConfigPath string
	SkaffoldVersion    string
}

// Client defines Cloud Deploy operations.
type Client interface {
	ListDeliveryPipelines(ctx context.Context, project, location string) ([]*DeliveryPipeline, error)
	GetDeliveryPipeline(ctx context.Context, project, location, name string) (*DeliveryPipeline, error)
	CreateDeliveryPipeline(ctx context.Context, project, location string, req *CreateDeliveryPipelineRequest) (*DeliveryPipeline, error)
	DeleteDeliveryPipeline(ctx context.Context, project, location, name string) error

	ListReleases(ctx context.Context, project, location, pipeline string) ([]*Release, error)
	GetRelease(ctx context.Context, project, location, pipeline, name string) (*Release, error)
	CreateRelease(ctx context.Context, project, location, pipeline string, req *CreateReleaseRequest) (*Release, error)
}

type gcpClient struct {
	svc *clouddeploy.Service
}

// NewClient creates a Client backed by the real Cloud Deploy API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := clouddeploy.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud deploy client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListDeliveryPipelines(ctx context.Context, project, location string) ([]*DeliveryPipeline, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.svc.Projects.Locations.DeliveryPipelines.List(parent).Context(ctx)

	var pipelines []*DeliveryPipeline
	if err := call.Pages(ctx, func(resp *clouddeploy.ListDeliveryPipelinesResponse) error {
		for _, pipeline := range resp.DeliveryPipelines {
			pipelines = append(pipelines, deliveryPipelineFromAPI(pipeline))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list delivery pipelines: %w", err)
	}
	return pipelines, nil
}

func (c *gcpClient) GetDeliveryPipeline(ctx context.Context, project, location, name string) (*DeliveryPipeline, error) {
	fullName := fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", project, location, name)
	pipeline, err := c.svc.Projects.Locations.DeliveryPipelines.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get delivery pipeline %s: %w", name, err)
	}
	return deliveryPipelineFromAPI(pipeline), nil
}

func (c *gcpClient) CreateDeliveryPipeline(ctx context.Context, project, location string, req *CreateDeliveryPipelineRequest) (*DeliveryPipeline, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	pipeline := &clouddeploy.DeliveryPipeline{
		Name:        fmt.Sprintf("%s/deliveryPipelines/%s", parent, req.Name),
		Description: req.Description,
		Suspended:   req.Suspended,
		SerialPipeline: &clouddeploy.SerialPipeline{
			Stages: buildStages(req.Targets),
		},
	}

	if _, err := c.svc.Projects.Locations.DeliveryPipelines.Create(parent, pipeline).DeliveryPipelineId(req.Name).Context(ctx).Do(); err != nil {
		return nil, fmt.Errorf("create delivery pipeline %s: %w", req.Name, err)
	}

	return &DeliveryPipeline{
		Name:        pipeline.Name,
		Description: req.Description,
		Suspended:   req.Suspended,
	}, nil
}

func (c *gcpClient) DeleteDeliveryPipeline(ctx context.Context, project, location, name string) error {
	fullName := fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", project, location, name)
	if _, err := c.svc.Projects.Locations.DeliveryPipelines.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete delivery pipeline %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListReleases(ctx context.Context, project, location, pipeline string) ([]*Release, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", project, location, pipeline)
	call := c.svc.Projects.Locations.DeliveryPipelines.Releases.List(parent).Context(ctx)

	var releases []*Release
	if err := call.Pages(ctx, func(resp *clouddeploy.ListReleasesResponse) error {
		for _, release := range resp.Releases {
			releases = append(releases, releaseFromAPI(release))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list releases: %w", err)
	}
	return releases, nil
}

func (c *gcpClient) GetRelease(ctx context.Context, project, location, pipeline, name string) (*Release, error) {
	fullName := fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s/releases/%s", project, location, pipeline, name)
	release, err := c.svc.Projects.Locations.DeliveryPipelines.Releases.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get release %s: %w", name, err)
	}
	return releaseFromAPI(release), nil
}

func (c *gcpClient) CreateRelease(ctx context.Context, project, location, pipeline string, req *CreateReleaseRequest) (*Release, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/deliveryPipelines/%s", project, location, pipeline)
	release := &clouddeploy.Release{
		Name:               fmt.Sprintf("%s/releases/%s", parent, req.Name),
		Description:        req.Description,
		SkaffoldConfigUri:  req.SkaffoldConfigUri,
		SkaffoldConfigPath: req.SkaffoldConfigPath,
		SkaffoldVersion:    req.SkaffoldVersion,
	}

	if _, err := c.svc.Projects.Locations.DeliveryPipelines.Releases.Create(parent, release).ReleaseId(req.Name).Context(ctx).Do(); err != nil {
		return nil, fmt.Errorf("create release %s: %w", req.Name, err)
	}

	return &Release{
		Name:               release.Name,
		Description:        req.Description,
		SkaffoldConfigUri:  req.SkaffoldConfigUri,
		SkaffoldConfigPath: req.SkaffoldConfigPath,
		SkaffoldVersion:    req.SkaffoldVersion,
	}, nil
}

func deliveryPipelineFromAPI(pipeline *clouddeploy.DeliveryPipeline) *DeliveryPipeline {
	if pipeline == nil {
		return nil
	}
	return &DeliveryPipeline{
		Name:        pipeline.Name,
		Description: pipeline.Description,
		Suspended:   pipeline.Suspended,
		CreateTime:  pipeline.CreateTime,
		UpdateTime:  pipeline.UpdateTime,
		Uid:         pipeline.Uid,
	}
}

func releaseFromAPI(release *clouddeploy.Release) *Release {
	if release == nil {
		return nil
	}
	return &Release{
		Name:               release.Name,
		Description:        release.Description,
		RenderState:        release.RenderState,
		SkaffoldConfigUri:  release.SkaffoldConfigUri,
		SkaffoldConfigPath: release.SkaffoldConfigPath,
		SkaffoldVersion:    release.SkaffoldVersion,
		CreateTime:         release.CreateTime,
		Uid:                release.Uid,
	}
}

func buildStages(targets []string) []*clouddeploy.Stage {
	stages := make([]*clouddeploy.Stage, 0, len(targets))
	for _, target := range targets {
		if target == "" {
			continue
		}
		stages = append(stages, &clouddeploy.Stage{TargetId: target})
	}
	return stages
}
