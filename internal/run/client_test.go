package run

import (
	"context"
	"fmt"
	"testing"

	"cloud.google.com/go/iam/apiv1/iampb"
)

type mockClient struct {
	services    []*Service
	serviceMap  map[string]*Service
	revisions   []*Revision
	revisionMap map[string]*Revision
	listErr     error
	getErr      error
	deployErr   error
	deleteErr   error
	revisionErr error
	trafficErr  error
	policy      *iampb.Policy
	permErr     error
	setPolicy   *iampb.Policy
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

func (m *mockClient) ListRevisions(_ context.Context, _, _, _ string) ([]*Revision, error) {
	return m.revisions, m.revisionErr
}

func (m *mockClient) GetRevision(_ context.Context, _, _, name string) (*Revision, error) {
	if m.revisionErr != nil {
		return nil, m.revisionErr
	}
	rev, ok := m.revisionMap[name]
	if !ok {
		return nil, fmt.Errorf("revision %q not found", name)
	}
	return rev, nil
}

func (m *mockClient) UpdateTraffic(_ context.Context, _, _, _ string, _ *UpdateTrafficRequest) (*Service, error) {
	if m.trafficErr != nil {
		return nil, m.trafficErr
	}
	return &Service{Name: "svc-1"}, nil
}

func (m *mockClient) GetServicePolicy(_ context.Context, _, _, _ string) (*iampb.Policy, error) {
	return m.policy, nil
}

func (m *mockClient) SetServicePolicy(_ context.Context, _, _, _ string, policy *iampb.Policy) (*iampb.Policy, error) {
	m.setPolicy = policy
	return policy, nil
}

func (m *mockClient) TestServicePermissions(_ context.Context, _, _, _ string, permissions []string) ([]string, error) {
	if m.permErr != nil {
		return nil, m.permErr
	}
	return permissions, nil
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

func TestMockRevisionLookup(t *testing.T) {
	mock := &mockClient{
		revisionMap: map[string]*Revision{
			"rev-1": {Name: "rev-1", Service: "svc-1"},
		},
	}

	rev, err := mock.GetRevision(context.Background(), "proj", "us-central1", "rev-1")
	if err != nil {
		t.Fatalf("get revision: %v", err)
	}
	if rev.Name != "rev-1" {
		t.Fatalf("name: got %q", rev.Name)
	}
}

func TestMockServicePolicyMutation(t *testing.T) {
	mock := &mockClient{
		policy: &iampb.Policy{
			Bindings: []*iampb.Binding{{Role: "roles/run.invoker", Members: []string{"user:a@example.com"}}},
		},
	}

	policy, err := mock.GetServicePolicy(context.Background(), "proj", "us-central1", "svc-1")
	if err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if len(policy.GetBindings()) != 1 {
		t.Fatalf("bindings: got %d", len(policy.GetBindings()))
	}

	updated, err := mock.SetServicePolicy(context.Background(), "proj", "us-central1", "svc-1", policy)
	if err != nil {
		t.Fatalf("set policy: %v", err)
	}
	if updated == nil {
		t.Fatal("expected policy")
	}
}

func TestMockServicePermissions(t *testing.T) {
	mock := &mockClient{}
	allowed, err := mock.TestServicePermissions(context.Background(), "proj", "us-central1", "svc-1", []string{"run.services.get"})
	if err != nil {
		t.Fatalf("test permissions: %v", err)
	}
	if len(allowed) != 1 || allowed[0] != "run.services.get" {
		t.Fatalf("allowed = %v", allowed)
	}
}
