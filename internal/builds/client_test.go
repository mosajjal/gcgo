package builds

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	builds     []*Build
	buildMap   map[string]*Build
	triggers   []*Trigger
	triggerMap map[string]*Trigger
	listErr    error
	getErr     error
	cancelErr  error
	tListErr   error
	tGetErr    error
	tCreateErr error
	tDeleteErr error
	tRunErr    error
}

func (m *mockClient) ListBuilds(_ context.Context, _ string) ([]*Build, error) {
	return m.builds, m.listErr
}

func (m *mockClient) GetBuild(_ context.Context, _, id string) (*Build, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	b, ok := m.buildMap[id]
	if !ok {
		return nil, fmt.Errorf("build %q not found", id)
	}
	return b, nil
}

func (m *mockClient) CancelBuild(_ context.Context, _, _ string) error {
	return m.cancelErr
}

func (m *mockClient) ListTriggers(_ context.Context, _ string) ([]*Trigger, error) {
	return m.triggers, m.tListErr
}

func (m *mockClient) GetTrigger(_ context.Context, _, id string) (*Trigger, error) {
	if m.tGetErr != nil {
		return nil, m.tGetErr
	}
	t, ok := m.triggerMap[id]
	if !ok {
		return nil, fmt.Errorf("trigger %q not found", id)
	}
	return t, nil
}

func (m *mockClient) CreateTrigger(_ context.Context, _ string, req *CreateTriggerRequest) (*Trigger, error) {
	if m.tCreateErr != nil {
		return nil, m.tCreateErr
	}
	return &Trigger{ID: "new-id", Name: req.Name}, nil
}

func (m *mockClient) DeleteTrigger(_ context.Context, _, _ string) error {
	return m.tDeleteErr
}

func (m *mockClient) RunTrigger(_ context.Context, _, _ string) error {
	return m.tRunErr
}

func TestMockListBuilds(t *testing.T) {
	mock := &mockClient{
		builds: []*Build{
			{ID: "b-1", Status: "SUCCESS", Source: "my-repo"},
			{ID: "b-2", Status: "WORKING", Source: "my-repo"},
		},
	}

	builds, err := mock.ListBuilds(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(builds) != 2 {
		t.Errorf("expected 2 builds, got %d", len(builds))
	}
}

func TestMockListBuildsError(t *testing.T) {
	mock := &mockClient{listErr: fmt.Errorf("permission denied")}

	_, err := mock.ListBuilds(context.Background(), "proj")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetBuild(t *testing.T) {
	mock := &mockClient{
		buildMap: map[string]*Build{
			"b-1": {ID: "b-1", Status: "SUCCESS"},
		},
	}

	b, err := mock.GetBuild(context.Background(), "proj", "b-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if b.Status != "SUCCESS" {
		t.Errorf("status: got %q", b.Status)
	}

	_, err = mock.GetBuild(context.Background(), "proj", "nope")
	if err == nil {
		t.Fatal("expected error for missing build")
	}
}

func TestMockCancelBuild(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"error", fmt.Errorf("build already finished")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{cancelErr: tt.err}
			err := mock.CancelBuild(context.Background(), "proj", "b-1")
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockListTriggers(t *testing.T) {
	mock := &mockClient{
		triggers: []*Trigger{
			{ID: "t-1", Name: "deploy-trigger"},
		},
	}

	triggers, err := mock.ListTriggers(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(triggers) != 1 {
		t.Errorf("expected 1 trigger, got %d", len(triggers))
	}
}

func TestMockGetTrigger(t *testing.T) {
	mock := &mockClient{
		triggerMap: map[string]*Trigger{
			"t-1": {ID: "t-1", Name: "deploy-trigger"},
		},
	}

	tr, err := mock.GetTrigger(context.Background(), "proj", "t-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if tr.Name != "deploy-trigger" {
		t.Errorf("name: got %q", tr.Name)
	}

	_, err = mock.GetTrigger(context.Background(), "proj", "nope")
	if err == nil {
		t.Fatal("expected error for missing trigger")
	}
}

func TestMockCreateTrigger(t *testing.T) {
	mock := &mockClient{}
	tr, err := mock.CreateTrigger(context.Background(), "proj", &CreateTriggerRequest{
		Name:     "new-trigger",
		Filename: "cloudbuild.yaml",
	})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if tr.Name != "new-trigger" {
		t.Errorf("name: got %q", tr.Name)
	}
}

func TestMockCreateTriggerError(t *testing.T) {
	mock := &mockClient{tCreateErr: fmt.Errorf("quota exceeded")}
	_, err := mock.CreateTrigger(context.Background(), "proj", &CreateTriggerRequest{Name: "x"})
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockDeleteTrigger(t *testing.T) {
	mock := &mockClient{}
	err := mock.DeleteTrigger(context.Background(), "proj", "t-1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestMockRunTrigger(t *testing.T) {
	tests := []struct {
		name string
		err  error
	}{
		{"success", nil},
		{"error", fmt.Errorf("trigger disabled")},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{tRunErr: tt.err}
			err := mock.RunTrigger(context.Background(), "proj", "t-1")
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}
