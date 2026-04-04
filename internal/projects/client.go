package projects

import (
	"context"
	"errors"
	"fmt"

	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Project holds the fields we care about.
type Project struct {
	ID     string `json:"project_id"`
	Name   string `json:"name"`
	Number string `json:"project_number"`
	State  string `json:"state"`
}

// Client defines the operations we use for projects.
type Client interface {
	List(ctx context.Context) ([]*Project, error)
	Get(ctx context.Context, projectID string) (*Project, error)
	CreateProject(ctx context.Context, projectID, name string, labels map[string]string) error
	DeleteProject(ctx context.Context, projectID string) error
}

type gcpClient struct {
	rm *resourcemanager.ProjectsClient
}

// NewClient creates a Client backed by the real GCP API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	rm, err := resourcemanager.NewProjectsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create resource manager client: %w", err)
	}
	return &gcpClient{rm: rm}, nil
}

func (c *gcpClient) List(ctx context.Context) ([]*Project, error) {
	it := c.rm.SearchProjects(ctx, &resourcemanagerpb.SearchProjectsRequest{})

	var projects []*Project
	for {
		p, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list projects: %w", err)
		}
		projects = append(projects, fromProto(p))
	}
	return projects, nil
}

func (c *gcpClient) Get(ctx context.Context, projectID string) (*Project, error) {
	p, err := c.rm.GetProject(ctx, &resourcemanagerpb.GetProjectRequest{
		Name: "projects/" + projectID,
	})
	if err != nil {
		return nil, fmt.Errorf("get project %s: %w", projectID, err)
	}
	return fromProto(p), nil
}

func (c *gcpClient) CreateProject(ctx context.Context, projectID, name string, labels map[string]string) error {
	op, err := c.rm.CreateProject(ctx, &resourcemanagerpb.CreateProjectRequest{
		Project: &resourcemanagerpb.Project{
			ProjectId:   projectID,
			DisplayName: name,
			Labels:      labels,
		},
	})
	if err != nil {
		return fmt.Errorf("create project %s: %w", projectID, err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait create project %s: %w", projectID, err)
	}
	return nil
}

func (c *gcpClient) DeleteProject(ctx context.Context, projectID string) error {
	op, err := c.rm.DeleteProject(ctx, &resourcemanagerpb.DeleteProjectRequest{
		Name: "projects/" + projectID,
	})
	if err != nil {
		return fmt.Errorf("delete project %s: %w", projectID, err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait delete project %s: %w", projectID, err)
	}
	return nil
}

func fromProto(p *resourcemanagerpb.Project) *Project {
	return &Project{
		ID:     p.GetProjectId(),
		Name:   p.GetDisplayName(),
		Number: p.GetName(), // "projects/NUMBER"
		State:  p.GetState().String(),
	}
}
