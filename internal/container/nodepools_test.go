package container

import (
	"context"
	"fmt"
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

type mockNodePoolClient struct {
	pools      []*NodePool
	poolMap    map[string]*NodePool
	listErr    error
	getErr     error
	createErr  error
	deleteErr  error
	updateErr  error
	upgradeErr error
}

func (m *mockNodePoolClient) ListNodePools(_ context.Context, _, _, _ string) ([]*NodePool, error) {
	return m.pools, m.listErr
}

func (m *mockNodePoolClient) GetNodePool(_ context.Context, _, _, _, name string) (*NodePool, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	p, ok := m.poolMap[name]
	if !ok {
		return nil, fmt.Errorf("node pool %q not found", name)
	}
	return p, nil
}

func (m *mockNodePoolClient) CreateNodePool(_ context.Context, _, _, _ string, _ *CreateNodePoolRequest) error {
	return m.createErr
}

func (m *mockNodePoolClient) DeleteNodePool(_ context.Context, _, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockNodePoolClient) UpdateNodePool(_ context.Context, _, _, _, _ string, _ *UpdateNodePoolRequest) error {
	return m.updateErr
}

func (m *mockNodePoolClient) UpgradeNodePool(_ context.Context, _, _, _, _ string, _ *UpgradeNodePoolRequest) error {
	return m.upgradeErr
}

func TestMockListNodePools(t *testing.T) {
	mock := &mockNodePoolClient{
		pools: []*NodePool{
			{Name: "default-pool", MachineType: "e2-medium", NodeCount: 3, Status: "RUNNING"},
			{Name: "gpu-pool", MachineType: "n1-standard-4", NodeCount: 2, Status: "RUNNING"},
		},
	}

	pools, err := mock.ListNodePools(context.Background(), "proj", "us-central1", "cluster-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(pools) != 2 {
		t.Errorf("expected 2, got %d", len(pools))
	}
	if pools[0].Name != "default-pool" {
		t.Errorf("name: got %q", pools[0].Name)
	}
	if pools[1].MachineType != "n1-standard-4" {
		t.Errorf("machine type: got %q", pools[1].MachineType)
	}
}

func TestMockListNodePoolsError(t *testing.T) {
	mock := &mockNodePoolClient{listErr: fmt.Errorf("permission denied")}
	_, err := mock.ListNodePools(context.Background(), "proj", "us-central1", "c1")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetNodePool(t *testing.T) {
	mock := &mockNodePoolClient{
		poolMap: map[string]*NodePool{
			"pool-1": {Name: "pool-1", MachineType: "e2-small", NodeCount: 1, Status: "RUNNING"},
		},
	}

	p, err := mock.GetNodePool(context.Background(), "proj", "us-central1", "c1", "pool-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if p.Name != "pool-1" {
		t.Errorf("name: got %q", p.Name)
	}
	if p.NodeCount != 1 {
		t.Errorf("node count: got %d", p.NodeCount)
	}

	_, err = mock.GetNodePool(context.Background(), "proj", "us-central1", "c1", "missing")
	if err == nil {
		t.Fatal("expected error for missing pool")
	}
}

func TestMockCreateNodePool(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"success", nil, false},
		{"quota error", fmt.Errorf("quota exceeded"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNodePoolClient{createErr: tt.err}
			err := mock.CreateNodePool(context.Background(), "proj", "us-central1", "c1", &CreateNodePoolRequest{
				Name:        "new-pool",
				MachineType: "e2-medium",
				NumNodes:    3,
			})
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockDeleteNodePool(t *testing.T) {
	mock := &mockNodePoolClient{}
	err := mock.DeleteNodePool(context.Background(), "proj", "us-central1", "c1", "pool-1")
	if err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestMockDeleteNodePoolError(t *testing.T) {
	mock := &mockNodePoolClient{deleteErr: fmt.Errorf("not found")}
	err := mock.DeleteNodePool(context.Background(), "proj", "us-central1", "c1", "missing")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockUpdateNodePool(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"success", nil, false},
		{"error", fmt.Errorf("invalid node count"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNodePoolClient{updateErr: tt.err}
			err := mock.UpdateNodePool(context.Background(), "proj", "us-central1", "c1", "pool-1", &UpdateNodePoolRequest{
				NumNodes: 5,
			})
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestMockUpgradeNodePool(t *testing.T) {
	tests := []struct {
		name    string
		err     error
		wantErr bool
	}{
		{"success", nil, false},
		{"error", fmt.Errorf("upgrade failed"), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockNodePoolClient{upgradeErr: tt.err}
			err := mock.UpgradeNodePool(context.Background(), "proj", "us-central1", "c1", "pool-1", &UpgradeNodePoolRequest{
				NodeVersion: "1.30.1-gke.1",
				ImageType:   "cos_containerd",
			})
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestContainerCommandTreeIncludesNodePools(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))
	if cmd == nil {
		t.Fatal("expected command")
	}

	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "node-pools" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected node-pools command to be wired")
	}
}

func TestNodePoolsCommandIncludesUpgrade(t *testing.T) {
	cmd := newNodePoolsCommand(&config.Config{}, auth.New(""))
	found := false
	for _, sub := range cmd.Commands() {
		if sub.Name() == "upgrade" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("expected upgrade command to be wired")
	}
}
