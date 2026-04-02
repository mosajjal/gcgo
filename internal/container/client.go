package container

import (
	"context"
	"encoding/base64"
	"fmt"

	gke "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"google.golang.org/api/option"
)

// Cluster holds GKE cluster fields.
type Cluster struct {
	Name      string `json:"name"`
	Location  string `json:"location"`
	Status    string `json:"status"`
	NodeCount int32  `json:"node_count"`
	Endpoint  string `json:"endpoint"`
}

// ClusterAuth holds kubeconfig auth details.
type ClusterAuth struct {
	Endpoint string
	CACert   []byte
}

// Client defines GKE operations.
type Client interface {
	ListClusters(ctx context.Context, project, location string) ([]*Cluster, error)
	GetCluster(ctx context.Context, project, location, name string) (*Cluster, error)
	GetClusterAuth(ctx context.Context, project, location, name string) (*ClusterAuth, error)
}

type gcpClient struct {
	cm *gke.ClusterManagerClient
}

// NewClient creates a Client backed by the real GKE API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	cm, err := gke.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create gke client: %w", err)
	}
	return &gcpClient{cm: cm}, nil
}

func (c *gcpClient) ListClusters(ctx context.Context, project, location string) ([]*Cluster, error) {
	resp, err := c.cm.ListClusters(ctx, &containerpb.ListClustersRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, location),
	})
	if err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}

	var clusters []*Cluster
	for _, cl := range resp.GetClusters() {
		clusters = append(clusters, clusterFromProto(cl))
	}
	return clusters, nil
}

func (c *gcpClient) GetCluster(ctx context.Context, project, location, name string) (*Cluster, error) {
	cl, err := c.cm.GetCluster(ctx, &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name),
	})
	if err != nil {
		return nil, fmt.Errorf("get cluster %s: %w", name, err)
	}
	return clusterFromProto(cl), nil
}

func (c *gcpClient) GetClusterAuth(ctx context.Context, project, location, name string) (*ClusterAuth, error) {
	cl, err := c.cm.GetCluster(ctx, &containerpb.GetClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name),
	})
	if err != nil {
		return nil, fmt.Errorf("get cluster auth %s: %w", name, err)
	}

	caCert, err := base64.StdEncoding.DecodeString(cl.GetMasterAuth().GetClusterCaCertificate())
	if err != nil {
		return nil, fmt.Errorf("decode ca cert: %w", err)
	}

	return &ClusterAuth{
		Endpoint: cl.GetEndpoint(),
		CACert:   caCert,
	}, nil
}

func clusterFromProto(cl *containerpb.Cluster) *Cluster {
	return &Cluster{
		Name:      cl.GetName(),
		Location:  cl.GetLocation(),
		Status:    cl.GetStatus().String(),
		NodeCount: cl.GetCurrentNodeCount(), //nolint:staticcheck // no replacement available yet
		Endpoint:  cl.GetEndpoint(),
	}
}
