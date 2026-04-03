package compute

import (
	"context"
	"fmt"
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/spf13/cobra"
)

type mockNetworkClient struct {
	networks  []*Network
	networkM  map[string]*Network
	subnets   []*Subnet
	subnetM   map[string]*Subnet
	addresses []*Address
	addressM  map[string]*Address
	routers   []*Router
	routerM   map[string]*Router
	routes    []*Route
	routeM    map[string]*Route

	listErr   error
	getErr    error
	createErr error
	deleteErr error
}

func (m *mockNetworkClient) ListNetworks(_ context.Context, _ string) ([]*Network, error) {
	return m.networks, m.listErr
}
func (m *mockNetworkClient) GetNetwork(_ context.Context, _, name string) (*Network, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	n, ok := m.networkM[name]
	if !ok {
		return nil, fmt.Errorf("network %q not found", name)
	}
	return n, nil
}
func (m *mockNetworkClient) CreateNetwork(_ context.Context, _ string, _ *CreateNetworkRequest) error {
	return m.createErr
}
func (m *mockNetworkClient) DeleteNetwork(_ context.Context, _, _ string) error {
	return m.deleteErr
}

func (m *mockNetworkClient) ListSubnets(_ context.Context, _, _ string) ([]*Subnet, error) {
	return m.subnets, m.listErr
}
func (m *mockNetworkClient) GetSubnet(_ context.Context, _, _, name string) (*Subnet, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	s, ok := m.subnetM[name]
	if !ok {
		return nil, fmt.Errorf("subnet %q not found", name)
	}
	return s, nil
}
func (m *mockNetworkClient) CreateSubnet(_ context.Context, _, _ string, _ *CreateSubnetRequest) error {
	return m.createErr
}
func (m *mockNetworkClient) DeleteSubnet(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}
func (m *mockNetworkClient) ExpandSubnetIPRange(_ context.Context, _, _, _, _ string) error {
	return m.createErr
}

func (m *mockNetworkClient) ListAddresses(_ context.Context, _, _ string) ([]*Address, error) {
	return m.addresses, m.listErr
}
func (m *mockNetworkClient) GetAddress(_ context.Context, _, _, name string) (*Address, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	a, ok := m.addressM[name]
	if !ok {
		return nil, fmt.Errorf("address %q not found", name)
	}
	return a, nil
}
func (m *mockNetworkClient) CreateAddress(_ context.Context, _, _ string, _ *CreateAddressRequest) error {
	return m.createErr
}
func (m *mockNetworkClient) DeleteAddress(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockNetworkClient) ListRouters(_ context.Context, _, _ string) ([]*Router, error) {
	return m.routers, m.listErr
}
func (m *mockNetworkClient) GetRouter(_ context.Context, _, _, name string) (*Router, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	r, ok := m.routerM[name]
	if !ok {
		return nil, fmt.Errorf("router %q not found", name)
	}
	return r, nil
}
func (m *mockNetworkClient) CreateRouter(_ context.Context, _, _ string, _ *CreateRouterRequest) error {
	return m.createErr
}
func (m *mockNetworkClient) DeleteRouter(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockNetworkClient) ListRoutes(_ context.Context, _ string) ([]*Route, error) {
	return m.routes, m.listErr
}
func (m *mockNetworkClient) GetRoute(_ context.Context, _, name string) (*Route, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	r, ok := m.routeM[name]
	if !ok {
		return nil, fmt.Errorf("route %q not found", name)
	}
	return r, nil
}

func TestMockListNetworks(t *testing.T) {
	tests := []struct {
		name    string
		mock    *mockNetworkClient
		want    int
		wantErr bool
	}{
		{
			name: "two networks",
			mock: &mockNetworkClient{
				networks: []*Network{
					{Name: "default", RoutingMode: "REGIONAL"},
					{Name: "custom-vpc", RoutingMode: "GLOBAL"},
				},
			},
			want: 2,
		},
		{
			name:    "empty list",
			mock:    &mockNetworkClient{},
			want:    0,
			wantErr: false,
		},
		{
			name:    "list error",
			mock:    &mockNetworkClient{listErr: fmt.Errorf("api down")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			networks, err := tt.mock.ListNetworks(context.Background(), "proj")
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && len(networks) != tt.want {
				t.Errorf("got %d networks, want %d", len(networks), tt.want)
			}
		})
	}
}

func TestMockGetNetwork(t *testing.T) {
	mock := &mockNetworkClient{
		networkM: map[string]*Network{
			"default": {Name: "default", RoutingMode: "REGIONAL"},
		},
	}

	net, err := mock.GetNetwork(context.Background(), "proj", "default")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if net.Name != "default" {
		t.Errorf("name: got %q", net.Name)
	}

	_, err = mock.GetNetwork(context.Background(), "proj", "nope")
	if err == nil {
		t.Fatal("expected error for missing network")
	}
}

func TestMockListSubnets(t *testing.T) {
	tests := []struct {
		name    string
		mock    *mockNetworkClient
		want    int
		wantErr bool
	}{
		{
			name: "one subnet",
			mock: &mockNetworkClient{
				subnets: []*Subnet{
					{Name: "subnet-1", Region: "us-central1", IPCIDRRange: "10.0.0.0/24"},
				},
			},
			want: 1,
		},
		{
			name:    "list error",
			mock:    &mockNetworkClient{listErr: fmt.Errorf("fail")},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			subnets, err := tt.mock.ListSubnets(context.Background(), "proj", "us-central1")
			if (err != nil) != tt.wantErr {
				t.Fatalf("err=%v, wantErr=%v", err, tt.wantErr)
			}
			if !tt.wantErr && len(subnets) != tt.want {
				t.Errorf("got %d subnets, want %d", len(subnets), tt.want)
			}
		})
	}
}

