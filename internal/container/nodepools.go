package container

import (
	"context"
	"fmt"

	gke "cloud.google.com/go/container/apiv1"
	"cloud.google.com/go/container/apiv1/containerpb"
	"google.golang.org/api/option"
)

// NodePool holds GKE node pool fields.
type NodePool struct {
	Name        string `json:"name"`
	MachineType string `json:"machine_type"`
	NodeCount   int32  `json:"node_count"`
	Status      string `json:"status"`
	Version     string `json:"version"`
}

// NodePoolClient defines GKE node pool operations.
type NodePoolClient interface {
	ListNodePools(ctx context.Context, project, location, cluster string) ([]*NodePool, error)
	GetNodePool(ctx context.Context, project, location, cluster, name string) (*NodePool, error)
	CreateNodePool(ctx context.Context, project, location, cluster string, req *CreateNodePoolRequest) error
	DeleteNodePool(ctx context.Context, project, location, cluster, name string) error
	UpdateNodePool(ctx context.Context, project, location, cluster, name string, req *UpdateNodePoolRequest) error
	UpgradeNodePool(ctx context.Context, project, location, cluster, name string, req *UpgradeNodePoolRequest) error
}

// CreateNodePoolRequest holds parameters for creating a node pool.
type CreateNodePoolRequest struct {
	Name        string
	MachineType string
	NumNodes    int32
}

// UpdateNodePoolRequest holds parameters for updating a node pool.
type UpdateNodePoolRequest struct {
	NumNodes int32
}

// UpgradeNodePoolRequest holds parameters for upgrading a node pool.
type UpgradeNodePoolRequest struct {
	NodeVersion string
	ImageType   string
}

type gcpNodePoolClient struct {
	cm *gke.ClusterManagerClient
}

// NewNodePoolClient creates a NodePoolClient backed by the real GKE API.
func NewNodePoolClient(ctx context.Context, opts ...option.ClientOption) (NodePoolClient, error) {
	cm, err := gke.NewClusterManagerClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create gke node pool client: %w", err)
	}
	return &gcpNodePoolClient{cm: cm}, nil
}

func (c *gcpNodePoolClient) ListNodePools(ctx context.Context, project, location, cluster string) ([]*NodePool, error) {
	resp, err := c.cm.ListNodePools(ctx, &containerpb.ListNodePoolsRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, cluster),
	})
	if err != nil {
		return nil, fmt.Errorf("list node pools: %w", err)
	}

	var pools []*NodePool
	for _, np := range resp.GetNodePools() {
		pools = append(pools, nodePoolFromProto(np))
	}
	return pools, nil
}

func (c *gcpNodePoolClient) GetNodePool(ctx context.Context, project, location, cluster, name string) (*NodePool, error) {
	np, err := c.cm.GetNodePool(ctx, &containerpb.GetNodePoolRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", project, location, cluster, name),
	})
	if err != nil {
		return nil, fmt.Errorf("get node pool %s: %w", name, err)
	}
	return nodePoolFromProto(np), nil
}

func (c *gcpNodePoolClient) CreateNodePool(ctx context.Context, project, location, cluster string, req *CreateNodePoolRequest) error {
	machineType := req.MachineType
	if machineType == "" {
		machineType = "e2-medium"
	}
	numNodes := req.NumNodes
	if numNodes <= 0 {
		numNodes = 3
	}

	op, err := c.cm.CreateNodePool(ctx, &containerpb.CreateNodePoolRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, cluster),
		NodePool: &containerpb.NodePool{
			Name:             req.Name,
			InitialNodeCount: numNodes,
			Config: &containerpb.NodeConfig{
				MachineType: machineType,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create node pool %s: %w", req.Name, err)
	}
	_ = op
	return nil
}

func (c *gcpNodePoolClient) DeleteNodePool(ctx context.Context, project, location, cluster, name string) error {
	_, err := c.cm.DeleteNodePool(ctx, &containerpb.DeleteNodePoolRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", project, location, cluster, name),
	})
	if err != nil {
		return fmt.Errorf("delete node pool %s: %w", name, err)
	}
	return nil
}

func (c *gcpNodePoolClient) UpdateNodePool(ctx context.Context, project, location, cluster, name string, req *UpdateNodePoolRequest) error {
	op, err := c.cm.SetNodePoolSize(ctx, &containerpb.SetNodePoolSizeRequest{
		Name:      fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", project, location, cluster, name),
		NodeCount: req.NumNodes,
	})
	if err != nil {
		return fmt.Errorf("update node pool %s: %w", name, err)
	}
	return waitForOperation(ctx, c.cm, op)
}

func (c *gcpNodePoolClient) UpgradeNodePool(ctx context.Context, project, location, cluster, name string, req *UpgradeNodePoolRequest) error {
	update := &containerpb.UpdateNodePoolRequest{
		Name: fmt.Sprintf("projects/%s/locations/%s/clusters/%s/nodePools/%s", project, location, cluster, name),
	}
	if req.NodeVersion != "" {
		update.NodeVersion = req.NodeVersion
	}
	if req.ImageType != "" {
		update.ImageType = req.ImageType
	}
	op, err := c.cm.UpdateNodePool(ctx, update)
	if err != nil {
		return fmt.Errorf("upgrade node pool %s: %w", name, err)
	}
	return waitForOperation(ctx, c.cm, op)
}

func nodePoolFromProto(np *containerpb.NodePool) *NodePool {
	machineType := ""
	if np.GetConfig() != nil {
		machineType = np.GetConfig().GetMachineType()
	}
	return &NodePool{
		Name:        np.GetName(),
		MachineType: machineType,
		NodeCount:   np.GetInitialNodeCount(),
		Status:      np.GetStatus().String(),
		Version:     np.GetVersion(),
	}
}
