package container

import (
	"context"
	"fmt"
	"testing"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/spf13/cobra"
)

type mockClient struct {
	clusters   []*Cluster
	clusterMap map[string]*Cluster
	authMap    map[string]*ClusterAuth
	ops        []*Operation
	opMap      map[string]*Operation
	listErr    error
	getErr     error
	authErr    error
	createErr  error
	deleteErr  error
	updateErr  error
	upgradeErr error
	resizeErr  error
	opErr      error
}

func (m *mockClient) ListClusters(_ context.Context, _, _ string) ([]*Cluster, error) {
	return m.clusters, m.listErr
}

func (m *mockClient) GetCluster(_ context.Context, _, _, name string) (*Cluster, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	c, ok := m.clusterMap[name]
	if !ok {
		return nil, fmt.Errorf("cluster %q not found", name)
	}
	return c, nil
}

func (m *mockClient) CreateCluster(_ context.Context, _, _ string, _ *CreateClusterRequest) error {
	return m.createErr
}

func (m *mockClient) DeleteCluster(_ context.Context, _, _, _ string) error {
	return m.deleteErr
}

func (m *mockClient) UpdateCluster(_ context.Context, _, _, _ string, _ *UpdateClusterRequest) error {
	return m.updateErr
}

func (m *mockClient) UpgradeCluster(_ context.Context, _, _, _ string, _ *UpgradeClusterRequest) error {
	return m.upgradeErr
}

func (m *mockClient) ResizeCluster(_ context.Context, _, _, _, _ string, _ int32) error {
	return m.resizeErr
}

func (m *mockClient) ListOperations(_ context.Context, _, _ string) ([]*Operation, error) {
	return m.ops, m.opErr
}

func (m *mockClient) GetOperation(_ context.Context, _, _, name string) (*Operation, error) {
	if m.opErr != nil {
		return nil, m.opErr
	}
	op, ok := m.opMap[name]
	if !ok {
		return nil, fmt.Errorf("operation %q not found", name)
	}
	return op, nil
}

func (m *mockClient) GetClusterAuth(_ context.Context, _, _, name string) (*ClusterAuth, error) {
	if m.authErr != nil {
		return nil, m.authErr
	}
	a, ok := m.authMap[name]
	if !ok {
		return nil, fmt.Errorf("cluster %q not found", name)
	}
	return a, nil
}

func TestMockListClusters(t *testing.T) {
	mock := &mockClient{
		clusters: []*Cluster{
			{Name: "cluster-1", Location: "us-central1", Status: "RUNNING", NodeCount: 3},
		},
	}

	clusters, err := mock.ListClusters(context.Background(), "proj", "-")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(clusters) != 1 {
		t.Errorf("expected 1, got %d", len(clusters))
	}
	if clusters[0].NodeCount != 3 {
		t.Errorf("node count: got %d", clusters[0].NodeCount)
	}
}

func TestMockGetCluster(t *testing.T) {
	mock := &mockClient{
		clusterMap: map[string]*Cluster{
			"c1": {Name: "c1", Status: "RUNNING"},
		},
	}

	c, err := mock.GetCluster(context.Background(), "proj", "us-central1", "c1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if c.Name != "c1" {
		t.Errorf("name: got %q", c.Name)
	}

	_, err = mock.GetCluster(context.Background(), "proj", "us-central1", "nope")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetClusterAuth(t *testing.T) {
	mock := &mockClient{
		authMap: map[string]*ClusterAuth{
			"c1": {Endpoint: "35.1.2.3", CACert: []byte("fake-cert")},
		},
	}

	auth, err := mock.GetClusterAuth(context.Background(), "proj", "us-central1", "c1")
	if err != nil {
		t.Fatalf("auth: %v", err)
	}
	if auth.Endpoint != "35.1.2.3" {
		t.Errorf("endpoint: got %q", auth.Endpoint)
	}
}

func TestMockClusterLifecycleAndOperations(t *testing.T) {
	mock := &mockClient{
		ops: []*Operation{
			{Name: "op-1", Location: "us-central1", Status: "RUNNING", OperationType: "CREATE_CLUSTER"},
		},
		opMap: map[string]*Operation{
			"op-1": {Name: "op-1", Location: "us-central1", Status: "DONE"},
		},
	}

	if err := mock.CreateCluster(context.Background(), "proj", "us-central1", &CreateClusterRequest{Name: "c1"}); err != nil {
		t.Fatalf("create: %v", err)
	}
	if err := mock.DeleteCluster(context.Background(), "proj", "us-central1", "c1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	if err := mock.UpdateCluster(context.Background(), "proj", "us-central1", "c1", &UpdateClusterRequest{}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if err := mock.UpgradeCluster(context.Background(), "proj", "us-central1", "c1", &UpgradeClusterRequest{}); err != nil {
		t.Fatalf("upgrade: %v", err)
	}

	ops, err := mock.ListOperations(context.Background(), "proj", "us-central1")
	if err != nil {
		t.Fatalf("list ops: %v", err)
	}
	if len(ops) != 1 || ops[0].Name != "op-1" {
		t.Fatalf("ops = %#v", ops)
	}

	op, err := mock.GetOperation(context.Background(), "proj", "us-central1", "op-1")
	if err != nil {
		t.Fatalf("get op: %v", err)
	}
	if op.Status != "DONE" {
		t.Fatalf("status = %q", op.Status)
	}
}

func TestContainerCommandTreeIncludesClusterLifecycle(t *testing.T) {
	cmd := newClustersCommand(&config.Config{}, auth.New(""))
	got := subcommandNames(cmd)

	for _, name := range []string{"create", "delete", "update", "upgrade", "operations", "get-credentials"} {
		if _, ok := got[name]; !ok {
			t.Fatalf("missing clusters subcommand %q", name)
		}
	}
}

func TestClusterOperationsCommandIncludesListAndDescribe(t *testing.T) {
	cmd := newClustersOperationsCommand(&config.Config{}, auth.New(""))
	got := subcommandNames(cmd)

	for _, name := range []string{"list", "describe"} {
		if _, ok := got[name]; !ok {
			t.Fatalf("missing operations subcommand %q", name)
		}
	}
}

func subcommandNames(cmd *cobra.Command) map[string]struct{} {
	out := make(map[string]struct{}, len(cmd.Commands()))
	for _, sub := range cmd.Commands() {
		out[sub.Name()] = struct{}{}
	}
	return out
}
