package artifacts

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	repos         []*Repository
	repoMap       map[string]*Repository
	packages      []*Package
	pkgMap        map[string]*Package
	versions      []*Version
	repoListErr   error
	repoGetErr    error
	repoCreateErr error
	repoDeleteErr error
	pkgListErr    error
	pkgDeleteErr  error
	verListErr    error
	verDeleteErr  error
}

func (m *mockClient) ListRepositories(_ context.Context, _, _ string) ([]*Repository, error) {
	return m.repos, m.repoListErr
}

func (m *mockClient) GetRepository(_ context.Context, _, _, name string) (*Repository, error) {
	if m.repoGetErr != nil {
		return nil, m.repoGetErr
	}
	r, ok := m.repoMap[name]
	if !ok {
		return nil, fmt.Errorf("repository %q not found", name)
	}
	return r, nil
}

func (m *mockClient) CreateRepository(_ context.Context, _, _ string, req *CreateRepositoryRequest) (*Repository, error) {
	if m.repoCreateErr != nil {
		return nil, m.repoCreateErr
	}
	return &Repository{Name: req.RepositoryID, Format: req.Format}, nil
}

func (m *mockClient) DeleteRepository(_ context.Context, _, _, _ string) error {
	return m.repoDeleteErr
}

func (m *mockClient) ListPackages(_ context.Context, _, _, _ string) ([]*Package, error) {
	return m.packages, m.pkgListErr
}

func (m *mockClient) DeletePackage(_ context.Context, _, _, _, _ string) error {
	return m.pkgDeleteErr
}

func (m *mockClient) ListVersions(_ context.Context, _, _, _, _ string) ([]*Version, error) {
	return m.versions, m.verListErr
}

func (m *mockClient) DeleteVersion(_ context.Context, _, _, _, _, _ string) error {
	return m.verDeleteErr
}

func TestMockListRepositories(t *testing.T) {
	mock := &mockClient{
		repos: []*Repository{
			{Name: "my-repo", Format: "DOCKER", Description: "Docker images"},
			{Name: "npm-repo", Format: "NPM", Description: "NPM packages"},
		},
	}

	repos, err := mock.ListRepositories(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(repos) != 2 {
		t.Errorf("expected 2 repos, got %d", len(repos))
	}
}

func TestMockListRepositoriesError(t *testing.T) {
	mock := &mockClient{repoListErr: fmt.Errorf("permission denied")}

	_, err := mock.ListRepositories(context.Background(), "proj", "us-central1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetRepository(t *testing.T) {
	mock := &mockClient{
		repoMap: map[string]*Repository{
			"my-repo": {Name: "my-repo", Format: "DOCKER"},
		},
	}

	r, err := mock.GetRepository(context.Background(), "proj", "us-central1", "my-repo")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if r.Format != "DOCKER" {
		t.Errorf("format: got %q", r.Format)
	}

	_, err = mock.GetRepository(context.Background(), "proj", "us-central1", "nope")
	if err == nil {
		t.Fatal("expected error for missing repo")
	}
}

func TestMockCreateRepository(t *testing.T) {
	mock := &mockClient{}
	r, err := mock.CreateRepository(context.Background(), "proj", "us-central1", &CreateRepositoryRequest{
		RepositoryID: "new-repo",
		Format:       "docker",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if r.Name != "new-repo" {
		t.Errorf("name: got %q", r.Name)
	}
}

func TestMockCreateRepositoryError(t *testing.T) {
	mock := &mockClient{repoCreateErr: fmt.Errorf("already exists")}
	_, err := mock.CreateRepository(context.Background(), "proj", "us-central1", &CreateRepositoryRequest{
		RepositoryID: "x",
	})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockDeleteRepository(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"error", fmt.Errorf("not found")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{repoDeleteErr: tt.err}
			err := mock.DeleteRepository(context.Background(), "proj", "us-central1", "repo")
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockListPackages(t *testing.T) {
	mock := &mockClient{
		packages: []*Package{
			{Name: "pkg-1"},
			{Name: "pkg-2"},
		},
	}

	pkgs, err := mock.ListPackages(context.Background(), "proj", "us-central1", "repo")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(pkgs) != 2 {
		t.Errorf("expected 2 packages, got %d", len(pkgs))
	}
}

func TestMockDeletePackage(t *testing.T) {
	mock := &mockClient{}
	err := mock.DeletePackage(context.Background(), "proj", "us-central1", "repo", "pkg-1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestMockListVersions(t *testing.T) {
	mock := &mockClient{
		versions: []*Version{
			{Name: "v1.0.0"},
			{Name: "v1.1.0"},
		},
	}

	versions, err := mock.ListVersions(context.Background(), "proj", "us-central1", "repo", "pkg")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(versions) != 2 {
		t.Errorf("expected 2 versions, got %d", len(versions))
	}
}

func TestMockDeleteVersion(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"error", fmt.Errorf("not found")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{verDeleteErr: tt.err}
			err := mock.DeleteVersion(context.Background(), "proj", "us-central1", "repo", "pkg", "v1")
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
