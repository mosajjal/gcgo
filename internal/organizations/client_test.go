package organizations

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	orgs      []*Organization
	getMap    map[string]*Organization
	bindings  []*IAMBinding
	listErr   error
	getErr    error
	policyErr error
}

func (m *mockClient) List(_ context.Context) ([]*Organization, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.orgs, nil
}

func (m *mockClient) Get(_ context.Context, id string) (*Organization, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	o, ok := m.getMap[id]
	if !ok {
		return nil, fmt.Errorf("organization %q not found", id)
	}
	return o, nil
}

func (m *mockClient) GetIAMPolicy(_ context.Context, _ string) ([]*IAMBinding, error) {
	if m.policyErr != nil {
		return nil, m.policyErr
	}
	return m.bindings, nil
}

func TestMockClientList(t *testing.T) {
	mock := &mockClient{
		orgs: []*Organization{
			{Name: "organizations/123", DisplayName: "Org One", State: "ACTIVE"},
			{Name: "organizations/456", DisplayName: "Org Two", State: "ACTIVE"},
		},
	}

	orgs, err := mock.List(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(orgs) != 2 {
		t.Errorf("expected 2 orgs, got %d", len(orgs))
	}
}

func TestMockClientListError(t *testing.T) {
	mock := &mockClient{listErr: fmt.Errorf("permission denied")}
	_, err := mock.List(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClientGet(t *testing.T) {
	mock := &mockClient{
		getMap: map[string]*Organization{
			"123": {Name: "organizations/123", DisplayName: "Org One", State: "ACTIVE"},
		},
	}

	o, err := mock.Get(context.Background(), "123")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if o.DisplayName != "Org One" {
		t.Errorf("display name: got %q", o.DisplayName)
	}
}

func TestMockClientGetNotFound(t *testing.T) {
	mock := &mockClient{getMap: map[string]*Organization{}}
	_, err := mock.Get(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error for missing org")
	}
}

func TestMockClientGetIAMPolicy(t *testing.T) {
	mock := &mockClient{
		bindings: []*IAMBinding{
			{Role: "roles/owner", Members: []string{"user:admin@example.com"}},
			{Role: "roles/viewer", Members: []string{"group:viewers@example.com"}},
		},
	}

	bindings, err := mock.GetIAMPolicy(context.Background(), "123")
	if err != nil {
		t.Fatalf("get iam policy: %v", err)
	}
	if len(bindings) != 2 {
		t.Errorf("expected 2 bindings, got %d", len(bindings))
	}
}

func TestMockClientGetIAMPolicyError(t *testing.T) {
	mock := &mockClient{policyErr: fmt.Errorf("access denied")}
	_, err := mock.GetIAMPolicy(context.Background(), "123")
	if err == nil {
		t.Fatal("expected error")
	}
}
