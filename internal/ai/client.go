package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	aiplatform "google.golang.org/api/aiplatform/v1"
	"google.golang.org/api/googleapi"
	"google.golang.org/api/option"
)

// Model holds Vertex AI model fields.
type Model struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
}

// Endpoint holds Vertex AI endpoint fields.
type Endpoint struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Region      string `json:"region"`
}

// DeployedModel holds Vertex AI deployed model fields.
type DeployedModel struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	Model       string `json:"model"`
	CreateTime  string `json:"create_time"`
}

// CustomJob holds Vertex AI custom job fields.
type CustomJob struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	CreateTime  string `json:"create_time"`
}

// Dataset holds Vertex AI dataset fields.
type Dataset struct {
	Name              string `json:"name"`
	DisplayName       string `json:"display_name"`
	Description       string `json:"description"`
	MetadataSchemaURI string `json:"metadata_schema_uri"`
	CreateTime        string `json:"create_time"`
	UpdateTime        string `json:"update_time"`
	DataItemCount     int64  `json:"data_item_count"`
}

// PipelineJob holds Vertex AI pipeline job fields.
type PipelineJob struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	CreateTime  string `json:"create_time"`
	EndTime     string `json:"end_time"`
	TemplateURI string `json:"template_uri"`
}

// BatchPredictionJob holds Vertex AI batch prediction job fields.
type BatchPredictionJob struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	CreateTime  string `json:"create_time"`
	EndTime     string `json:"end_time"`
	Model       string `json:"model"`
}

