package container

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	clusters   []*Cluster
	clusterMap map[string]*Cluster
	authMap    map[string]*ClusterAuth
	listErr    error
	getErr     error
	authErr    error
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
