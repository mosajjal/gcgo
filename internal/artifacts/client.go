package artifacts

import (
	"context"
	"fmt"
	"strings"

	artifactregistry "google.golang.org/api/artifactregistry/v1"
	"google.golang.org/api/option"
)

// Repository holds artifact repository fields.
type Repository struct {
	Name        string `json:"name"`
	Format      string `json:"format"`
	Description string `json:"description"`
	SizeBytes   int64  `json:"size_bytes"`
	CreateTime  string `json:"create_time"`
	UpdateTime  string `json:"update_time"`
}

// Package holds artifact package fields.
type Package struct {
	Name       string `json:"name"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

// Version holds artifact version fields.
type Version struct {
	Name       string `json:"name"`
	CreateTime string `json:"create_time"`
	UpdateTime string `json:"update_time"`
}

// CreateRepositoryRequest holds parameters for repository creation.
type CreateRepositoryRequest struct {
	RepositoryID string
	Format       string
	Description  string
}

// Client defines the operations we use for Artifact Registry.
type Client interface {
	ListRepositories(ctx context.Context, project, location string) ([]*Repository, error)
	GetRepository(ctx context.Context, project, location, repo string) (*Repository, error)
	CreateRepository(ctx context.Context, project, location string, req *CreateRepositoryRequest) (*Repository, error)
	DeleteRepository(ctx context.Context, project, location, repo string) error
	ListPackages(ctx context.Context, project, location, repo string) ([]*Package, error)
	DeletePackage(ctx context.Context, project, location, repo, pkg string) error
	ListVersions(ctx context.Context, project, location, repo, pkg string) ([]*Version, error)
	DeleteVersion(ctx context.Context, project, location, repo, pkg, version string) error
}

type gcpClient struct {
	svc *artifactregistry.Service
}

// NewClient creates a Client backed by the real Artifact Registry API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := artifactregistry.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create artifact registry client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListRepositories(ctx context.Context, project, location string) ([]*Repository, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)

	var repos []*Repository
	if err := c.svc.Projects.Locations.Repositories.List(parent).Context(ctx).Pages(ctx, func(resp *artifactregistry.ListRepositoriesResponse) error {
		for _, r := range resp.Repositories {
			repos = append(repos, repoFromAPI(r))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list repositories: %w", err)
	}
	return repos, nil
}

func (c *gcpClient) GetRepository(ctx context.Context, project, location, repo string) (*Repository, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repo)
	r, err := c.svc.Projects.Locations.Repositories.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get repository %s: %w", repo, err)
	}
	return repoFromAPI(r), nil
}

func (c *gcpClient) CreateRepository(ctx context.Context, project, location string, req *CreateRepositoryRequest) (*Repository, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	repository := &artifactregistry.Repository{
		Description: req.Description,
		Format:      strings.ToUpper(req.Format),
	}

	if _, err := c.svc.Projects.Locations.Repositories.Create(parent, repository).
		RepositoryId(req.RepositoryID).
		Context(ctx).
		Do(); err != nil {
		return nil, fmt.Errorf("create repository %s: %w", req.RepositoryID, err)
	}
	return &Repository{
		Name:        fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, req.RepositoryID),
		Format:      strings.ToUpper(req.Format),
		Description: req.Description,
	}, nil
}

func (c *gcpClient) DeleteRepository(ctx context.Context, project, location, repo string) error {
	name := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repo)
	if _, err := c.svc.Projects.Locations.Repositories.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete repository %s: %w", repo, err)
	}
	return nil
}

func (c *gcpClient) ListPackages(ctx context.Context, project, location, repo string) ([]*Package, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/repositories/%s", project, location, repo)

	var pkgs []*Package
	if err := c.svc.Projects.Locations.Repositories.Packages.List(parent).Context(ctx).Pages(ctx, func(resp *artifactregistry.ListPackagesResponse) error {
		for _, p := range resp.Packages {
			pkgs = append(pkgs, packageFromAPI(p))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list packages: %w", err)
	}
	return pkgs, nil
}

func (c *gcpClient) DeletePackage(ctx context.Context, project, location, repo, pkg string) error {
	name := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s", project, location, repo, pkg)
	if _, err := c.svc.Projects.Locations.Repositories.Packages.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete package %s: %w", pkg, err)
	}
	return nil
}

func (c *gcpClient) ListVersions(ctx context.Context, project, location, repo, pkg string) ([]*Version, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s", project, location, repo, pkg)

	var versions []*Version
	if err := c.svc.Projects.Locations.Repositories.Packages.Versions.List(parent).Context(ctx).Pages(ctx, func(resp *artifactregistry.ListVersionsResponse) error {
		for _, v := range resp.Versions {
			versions = append(versions, versionFromAPI(v))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list versions: %w", err)
	}
	return versions, nil
}

func (c *gcpClient) DeleteVersion(ctx context.Context, project, location, repo, pkg, version string) error {
	name := fmt.Sprintf("projects/%s/locations/%s/repositories/%s/packages/%s/versions/%s",
		project, location, repo, pkg, version)
	if _, err := c.svc.Projects.Locations.Repositories.Packages.Versions.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete version %s: %w", version, err)
	}
	return nil
}

func repoFromAPI(r *artifactregistry.Repository) *Repository {
	return &Repository{
		Name:        r.Name,
		Format:      strings.ToUpper(r.Format),
		Description: r.Description,
		SizeBytes:   r.SizeBytes,
		CreateTime:  r.CreateTime,
		UpdateTime:  r.UpdateTime,
	}
}

func packageFromAPI(p *artifactregistry.Package) *Package {
	return &Package{
		Name:       p.Name,
		CreateTime: p.CreateTime,
		UpdateTime: p.UpdateTime,
	}
}

func versionFromAPI(v *artifactregistry.Version) *Version {
	return &Version{
		Name:       v.Name,
		CreateTime: v.CreateTime,
		UpdateTime: v.UpdateTime,
	}
}
