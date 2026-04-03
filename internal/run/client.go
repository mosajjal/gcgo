package run

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/iam/apiv1/iampb"
	runpb "cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	run "cloud.google.com/go/run/apiv2"
)

// Service holds Cloud Run service fields.
type Service struct {
	Name                  string           `json:"name"`
	URI                   string           `json:"uri"`
	Region                string           `json:"region"`
	LatestReadyRevision   string           `json:"latest_ready_revision,omitempty"`
	LatestCreatedRevision string           `json:"latest_created_revision,omitempty"`
	Traffic               []*TrafficTarget `json:"traffic,omitempty"`
}

// TrafficTarget holds Cloud Run traffic routing fields.
type TrafficTarget struct {
	Type     string `json:"type"`
	Revision string `json:"revision,omitempty"`
	Percent  int32  `json:"percent"`
	Tag      string `json:"tag,omitempty"`
}

// Revision holds Cloud Run revision fields.
type Revision struct {
	Name        string `json:"name"`
	Service     string `json:"service"`
	Image       string `json:"image,omitempty"`
	CreateTime  string `json:"create_time,omitempty"`
	Generation  int64  `json:"generation"`
	Reconciling bool   `json:"reconciling"`
}

// Client defines Cloud Run operations.
type Client interface {
	ListServices(ctx context.Context, project, region string) ([]*Service, error)
	GetService(ctx context.Context, project, region, name string) (*Service, error)
	Deploy(ctx context.Context, project, region string, req *DeployRequest) error
	DeleteService(ctx context.Context, project, region, name string) error
	ListRevisions(ctx context.Context, project, region, service string) ([]*Revision, error)
	GetRevision(ctx context.Context, project, region, name string) (*Revision, error)
	UpdateTraffic(ctx context.Context, project, region, service string, req *UpdateTrafficRequest) (*Service, error)
	GetServicePolicy(ctx context.Context, project, region, service string) (*iampb.Policy, error)
	SetServicePolicy(ctx context.Context, project, region, service string, policy *iampb.Policy) (*iampb.Policy, error)
	TestServicePermissions(ctx context.Context, project, region, service string, permissions []string) ([]string, error)
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

// UpdateTrafficRequest holds Cloud Run traffic mutation parameters.
type UpdateTrafficRequest struct {
	ToLatest bool
	Revision string
	Percent  int32
	Tag      string
}

type gcpClient struct {
	services  *run.ServicesClient
	revisions *run.RevisionsClient
}

// NewClient creates a Client backed by the real Cloud Run API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	sc, err := run.NewServicesClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud run client: %w", err)
	}
	rc, err := run.NewRevisionsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create cloud run revisions client: %w", err)
	}
	return &gcpClient{services: sc, revisions: rc}, nil
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
	svcName := serviceName(project, region, req.Name)
	existing, err := c.services.GetService(ctx, &runpb.GetServiceRequest{Name: svcName})
	notFound := status.Code(err) == codes.NotFound
	if err != nil && !notFound {
		return fmt.Errorf("get service %s: %w", req.Name, err)
	}

	service := serviceFromExisting(svcName, existing)
	container := ensurePrimaryContainer(service)

	if req.Image != "" {
		container.Image = req.Image
	}
	if container.GetImage() == "" {
		return fmt.Errorf("image is required for new services or when no existing image is present")
	}

	port := req.Port
	if port == 0 && len(container.Ports) == 0 {
		port = 8080
	}
	if port != 0 {
		container.Ports = []*runpb.ContainerPort{{ContainerPort: port}}
	}

	if len(req.Env) > 0 {
		container.Env = envVarsFromMap(req.Env)
	}

	limits := map[string]string{}
	if container.GetResources() != nil {
		for k, v := range container.GetResources().GetLimits() {
			limits[k] = v
		}
	}
	if req.Memory != "" {
		limits["memory"] = req.Memory
	}
	if req.CPU != "" {
		limits["cpu"] = req.CPU
	}
	if len(limits) > 0 {
		container.Resources = &runpb.ResourceRequirements{
			Limits:  limits,
			CpuIdle: true,
		}
	}

	if req.AllowUnauthenticated {
		service.InvokerIamDisabled = true
	}

	if notFound {
		op, err := c.services.CreateService(ctx, &runpb.CreateServiceRequest{
			Parent:    fmt.Sprintf("projects/%s/locations/%s", project, region),
			ServiceId: req.Name,
			Service:   service,
		})
		if err != nil {
			return fmt.Errorf("create service %s: %w", req.Name, err)
		}
		if _, err := op.Wait(ctx); err != nil {
			return fmt.Errorf("wait for create service %s: %w", req.Name, err)
		}
		return nil
	}

	op, err := c.services.UpdateService(ctx, &runpb.UpdateServiceRequest{Service: service})
	if err != nil {
		return fmt.Errorf("update service %s: %w", req.Name, err)
	}
	if _, err := op.Wait(ctx); err != nil {
		return fmt.Errorf("wait for update service %s: %w", req.Name, err)
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

func (c *gcpClient) ListRevisions(ctx context.Context, project, region, service string) ([]*Revision, error) {
	it := c.revisions.ListRevisions(ctx, &runpb.ListRevisionsRequest{
		Parent: serviceName(project, region, service),
	})

	var revisions []*Revision
	for {
		rev, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list revisions for service %s: %w", service, err)
		}
		revisions = append(revisions, revisionFromProto(rev))
	}
	return revisions, nil
}