func TestMockGetSubnet(t *testing.T) {
	mock := &mockNetworkClient{
		subnetM: map[string]*Subnet{
			"subnet-1": {Name: "subnet-1", IPCIDRRange: "10.0.0.0/24"},
		},
	}

	s, err := mock.GetSubnet(context.Background(), "proj", "us-central1", "subnet-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if s.IPCIDRRange != "10.0.0.0/24" {
		t.Errorf("cidr: got %q", s.IPCIDRRange)
	}
}

func TestMockListAddresses(t *testing.T) {
	mock := &mockNetworkClient{
		addresses: []*Address{
			{Name: "addr-1", Address: "35.1.2.3", Status: "RESERVED"},
			{Name: "addr-2", Address: "35.4.5.6", Status: "IN_USE"},
		},
	}

	addrs, err := mock.ListAddresses(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(addrs) != 2 {
		t.Errorf("got %d addresses, want 2", len(addrs))
	}
}

func TestMockGetAddress(t *testing.T) {
	mock := &mockNetworkClient{
		addressM: map[string]*Address{
			"addr-1": {Name: "addr-1", Address: "35.1.2.3"},
		},
	}

	a, err := mock.GetAddress(context.Background(), "proj", "us-central1", "addr-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if a.Address != "35.1.2.3" {
		t.Errorf("address: got %q", a.Address)
	}

	_, err = mock.GetAddress(context.Background(), "proj", "us-central1", "missing")
	if err == nil {
		t.Fatal("expected error for missing address")
	}
}

func TestMockListRouters(t *testing.T) {
	mock := &mockNetworkClient{
		routers: []*Router{
			{Name: "router-1", Region: "us-central1", BGPAsn: 64512},
		},
	}

	routers, err := mock.ListRouters(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(routers) != 1 {
		t.Errorf("got %d routers, want 1", len(routers))
	}
}

func TestMockGetRouter(t *testing.T) {
	mock := &mockNetworkClient{
		routerM: map[string]*Router{
			"router-1": {Name: "router-1", BGPAsn: 64512},
		},
	}

	r, err := mock.GetRouter(context.Background(), "proj", "us-central1", "router-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if r.BGPAsn != 64512 {
		t.Errorf("asn: got %d", r.BGPAsn)
	}
}

func TestMockListRoutes(t *testing.T) {
	mock := &mockNetworkClient{
		routes: []*Route{
			{Name: "default-route", DestRange: "0.0.0.0/0", Priority: 1000},
		},
	}

	routes, err := mock.ListRoutes(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(routes) != 1 {
		t.Errorf("got %d routes, want 1", len(routes))
	}
}

func TestMockGetRoute(t *testing.T) {
	mock := &mockNetworkClient{
		routeM: map[string]*Route{
			"default-route": {Name: "default-route", DestRange: "0.0.0.0/0"},
		},
	}

	r, err := mock.GetRoute(context.Background(), "proj", "default-route")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if r.DestRange != "0.0.0.0/0" {
		t.Errorf("dest: got %q", r.DestRange)
	}

	_, err = mock.GetRoute(context.Background(), "proj", "missing")
	if err == nil {
		t.Fatal("expected error for missing route")
	}
}

func TestComputeCommandIncludesNetworkingGroups(t *testing.T) {
	t.Parallel()

	root := NewCommand(&config.Config{}, &auth.Credentials{})
	got := subcommandNames(root)

	want := []string{
		"instances",
		"disks",
		"snapshots",
		"firewall-rules",
		"networks",
		"subnets",
		"addresses",
		"routers",
		"routes",
		"ssh",
		"scp",
	}

	for _, name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("missing top-level compute subcommand %q", name)
		}
	}
}

func TestNetworksCommandWiresSubcommands(t *testing.T) {
	t.Parallel()

	cmd := newNetworksCommand(&config.Config{}, &auth.Credentials{})
	got := subcommandNames(cmd)

	want := []string{"list", "describe", "create", "delete"}
	for _, name := range want {
		if _, ok := got[name]; !ok {
			t.Errorf("missing networks subcommand %q", name)
		}
	}
}

func TestDisksAndSnapshotsCommandsWireSubcommands(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cmd  *cobra.Command
		want []string
	}{
		{
			name: "disks",
			cmd:  newDisksCommand(&config.Config{}, &auth.Credentials{}),
			want: []string{"list", "describe", "create", "delete"},
		},
		{
			name: "snapshots",
			cmd:  newSnapshotsCommand(&config.Config{}, &auth.Credentials{}),
			want: []string{"list", "describe", "create", "delete"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := subcommandNames(tt.cmd)
			for _, name := range tt.want {
				if _, ok := got[name]; !ok {
					t.Errorf("missing %s subcommand %q", tt.name, name)
				}
			}
		})
	}
}

func subcommandNames(cmd *cobra.Command) map[string]struct{} {
	out := make(map[string]struct{}, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		out[sub.Name()] = struct{}{}
	}
	return out
}