// Operation holds Vertex AI long-running operation fields.
type Operation struct {
	Name  string `json:"name"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// PredictResponse holds the prediction result.
type PredictResponse struct {
	Predictions []any `json:"predictions"`
}

// Client defines Vertex AI operations.
type Client interface {
	ListModels(ctx context.Context, project, region string) ([]*Model, error)
	GetModel(ctx context.Context, project, region, modelID string) (*Model, error)
	DeleteModel(ctx context.Context, project, region, modelID string) error
	UploadModel(ctx context.Context, project, region string, req *UploadModelRequest) (*Model, error)

	ListEndpoints(ctx context.Context, project, region string) ([]*Endpoint, error)
	GetEndpoint(ctx context.Context, project, region, endpointID string) (*Endpoint, error)
	CreateEndpoint(ctx context.Context, project, region string, req *CreateEndpointRequest) (*Endpoint, error)
	DeleteEndpoint(ctx context.Context, project, region, endpointID string) error
	DeployModel(ctx context.Context, project, region, endpointID string, req *DeployModelRequest) (*DeployedModel, error)
	UndeployModel(ctx context.Context, project, region, endpointID, deployedModelID string) error
	Predict(ctx context.Context, project, region, endpointID string, instances []any) (*PredictResponse, error)

	ListCustomJobs(ctx context.Context, project, region string) ([]*CustomJob, error)
	GetCustomJob(ctx context.Context, project, region, jobID string) (*CustomJob, error)
	CreateCustomJob(ctx context.Context, project, region string, req *CreateCustomJobRequest) (*CustomJob, error)
	CancelCustomJob(ctx context.Context, project, region, jobID string) error
	ListDatasets(ctx context.Context, project, region string) ([]*Dataset, error)
	GetDataset(ctx context.Context, project, region, datasetID string) (*Dataset, error)
	CreateDataset(ctx context.Context, project, region string, req *CreateDatasetRequest) (string, error)
	UpdateDataset(ctx context.Context, project, region, datasetID string, req *UpdateDatasetRequest) (string, error)
	DeleteDataset(ctx context.Context, project, region, datasetID string) (string, error)
	ListPipelineJobs(ctx context.Context, project, region string) ([]*PipelineJob, error)
	GetPipelineJob(ctx context.Context, project, region, jobID string) (*PipelineJob, error)
	CreatePipelineJob(ctx context.Context, project, region string, req *CreatePipelineJobRequest) (string, error)
	DeletePipelineJob(ctx context.Context, project, region, jobID string) error
	CancelPipelineJob(ctx context.Context, project, region, jobID string) error
	ListBatchPredictionJobs(ctx context.Context, project, region string) ([]*BatchPredictionJob, error)
	GetBatchPredictionJob(ctx context.Context, project, region, jobID string) (*BatchPredictionJob, error)
	CreateBatchPredictionJob(ctx context.Context, project, region string, req *CreateBatchPredictionJobRequest) (string, error)
	DeleteBatchPredictionJob(ctx context.Context, project, region, jobID string) error
	CancelBatchPredictionJob(ctx context.Context, project, region, jobID string) error
	ListOperations(ctx context.Context, project, region, filter string) ([]*Operation, error)
	GetOperation(ctx context.Context, name string) (*Operation, error)
}

// UploadModelRequest holds parameters for model upload.
type UploadModelRequest struct {
	ModelID      string
	ParentModel  string
	DisplayName  string
	ArtifactURI  string
	ContainerURI string
}

// CreateEndpointRequest holds parameters for endpoint creation.
type CreateEndpointRequest struct {
	DisplayName string
}

// DeployModelRequest holds parameters for endpoint model deployment.
type DeployModelRequest struct {
	Model          string
	DisplayName    string
	MachineType    string
	MinReplicas    int64
	MaxReplicas    int64
	TrafficPercent int64
}

// CreateCustomJobRequest holds parameters for custom job creation.
type CreateCustomJobRequest struct {
	DisplayName  string
	ContainerURI string
	Args         []string
	MachineType  string
}

// CreateDatasetRequest holds parameters for dataset creation.
type CreateDatasetRequest struct {
	DisplayName       string
	Description       string
	MetadataSchemaURI string
	MetadataJSON      string
	Labels            map[string]string
}

// UpdateDatasetRequest holds parameters for dataset updates.
type UpdateDatasetRequest struct {
	DisplayName       string
	Description       string
	MetadataSchemaURI string
	MetadataJSON      string
	Labels            map[string]string
}

// CreatePipelineJobRequest holds parameters for pipeline job creation.
type CreatePipelineJobRequest struct {
	DisplayName         string
	TemplateURI         string
	PipelineSpecJSON    string
	GCSOutputDirectory  string
	ServiceAccount      string
	Network             string
	ParameterValuesJSON string
}

// CreateBatchPredictionJobRequest holds parameters for batch prediction job creation.
type CreateBatchPredictionJobRequest struct {
	DisplayName         string
	Model               string
	InstancesFormat     string
	GCSSource           string
	BigQuerySource      string
	PredictionsFormat   string
	GCSDestination      string
	BigQueryDestination string
	ServiceAccount      string
	MachineType         string
	StartingReplicas    int64
	MaxReplicas         int64
}

type gcpClient struct {
	service *aiplatform.Service
}

// NewClient creates a Client backed by the real Vertex AI API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := aiplatform.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create vertex ai client: %w", err)
	}
	return &gcpClient{service: svc}, nil
}

func locationParent(project, region string) string {
	return fmt.Sprintf("projects/%s/locations/%s", project, region)
}

func modelName(project, region, modelID string) string {
	return fmt.Sprintf("%s/models/%s", locationParent(project, region), modelID)
}

func endpointName(project, region, endpointID string) string {
	return fmt.Sprintf("%s/endpoints/%s", locationParent(project, region), endpointID)
}

func customJobName(project, region, jobID string) string {
	return fmt.Sprintf("%s/customJobs/%s", locationParent(project, region), jobID)
}

func datasetName(project, region, datasetID string) string {
	return fmt.Sprintf("%s/datasets/%s", locationParent(project, region), datasetID)
}

func pipelineJobName(project, region, jobID string) string {
	return fmt.Sprintf("%s/pipelineJobs/%s", locationParent(project, region), jobID)
}

func batchPredictionJobName(project, region, jobID string) string {
	return fmt.Sprintf("%s/batchPredictionJobs/%s", locationParent(project, region), jobID)
}

func (c *gcpClient) ListModels(ctx context.Context, project, region string) ([]*Model, error) {
	var models []*Model
	err := c.service.Projects.Locations.Models.List(locationParent(project, region)).
		Context(ctx).
		Pages(ctx, func(resp *aiplatform.GoogleCloudAiplatformV1ListModelsResponse) error {
			for _, model := range resp.Models {
				models = append(models, modelFromProto(model))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list models: %w", err)
	}
	return models, nil
}

func (c *gcpClient) GetModel(ctx context.Context, project, region, modelID string) (*Model, error) {
	model, err := c.service.Projects.Locations.Models.Get(modelName(project, region, modelID)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get model %s: %w", modelID, err)
	}
	return modelFromProto(model), nil
}

func (c *gcpClient) DeleteModel(ctx context.Context, project, region, modelID string) error {
	op, err := c.service.Projects.Locations.Models.Delete(modelName(project, region, modelID)).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("delete model %s: %w", modelID, err)
	}
	if _, err := c.waitForOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for model delete %s: %w", modelID, err)
	}
	return nil
}

func (c *gcpClient) UploadModel(ctx context.Context, project, region string, req *UploadModelRequest) (*Model, error) {
	model := &aiplatform.GoogleCloudAiplatformV1Model{
		DisplayName: req.DisplayName,
		ArtifactUri: req.ArtifactURI,
		ContainerSpec: &aiplatform.GoogleCloudAiplatformV1ModelContainerSpec{
			ImageUri: req.ContainerURI,
		},
	}

	op, err := c.service.Projects.Locations.Models.Upload(locationParent(project, region), &aiplatform.GoogleCloudAiplatformV1UploadModelRequest{
		Model:       model,
		ModelId:     req.ModelID,
		ParentModel: req.ParentModel,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("upload model %s: %w", req.DisplayName, err)
	}

	finished, err := c.waitForOperation(ctx, op.Name)
	if err != nil {
		return nil, fmt.Errorf("wait for model upload %s: %w", req.DisplayName, err)
	}

	var uploadResp aiplatform.GoogleCloudAiplatformV1UploadModelResponse
	if err := json.Unmarshal(finished.Response, &uploadResp); err != nil {
		return nil, fmt.Errorf("decode model upload response: %w", err)
	}
	if uploadResp.Model == "" {
		return nil, fmt.Errorf("upload model %s returned an empty model name", req.DisplayName)
	}

	uploaded, err := c.service.Projects.Locations.Models.Get(uploadResp.Model).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get uploaded model: %w", err)
	}
	return modelFromProto(uploaded), nil
}

func (c *gcpClient) ListEndpoints(ctx context.Context, project, region string) ([]*Endpoint, error) {
	var endpoints []*Endpoint
	err := c.service.Projects.Locations.Endpoints.List(locationParent(project, region)).
		Context(ctx).
		Pages(ctx, func(resp *aiplatform.GoogleCloudAiplatformV1ListEndpointsResponse) error {
			for _, endpoint := range resp.Endpoints {
				endpoints = append(endpoints, endpointFromProto(endpoint, region))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list endpoints: %w", err)
	}
	return endpoints, nil
}

func (c *gcpClient) GetEndpoint(ctx context.Context, project, region, endpointID string) (*Endpoint, error) {
	endpoint, err := c.service.Projects.Locations.Endpoints.Get(endpointName(project, region, endpointID)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get endpoint %s: %w", endpointID, err)
	}
	return endpointFromProto(endpoint, region), nil
}

func (c *gcpClient) CreateEndpoint(ctx context.Context, project, region string, req *CreateEndpointRequest) (*Endpoint, error) {
	op, err := c.service.Projects.Locations.Endpoints.Create(locationParent(project, region), &aiplatform.GoogleCloudAiplatformV1Endpoint{
		DisplayName: req.DisplayName,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create endpoint %s: %w", req.DisplayName, err)
	}

	finished, err := c.waitForOperation(ctx, op.Name)
	if err != nil {
		return nil, fmt.Errorf("wait for endpoint create %s: %w", req.DisplayName, err)
	}

	var endpoint aiplatform.GoogleCloudAiplatformV1Endpoint
	if err := json.Unmarshal(finished.Response, &endpoint); err != nil {
		return nil, fmt.Errorf("decode endpoint create response: %w", err)
	}
	return endpointFromProto(&endpoint, region), nil
}

func (c *gcpClient) DeleteEndpoint(ctx context.Context, project, region, endpointID string) error {
	op, err := c.service.Projects.Locations.Endpoints.Delete(endpointName(project, region, endpointID)).
		Context(ctx).
		Do()
	if err != nil {
		return fmt.Errorf("delete endpoint %s: %w", endpointID, err)
	}
	if _, err := c.waitForOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for endpoint delete %s: %w", endpointID, err)
	}
	return nil
}

func (c *gcpClient) DeployModel(ctx context.Context, project, region, endpointID string, req *DeployModelRequest) (*DeployedModel, error) {
	machineType := req.MachineType
	if machineType == "" {
		machineType = "n1-standard-2"
	}
	minReplicas := req.MinReplicas
	if minReplicas == 0 {
		minReplicas = 1
	}
	maxReplicas := req.MaxReplicas
	if maxReplicas == 0 {
		maxReplicas = minReplicas
	}
	trafficPercent := req.TrafficPercent
	if trafficPercent == 0 {
		trafficPercent = 100
	}

	op, err := c.service.Projects.Locations.Endpoints.DeployModel(endpointName(project, region, endpointID), &aiplatform.GoogleCloudAiplatformV1DeployModelRequest{
		DeployedModel: &aiplatform.GoogleCloudAiplatformV1DeployedModel{
			Model:       resolveModelResource(project, region, req.Model),
			DisplayName: req.DisplayName,
			DedicatedResources: &aiplatform.GoogleCloudAiplatformV1DedicatedResources{
				MachineSpec: &aiplatform.GoogleCloudAiplatformV1MachineSpec{
					MachineType: machineType,
				},
				MinReplicaCount: minReplicas,
				MaxReplicaCount: maxReplicas,
			},
		},
		TrafficSplit: map[string]int64{"0": trafficPercent},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("deploy model %s to endpoint %s: %w", req.Model, endpointID, err)
	}

	finished, err := c.waitForOperation(ctx, op.Name)
	if err != nil {
		return nil, fmt.Errorf("wait for model deploy to endpoint %s: %w", endpointID, err)
	}

	var deployResp aiplatform.GoogleCloudAiplatformV1DeployModelResponse
	if err := json.Unmarshal(finished.Response, &deployResp); err != nil {
		return nil, fmt.Errorf("decode deploy model response: %w", err)
	}
	if deployResp.DeployedModel == nil {
		return nil, fmt.Errorf("deploy model to endpoint %s returned an empty deployed model", endpointID)
	}
	return deployedModelFromProto(deployResp.DeployedModel), nil
}

func (c *gcpClient) UndeployModel(ctx context.Context, project, region, endpointID, deployedModelID string) error {
	op, err := c.service.Projects.Locations.Endpoints.UndeployModel(endpointName(project, region, endpointID), &aiplatform.GoogleCloudAiplatformV1UndeployModelRequest{
		DeployedModelId: deployedModelID,
		TrafficSplit:    map[string]int64{},
		ForceSendFields: []string{"TrafficSplit"},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("undeploy model %s from endpoint %s: %w", deployedModelID, endpointID, err)
	}
	if _, err := c.waitForOperation(ctx, op.Name); err != nil {
		return fmt.Errorf("wait for model undeploy from endpoint %s: %w", endpointID, err)
	}
	return nil
}

func (c *gcpClient) Predict(ctx context.Context, project, region, endpointID string, instances []any) (*PredictResponse, error) {
	resp, err := c.service.Projects.Locations.Endpoints.Predict(endpointName(project, region, endpointID), &aiplatform.GoogleCloudAiplatformV1PredictRequest{
		Instances: instances,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("predict on endpoint %s: %w", endpointID, err)
	}

	predictions := make([]any, 0, len(resp.Predictions))
	for _, prediction := range resp.Predictions {
		predictions = append(predictions, prediction)
	}
	return &PredictResponse{Predictions: predictions}, nil
}

func (c *gcpClient) ListCustomJobs(ctx context.Context, project, region string) ([]*CustomJob, error) {
	var jobs []*CustomJob
	err := c.service.Projects.Locations.CustomJobs.List(locationParent(project, region)).
		Context(ctx).
		Pages(ctx, func(resp *aiplatform.GoogleCloudAiplatformV1ListCustomJobsResponse) error {
			for _, job := range resp.CustomJobs {
				jobs = append(jobs, customJobFromProto(job))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list custom jobs: %w", err)
	}
	return jobs, nil
}

func (c *gcpClient) GetCustomJob(ctx context.Context, project, region, jobID string) (*CustomJob, error) {
	job, err := c.service.Projects.Locations.CustomJobs.Get(customJobName(project, region, jobID)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get custom job %s: %w", jobID, err)
	}
	return customJobFromProto(job), nil
}

func (c *gcpClient) CreateCustomJob(ctx context.Context, project, region string, req *CreateCustomJobRequest) (*CustomJob, error) {
	machineType := req.MachineType
	if machineType == "" {
		machineType = "n1-standard-4"
	}

	job, err := c.service.Projects.Locations.CustomJobs.Create(locationParent(project, region), &aiplatform.GoogleCloudAiplatformV1CustomJob{
		DisplayName: req.DisplayName,
		JobSpec: &aiplatform.GoogleCloudAiplatformV1CustomJobSpec{
			WorkerPoolSpecs: []*aiplatform.GoogleCloudAiplatformV1WorkerPoolSpec{
				{
					ReplicaCount: 1,
					MachineSpec: &aiplatform.GoogleCloudAiplatformV1MachineSpec{
						MachineType: machineType,
					},
					ContainerSpec: &aiplatform.GoogleCloudAiplatformV1ContainerSpec{
						ImageUri: req.ContainerURI,
						Args:     req.Args,
					},
				},
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create custom job %s: %w", req.DisplayName, err)
	}
	return customJobFromProto(job), nil
}

func (c *gcpClient) CancelCustomJob(ctx context.Context, project, region, jobID string) error {
	if _, err := c.service.Projects.Locations.CustomJobs.Cancel(customJobName(project, region, jobID), &aiplatform.GoogleCloudAiplatformV1CancelCustomJobRequest{}).
		Context(ctx).
		Do(); err != nil {
		return fmt.Errorf("cancel custom job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) ListDatasets(ctx context.Context, project, region string) ([]*Dataset, error) {
	var datasets []*Dataset
	err := c.service.Projects.Locations.Datasets.List(locationParent(project, region)).
		Context(ctx).
		Pages(ctx, func(resp *aiplatform.GoogleCloudAiplatformV1ListDatasetsResponse) error {
			for _, dataset := range resp.Datasets {
				datasets = append(datasets, datasetFromProto(dataset))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list datasets: %w", err)
	}
	return datasets, nil
}

func (c *gcpClient) GetDataset(ctx context.Context, project, region, datasetID string) (*Dataset, error) {
	dataset, err := c.service.Projects.Locations.Datasets.Get(datasetName(project, region, datasetID)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get dataset %s: %w", datasetID, err)
	}
	return datasetFromProto(dataset), nil
}

func (c *gcpClient) CreateDataset(ctx context.Context, project, region string, req *CreateDatasetRequest) (string, error) {
	dataset := &aiplatform.GoogleCloudAiplatformV1Dataset{
		DisplayName:       req.DisplayName,
		Description:       req.Description,
		MetadataSchemaUri: req.MetadataSchemaURI,
		Labels:            req.Labels,
	}
	if req.MetadataJSON != "" {
		var metadata any
		if err := json.Unmarshal([]byte(req.MetadataJSON), &metadata); err != nil {
			return "", fmt.Errorf("decode dataset metadata: %w", err)
		}
		dataset.Metadata = metadata
	} else {
		dataset.Metadata = map[string]any{}
	}

	op, err := c.service.Projects.Locations.Datasets.Create(locationParent(project, region), dataset).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create dataset %s: %w", req.DisplayName, err)
	}
	return op.Name, nil
}

func (c *gcpClient) UpdateDataset(ctx context.Context, project, region, datasetID string, req *UpdateDatasetRequest) (string, error) {
	dataset := &aiplatform.GoogleCloudAiplatformV1Dataset{}
	if req.DisplayName != "" {
		dataset.DisplayName = req.DisplayName
	}
	if req.Description != "" {
		dataset.Description = req.Description
	}
	if req.MetadataSchemaURI != "" {
		dataset.MetadataSchemaUri = req.MetadataSchemaURI
	}
	if req.Labels != nil {
		dataset.Labels = req.Labels
	}
	if req.MetadataJSON != "" {
		var metadata any
		if err := json.Unmarshal([]byte(req.MetadataJSON), &metadata); err != nil {
			return "", fmt.Errorf("decode dataset metadata: %w", err)
		}
		dataset.Metadata = metadata
	}
	if dataset.DisplayName == "" && dataset.Description == "" && dataset.MetadataSchemaUri == "" && dataset.Labels == nil && dataset.Metadata == nil {
		return "", fmt.Errorf("no update fields provided")
	}

	op, err := c.service.Projects.Locations.Datasets.Patch(datasetName(project, region, datasetID), dataset).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("update dataset %s: %w", datasetID, err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeleteDataset(ctx context.Context, project, region, datasetID string) (string, error) {
	op, err := c.service.Projects.Locations.Datasets.Delete(datasetName(project, region, datasetID)).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("delete dataset %s: %w", datasetID, err)
	}
	return op.Name, nil
}

func (c *gcpClient) ListPipelineJobs(ctx context.Context, project, region string) ([]*PipelineJob, error) {
	var jobs []*PipelineJob
	err := c.service.Projects.Locations.PipelineJobs.List(locationParent(project, region)).
		Context(ctx).
		Pages(ctx, func(resp *aiplatform.GoogleCloudAiplatformV1ListPipelineJobsResponse) error {
			for _, job := range resp.PipelineJobs {
				jobs = append(jobs, pipelineJobFromProto(job))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list pipeline jobs: %w", err)
	}
	return jobs, nil
}

func (c *gcpClient) GetPipelineJob(ctx context.Context, project, region, jobID string) (*PipelineJob, error) {
	job, err := c.service.Projects.Locations.PipelineJobs.Get(pipelineJobName(project, region, jobID)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get pipeline job %s: %w", jobID, err)
	}
	return pipelineJobFromProto(job), nil
}

func (c *gcpClient) CreatePipelineJob(ctx context.Context, project, region string, req *CreatePipelineJobRequest) (string, error) {
	job := &aiplatform.GoogleCloudAiplatformV1PipelineJob{
		DisplayName:    req.DisplayName,
		TemplateUri:    req.TemplateURI,
		Network:        req.Network,
		ServiceAccount: req.ServiceAccount,
	}
	if req.PipelineSpecJSON != "" {
		job.PipelineSpec = googleapi.RawMessage([]byte(req.PipelineSpecJSON))
	}
	if req.ParameterValuesJSON != "" || req.GCSOutputDirectory != "" {
		job.RuntimeConfig = &aiplatform.GoogleCloudAiplatformV1PipelineJobRuntimeConfig{
			GcsOutputDirectory: req.GCSOutputDirectory,
		}
		if req.ParameterValuesJSON != "" {
			job.RuntimeConfig.ParameterValues = googleapi.RawMessage([]byte(req.ParameterValuesJSON))
		}
	}

	op, err := c.service.Projects.Locations.PipelineJobs.Create(locationParent(project, region), job).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create pipeline job %s: %w", req.DisplayName, err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeletePipelineJob(ctx context.Context, project, region, jobID string) error {
	_, err := c.service.Projects.Locations.PipelineJobs.Delete(pipelineJobName(project, region, jobID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete pipeline job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) CancelPipelineJob(ctx context.Context, project, region, jobID string) error {
	_, err := c.service.Projects.Locations.PipelineJobs.Cancel(pipelineJobName(project, region, jobID), &aiplatform.GoogleCloudAiplatformV1CancelPipelineJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancel pipeline job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) ListBatchPredictionJobs(ctx context.Context, project, region string) ([]*BatchPredictionJob, error) {
	var jobs []*BatchPredictionJob
	err := c.service.Projects.Locations.BatchPredictionJobs.List(locationParent(project, region)).
		Context(ctx).
		Pages(ctx, func(resp *aiplatform.GoogleCloudAiplatformV1ListBatchPredictionJobsResponse) error {
			for _, job := range resp.BatchPredictionJobs {
				jobs = append(jobs, batchPredictionJobFromProto(job))
			}
			return nil
		})
	if err != nil {
		return nil, fmt.Errorf("list batch prediction jobs: %w", err)
	}
	return jobs, nil
}

func (c *gcpClient) GetBatchPredictionJob(ctx context.Context, project, region, jobID string) (*BatchPredictionJob, error) {
	job, err := c.service.Projects.Locations.BatchPredictionJobs.Get(batchPredictionJobName(project, region, jobID)).
		Context(ctx).
		Do()
	if err != nil {
		return nil, fmt.Errorf("get batch prediction job %s: %w", jobID, err)
	}
	return batchPredictionJobFromProto(job), nil
}

func (c *gcpClient) CreateBatchPredictionJob(ctx context.Context, project, region string, req *CreateBatchPredictionJobRequest) (string, error) {
	job := &aiplatform.GoogleCloudAiplatformV1BatchPredictionJob{
		DisplayName:    req.DisplayName,
		Model:          resolveModelResource(project, region, req.Model),
		ServiceAccount: req.ServiceAccount,
	}
	if req.InstancesFormat != "" {
		job.InputConfig = &aiplatform.GoogleCloudAiplatformV1BatchPredictionJobInputConfig{
			InstancesFormat: req.InstancesFormat,
		}
		if req.GCSSource != "" {
			job.InputConfig.GcsSource = &aiplatform.GoogleCloudAiplatformV1GcsSource{Uris: []string{req.GCSSource}}
		}
		if req.BigQuerySource != "" {
			job.InputConfig.BigquerySource = &aiplatform.GoogleCloudAiplatformV1BigQuerySource{InputUri: req.BigQuerySource}
		}
	}
	if req.PredictionsFormat != "" {
		job.OutputConfig = &aiplatform.GoogleCloudAiplatformV1BatchPredictionJobOutputConfig{
			PredictionsFormat: req.PredictionsFormat,
		}
		if req.GCSDestination != "" {
			job.OutputConfig.GcsDestination = &aiplatform.GoogleCloudAiplatformV1GcsDestination{OutputUriPrefix: req.GCSDestination}
		}
		if req.BigQueryDestination != "" {
			job.OutputConfig.BigqueryDestination = &aiplatform.GoogleCloudAiplatformV1BigQueryDestination{OutputUri: req.BigQueryDestination}
		}
	}
	if req.MachineType != "" {
		job.DedicatedResources = &aiplatform.GoogleCloudAiplatformV1BatchDedicatedResources{
			MachineSpec:          &aiplatform.GoogleCloudAiplatformV1MachineSpec{MachineType: req.MachineType},
			StartingReplicaCount: req.StartingReplicas,
			MaxReplicaCount:      req.MaxReplicas,
		}
	}

	op, err := c.service.Projects.Locations.BatchPredictionJobs.Create(locationParent(project, region), job).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create batch prediction job %s: %w", req.DisplayName, err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeleteBatchPredictionJob(ctx context.Context, project, region, jobID string) error {
	_, err := c.service.Projects.Locations.BatchPredictionJobs.Delete(batchPredictionJobName(project, region, jobID)).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete batch prediction job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) CancelBatchPredictionJob(ctx context.Context, project, region, jobID string) error {
	_, err := c.service.Projects.Locations.BatchPredictionJobs.Cancel(batchPredictionJobName(project, region, jobID), &aiplatform.GoogleCloudAiplatformV1CancelBatchPredictionJobRequest{}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("cancel batch prediction job %s: %w", jobID, err)
	}
	return nil
}

func (c *gcpClient) ListOperations(ctx context.Context, project, region, filter string) ([]*Operation, error) {
	call := c.service.Projects.Locations.Operations.List(locationParent(project, region)).Context(ctx)
	if filter != "" {
		call = call.Filter(filter)
	}

	var operations []*Operation
	if err := call.Pages(ctx, func(resp *aiplatform.GoogleLongrunningListOperationsResponse) error {
		for _, op := range resp.Operations {
			operations = append(operations, operationFromProto(op))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list vertex ai operations: %w", err)
	}
	return operations, nil
}

func (c *gcpClient) GetOperation(ctx context.Context, name string) (*Operation, error) {
	op, err := c.service.Projects.Locations.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get vertex ai operation %s: %w", name, err)
	}
	return operationFromProto(op), nil
}

func (c *gcpClient) waitForOperation(ctx context.Context, name string) (*aiplatform.GoogleLongrunningOperation, error) {
	for {
		op, err := c.service.Projects.Locations.Operations.Wait(name).Context(ctx).Timeout("600s").Do()
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

func modelFromProto(m *aiplatform.GoogleCloudAiplatformV1Model) *Model {
	return &Model{
		Name:        m.Name,
		DisplayName: m.DisplayName,
		State:       "",
	}
}

func endpointFromProto(e *aiplatform.GoogleCloudAiplatformV1Endpoint, region string) *Endpoint {
	return &Endpoint{
		Name:        e.Name,
		DisplayName: e.DisplayName,
		Region:      region,
	}
}

func deployedModelFromProto(dm *aiplatform.GoogleCloudAiplatformV1DeployedModel) *DeployedModel {
	if dm == nil {
		return nil
	}
	return &DeployedModel{
		ID:          dm.Id,
		DisplayName: dm.DisplayName,
		Model:       dm.Model,
		CreateTime:  dm.CreateTime,
	}
}

func customJobFromProto(j *aiplatform.GoogleCloudAiplatformV1CustomJob) *CustomJob {
	return &CustomJob{
		Name:        j.Name,
		DisplayName: j.DisplayName,
		State:       j.State,
		CreateTime:  j.CreateTime,
	}
}

func datasetFromProto(d *aiplatform.GoogleCloudAiplatformV1Dataset) *Dataset {
	if d == nil {
		return nil
	}
	return &Dataset{
		Name:              d.Name,
		DisplayName:       d.DisplayName,
		Description:       d.Description,
		MetadataSchemaURI: d.MetadataSchemaUri,
		CreateTime:        d.CreateTime,
		UpdateTime:        d.UpdateTime,
		DataItemCount:     d.DataItemCount,
	}
}

func pipelineJobFromProto(j *aiplatform.GoogleCloudAiplatformV1PipelineJob) *PipelineJob {
	if j == nil {
		return nil
	}
	return &PipelineJob{
		Name:        j.Name,
		DisplayName: j.DisplayName,
		State:       j.State,
		CreateTime:  j.CreateTime,
		EndTime:     j.EndTime,
		TemplateURI: j.TemplateUri,
	}
}

func batchPredictionJobFromProto(j *aiplatform.GoogleCloudAiplatformV1BatchPredictionJob) *BatchPredictionJob {
	if j == nil {
		return nil
	}
	return &BatchPredictionJob{
		Name:        j.Name,
		DisplayName: j.DisplayName,
		State:       j.State,
		CreateTime:  j.CreateTime,
		EndTime:     j.EndTime,
		Model:       j.Model,
	}
}

func operationFromProto(op *aiplatform.GoogleLongrunningOperation) *Operation {
	if op == nil {
		return nil
	}
	out := &Operation{
		Name: op.Name,
		Done: op.Done,
	}
	if op.Error != nil {
		out.Error = op.Error.Message
	}
	return out
}

func resolveModelResource(project, region, model string) string {
	if strings.HasPrefix(model, "projects/") {
		return model
	}
	return modelName(project, region, model)
}
