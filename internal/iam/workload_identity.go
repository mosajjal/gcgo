package iam

import (
	"context"
	"fmt"

	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// WorkloadIdentityPool holds pool metadata.
type WorkloadIdentityPool struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	State       string `json:"state"`
	Disabled    bool   `json:"disabled"`
}

// WorkloadIdentityProvider holds provider metadata.
type WorkloadIdentityProvider struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	State       string `json:"state"`
	Disabled    bool   `json:"disabled"`
	Issuer      string `json:"issuer,omitempty"`
}

// WorkloadIdentityClient defines workload identity federation operations.
type WorkloadIdentityClient interface {
	ListPools(ctx context.Context, project, location string) ([]*WorkloadIdentityPool, error)
	CreatePool(ctx context.Context, project, location, poolID, displayName, description string) (*WorkloadIdentityPool, error)
	DeletePool(ctx context.Context, name string) error
	DescribePool(ctx context.Context, name string) (*WorkloadIdentityPool, error)
	ListProviders(ctx context.Context, poolName string) ([]*WorkloadIdentityProvider, error)
	CreateProvider(ctx context.Context, poolName, providerID, displayName, issuerURI string) (*WorkloadIdentityProvider, error)
	DeleteProvider(ctx context.Context, name string) error
	DescribeProvider(ctx context.Context, name string) (*WorkloadIdentityProvider, error)
}

type gcpWorkloadIdentityClient struct {
	svc *iamv1.Service
}

// NewWorkloadIdentityClient creates a WorkloadIdentityClient backed by the real GCP IAM API.
func NewWorkloadIdentityClient(ctx context.Context, opts ...option.ClientOption) (WorkloadIdentityClient, error) {
	svc, err := iamv1.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create workload identity client: %w", err)
	}
	return &gcpWorkloadIdentityClient{svc: svc}, nil
}

func (c *gcpWorkloadIdentityClient) ListPools(ctx context.Context, project, location string) ([]*WorkloadIdentityPool, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.svc.Projects.Locations.WorkloadIdentityPools.List(parent)

	var pools []*WorkloadIdentityPool
	if err := call.Pages(ctx, func(resp *iamv1.ListWorkloadIdentityPoolsResponse) error {
		for _, p := range resp.WorkloadIdentityPools {
			pools = append(pools, poolFromAPI(p))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list workload identity pools: %w", err)
	}
	return pools, nil
}

func (c *gcpWorkloadIdentityClient) CreatePool(ctx context.Context, project, location, poolID, displayName, description string) (*WorkloadIdentityPool, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	_, err := c.svc.Projects.Locations.WorkloadIdentityPools.Create(parent, &iamv1.WorkloadIdentityPool{
		DisplayName: displayName,
		Description: description,
	}).WorkloadIdentityPoolId(poolID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create workload identity pool %s: %w", poolID, err)
	}
	return &WorkloadIdentityPool{
		Name:        fmt.Sprintf("%s/workloadIdentityPools/%s", parent, poolID),
		DisplayName: displayName,
		Description: description,
		State:       "ACTIVE",
	}, nil
}

func (c *gcpWorkloadIdentityClient) DeletePool(ctx context.Context, name string) error {
	if _, err := c.svc.Projects.Locations.WorkloadIdentityPools.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete workload identity pool %s: %w", name, err)
	}
	return nil
}

func (c *gcpWorkloadIdentityClient) DescribePool(ctx context.Context, name string) (*WorkloadIdentityPool, error) {
	p, err := c.svc.Projects.Locations.WorkloadIdentityPools.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get workload identity pool %s: %w", name, err)
	}
	return poolFromAPI(p), nil
}

func (c *gcpWorkloadIdentityClient) ListProviders(ctx context.Context, poolName string) ([]*WorkloadIdentityProvider, error) {
	call := c.svc.Projects.Locations.WorkloadIdentityPools.Providers.List(poolName)

	var providers []*WorkloadIdentityProvider
	if err := call.Pages(ctx, func(resp *iamv1.ListWorkloadIdentityPoolProvidersResponse) error {
		for _, p := range resp.WorkloadIdentityPoolProviders {
			providers = append(providers, providerFromAPI(p))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list workload identity providers: %w", err)
	}
	return providers, nil
}

func (c *gcpWorkloadIdentityClient) CreateProvider(ctx context.Context, poolName, providerID, displayName, issuerURI string) (*WorkloadIdentityProvider, error) {
	_, err := c.svc.Projects.Locations.WorkloadIdentityPools.Providers.Create(poolName, &iamv1.WorkloadIdentityPoolProvider{
		DisplayName: displayName,
		Oidc: &iamv1.Oidc{
			IssuerUri: issuerURI,
		},
	}).WorkloadIdentityPoolProviderId(providerID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create workload identity provider %s: %w", providerID, err)
	}
	return &WorkloadIdentityProvider{
		Name:        fmt.Sprintf("%s/providers/%s", poolName, providerID),
		DisplayName: displayName,
		State:       "ACTIVE",
		Issuer:      issuerURI,
	}, nil
}

func (c *gcpWorkloadIdentityClient) DeleteProvider(ctx context.Context, name string) error {
	if _, err := c.svc.Projects.Locations.WorkloadIdentityPools.Providers.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete workload identity provider %s: %w", name, err)
	}
	return nil
}

func (c *gcpWorkloadIdentityClient) DescribeProvider(ctx context.Context, name string) (*WorkloadIdentityProvider, error) {
	p, err := c.svc.Projects.Locations.WorkloadIdentityPools.Providers.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get workload identity provider %s: %w", name, err)
	}
	return providerFromAPI(p), nil
}

func poolFromAPI(p *iamv1.WorkloadIdentityPool) *WorkloadIdentityPool {
	return &WorkloadIdentityPool{
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Description: p.Description,
		State:       p.State,
		Disabled:    p.Disabled,
	}
}

func providerFromAPI(p *iamv1.WorkloadIdentityPoolProvider) *WorkloadIdentityProvider {
	prov := &WorkloadIdentityProvider{
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Description: p.Description,
		State:       p.State,
		Disabled:    p.Disabled,
	}
	if p.Oidc != nil {
		prov.Issuer = p.Oidc.IssuerUri
	}
	return prov
}
