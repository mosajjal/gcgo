package projects

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	projects    []*Project
	getMap      map[string]*Project
	listErr     error
	getErr      error
	createErr   error
	deleteErr   error
	createdID   string
	deletedID   string
}

func (m *mockClient) List(_ context.Context) ([]*Project, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.projects, nil
}

func (m *mockClient) Get(_ context.Context, id string) (*Project, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.getMap[id]
	if !ok {
		return nil, fmt.Errorf("project %q not found", id)
	}
	return p, nil
}

func (m *mockClient) CreateProject(_ context.Context, projectID, _ string, _ map[string]string) error {
	if m.createErr != nil {
		return m.createErr
	}
	m.createdID = projectID
	return nil
}

func (m *mockClient) DeleteProject(_ context.Context, projectID string) error {
	if m.deleteErr != nil {
		return m.deleteErr
	}
	m.deletedID = projectID
	return nil
}

func TestMockClientList(t *testing.T) {
	mock := &mockClient{
		projects: []*Project{
			{ID: "proj-1", Name: "Project One", State: "ACTIVE"},
			{ID: "proj-2", Name: "Project Two", State: "ACTIVE"},
		},
	}

	projects, err := mock.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(projects) != 2 {
		t.Errorf("expected 2 projects, got %d", len(projects))
	}
}

func TestMockClientListError(t *testing.T) {
	mock := &mockClient{
		listErr: fmt.Errorf("permission denied"),
	}

	_, err := mock.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClientGet(t *testing.T) {
	mock := &mockClient{
		getMap: map[string]*Project{
			"proj-1": {ID: "proj-1", Name: "Project One", State: "ACTIVE"},
		},
	}

	p, err := mock.Get(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if p.ID != "proj-1" {
		t.Errorf("ID: got %q", p.ID)
	}
}

func TestMockClientGetNotFound(t *testing.T) {
	mock := &mockClient{
		getMap: map[string]*Project{},
	}

	_, err := mock.Get(context.Background(), "nope")
	if err == nil {
		t.Fatal("expected error for missing project")
	}
}

func TestMockClientCreateProject(t *testing.T) {
	mock := &mockClient{}

	if err := mock.CreateProject(context.Background(), "new-proj", "New Project", nil); err != nil {
		t.Fatalf("create: %v", err)
	}
	if mock.createdID != "new-proj" {
		t.Errorf("createdID: got %q, want %q", mock.createdID, "new-proj")
	}
}

func TestMockClientCreateProjectError(t *testing.T) {
	mock := &mockClient{createErr: fmt.Errorf("quota exceeded")}

	if err := mock.CreateProject(context.Background(), "x", "", nil); err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClientDeleteProject(t *testing.T) {
	mock := &mockClient{}

	if err := mock.DeleteProject(context.Background(), "old-proj"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if mock.deletedID != "old-proj" {
		t.Errorf("deletedID: got %q, want %q", mock.deletedID, "old-proj")
	}
}

func TestMockClientDeleteProjectError(t *testing.T) {
	mock := &mockClient{deleteErr: fmt.Errorf("permission denied")}

	if err := mock.DeleteProject(context.Background(), "x"); err == nil {
		t.Fatal("expected error")
	}
}
