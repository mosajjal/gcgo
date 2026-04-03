package workflows

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
	workflowsapi "google.golang.org/api/workflows/v1"
)

// Workflow holds the fields we display.
type Workflow struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	State          string `json:"state"`
	RevisionID     string `json:"revision_id"`
	CreateTime     string `json:"create_time"`
	UpdateTime     string `json:"update_time"`
	ServiceAccount string `json:"service_account"`
}

// DeployRequest holds workflow deployment parameters.
type DeployRequest struct {
	Name           string
	Description    string
	SourceContents string
	ServiceAccount string
}

// Client defines workflow operations.
type Client interface {
	List(ctx context.Context, project, location string) ([]*Workflow, error)
	Get(ctx context.Context, project, location, name string) (*Workflow, error)
	Deploy(ctx context.Context, project, location string, req *DeployRequest) (*Workflow, error)
	Delete(ctx context.Context, project, location, name string) error
}

type gcpClient struct {
	svc *workflowsapi.Service
}

// NewClient creates a Client backed by the real Workflows API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := workflowsapi.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create workflows client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) List(ctx context.Context, project, location string) ([]*Workflow, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.svc.Projects.Locations.Workflows.List(parent).Context(ctx)

	var workflows []*Workflow
	if err := call.Pages(ctx, func(resp *workflowsapi.ListWorkflowsResponse) error {
		for _, workflow := range resp.Workflows {
			workflows = append(workflows, workflowFromAPI(workflow))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list workflows: %w", err)
	}
	return workflows, nil
}

func (c *gcpClient) Get(ctx context.Context, project, location, name string) (*Workflow, error) {
	fullName := workflowName(project, location, name)
	workflow, err := c.svc.Projects.Locations.Workflows.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get workflow %s: %w", name, err)
	}
	return workflowFromAPI(workflow), nil
}

func (c *gcpClient) Deploy(ctx context.Context, project, location string, req *DeployRequest) (*Workflow, error) {
	fullName := workflowName(project, location, req.Name)
	workflow := &workflowsapi.Workflow{
		Name:           fullName,
		Description:    req.Description,
		SourceContents: req.SourceContents,
		ServiceAccount: req.ServiceAccount,
	}

	_, err := c.svc.Projects.Locations.Workflows.Get(fullName).Context(ctx).Do()
	switch {
	case err == nil:
		if _, err := c.svc.Projects.Locations.Workflows.Patch(fullName, workflow).
			UpdateMask("description,source_contents,service_account").
			Context(ctx).
			Do(); err != nil {
			return nil, fmt.Errorf("update workflow %s: %w", req.Name, err)
		}
	case isNotFound(err):
		parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
		if _, err := c.svc.Projects.Locations.Workflows.Create(parent, workflow).WorkflowId(req.Name).Context(ctx).Do(); err != nil {
			return nil, fmt.Errorf("create workflow %s: %w", req.Name, err)
		}
	default:
		return nil, fmt.Errorf("check workflow %s: %w", req.Name, err)
	}

	return &Workflow{
		Name:           fullName,
		Description:    req.Description,
		ServiceAccount: req.ServiceAccount,
	}, nil
}

func (c *gcpClient) Delete(ctx context.Context, project, location, name string) error {
	fullName := workflowName(project, location, name)
	if _, err := c.svc.Projects.Locations.Workflows.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete workflow %s: %w", name, err)
	}
	return nil
}

func workflowFromAPI(workflow *workflowsapi.Workflow) *Workflow {
	if workflow == nil {
		return nil
	}
	return &Workflow{
		Name:           workflow.Name,
		Description:    workflow.Description,
		State:          workflow.State,
		RevisionID:     workflow.RevisionId,
		CreateTime:     workflow.CreateTime,
		UpdateTime:     workflow.UpdateTime,
		ServiceAccount: workflow.ServiceAccount,
	}
}

func workflowName(project, location, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/workflows/%s", project, location, name)
}

func isNotFound(err error) bool {
	var apiErr *googleapi.Error
	return errors.As(err, &apiErr) && apiErr.Code == 404
}
