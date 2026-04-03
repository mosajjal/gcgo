package dataproc

import (
	"context"
	"fmt"
	"time"

	dataproc "google.golang.org/api/dataproc/v1"
	"google.golang.org/api/option"
)

// Cluster holds Dataproc cluster fields.
type Cluster struct {
	Name   string            `json:"name"`
	Region string            `json:"region"`
	Status string            `json:"status"`
	Config string            `json:"config"`
	Labels map[string]string `json:"labels,omitempty"`
}

// Job holds Dataproc job fields.
type Job struct {
	ID     string `json:"job_id"`
	Status string `json:"status"`
	Type   string `json:"type"`
}

// Batch holds Dataproc batch fields.
type Batch struct {
	Name   string `json:"name"`
	State  string `json:"state"`
	Create string `json:"create_time"`
}

// Client defines Dataproc operations.
type Client interface {
	ListClusters(ctx context.Context, project, region string) ([]*Cluster, error)
	GetCluster(ctx context.Context, project, region, name string) (*Cluster, error)
	CreateCluster(ctx context.Context, project, region string, req *CreateClusterRequest) error
	DeleteCluster(ctx context.Context, project, region, name string) error
	StartCluster(ctx context.Context, project, region, name string) error
	StopCluster(ctx context.Context, project, region, name string) error

	ListJobs(ctx context.Context, project, region string) ([]*Job, error)
	GetJob(ctx context.Context, project, region, jobID string) (*Job, error)
	SubmitJob(ctx context.Context, project, region string, req *SubmitJobRequest) (*Job, error)
	CancelJob(ctx context.Context, project, region, jobID string) error

	ListBatches(ctx context.Context, project, region string) ([]*Batch, error)
	GetBatch(ctx context.Context, project, region, name string) (*Batch, error)
	CreateBatch(ctx context.Context, project, region string, req *CreateBatchRequest) error
	DeleteBatch(ctx context.Context, project, region, name string) error
}

// CreateClusterRequest holds parameters for cluster creation.
type CreateClusterRequest struct {
	Name         string
	MachineType  string
	NumWorkers   int64
	ImageVersion string
}

// SubmitJobRequest holds parameters for job submission.
type SubmitJobRequest struct {
	ClusterName string
	MainClass   string
	JarFileURIs []string
	Args        []string
}

// CreateBatchRequest holds parameters for batch creation.
type CreateBatchRequest struct {
	BatchID     string
	MainClass   string
	JarFileURIs []string
	Args        []string
}

type gcpClient struct {
	service *dataproc.Service
}

// NewClient creates a Client backed by the real Dataproc API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := dataproc.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create dataproc client: %w", err)
	}
	return &gcpClient{service: svc}, nil
}

