package compute

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	instances      []*Instance
	instanceMap    map[string]*Instance
	firewalls      []*FirewallRule
	disks          []*Disk
	diskMap        map[string]*Disk
	snapshots      []*Snapshot
	snapshotMap    map[string]*Snapshot
	templates      []*InstanceTemplate
	templateMap    map[string]*InstanceTemplate
	groups         []*ManagedInstanceGroup
	groupMap       map[string]*ManagedInstanceGroup
	autoscalers    []*Autoscaler
	autoscalerMap  map[string]*Autoscaler
	images         []*Image
	imageMap       map[string]*Image
	vpnTunnels     []*VPNTunnel
	vpnTunnelMap   map[string]*VPNTunnel
	uigs           []*UnmanagedInstanceGroup
	uigMap         map[string]*UnmanagedInstanceGroup
	listErr        error
	getErr         error
	createErr      error
	deleteErr      error
	startErr       error
	stopErr        error
	resetErr       error
	fwListErr      error
	fwCreateErr    error
	fwDeleteErr    error
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

func (m *mockClient) ListDisks(_ context.Context, _, _ string) ([]*Disk, error) {
	return m.disks, m.listErr
}

func (m *mockClient) GetDisk(_ context.Context, _, _, name string) (*Disk, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if disk, ok := m.diskMap[name]; ok {
		return disk, nil
	}
	return nil, fmt.Errorf("disk %q not found", name)
}

func (m *mockClient) CreateDisk(_ context.Context, _, _ string, _ *CreateDiskRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteDisk(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListSnapshots(_ context.Context, _ string) ([]*Snapshot, error) {
	return m.snapshots, m.listErr
}

func (m *mockClient) GetSnapshot(_ context.Context, _, name string) (*Snapshot, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if snapshot, ok := m.snapshotMap[name]; ok {
		return snapshot, nil
	}
	return nil, fmt.Errorf("snapshot %q not found", name)
}

func (m *mockClient) CreateSnapshot(_ context.Context, _, _ string, _ *CreateSnapshotRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteSnapshot(_ context.Context, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListInstanceTemplates(_ context.Context, _ string) ([]*InstanceTemplate, error) {
	return m.templates, m.listErr
}

func (m *mockClient) GetInstanceTemplate(_ context.Context, _, name string) (*InstanceTemplate, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if tpl, ok := m.templateMap[name]; ok {
		return tpl, nil
	}
	return nil, fmt.Errorf("instance template %q not found", name)
}

func (m *mockClient) CreateInstanceTemplate(_ context.Context, _ string, _ *CreateInstanceTemplateRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteInstanceTemplate(_ context.Context, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListInstanceGroupManagers(_ context.Context, _, _ string) ([]*ManagedInstanceGroup, error) {
	return m.groups, m.listErr
}

func (m *mockClient) GetInstanceGroupManager(_ context.Context, _, _, name string) (*ManagedInstanceGroup, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if group, ok := m.groupMap[name]; ok {
		return group, nil
	}
	return nil, fmt.Errorf("managed instance group %q not found", name)
}

func (m *mockClient) CreateInstanceGroupManager(_ context.Context, _, _ string, _ *CreateInstanceGroupManagerRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteInstanceGroupManager(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListAutoscalers(_ context.Context, _, _ string) ([]*Autoscaler, error) {
	return m.autoscalers, m.listErr
}

func (m *mockClient) GetAutoscaler(_ context.Context, _, _, name string) (*Autoscaler, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if autoscaler, ok := m.autoscalerMap[name]; ok {
		return autoscaler, nil
	}
	return nil, fmt.Errorf("autoscaler %q not found", name)
}

func (m *mockClient) CreateAutoscaler(_ context.Context, _, _ string, _ *CreateAutoscalerRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteAutoscaler(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListImages(_ context.Context, _ string) ([]*Image, error) {
	return m.images, m.listErr
}

func (m *mockClient) GetImage(_ context.Context, _, name string) (*Image, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if img, ok := m.imageMap[name]; ok {
		return img, nil
	}
	return nil, fmt.Errorf("image %q not found", name)
}

func (m *mockClient) CreateImage(_ context.Context, _ string, _ *CreateImageRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteImage(_ context.Context, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListVPNTunnels(_ context.Context, _, _ string) ([]*VPNTunnel, error) {
	return m.vpnTunnels, m.listErr
}

func (m *mockClient) GetVPNTunnel(_ context.Context, _, _, name string) (*VPNTunnel, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if t, ok := m.vpnTunnelMap[name]; ok {
		return t, nil
	}
	return nil, fmt.Errorf("vpn tunnel %q not found", name)
}

func (m *mockClient) CreateVPNTunnel(_ context.Context, _, _ string, _ *CreateVPNTunnelRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteVPNTunnel(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListUnmanagedInstanceGroups(_ context.Context, _, _ string) ([]*UnmanagedInstanceGroup, error) {
	return m.uigs, m.listErr
}

func (m *mockClient) GetUnmanagedInstanceGroup(_ context.Context, _, _, name string) (*UnmanagedInstanceGroup, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	if g, ok := m.uigMap[name]; ok {
		return g, nil
	}
	return nil, fmt.Errorf("unmanaged instance group %q not found", name)
}

func (m *mockClient) CreateUnmanagedInstanceGroup(_ context.Context, _, _ string, _ *CreateUnmanagedInstanceGroupRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteUnmanagedInstanceGroup(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) SetTags(_ context.Context, _, _, _ string, _ []string) error {
	return nil
}

func (m *mockClient) SetMachineType(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (m *mockClient) AttachDisk(_ context.Context, _, _, _, _ string, _ bool) error {
	return nil
}

func (m *mockClient) DetachDisk(_ context.Context, _, _, _, _ string) error {
	return nil
}

func (m *mockClient) ListSSLCertificates(_ context.Context, _ string) ([]*SSLCertificate, error) {
	return nil, nil
}

func (m *mockClient) GetSSLCertificate(_ context.Context, _, _ string) (*SSLCertificate, error) {
	return nil, nil
}

func (m *mockClient) CreateSSLCertificate(_ context.Context, _ string, _ *CreateSSLCertificateRequest) error {
	return nil
}

func (m *mockClient) DeleteSSLCertificate(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClient) ListSecurityPolicies(_ context.Context, _ string) ([]*SecurityPolicy, error) {
	return nil, nil
}

func (m *mockClient) GetSecurityPolicy(_ context.Context, _, _ string) (*SecurityPolicy, error) {
	return nil, nil
}

func (m *mockClient) CreateSecurityPolicy(_ context.Context, _ string, _ *CreateSecurityPolicyRequest) error {
	return nil
}

func (m *mockClient) DeleteSecurityPolicy(_ context.Context, _, _ string) error {
	return nil
}

func (m *mockClient) AddSecurityPolicyRule(_ context.Context, _, _ string, _ *SecurityPolicyRuleRequest) error {
	return nil
}

func (m *mockClient) RemoveSecurityPolicyRule(_ context.Context, _, _ string, _ int32) error {
	return nil
}

func (m *mockClient) ListZones(_ context.Context, _, _ string) ([]*Zone, error) {
	return nil, nil
}

func (m *mockClient) ListRegions(_ context.Context, _ string) ([]*Region, error) {
	return nil, nil
}

func (m *mockClient) ListMachineTypes(_ context.Context, _, _ string) ([]*MachineType, error) {
	return nil, nil
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
