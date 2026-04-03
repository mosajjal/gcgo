package alloydb

import (
	"context"
	"fmt"

	alloydbapi "google.golang.org/api/alloydb/v1"
	"google.golang.org/api/option"
)

// Cluster holds the fields we display.
type Cluster struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	DatabaseVersion string `json:"database_version"`
	Network         string `json:"network"`
	CreateTime      string `json:"create_time"`
}

// Instance holds the fields we display.
type Instance struct {
	Name             string `json:"name"`
	DisplayName      string `json:"display_name"`
	InstanceType     string `json:"instance_type"`
	AvailabilityType string `json:"availability_type"`
	CpuCount         int64  `json:"cpu_count"`
	IPAddress        string `json:"ip_address"`
	State            string `json:"state"`
	CreateTime       string `json:"create_time"`
}

// CreateClusterRequest holds cluster creation parameters.
type CreateClusterRequest struct {
	Name             string
	DisplayName      string
	DatabaseVersion  string
	Network          string
	AllocatedIPRange string
	Username         string
	Password         string
}

// CreateInstanceRequest holds instance creation parameters.
type CreateInstanceRequest struct {
	Name             string
	DisplayName      string
	InstanceType     string
	AvailabilityType string
	CPUCount         int64
	NodeCount        int64
	Zone             string
}

// Client defines AlloyDB operations.
type Client interface {
	ListClusters(ctx context.Context, project, location string) ([]*Cluster, error)
	GetCluster(ctx context.Context, project, location, name string) (*Cluster, error)
	CreateCluster(ctx context.Context, project, location string, req *CreateClusterRequest) (*Cluster, error)
	DeleteCluster(ctx context.Context, project, location, name string) error

	ListInstances(ctx context.Context, project, location, cluster string) ([]*Instance, error)
	GetInstance(ctx context.Context, project, location, cluster, name string) (*Instance, error)
	CreateInstance(ctx context.Context, project, location, cluster string, req *CreateInstanceRequest) (*Instance, error)
	DeleteInstance(ctx context.Context, project, location, cluster, name string) error
}

type gcpClient struct {
	svc *alloydbapi.Service
}