func (c *gcpClient) GetRevision(ctx context.Context, project, region, name string) (*Revision, error) {
	fullName := name
	if !strings.HasPrefix(fullName, "projects/") {
		fullName = fmt.Sprintf("projects/%s/locations/%s/revisions/%s", project, region, name)
	}
	rev, err := c.revisions.GetRevision(ctx, &runpb.GetRevisionRequest{Name: fullName})
	if err != nil {
		return nil, fmt.Errorf("get revision %s: %w", name, err)
	}
	return revisionFromProto(rev), nil
}

func (c *gcpClient) UpdateTraffic(ctx context.Context, project, region, service string, req *UpdateTrafficRequest) (*Service, error) {
	svc, err := c.services.GetService(ctx, &runpb.GetServiceRequest{Name: serviceName(project, region, service)})
	if err != nil {
		return nil, fmt.Errorf("get service %s: %w", service, err)
	}

	percent := req.Percent
	if percent == 0 {
		percent = 100
	}

	target := &runpb.TrafficTarget{
		Percent: percent,
		Tag:     req.Tag,
	}
	if req.ToLatest {
		target.Type = runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST
	} else {
		target.Type = runpb.TrafficTargetAllocationType_TRAFFIC_TARGET_ALLOCATION_TYPE_REVISION
		target.Revision = req.Revision
	}
	svc.Traffic = []*runpb.TrafficTarget{target}

	op, err := c.services.UpdateService(ctx, &runpb.UpdateServiceRequest{Service: svc})
	if err != nil {
		return nil, fmt.Errorf("update traffic for service %s: %w", service, err)
	}
	updated, err := op.Wait(ctx)
	if err != nil {
		return nil, fmt.Errorf("wait for traffic update %s: %w", service, err)
	}
	return svcFromProto(updated, region), nil
}

func (c *gcpClient) GetServicePolicy(ctx context.Context, project, region, service string) (*iampb.Policy, error) {
	policy, err := c.services.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: serviceName(project, region, service),
	})
	if err != nil {
		return nil, fmt.Errorf("get service iam policy for %s: %w", service, err)
	}
	return policy, nil
}

func (c *gcpClient) SetServicePolicy(ctx context.Context, project, region, service string, policy *iampb.Policy) (*iampb.Policy, error) {
	updated, err := c.services.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: serviceName(project, region, service),
		Policy:   policy,
	})
	if err != nil {
		return nil, fmt.Errorf("set service iam policy for %s: %w", service, err)
	}
	return updated, nil
}

func (c *gcpClient) TestServicePermissions(ctx context.Context, project, region, service string, permissions []string) ([]string, error) {
	resp, err := c.services.TestIamPermissions(ctx, &iampb.TestIamPermissionsRequest{
		Resource:    serviceName(project, region, service),
		Permissions: permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("test service iam permissions for %s: %w", service, err)
	}
	return resp.GetPermissions(), nil
}

func svcFromProto(svc *runpb.Service, region string) *Service {
	var traffic []*TrafficTarget
	for _, target := range svc.GetTrafficStatuses() {
		traffic = append(traffic, &TrafficTarget{
			Type:     target.GetType().String(),
			Revision: target.GetRevision(),
			Percent:  target.GetPercent(),
			Tag:      target.GetTag(),
		})
	}
	return &Service{
		Name:                  svc.GetName(),
		URI:                   svc.GetUri(),
		Region:                region,
		LatestReadyRevision:   svc.GetLatestReadyRevision(),
		LatestCreatedRevision: svc.GetLatestCreatedRevision(),
		Traffic:               traffic,
	}
}

func revisionFromProto(rev *runpb.Revision) *Revision {
	image := ""
	containers := rev.GetContainers()
	if len(containers) > 0 {
		image = containers[0].GetImage()
	}
	createTime := ""
	if rev.GetCreateTime() != nil {
		createTime = rev.GetCreateTime().AsTime().Format(time.RFC3339)
	}
	return &Revision{
		Name:        rev.GetName(),
		Service:     rev.GetService(),
		Image:       image,
		CreateTime:  createTime,
		Generation:  rev.GetGeneration(),
		Reconciling: rev.GetReconciling(),
	}
}

func serviceName(project, region, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/services/%s", project, region, name)
}

func serviceFromExisting(name string, existing *runpb.Service) *runpb.Service {
	if existing != nil {
		return &runpb.Service{
			Name:               name,
			Ingress:            existing.GetIngress(),
			Template:           existing.GetTemplate(),
			Traffic:            existing.GetTraffic(),
			Scaling:            existing.GetScaling(),
			InvokerIamDisabled: existing.GetInvokerIamDisabled(),
			DefaultUriDisabled: existing.GetDefaultUriDisabled(),
		}
	}
	return &runpb.Service{
		Name:     name,
		Template: &runpb.RevisionTemplate{},
	}
}

func ensurePrimaryContainer(service *runpb.Service) *runpb.Container {
	if service.Template == nil {
		service.Template = &runpb.RevisionTemplate{}
	}
	if len(service.Template.Containers) == 0 {
		service.Template.Containers = []*runpb.Container{{}}
	}
	return service.Template.Containers[0]
}

func envVarsFromMap(env map[string]string) []*runpb.EnvVar {
	var envVars []*runpb.EnvVar
	for k, v := range env {
		envVars = append(envVars, &runpb.EnvVar{
			Name:   k,
			Values: &runpb.EnvVar_Value{Value: v},
		})
	}
	return envVars
}
