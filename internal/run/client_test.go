package run

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	services   []*Service
	serviceMap map[string]*Service
	listErr    error
	getErr     error
	deployErr  error
	deleteErr  error
}

func (m *mockClient) ListServices(_ context.Context, _, _ string) ([]*Service, error) {
	return m.services, m.listErr
}

func (m *mockClient) GetService(_ context.Context, _, _, name string) (*Service, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	s, ok := m.serviceMap[name]
	if !ok {
		return nil, fmt.Errorf("service %q not found", name)
	}
	return s, nil
}

func (m *mockClient) Deploy(_ context.Context, _, _ string, _ *DeployRequest) error {
	return m.deployErr
}

func (m *mockClient) DeleteService(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func TestMockListServices(t *testing.T) {
	mock := &mockClient{
		services: []*Service{
			{Name: "svc-1", Region: "us-central1", URI: "https://svc-1.run.app"},
		},
	}

	svcs, err := mock.ListServices(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(svcs) != 1 {
		t.Errorf("expected 1, got %d", len(svcs))
	}
}

func TestMockGetService(t *testing.T) {
	mock := &mockClient{
		serviceMap: map[string]*Service{
			"svc-1": {Name: "svc-1", URI: "https://svc-1.run.app"},
		},
	}

	s, err := mock.GetService(context.Background(), "proj", "us-central1", "svc-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s.URI != "https://svc-1.run.app" {
		t.Errorf("uri: got %q", s.URI)
	}

	_, err = mock.GetService(context.Background(), "proj", "us-central1", "nope")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockDeploy(t *testing.T) {
	mock := &mockClient{}
	err := mock.Deploy(context.Background(), "proj", "us-central1", &DeployRequest{
		Name:  "svc-1",
		Image: "gcr.io/proj/img:latest",
	})
	if err != nil {
		t.Fatalf("deploy: %v", err)
	}
}

func TestMockDeployError(t *testing.T) {
	mock := &mockClient{deployErr: fmt.Errorf("quota exceeded")}
	err := mock.Deploy(context.Background(), "proj", "us-central1", &DeployRequest{})
	if err == nil {
		t.Fatal("expected error")
	}
}
