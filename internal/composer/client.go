package composer

import (
	"context"
	"fmt"

	"google.golang.org/api/composer/v1"
	"google.golang.org/api/option"
)

// Environment holds Cloud Composer environment fields.
type Environment struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Config string `json:"config"`
	Region string `json:"region"`
}

// Client defines Cloud Composer operations.
type Client interface {
	ListEnvironments(ctx context.Context, project, region string) ([]*Environment, error)
	GetEnvironment(ctx context.Context, project, region, name string) (*Environment, error)
	CreateEnvironment(ctx context.Context, project, region string, req *CreateEnvironmentRequest) error
	DeleteEnvironment(ctx context.Context, project, region, name string) error
	UpdateEnvironment(ctx context.Context, project, region, name string, req *UpdateEnvironmentRequest) error
}

// CreateEnvironmentRequest holds parameters for environment creation.
type CreateEnvironmentRequest struct {
	Name         string
	NodeCount    int64
	MachineType  string
	ImageVersion string
}

// UpdateEnvironmentRequest holds parameters for environment update.
type UpdateEnvironmentRequest struct {
	NodeCount int64
	Labels    map[string]string
}

type gcpClient struct {
	envs *composer.Service
}

// NewClient creates a Client backed by the real Cloud Composer API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := composer.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create composer client: %w", err)
	}
	return &gcpClient{envs: svc}, nil
}

func (c *gcpClient) ListEnvironments(ctx context.Context, project, region string) ([]*Environment, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	var envs []*Environment
	err := c.envs.Projects.Locations.Environments.List(parent).Context(ctx).Pages(ctx, func(resp *composer.ListEnvironmentsResponse) error {
		for _, e := range resp.Environments {
			envs = append(envs, envFromAPI(e, region))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list environments: %w", err)
	}
	return envs, nil
}

func (c *gcpClient) GetEnvironment(ctx context.Context, project, region, name string) (*Environment, error) {
	fullName := fmt.Sprintf("projects/%s/locations/%s/environments/%s", project, region, name)
	e, err := c.envs.Projects.Locations.Environments.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get environment %s: %w", name, err)
	}
	return envFromAPI(e, region), nil
}

func (c *gcpClient) CreateEnvironment(ctx context.Context, project, region string, req *CreateEnvironmentRequest) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	env := &composer.Environment{
		Name: fmt.Sprintf("projects/%s/locations/%s/environments/%s", project, region, req.Name),
		Config: &composer.EnvironmentConfig{
			NodeCount: req.NodeCount,
		},
	}
	if req.ImageVersion != "" {
		env.Config.SoftwareConfig = &composer.SoftwareConfig{ImageVersion: req.ImageVersion}
	}
	if req.MachineType != "" {
		env.Config.NodeConfig = &composer.NodeConfig{MachineType: req.MachineType}
	}
	if _, err := c.envs.Projects.Locations.Environments.Create(parent, env).Context(ctx).Do(); err != nil {
		return fmt.Errorf("create environment %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpClient) DeleteEnvironment(ctx context.Context, project, region, name string) error {
	fullName := fmt.Sprintf("projects/%s/locations/%s/environments/%s", project, region, name)
	if _, err := c.envs.Projects.Locations.Environments.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete environment %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) UpdateEnvironment(ctx context.Context, project, region, name string, req *UpdateEnvironmentRequest) error {
	fullName := fmt.Sprintf("projects/%s/locations/%s/environments/%s", project, region, name)
	env := &composer.Environment{}
	if req.NodeCount > 0 {
		env.Config = &composer.EnvironmentConfig{NodeCount: req.NodeCount}
	}
	if len(req.Labels) > 0 {
		env.Labels = req.Labels
	}
	if _, err := c.envs.Projects.Locations.Environments.Patch(fullName, env).Context(ctx).Do(); err != nil {
		return fmt.Errorf("update environment %s: %w", name, err)
	}
	return nil
}

func envFromAPI(e *composer.Environment, region string) *Environment {
	configDesc := ""
	if e.Config != nil {
		if e.Config.SoftwareConfig != nil {
			configDesc = e.Config.SoftwareConfig.ImageVersion
		}
	}
	return &Environment{
		Name:   e.Name,
		State:  e.State,
		Config: configDesc,
		Region: region,
	}
}
