package compute

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	instances   []*Instance
	instanceMap map[string]*Instance
	firewalls   []*FirewallRule
	listErr     error
	getErr      error
	createErr   error
	deleteErr   error
	startErr    error
	stopErr     error
	resetErr    error
	fwListErr   error
	fwCreateErr error
	fwDeleteErr error
}

func (m *mockClient) ListInstances(_ context.Context, _, _ string) ([]*Instance, error) {
	return m.instances, m.listErr
}

func (m *mockClient) GetInstance(_ context.Context, _, _, name string) (*Instance, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	inst, ok := m.instanceMap[name]
	if !ok {
		return nil, fmt.Errorf("instance %q not found", name)
	}
	return inst, nil
}

func (m *mockClient) CreateInstance(_ context.Context, _, _ string, _ *CreateInstanceRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteInstance(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) StartInstance(_ context.Context, _, _, _ string) error {
	return m.startErr
}

func (m *mockClient) StopInstance(_ context.Context, _, _, _ string) error {
	return m.stopErr
}

func (m *mockClient) ResetInstance(_ context.Context, _, _, _ string) error {
	return m.resetErr
}

func (m *mockClient) ListFirewallRules(_ context.Context, _ string) ([]*FirewallRule, error) {
	return m.firewalls, m.fwListErr
}

func (m *mockClient) CreateFirewallRule(_ context.Context, _ string, _ *CreateFirewallRequest) error {
	return m.fwCreateErr
}

func (m *mockClient) DeleteFirewallRule(_ context.Context, _, _ string) error {
	return m.fwDeleteErr
}

func TestMockListInstances(t *testing.T) {
	mock := &mockClient{
		instances: []*Instance{
			{Name: "vm-1", Zone: "us-central1-a", Status: "RUNNING", InternalIP: "10.0.0.1"},
			{Name: "vm-2", Zone: "us-central1-a", Status: "TERMINATED", InternalIP: "10.0.0.2"},
		},
	}

	instances, err := mock.ListInstances(context.Background(), "proj", "us-central1-a")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}
}

func TestMockGetInstance(t *testing.T) {
	mock := &mockClient{
		instanceMap: map[string]*Instance{
			"vm-1": {Name: "vm-1", Status: "RUNNING"},
		},
	}

	inst, err := mock.GetInstance(context.Background(), "proj", "zone", "vm-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if inst.Name != "vm-1" {
		t.Errorf("name: got %q", inst.Name)
	}

	_, err = mock.GetInstance(context.Background(), "proj", "zone", "nope")
	if err == nil {
		t.Fatal("expected error for missing instance")
	}
}

func TestMockLifecycleOps(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func(Client) error
		err  error
	}{
		{
			name: "start success",
			fn:   func(c Client) error { return c.StartInstance(ctx, "p", "z", "vm") },
		},
		{
			name: "stop success",
			fn:   func(c Client) error { return c.StopInstance(ctx, "p", "z", "vm") },
		},
		{
			name: "reset success",
			fn:   func(c Client) error { return c.ResetInstance(ctx, "p", "z", "vm") },
		},
		{
			name: "start error",
			fn:   func(c Client) error { return c.StartInstance(ctx, "p", "z", "vm") },
			err:  fmt.Errorf("nope"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				startErr: tt.err,
				stopErr:  tt.err,
				resetErr: tt.err,
			}
			err := tt.fn(mock)
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockListFirewallRules(t *testing.T) {
	mock := &mockClient{
		firewalls: []*FirewallRule{
			{Name: "allow-http", Direction: "INGRESS", Allowed: []string{"tcp:80"}},
		},
	}

	rules, err := mock.ListFirewallRules(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(rules))
	}
}

func TestParseAllowed(t *testing.T) {
	tests := []struct {
		input     string
		wantProto string
		wantPorts int
	}{
		{"tcp:80", "tcp", 1},
		{"udp:53", "udp", 1},
		{"icmp", "icmp", 0},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			a := parseAllowed(tt.input)
			if a.GetIPProtocol() != tt.wantProto {
				t.Errorf("proto: got %q, want %q", a.GetIPProtocol(), tt.wantProto)
			}
			if len(a.GetPorts()) != tt.wantPorts {
				t.Errorf("ports: got %d, want %d", len(a.GetPorts()), tt.wantPorts)
			}
		})
	}
}