// NewClient creates a Client backed by the real AlloyDB API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := alloydbapi.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create alloydb client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListClusters(ctx context.Context, project, location string) ([]*Cluster, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.svc.Projects.Locations.Clusters.List(parent).Context(ctx)

	var clusters []*Cluster
	if err := call.Pages(ctx, func(resp *alloydbapi.ListClustersResponse) error {
		for _, cluster := range resp.Clusters {
			clusters = append(clusters, clusterFromAPI(cluster))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	return clusters, nil
}

func (c *gcpClient) GetCluster(ctx context.Context, project, location, name string) (*Cluster, error) {
	fullName := clusterName(project, location, name)
	cluster, err := c.svc.Projects.Locations.Clusters.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get cluster %s: %w", name, err)
	}
	return clusterFromAPI(cluster), nil
}

func (c *gcpClient) CreateCluster(ctx context.Context, project, location string, req *CreateClusterRequest) (*Cluster, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	cluster := &alloydbapi.Cluster{
		DisplayName:     req.DisplayName,
		DatabaseVersion: req.DatabaseVersion,
		InitialUser:     &alloydbapi.UserPassword{User: req.Username, Password: req.Password},
		NetworkConfig: &alloydbapi.NetworkConfig{
			Network:          req.Network,
			AllocatedIpRange: req.AllocatedIPRange,
		},
	}
	if _, err := c.svc.Projects.Locations.Clusters.Create(parent, cluster).ClusterId(req.Name).Context(ctx).Do(); err != nil {
		return nil, fmt.Errorf("create cluster %s: %w", req.Name, err)
	}
	return &Cluster{
		Name:            clusterName(project, location, req.Name),
		DisplayName:     req.DisplayName,
		DatabaseVersion: req.DatabaseVersion,
		Network:         req.Network,
	}, nil
}

func (c *gcpClient) DeleteCluster(ctx context.Context, project, location, name string) error {
	fullName := clusterName(project, location, name)
	if _, err := c.svc.Projects.Locations.Clusters.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete cluster %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListInstances(ctx context.Context, project, location, cluster string) ([]*Instance, error) {
	parent := clusterName(project, location, cluster)
	call := c.svc.Projects.Locations.Clusters.Instances.List(parent).Context(ctx)

	var instances []*Instance
	if err := call.Pages(ctx, func(resp *alloydbapi.ListInstancesResponse) error {
		for _, instance := range resp.Instances {
			instances = append(instances, instanceFromAPI(instance))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	return instances, nil
}

func (c *gcpClient) GetInstance(ctx context.Context, project, location, cluster, name string) (*Instance, error) {
	fullName := instanceName(project, location, cluster, name)
	instance, err := c.svc.Projects.Locations.Clusters.Instances.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get instance %s: %w", name, err)
	}
	return instanceFromAPI(instance), nil
}

func (c *gcpClient) CreateInstance(ctx context.Context, project, location, cluster string, req *CreateInstanceRequest) (*Instance, error) {
	parent := clusterName(project, location, cluster)
	instance := &alloydbapi.Instance{
		DisplayName:      req.DisplayName,
		InstanceType:     req.InstanceType,
		AvailabilityType: req.AvailabilityType,
		GceZone:          req.Zone,
		MachineConfig:    &alloydbapi.MachineConfig{CpuCount: req.CPUCount},
	}
	if req.InstanceType == "READ_POOL" && req.NodeCount > 0 {
		instance.ReadPoolConfig = &alloydbapi.ReadPoolConfig{NodeCount: req.NodeCount}
	}
	if _, err := c.svc.Projects.Locations.Clusters.Instances.Create(parent, instance).InstanceId(req.Name).Context(ctx).Do(); err != nil {
		return nil, fmt.Errorf("create instance %s: %w", req.Name, err)
	}
	return &Instance{
		Name:             instanceName(project, location, cluster, req.Name),
		DisplayName:      req.DisplayName,
		InstanceType:     req.InstanceType,
		AvailabilityType: req.AvailabilityType,
		CpuCount:         req.CPUCount,
	}, nil
}

func (c *gcpClient) DeleteInstance(ctx context.Context, project, location, cluster, name string) error {
	fullName := instanceName(project, location, cluster, name)
	if _, err := c.svc.Projects.Locations.Clusters.Instances.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete instance %s: %w", name, err)
	}
	return nil
}

func clusterFromAPI(cluster *alloydbapi.Cluster) *Cluster {
	if cluster == nil {
		return nil
	}
	item := &Cluster{
		Name:            cluster.Name,
		DisplayName:     cluster.DisplayName,
		DatabaseVersion: cluster.DatabaseVersion,
		CreateTime:      cluster.CreateTime,
	}
	if cluster.NetworkConfig != nil {
		item.Network = cluster.NetworkConfig.Network
	}
	return item
}

func instanceFromAPI(instance *alloydbapi.Instance) *Instance {
	if instance == nil {
		return nil
	}
	item := &Instance{
		Name:             instance.Name,
		DisplayName:      instance.DisplayName,
		InstanceType:     instance.InstanceType,
		AvailabilityType: instance.AvailabilityType,
		IPAddress:        instance.IpAddress,
		State:            instance.State,
		CreateTime:       instance.CreateTime,
	}
	if instance.MachineConfig != nil {
		item.CpuCount = instance.MachineConfig.CpuCount
	}
	return item
}

func clusterName(project, location, name string) string {
	return fmt.Sprintf("projects/%s/locations/%s/clusters/%s", project, location, name)
}

func instanceName(project, location, cluster, name string) string {
	return fmt.Sprintf("%s/instances/%s", clusterName(project, location, cluster), name)
}
