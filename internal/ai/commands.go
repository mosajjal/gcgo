package ai

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the Vertex AI command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ai",
		Short: "Manage Vertex AI resources",
	}

	cmd.AddCommand(
		newModelsCommand(cfg, creds),
		newEndpointsCommand(cfg, creds),
		newCustomJobsCommand(cfg, creds),
		newDatasetsCommand(cfg, creds),
		newPipelineJobsCommand(cfg, creds),
		newBatchPredictionJobsCommand(cfg, creds),
		newOperationsCommand(cfg, creds),
	)

	return cmd
}

func newOperationsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Manage Vertex AI operations",
	}
	cmd.AddCommand(
		newOperationsListCommand(cfg, creds),
		newOperationsDescribeCommand(cfg, creds),
	)
	return cmd
}

func newOperationsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var filter string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Vertex AI operations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			operations, err := client.ListOperations(ctx, project, region, filter)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), operations)
			}

			headers := []string{"NAME", "DONE", "ERROR"}
			rows := make([][]string, len(operations))
			for i, operation := range operations {
				rows[i] = []string{operation.Name, fmt.Sprintf("%t", operation.Done), operation.Error}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&filter, "filter", "", "Operations filter expression")
	return cmd
}

func newOperationsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe OPERATION_NAME",
		Short: "Describe a Vertex AI operation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			_, err = requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			operation, err := client.GetOperation(ctx, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), operation)
		},
	}
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func requireRegion(cmd *cobra.Command, cfg *config.Config) (string, error) {
	region, _ := cmd.Flags().GetString("region")
	if region == "" {
		region = cfg.Region()
	}
	if region == "" {
		return "", fmt.Errorf("no region set (use --region or 'gcgo config set region REGION')")
	}
	return region, nil
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newModelsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "models",
		Short: "Manage Vertex AI models",
	}

	cmd.AddCommand(
		newModelsListCommand(cfg, creds),
		newModelsDescribeCommand(cfg, creds),
		newModelsUploadCommand(cfg, creds),
		newModelsDeleteCommand(cfg, creds),
	)

	return cmd
}

func newModelsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List models",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			models, err := client.ListModels(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), models)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "STATE"}
			rows := make([][]string, len(models))
			for i, model := range models {
				rows[i] = []string{model.Name, model.DisplayName, model.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newModelsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe MODEL",
		Short: "Describe a model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			model, err := client.GetModel(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), model)
		},
	}
}

func newModelsUploadCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req UploadModelRequest

	cmd := &cobra.Command{
		Use:   "upload MODEL_ID",
		Short: "Upload a model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" || req.ArtifactURI == "" || req.ContainerURI == "" {
				return fmt.Errorf("--display-name, --artifact-uri, and --container-uri are required")
			}
			req.ModelID = args[0]

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			model, err := client.UploadModel(ctx, project, region, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Uploaded model %s.\n", model.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Model display name")
	cmd.Flags().StringVar(&req.ParentModel, "parent-model", "", "Parent model resource")
	cmd.Flags().StringVar(&req.ArtifactURI, "artifact-uri", "", "Artifact URI")
	cmd.Flags().StringVar(&req.ContainerURI, "container-uri", "", "Serving container image URI")

	return cmd
}

func newModelsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete MODEL",
		Short: "Delete a model",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteModel(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted model %s.\n", args[0])
			return nil
		},
	}
}

func newEndpointsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "endpoints",
		Short: "Manage Vertex AI endpoints",
	}

	cmd.AddCommand(
		newEndpointsListCommand(cfg, creds),
		newEndpointsDescribeCommand(cfg, creds),
		newEndpointsCreateCommand(cfg, creds),
		newEndpointsDeleteCommand(cfg, creds),
		newEndpointsDeployModelCommand(cfg, creds),
		newEndpointsUndeployModelCommand(cfg, creds),
		newEndpointsPredictCommand(cfg, creds),
	)

	return cmd
}

func newEndpointsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List endpoints",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			endpoints, err := client.ListEndpoints(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), endpoints)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "REGION"}
			rows := make([][]string, len(endpoints))
			for i, endpoint := range endpoints {
				rows[i] = []string{endpoint.Name, endpoint.DisplayName, endpoint.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newEndpointsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ENDPOINT",
		Short: "Describe an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			endpoint, err := client.GetEndpoint(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), endpoint)
		},
	}
}

func newEndpointsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateEndpointRequest

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create an endpoint",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" {
				return fmt.Errorf("--display-name is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			endpoint, err := client.CreateEndpoint(ctx, project, region, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created endpoint %s.\n", endpoint.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Endpoint display name")

	return cmd
}

func newEndpointsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENDPOINT",
		Short: "Delete an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteEndpoint(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted endpoint %s.\n", args[0])
			return nil
		},
	}
}

func newEndpointsDeployModelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req DeployModelRequest

	cmd := &cobra.Command{
		Use:   "deploy-model ENDPOINT",
		Short: "Deploy a model to an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.Model == "" {
				return fmt.Errorf("--model is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			deployedModel, err := client.DeployModel(ctx, project, region, args[0], &req)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), deployedModel)
		},
	}
	cmd.Flags().StringVar(&req.Model, "model", "", "Model resource name or model ID")
	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Deployed model display name")
	cmd.Flags().StringVar(&req.MachineType, "machine-type", "", "Machine type for deployment")
	cmd.Flags().Int64Var(&req.MinReplicas, "min-replicas", 1, "Minimum replica count")
	cmd.Flags().Int64Var(&req.MaxReplicas, "max-replicas", 1, "Maximum replica count")
	cmd.Flags().Int64Var(&req.TrafficPercent, "traffic-percent", 100, "Traffic percentage to send to the deployed model")
	return cmd
}

func newEndpointsUndeployModelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var deployedModelID string

	cmd := &cobra.Command{
		Use:   "undeploy-model ENDPOINT",
		Short: "Undeploy a model from an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if deployedModelID == "" {
				return fmt.Errorf("--deployed-model-id is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.UndeployModel(ctx, project, region, args[0], deployedModelID); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Undeployed model %s from endpoint %s.\n", deployedModelID, args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&deployedModelID, "deployed-model-id", "", "Deployed model ID")
	return cmd
}

func newEndpointsPredictCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var instancesRaw string

	cmd := &cobra.Command{
		Use:   "predict ENDPOINT --instances=JSON",
		Short: "Run prediction against an endpoint",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if instancesRaw == "" {
				return fmt.Errorf("--instances is required")
			}

			var instances []any
			if err := json.Unmarshal([]byte(instancesRaw), &instances); err != nil {
				return fmt.Errorf("parse --instances JSON: %w", err)
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			resp, err := client.Predict(ctx, project, region, args[0], instances)
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), resp)
		},
	}

	cmd.Flags().StringVar(&instancesRaw, "instances", "", "Prediction instances as a JSON array")

	return cmd
}

func newCustomJobsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "custom-jobs",
		Short: "Manage Vertex AI custom jobs",
	}

	cmd.AddCommand(
		newCustomJobsListCommand(cfg, creds),
		newCustomJobsDescribeCommand(cfg, creds),
		newCustomJobsCreateCommand(cfg, creds),
		newCustomJobsCancelCommand(cfg, creds),
	)

	return cmd
}

func newCustomJobsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List custom jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			jobs, err := client.ListCustomJobs(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), jobs)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "CREATE_TIME"}
			rows := make([][]string, len(jobs))
			for i, job := range jobs {
				rows[i] = []string{job.Name, job.DisplayName, job.State, job.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newCustomJobsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe JOB",
		Short: "Describe a custom job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			job, err := client.GetCustomJob(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), job)
		},
	}
}

func newCustomJobsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateCustomJobRequest

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a custom job",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" || req.ContainerURI == "" {
				return fmt.Errorf("--display-name and --container-uri are required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			job, err := client.CreateCustomJob(ctx, project, region, &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created custom job %s.\n", job.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Job display name")
	cmd.Flags().StringVar(&req.ContainerURI, "container-uri", "", "Container image URI")
	cmd.Flags().StringVar(&req.MachineType, "machine-type", "", "Machine type")
	cmd.Flags().StringSliceVar(&req.Args, "arg", nil, "Container arguments")

	return cmd
}

func newCustomJobsCancelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel JOB",
		Short: "Cancel a custom job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.CancelCustomJob(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cancelled custom job %s.\n", args[0])
			return nil
		},
	}
}

func newDatasetsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datasets",
		Short: "Manage Vertex AI datasets",
	}
	cmd.AddCommand(
		newDatasetsListCommand(cfg, creds),
		newDatasetsDescribeCommand(cfg, creds),
		newDatasetsCreateCommand(cfg, creds),
		newDatasetsUpdateCommand(cfg, creds),
		newDatasetsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newDatasetsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List datasets",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			datasets, err := client.ListDatasets(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), datasets)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "METADATA_SCHEMA_URI", "DATA_ITEMS"}
			rows := make([][]string, len(datasets))
			for i, dataset := range datasets {
				rows[i] = []string{dataset.Name, dataset.DisplayName, dataset.MetadataSchemaURI, fmt.Sprintf("%d", dataset.DataItemCount)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDatasetsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe DATASET",
		Short: "Describe a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			dataset, err := client.GetDataset(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), dataset)
		},
	}
}

func newDatasetsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateDatasetRequest
	var labels map[string]string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a dataset",
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" || req.MetadataSchemaURI == "" {
				return fmt.Errorf("--display-name and --metadata-schema-uri are required")
			}
			req.Labels = labels

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CreateDataset(ctx, project, region, &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started dataset create operation %s.\n", opName)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Dataset display name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Dataset description")
	cmd.Flags().StringVar(&req.MetadataSchemaURI, "metadata-schema-uri", "", "Metadata schema URI")
	cmd.Flags().StringVar(&req.MetadataJSON, "metadata-json", "", "Dataset metadata JSON object")
	cmd.Flags().StringToStringVar(&labels, "label", nil, "Labels")

	return cmd
}

func newDatasetsUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req UpdateDatasetRequest
	var labels map[string]string

	cmd := &cobra.Command{
		Use:   "update DATASET",
		Short: "Update a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" && req.Description == "" && req.MetadataSchemaURI == "" && req.MetadataJSON == "" && len(labels) == 0 {
				return fmt.Errorf("at least one update flag is required")
			}
			req.Labels = labels

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.UpdateDataset(ctx, project, region, args[0], &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started dataset update operation %s.\n", opName)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Dataset display name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Dataset description")
	cmd.Flags().StringVar(&req.MetadataSchemaURI, "metadata-schema-uri", "", "Metadata schema URI")
	cmd.Flags().StringVar(&req.MetadataJSON, "metadata-json", "", "Dataset metadata JSON object")
	cmd.Flags().StringToStringVar(&labels, "label", nil, "Labels")

	return cmd
}

func newDatasetsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete DATASET",
		Short: "Delete a dataset",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.DeleteDataset(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started dataset delete operation %s.\n", opName)
			return nil
		},
	}
}

func newPipelineJobsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:     "pipeline-jobs",
		Aliases: []string{"pipelines"},
		Short:   "Manage Vertex AI pipeline jobs",
	}
	cmd.AddCommand(
		newPipelineJobsListCommand(cfg, creds),
		newPipelineJobsDescribeCommand(cfg, creds),
		newPipelineJobsCreateCommand(cfg, creds),
		newPipelineJobsDeleteCommand(cfg, creds),
		newPipelineJobsCancelCommand(cfg, creds),
	)
	return cmd
}

func newPipelineJobsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List pipeline jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			jobs, err := client.ListPipelineJobs(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), jobs)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "CREATE_TIME"}
			rows := make([][]string, len(jobs))
			for i, job := range jobs {
				rows[i] = []string{job.Name, job.DisplayName, job.State, job.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newPipelineJobsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe JOB",
		Short: "Describe a pipeline job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			job, err := client.GetPipelineJob(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), job)
		},
	}
}

func newPipelineJobsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreatePipelineJobRequest

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a pipeline job",
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" || req.GCSOutputDirectory == "" {
				return fmt.Errorf("--display-name and --gcs-output-directory are required")
			}
			if req.TemplateURI == "" && req.PipelineSpecJSON == "" {
				return fmt.Errorf("one of --template-uri or --pipeline-spec-json is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CreatePipelineJob(ctx, project, region, &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started pipeline job create operation %s.\n", opName)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Pipeline display name")
	cmd.Flags().StringVar(&req.TemplateURI, "template-uri", "", "Template URI")
	cmd.Flags().StringVar(&req.PipelineSpecJSON, "pipeline-spec-json", "", "Pipeline spec JSON")
	cmd.Flags().StringVar(&req.GCSOutputDirectory, "gcs-output-directory", "", "GCS output directory")
	cmd.Flags().StringVar(&req.ServiceAccount, "service-account", "", "Service account")
	cmd.Flags().StringVar(&req.Network, "network", "", "Compute network")
	cmd.Flags().StringVar(&req.ParameterValuesJSON, "parameter-values-json", "", "Runtime parameter values JSON")

	return cmd
}

func newPipelineJobsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete JOB",
		Short: "Delete a pipeline job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeletePipelineJob(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted pipeline job %s.\n", args[0])
			return nil
		},
	}
}

func newPipelineJobsCancelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel JOB",
		Short: "Cancel a pipeline job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CancelPipelineJob(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cancelled pipeline job %s.\n", args[0])
			return nil
		},
	}
}

func newBatchPredictionJobsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batch-prediction-jobs",
		Short: "Manage Vertex AI batch prediction jobs",
	}
	cmd.AddCommand(
		newBatchPredictionJobsListCommand(cfg, creds),
		newBatchPredictionJobsDescribeCommand(cfg, creds),
		newBatchPredictionJobsCreateCommand(cfg, creds),
		newBatchPredictionJobsDeleteCommand(cfg, creds),
		newBatchPredictionJobsCancelCommand(cfg, creds),
	)
	return cmd
}

func newBatchPredictionJobsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List batch prediction jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			jobs, err := client.ListBatchPredictionJobs(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), jobs)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "MODEL", "CREATE_TIME"}
			rows := make([][]string, len(jobs))
			for i, job := range jobs {
				rows[i] = []string{job.Name, job.DisplayName, job.State, job.Model, job.CreateTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newBatchPredictionJobsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe JOB",
		Short: "Describe a batch prediction job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			job, err := client.GetBatchPredictionJob(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), job)
		},
	}
}

func newBatchPredictionJobsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateBatchPredictionJobRequest

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a batch prediction job",
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" || req.Model == "" || req.InstancesFormat == "" || req.PredictionsFormat == "" {
				return fmt.Errorf("--display-name, --model, --instances-format, and --predictions-format are required")
			}
			if req.GCSSource == "" && req.BigQuerySource == "" {
				return fmt.Errorf("one of --gcs-source or --bigquery-source is required")
			}
			if req.GCSDestination == "" && req.BigQueryDestination == "" {
				return fmt.Errorf("one of --gcs-destination or --bigquery-destination is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CreateBatchPredictionJob(ctx, project, region, &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started batch prediction job create operation %s.\n", opName)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Job display name")
	cmd.Flags().StringVar(&req.Model, "model", "", "Model resource name or ID")
	cmd.Flags().StringVar(&req.InstancesFormat, "instances-format", "", "Input instances format")
	cmd.Flags().StringVar(&req.GCSSource, "gcs-source", "", "GCS source URI")
	cmd.Flags().StringVar(&req.BigQuerySource, "bigquery-source", "", "BigQuery source URI")
	cmd.Flags().StringVar(&req.PredictionsFormat, "predictions-format", "", "Predictions output format")
	cmd.Flags().StringVar(&req.GCSDestination, "gcs-destination", "", "GCS destination prefix")
	cmd.Flags().StringVar(&req.BigQueryDestination, "bigquery-destination", "", "BigQuery destination URI")
	cmd.Flags().StringVar(&req.ServiceAccount, "service-account", "", "Service account")
	cmd.Flags().StringVar(&req.MachineType, "machine-type", "", "Machine type")
	cmd.Flags().Int64Var(&req.StartingReplicas, "starting-replicas", 0, "Starting replica count")
	cmd.Flags().Int64Var(&req.MaxReplicas, "max-replicas", 0, "Maximum replica count")

	return cmd
}

func newBatchPredictionJobsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete JOB",
		Short: "Delete a batch prediction job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteBatchPredictionJob(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted batch prediction job %s.\n", args[0])
			return nil
		},
	}
}

func newBatchPredictionJobsCancelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "cancel JOB",
		Short: "Cancel a batch prediction job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CancelBatchPredictionJob(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cancelled batch prediction job %s.\n", args[0])
			return nil
		},
	}
}
