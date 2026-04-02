package projects

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	projects []*Project
	getMap   map[string]*Project
	listErr  error
	getErr   error
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
