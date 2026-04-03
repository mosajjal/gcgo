package dataproc

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the dataproc command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "dataproc",
		Short: "Manage Dataproc resources",
	}

	cmd.AddCommand(
		newClustersCommand(cfg, creds),
		newJobsCommand(cfg, creds),
		newBatchesCommand(cfg, creds),
	)

	return cmd
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

func newClustersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "Manage Dataproc clusters",
	}

	cmd.AddCommand(
		newClustersListCommand(cfg, creds),
		newClustersDescribeCommand(cfg, creds),
		newClustersCreateCommand(cfg, creds),
		newClustersDeleteCommand(cfg, creds),
		newClustersStartCommand(cfg, creds),
		newClustersStopCommand(cfg, creds),
	)

	return cmd
}

func newClustersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Dataproc clusters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			clusters, err := client.ListClusters(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), clusters)
			}
			headers := []string{"NAME", "REGION", "STATUS", "CONFIG"}
			rows := make([][]string, len(clusters))
			for i, cluster := range clusters {
				rows[i] = []string{cluster.Name, cluster.Region, cluster.Status, cluster.Config}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newClustersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "describe CLUSTER",
		Short: "Describe a Dataproc cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			cluster, err := client.GetCluster(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), cluster)
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newClustersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var req CreateClusterRequest
	cmd := &cobra.Command{
		Use:   "create CLUSTER",
		Short: "Create a Dataproc cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateCluster(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created cluster %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&req.MachineType, "machine-type", "n1-standard-4", "Machine type for nodes")
	cmd.Flags().Int64Var(&req.NumWorkers, "num-workers", 2, "Number of worker nodes")
	cmd.Flags().StringVar(&req.ImageVersion, "image-version", "", "Dataproc image version")
	return cmd
}

func newClustersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "delete CLUSTER",
		Short: "Delete a Dataproc cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteCluster(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted cluster %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newClustersStartCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "start CLUSTER",
		Short: "Start a Dataproc cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.StartCluster(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started cluster %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newClustersStopCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "stop CLUSTER",
		Short: "Stop a Dataproc cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.StopCluster(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Stopped cluster %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newJobsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "jobs",
		Short: "Manage Dataproc jobs",
	}

	cmd.AddCommand(
		newJobsListCommand(cfg, creds),
		newJobsDescribeCommand(cfg, creds),
		newJobsSubmitCommand(cfg, creds),
		newJobsCancelCommand(cfg, creds),
	)

	return cmd
}

func newJobsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Dataproc jobs",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			jobs, err := client.ListJobs(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), jobs)
			}
			headers := []string{"JOB_ID", "TYPE", "STATUS"}
			rows := make([][]string, len(jobs))
			for i, job := range jobs {
				rows[i] = []string{job.ID, job.Type, job.Status}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newJobsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "describe JOB_ID",
		Short: "Describe a Dataproc job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			job, err := client.GetJob(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), job)
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newJobsSubmitCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var req SubmitJobRequest
	cmd := &cobra.Command{
		Use:   "submit",
		Short: "Submit a Dataproc Spark job",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			if req.ClusterName == "" {
				return fmt.Errorf("--cluster is required")
			}
			if req.MainClass == "" {
				return fmt.Errorf("--class is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			job, err := client.SubmitJob(ctx, project, region, &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Submitted job %q.\n", job.ID)
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&req.ClusterName, "cluster", "", "Cluster name")
	cmd.Flags().StringVar(&req.MainClass, "class", "", "Main class")
	cmd.Flags().StringSliceVar(&req.JarFileURIs, "jars", nil, "Jar file URIs")
	cmd.Flags().StringSliceVar(&req.Args, "args", nil, "Job arguments")
	return cmd
}

func newJobsCancelCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "cancel JOB_ID",
		Short: "Cancel a Dataproc job",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CancelJob(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Cancelled job %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newBatchesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "batches",
		Short: "Manage Dataproc batches",
	}

	cmd.AddCommand(
		newBatchesListCommand(cfg, creds),
		newBatchesDescribeCommand(cfg, creds),
		newBatchesCreateCommand(cfg, creds),
		newBatchesDeleteCommand(cfg, creds),
	)

	return cmd
}

func newBatchesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Dataproc batches",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			batches, err := client.ListBatches(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), batches)
			}
			headers := []string{"NAME", "STATE", "CREATE_TIME"}
			rows := make([][]string, len(batches))
			for i, batch := range batches {
				rows[i] = []string{batch.Name, batch.State, batch.Create}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newBatchesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "describe BATCH",
		Short: "Describe a Dataproc batch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			batch, err := client.GetBatch(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), batch)
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}

func newBatchesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var req CreateBatchRequest
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Dataproc batch",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			if req.BatchID == "" {
				return fmt.Errorf("--batch-id is required")
			}
			if req.MainClass == "" {
				return fmt.Errorf("--class is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateBatch(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created batch %q.\n", req.BatchID)
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	cmd.Flags().StringVar(&req.BatchID, "batch-id", "", "Batch ID")
	cmd.Flags().StringVar(&req.MainClass, "class", "", "Main class")
	cmd.Flags().StringSliceVar(&req.JarFileURIs, "jars", nil, "Jar file URIs")
	cmd.Flags().StringSliceVar(&req.Args, "args", nil, "Batch arguments")
	return cmd
}

func newBatchesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	cmd := &cobra.Command{
		Use:   "delete BATCH",
		Short: "Delete a Dataproc batch",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region, err = requireRegion(cmd, cfg)
				if err != nil {
					return err
				}
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteBatch(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted batch %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&region, "region", "", "Region")
	return cmd
}