func (c *gcpClient) ListClusters(ctx context.Context, project, region string) ([]*Cluster, error) {
	var clusters []*Cluster
	err := c.service.Projects.Regions.Clusters.List(project, region).
		Context(ctx).
		Pages(ctx, func(resp *dataproc.ListClustersResponse) error {
			for _, cluster := range resp.Clusters {
				clusters = append(clusters, clusterFromProto(cluster, region))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list clusters: %w", err)
	}
	return clusters, nil
}

func (c *gcpClient) GetCluster(ctx context.Context, project, region, name string) (*Cluster, error) {
	cluster, err := c.service.Projects.Regions.Clusters.Get(project, region, name).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get cluster %s: %w", name, err)
	}
	return clusterFromProto(cluster, region), nil
}

func (c *gcpClient) CreateCluster(ctx context.Context, project, region string, req *CreateClusterRequest) error {
	machineType := req.MachineType
	if machineType == "" {
		machineType = "n1-standard-4"
	}
	numWorkers := req.NumWorkers
	if numWorkers == 0 {
		numWorkers = 2
	}

	cluster := &dataproc.Cluster{
		ProjectId:   project,
		ClusterName: req.Name,
		Config: &dataproc.ClusterConfig{
			MasterConfig: &dataproc.InstanceGroupConfig{
				MachineTypeUri: machineType,
				NumInstances:   1,
			},
			WorkerConfig: &dataproc.InstanceGroupConfig{
				MachineTypeUri: machineType,
				NumInstances:   numWorkers,
			},
		},
	}
	if req.ImageVersion != "" {
		cluster.Config.SoftwareConfig = &dataproc.SoftwareConfig{
			ImageVersion: req.ImageVersion,
		}
	}

	op, err := c.service.Projects.Regions.Clusters.Create(project, region, cluster).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("create cluster %s: %w", req.Name, err)
	}
	if _, err := c.waitForRegionOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for cluster create %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpClient) DeleteCluster(ctx context.Context, project, region, name string) error {
	op, err := c.service.Projects.Regions.Clusters.Delete(project, region, name).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("delete cluster %s: %w", name, err)
	}
	if _, err := c.waitForRegionOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for cluster delete %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) StartCluster(ctx context.Context, project, region, name string) error {
	op, err := c.service.Projects.Regions.Clusters.Start(project, region, name, &dataproc.StartClusterRequest{}).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("start cluster %s: %w", name, err)
	}
	if _, err := c.waitForRegionOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for cluster start %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) StopCluster(ctx context.Context, project, region, name string) error {
	op, err := c.service.Projects.Regions.Clusters.Stop(project, region, name, &dataproc.StopClusterRequest{}).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("stop cluster %s: %w", name, err)
	}
	if _, err := c.waitForRegionOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for cluster stop %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListJobs(ctx context.Context, project, region string) ([]*Job, error) {
	var jobs []*Job
	err := c.service.Projects.Regions.Jobs.List(project, region).
		Context(ctx).
		Pages(ctx, func(resp *dataproc.ListJobsResponse) error {
			for _, job := range resp.Jobs {
				jobs = append(jobs, jobFromProto(job))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list jobs: %w", err)
	}
	return jobs, nil
}

func (c *gcpClient) GetJob(ctx context.Context, project, region, jobID string) (*Job, error) {
	job, err := c.service.Projects.Regions.Jobs.Get(project, region, jobID).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get job %s: %w", jobID, err)
	}
	return jobFromProto(job), nil
}

func (c *gcpClient) SubmitJob(ctx context.Context, project, region string, req *SubmitJobRequest) (*Job, error) {
	job, err := c.service.Projects.Regions.Jobs.Submit(project, region, &dataproc.SubmitJobRequest{
		Job: &dataproc.Job{
			Placement: &dataproc.JobPlacement{
				ClusterName: req.ClusterName,
			},
			SparkJob: &dataproc.SparkJob{
				MainClass:   req.MainClass,
				JarFileUris: req.JarFileURIs,
				Args:        req.Args,
			},
		},
		RequestId: fmt.Sprintf("gcgo-%d", time.Now().UnixNano()),
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("submit job: %w", err)
	}
	return jobFromProto(job), nil
}

func (c *gcpClient) CancelJob(ctx context.Context, project, region, jobID string) error {
	if _, err := c.service.Projects.Regions.Jobs.Cancel(project, region, jobID, &dataproc.CancelJobRequest{}).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("cancel job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) ListBatches(ctx context.Context, project, region string) ([]*Batch, error) {
	var batches []*Batch
	err := c.service.Projects.Locations.Batches.List(locationParent(project, region)).
		Context(ctx).
		Pages(ctx, func(resp *dataproc.ListBatchesResponse) error {
			for _, batch := range resp.Batches {
				batches = append(batches, batchFromProto(batch))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list batches: %w", err)
	}
	return batches, nil
}

func (c *gcpClient) GetBatch(ctx context.Context, project, region, name string) (*Batch, error) {
	batch, err := c.service.Projects.Locations.Batches.Get(batchName(project, region, name)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get batch %s: %w", name, err)
	}
	return batchFromProto(batch), nil
}

func (c *gcpClient) CreateBatch(ctx context.Context, project, region string, req *CreateBatchRequest) error {
	op, err := c.service.Projects.Locations.Batches.Create(locationParent(project, region), &dataproc.Batch{
		SparkBatch: &dataproc.SparkBatch{
			MainClass:   req.MainClass,
			JarFileUris: req.JarFileURIs,
			Args:        req.Args,
		},
	}).BatchId(req.BatchID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create batch %s: %w", req.BatchID, err)
	}
	if _, err := c.waitForLocationOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for batch create %s: %w", req.BatchID, err)
	}
	return nil
}

func (c *gcpClient) DeleteBatch(ctx context.Context, project, region, name string) error {
	if _, err := c.service.Projects.Locations.Batches.Delete(batchName(project, region, name)).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("delete batch %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) waitForRegionOperation(ctx context.Context, name string) (*dataproc.Operation, error) {
	for {
		op, err := c.service.Projects.Regions.Operations.Get(name).
			Context(ctx).
			Do()
		if err != nil {
			return nil, err
		}
		if op.Done {
			if op.Error != nil {
				if op.Error.Message != "" {
					return nil, fmt.Errorf("operation %s failed: %s", name, op.Error.Message)
				}
				return nil, fmt.Errorf("operation %s failed", name)
			}
			return op, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func (c *gcpClient) waitForLocationOperation(ctx context.Context, name string) (*dataproc.Operation, error) {
	for {
		op, err := c.service.Projects.Locations.Operations.Get(name).
			Context(ctx).
			Do()
		if err != nil {
			return nil, err
		}
		if op.Done {
			if op.Error != nil {
				if op.Error.Message != "" {
					return nil, fmt.Errorf("operation %s failed: %s", name, op.Error.Message)
				}
				return nil, fmt.Errorf("operation %s failed", name)
			}
			return op, nil
		}
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(2 * time.Second):
		}
	}
}

func locationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func batchName(project, region, name string) string {
	return fmt.Sprintf("%s/batches/%s", locationParent(project, region), name)
}

func clusterFromProto(c *dataproc.Cluster, region string) *Cluster {
	status := ""
	if c.Status != nil {
		status = c.Status.State
	}
	configDesc := ""
	if cfg := c.Config; cfg != nil {
		if master := cfg.MasterConfig; master != nil {
			configDesc = master.MachineTypeUri
		}
	}
	return &Cluster{
		Name:   c.ClusterName,
		Region: region,
		Status: status,
		Config: configDesc,
		Labels: c.Labels,
	}
}

func jobFromProto(j *dataproc.Job) *Job {
	status := ""
	if j.Status != nil {
		status = j.Status.State
	}

	jobType := "UNKNOWN"
	switch {
	case j.SparkJob != nil:
		jobType = "SPARK"
	case j.PysparkJob != nil:
		jobType = "PYSPARK"
	case j.HiveJob != nil:
		jobType = "HIVE"
	case j.PigJob != nil:
		jobType = "PIG"
	case j.HadoopJob != nil:
		jobType = "HADOOP"
	case j.SparkRJob != nil:
		jobType = "SPARK_R"
	case j.SparkSqlJob != nil:
		jobType = "SPARK_SQL"
	case j.PrestoJob != nil:
		jobType = "PRESTO"
	case j.TrinoJob != nil:
		jobType = "TRINO"
	case j.FlinkJob != nil:
		jobType = "FLINK"
	}

	id := j.JobUuid
	if ref := j.Reference; ref != nil && ref.JobId != "" {
		id = ref.JobId
	}

	return &Job{
		ID:     id,
		Status: status,
		Type:   jobType,
	}
}

func batchFromProto(b *dataproc.Batch) *Batch {
	return &Batch{
		Name:   b.Name,
		State:  b.State,
		Create: b.CreateTime,
	}
}
