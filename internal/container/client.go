package container

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

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

// Operation holds GKE operation fields.
type Operation struct {
	Name          string `json:"name"`
	Location      string `json:"location"`
	Status        string `json:"status"`
	OperationType string `json:"operation_type"`
	Detail        string `json:"detail"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	TargetLink    string `json:"target_link"`
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
	CreateCluster(ctx context.Context, project, location string, req *CreateClusterRequest) error
	DeleteCluster(ctx context.Context, project, location, name string) error
	UpdateCluster(ctx context.Context, project, location, name string, req *UpdateClusterRequest) error
	UpgradeCluster(ctx context.Context, project, location, name string, req *UpgradeClusterRequest) error
	ResizeCluster(ctx context.Context, project, location, cluster, nodePool string, nodeCount int32) error
	ListOperations(ctx context.Context, project, location string) ([]*Operation, error)
	GetOperation(ctx context.Context, project, location, name string) (*Operation, error)
	GetClusterAuth(ctx context.Context, project, location, name string) (*ClusterAuth, error)
}

// CreateClusterRequest holds cluster creation parameters.
type CreateClusterRequest struct {
	Name        string
	NumNodes    int32
	MachineType string
}

// UpdateClusterRequest holds cluster update parameters.
type UpdateClusterRequest struct {
	MasterVersion string
	NodeVersion   string
}

// UpgradeClusterRequest holds cluster upgrade parameters.
type UpgradeClusterRequest struct {
	Version string
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

func (c *gcpClient) CreateCluster(ctx context.Context, project, location string, req *CreateClusterRequest) error {
	numNodes := req.NumNodes
	if numNodes <= 0 {
		numNodes = 3
	}
	machineType := req.MachineType
	if machineType == "" {
		machineType = "e2-medium"
	}

	op, err := c.cm.CreateCluster(ctx, &containerpb.CreateClusterRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, location),
		Cluster: &containerpb.Cluster{
			Name:             req.Name,
			InitialNodeCount: numNodes,
			NodeConfig: &containerpb.NodeConfig{
				MachineType: machineType,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create cluster %s: %w", req.Name, err)
	}
	return waitForOperation(ctx, c.cm, op)
}

func (c *gcpClient) DeleteCluster(ctx context.Context, project, location, name string) error {
	op, err := c.cm.DeleteCluster(ctx, &containerpb.DeleteClusterRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name),
	})
	if err != nil {
		return fmt.Errorf("delete cluster %s: %w", name, err)
	}
	return waitForOperation(ctx, c.cm, op)
}

func (c *gcpClient) UpdateCluster(ctx context.Context, project, location, name string, req *UpdateClusterRequest) error {
	update := &containerpb.ClusterUpdate{}
	if req.MasterVersion != "" {
		update.DesiredMasterVersion = req.MasterVersion
	}
	if req.NodeVersion != "" {
		update.DesiredNodeVersion = req.NodeVersion
	}
	op, err := c.cm.UpdateCluster(ctx, &containerpb.UpdateClusterRequest{
		Name:   fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name),
		Update: update,
	})
	if err != nil {
		return fmt.Errorf("update cluster %s: %w", name, err)
	}
	return waitForOperation(ctx, c.cm, op)
}

func (c *gcpClient) UpgradeCluster(ctx context.Context, project, location, name string, req *UpgradeClusterRequest) error {
	update := &UpdateClusterRequest{
		MasterVersion: req.Version,
		NodeVersion:   req.Version,
	}
	return c.UpdateCluster(ctx, project, location, name, update)
}

func (c *gcpClient) ResizeCluster(ctx context.Context, project, location, cluster, nodePool string, nodeCount int32) error {
	op, err := c.cm.SetNodePoolSize(ctx, &containerpb.SetNodePoolSizeRequest{
		Name:      fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", project, location, cluster, nodePool),
		NodeCount: nodeCount,
	})
	if err != nil {
		return fmt.Errorf("resize cluster %s: %w", cluster, err)
	}
	return waitForOperation(ctx, c.cm, op)
}

func (c *gcpClient) ListOperations(ctx context.Context, project, location string) ([]*Operation, error) {
	resp, err := c.cm.ListOperations(ctx, &containerpb.ListOperationsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", project, location),
	})
	if err != nil {
		return nil, fmt.Errorf("list operations: %w", err)
	}
	ops := make([]*Operation, 0, len(resp.GetOperations()))
	for _, op := range resp.GetOperations() {
		ops = append(ops, operationFromProto(op))
	}
	return ops, nil
}

func (c *gcpClient) GetOperation(ctx context.Context, project, location, name string) (*Operation, error) {
	op, err := c.cm.GetOperation(ctx, &containerpb.GetOperationRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/operations/%s", project, location, name),
	})
	if err != nil {
		return nil, fmt.Errorf("get operation %s: %w", name, err)
	}
	return operationFromProto(op), nil
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

func operationFromProto(op *containerpb.Operation) *Operation {
	return &Operation{
		Name:          op.GetName(),
		Location:      op.GetLocation(),
		Status:        op.GetStatus().String(),
		OperationType: op.GetOperationType().String(),
		Detail:        op.GetDetail(),
		StartTime:     op.GetStartTime(),
		EndTime:       op.GetEndTime(),
		TargetLink:    op.GetTargetLink(),
	}
}

func waitForOperation(ctx context.Context, cm *gke.ClusterManagerClient, op *containerpb.Operation) error {
	if op == nil {
		return fmt.Errorf("wait for operation: nil operation")
	}
	if op.GetName() == "" {
		return nil
	}

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		switch op.GetStatus() {
		case containerpb.Operation_DONE:
			return nil
		case containerpb.Operation_ABORTING:
			return fmt.Errorf("operation %s aborted", op.GetName())
		}

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-ticker.C:
		}

		refreshed, err := cm.GetOperation(ctx, &containerpb.GetOperationRequest{
			Name: op.GetName(),
		})
		if err != nil {
			return fmt.Errorf("wait for operation %s: %w", op.GetName(), err)
		}
		op = refreshed
	}
}
