package run

import (
	"context"
	"errors"
	"fmt"

	runpb "cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	run "cloud.google.com/go/run/apiv2"
)

// Service holds Cloud Run service fields.
type Service struct {
	Name   string `json:"name"`
	URI    string `json:"uri"`
	Region string `json:"region"`
}

// Client defines Cloud Run operations.
type Client interface {
	ListServices(ctx context.Context, project, region string) ([]*Service, error)
	GetService(ctx context.Context, project, region, name string) (*Service, error)
	Deploy(ctx context.Context, project, region string, req *DeployRequest) error
	DeleteService(ctx context.Context, project, region, name string) error
}

// DeployRequest holds deploy parameters.
type DeployRequest struct {
	Name                 string
	Image                string
	Memory               string
	CPU                  string
	Port                 int32
	Env                  map[string]string
	AllowUnauthenticated bool
}

type gcpClient struct {
	services *run.ServicesClient
}

// NewClient creates a Client backed by the real Cloud Run API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	sc, err := run.NewServicesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud run client: %w", err)
	}
	return &gcpClient{services: sc}, nil
}

func (c *gcpClient) ListServices(ctx context.Context, project, region string) ([]*Service, error) {
	it := c.services.ListServices(ctx, &runpb.ListServicesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, region),
	})

	var services []*Service
	for {
		svc, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list services: %w", err)
		}
		services = append(services, svcFromProto(svc, region))
	}
	return services, nil
}

func (c *gcpClient) GetService(ctx context.Context, project, region, name string) (*Service, error) {
	svc, err := c.services.GetService(ctx, &runpb.GetServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", project, region, name),
	})
	if err != nil {
		return nil, fmt.Errorf("get service %s: %w", name, err)
	}
	return svcFromProto(svc, region), nil
}

func (c *gcpClient) Deploy(ctx context.Context, project, region string, req *DeployRequest) error {
	var envVars []*runpb.EnvVar
	for k, v := range req.Env {
		envVars = append(envVars, &runpb.EnvVar{
			Name:   k,
			Values: &runpb.EnvVar_Value{Value: v},
		})
	}

	port := req.Port
	if port == 0 {
		port = 8080
	}

	svcName := fmt.Sprintf("projects/%s/locations/%s/services/%s", project, region, req.Name)

	pbReq := &runpb.UpdateServiceRequest{
		Service: &runpb.Service{
			Name: svcName,
			Template: &runpb.RevisionTemplate{
				Containers: []*runpb.Container{
					{
						Image: req.Image,
						Ports: []*runpb.ContainerPort{{ContainerPort: port}},
						Env:   envVars,
					},
				},
			},
		},
	}

	op, err := c.services.UpdateService(ctx, pbReq)
	if err != nil {
		return fmt.Errorf("deploy service %s: %w", req.Name, err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for deploy %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpClient) DeleteService(ctx context.Context, project, region, name string) error {
	op, err := c.services.DeleteService(ctx, &runpb.DeleteServiceRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/services/%s", project, region, name),
	})
	if err != nil {
		return fmt.Errorf("delete service %s: %w", name, err)
	}

	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for delete %s: %w", name, err)
	}
	return nil
}

func svcFromProto(svc *runpb.Service, region string) *Service {
	return &Service{
		Name:   svc.GetName(),
		URI:    svc.GetUri(),
		Region: region,
	}
}
