package redis

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	redisapi "google.golang.org/api/redis/v1"
)

// Instance holds the fields we display.
type Instance struct {
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	Tier              string `json:"tier"`
	MemorySizeGB      int64  `json:"memory_size_gb"`
	Host              string `json:"host"`
	LocationID        string `json:"location_id"`
	AuthorizedNetwork string `json:"authorized_network"`
	CreateTime        string `json:"create_time"`
	State             string `json:"state"`
}

// CreateInstanceRequest holds parameters for Redis instance creation.
type CreateInstanceRequest struct {
	Name              string
	DisplayName       string
	Tier              string
	MemorySizeGB      int64
	LocationID        string
	AuthorizedNetwork string
}

// Client defines Redis operations.
type Client interface {
	ListInstances(ctx context.Context, project, location string) ([]*Instance, error)
	GetInstance(ctx context.Context, project, location, name string) (*Instance, error)
	CreateInstance(ctx context.Context, project, location string, req *CreateInstanceRequest) (*Instance, error)
	DeleteInstance(ctx context.Context, project, location, name string) error
}

type gcpClient struct {
	svc *redisapi.Service
}

// NewClient creates a Client backed by the real Memorystore for Redis API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := redisapi.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create redis client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListInstances(ctx context.Context, project, location string) ([]*Instance, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.svc.Projects.Locations.Instances.List(parent).Context(ctx)

	var instances []*Instance
	if err := call.Pages(ctx, func(resp *redisapi.ListInstancesResponse) error {
		for _, instance := range resp.Instances {
			instances = append(instances, instanceFromAPI(instance))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	return instances, nil
}

func (c *gcpClient) GetInstance(ctx context.Context, project, location, name string) (*Instance, error) {
	fullName := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, location, name)
	instance, err := c.svc.Projects.Locations.Instances.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get instance %s: %w", name, err)
	}
	return instanceFromAPI(instance), nil
}

func (c *gcpClient) CreateInstance(ctx context.Context, project, location string, req *CreateInstanceRequest) (*Instance, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	instance := &redisapi.Instance{
		Name:              fmt.Sprintf("%s/instances/%s", parent, req.Name),
		DisplayName:       req.DisplayName,
		Tier:              req.Tier,
		MemorySizeGb:      req.MemorySizeGB,
		LocationId:        req.LocationID,
		AuthorizedNetwork: req.AuthorizedNetwork,
	}

	if _, err := c.svc.Projects.Locations.Instances.Create(parent, instance).InstanceId(req.Name).Context(ctx).Do(); err != nil {
		return nil, fmt.Errorf("create instance %s: %w", req.Name, err)
	}

	return &Instance{
		Name:              instance.Name,
		DisplayName:       req.DisplayName,
		Tier:              req.Tier,
		MemorySizeGB:      req.MemorySizeGB,
		LocationID:        req.LocationID,
		AuthorizedNetwork: req.AuthorizedNetwork,
		State:             "CREATING",
	}, nil
}

func (c *gcpClient) DeleteInstance(ctx context.Context, project, location, name string) error {
	fullName := fmt.Sprintf("projects/%s/locations/%s/instances/%s", project, location, name)
	if _, err := c.svc.Projects.Locations.Instances.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete instance %s: %w", name, err)
	}
	return nil
}

func instanceFromAPI(instance *redisapi.Instance) *Instance {
	if instance == nil {
		return nil
	}
	return &Instance{
		Name:              instance.Name,
		DisplayName:       instance.DisplayName,
		Tier:              instance.Tier,
		MemorySizeGB:      instance.MemorySizeGb,
		Host:              instance.Host,
		LocationID:        instance.LocationId,
		AuthorizedNetwork: instance.AuthorizedNetwork,
		CreateTime:        instance.CreateTime,
		State:             instance.State,
	}
}
