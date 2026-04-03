package container

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the container command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "container",
		Short: "Manage GKE clusters",
	}

	cmd.AddCommand(
		newClustersCommand(cfg, creds),
		newNodePoolsCommand(cfg, creds),
	)

	return cmd
}

func newClustersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "clusters",
		Short: "Manage GKE clusters",
	}

	cmd.AddCommand(
		newClustersListCommand(cfg, creds),
		newClustersDescribeCommand(cfg, creds),
		newClustersCreateCommand(cfg, creds),
		newClustersDeleteCommand(cfg, creds),
		newClustersUpdateCommand(cfg, creds),
		newClustersUpgradeCommand(cfg, creds),
		newClustersOperationsCommand(cfg, creds),
		newGetCredentialsCommand(cfg, creds),
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

func gkeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newClustersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List GKE clusters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = "-" // all locations
			}

			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}

			clusters, err := client.ListClusters(ctx, project, location)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), clusters)
			}

			headers := []string{"NAME", "LOCATION", "STATUS", "NODES"}
			rows := make([][]string, len(clusters))
			for i, c := range clusters {
				rows[i] = []string{c.Name, c.Location, c.Status, fmt.Sprintf("%d", c.NodeCount)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location (default: all)")

	return cmd
}

func newClustersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "describe CLUSTER",
		Short: "Describe a GKE cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}

			cluster, err := client.GetCluster(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), cluster)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")

	return cmd
}

func newGetCredentialsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "get-credentials CLUSTER",
		Short: "Write kubeconfig for a GKE cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}

			clAuth, err := client.GetClusterAuth(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			kubeconfig := kubeconfigPath()
			if err := writeKubeconfig(kubeconfig, args[0], project, location, clAuth); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "kubeconfig entry written to %s for cluster %s.\n", kubeconfig, args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")

	return cmd
}

func newClustersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location    string
		numNodes    int32
		machineType string
	)

	cmd := &cobra.Command{
		Use:   "create CLUSTER",
		Short: "Create a GKE cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}
			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateCluster(ctx, project, location, &CreateClusterRequest{
				Name:        args[0],
				NumNodes:    numNodes,
				MachineType: machineType,
			}); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created cluster %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().Int32Var(&numNodes, "num-nodes", 3, "Number of nodes")
	cmd.Flags().StringVar(&machineType, "machine-type", "e2-medium", "Node machine type")
	return cmd
}

func newClustersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "delete CLUSTER",
		Short: "Delete a GKE cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}
			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteCluster(ctx, project, location, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted cluster %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	return cmd
}

func newClustersUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location      string
		masterVersion string
		nodeVersion   string
	)

	cmd := &cobra.Command{
		Use:   "update CLUSTER",
		Short: "Update a GKE cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}
			if masterVersion == "" && nodeVersion == "" {
				return fmt.Errorf("at least one of --master-version or --node-version is required")
			}
			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.UpdateCluster(ctx, project, location, args[0], &UpdateClusterRequest{
				MasterVersion: masterVersion,
				NodeVersion:   nodeVersion,
			}); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated cluster %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&masterVersion, "master-version", "", "Desired master version")
	cmd.Flags().StringVar(&nodeVersion, "node-version", "", "Desired node version")
	return cmd
}

func newClustersUpgradeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location string
		version  string
	)

	cmd := &cobra.Command{
		Use:   "upgrade CLUSTER",
		Short: "Upgrade a GKE cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = cfg.Region()
			}
			if location == "" {
				return fmt.Errorf("--location is required (or set region in config)")
			}
			if version == "" {
				return fmt.Errorf("--cluster-version is required")
			}
			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.UpgradeCluster(ctx, project, location, args[0], &UpgradeClusterRequest{Version: version}); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Upgraded cluster %s to %s.\n", args[0], version)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&version, "cluster-version", "", "Desired cluster version")
	return cmd
}

func newClustersOperationsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Manage GKE operations",
	}

	cmd.AddCommand(
		newClustersOperationsListCommand(cfg, creds, &location),
		newClustersOperationsDescribeCommand(cfg, creds, &location),
	)

	cmd.Flags().StringVar(&location, "location", "", "Cluster location (default: all)")
	return cmd
}

func newClustersOperationsListCommand(cfg *config.Config, creds *auth.Credentials, location *string) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List GKE operations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			loc := *location
			if loc == "" {
				loc = "-"
			}
			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}
			ops, err := client.ListOperations(ctx, project, loc)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), ops)
			}
			headers := []string{"NAME", "LOCATION", "STATUS", "TYPE", "DETAIL"}
			rows := make([][]string, len(ops))
			for i, op := range ops {
				rows[i] = []string{op.Name, op.Location, op.Status, op.OperationType, op.Detail}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newClustersOperationsDescribeCommand(cfg *config.Config, creds *auth.Credentials, location *string) *cobra.Command {
	return &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe a GKE operation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			loc := *location
			if loc == "" {
				return fmt.Errorf("--location is required")
			}
			ctx := context.Background()
			client, err := gkeClient(ctx, creds)
			if err != nil {
				return err
			}
			op, err := client.GetOperation(ctx, project, loc, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), op)
		},
	}
}

func kubeconfigPath() string {
	if v := os.Getenv("KUBECONFIG"); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return filepath.Join(".", ".kube", "config")
	}
	return filepath.Join(home, ".kube", "config")
}

func writeKubeconfig(path, clusterName, project, location string, clAuth *ClusterAuth) error {
	contextName := fmt.Sprintf("gcgo_%s_%s_%s", project, location, clusterName)
	caCertB64 := base64.StdEncoding.EncodeToString(clAuth.CACert)

	content := fmt.Sprintf(`apiVersion: v1
kind: Config
clusters:
- cluster:
    certificate-authority-data: %s
    server: https://%s
  name: %s
contexts:
- context:
    cluster: %s
    user: %s
  name: %s
current-context: %s
users:
- name: %s
  user:
    exec:
      apiVersion: client.authentication.k8s.io/v1beta1
      command: gke-gcloud-auth-plugin
      installHint: Install gke-gcloud-auth-plugin for kubectl auth
`, caCertB64, clAuth.Endpoint, contextName,
		contextName, contextName, contextName, contextName,
		contextName)

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create kubeconfig dir: %w", err)
	}

	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		return fmt.Errorf("write kubeconfig: %w", err)
	}

	return nil
}
