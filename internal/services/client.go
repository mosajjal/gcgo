package services

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	serviceusage "google.golang.org/api/serviceusage/v1"
)

// Service holds the fields we display.
type Service struct {
	Name  string `json:"name"`
	Title string `json:"title"`
	State string `json:"state"`
}

// Client defines Service Usage operations.
type Client interface {
	ListServices(ctx context.Context, project string) ([]*Service, error)
	GetService(ctx context.Context, project, serviceID string) (*Service, error)
	EnableService(ctx context.Context, project, serviceID string) error
	DisableService(ctx context.Context, project, serviceID string) error
}

type gcpClient struct {
	svc *serviceusage.Service
}

// NewClient creates a Client backed by the real Service Usage API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := serviceusage.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create service usage client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListServices(ctx context.Context, project string) ([]*Service, error) {
	parent := fmt.Sprintf("projects/%s", project)
	call := c.svc.Services.List(parent).Context(ctx)

	var services []*Service
	if err := call.Pages(ctx, func(resp *serviceusage.ListServicesResponse) error {
		for _, service := range resp.Services {
			services = append(services, serviceFromAPI(service))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list services: %w", err)
	}
	return services, nil
}

func (c *gcpClient) GetService(ctx context.Context, project, serviceID string) (*Service, error) {
	name := fmt.Sprintf("projects/%s/services/%s", project, serviceID)
	service, err := c.svc.Services.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get service %s: %w", serviceID, err)
	}
	return serviceFromAPI(service), nil
}

func (c *gcpClient) EnableService(ctx context.Context, project, serviceID string) error {
	name := fmt.Sprintf("projects/%s/services/%s", project, serviceID)
	if _, err := c.svc.Services.Enable(name, &serviceusage.EnableServiceRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("enable service %s: %w", serviceID, err)
	}
	return nil
}

func (c *gcpClient) DisableService(ctx context.Context, project, serviceID string) error {
	name := fmt.Sprintf("projects/%s/services/%s", project, serviceID)
	if _, err := c.svc.Services.Disable(name, &serviceusage.DisableServiceRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("disable service %s: %w", serviceID, err)
	}
	return nil
}

func serviceFromAPI(service *serviceusage.GoogleApiServiceusageV1Service) *Service {
	if service == nil {
		return nil
	}
	title := ""
	if service.Config != nil {
		title = service.Config.Title
	}
	return &Service{
		Name:  service.Name,
		Title: title,
		State: service.State,
	}
}
