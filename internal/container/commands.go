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

	cmd.AddCommand(newClustersCommand(cfg, creds))

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
