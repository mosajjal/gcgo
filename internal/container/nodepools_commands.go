package container

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func nodePoolClient(ctx context.Context, creds *auth.Credentials) (NodePoolClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewNodePoolClient(ctx, opt)
}

func newNodePoolsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "node-pools",
		Short: "Manage GKE node pools",
	}

	cmd.AddCommand(
		newNodePoolsListCommand(cfg, creds),
		newNodePoolsDescribeCommand(cfg, creds),
		newNodePoolsCreateCommand(cfg, creds),
		newNodePoolsDeleteCommand(cfg, creds),
		newNodePoolsUpdateCommand(cfg, creds),
		newNodePoolsUpgradeCommand(cfg, creds),
	)

	return cmd
}

func newNodePoolsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location string
		cluster  string
	)

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List node pools in a cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
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
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}

			ctx := context.Background()
			client, err := nodePoolClient(ctx, creds)
			if err != nil {
				return err
			}

			pools, err := client.ListNodePools(ctx, project, location, cluster)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), pools)
			}

			headers := []string{"NAME", "MACHINE_TYPE", "NODES", "STATUS"}
			rows := make([][]string, len(pools))
			for i, p := range pools {
				rows[i] = []string{p.Name, p.MachineType, fmt.Sprintf("%d", p.NodeCount), p.Status}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")

	return cmd
}

func newNodePoolsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location string
		cluster  string
	)

	cmd := &cobra.Command{
		Use:   "describe NODE_POOL",
		Short: "Describe a node pool",
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
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}

			ctx := context.Background()
			client, err := nodePoolClient(ctx, creds)
			if err != nil {
				return err
			}

			pool, err := client.GetNodePool(ctx, project, location, cluster, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), pool)
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")

	return cmd
}

func newNodePoolsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location    string
		cluster     string
		numNodes    int32
		machineType string
	)

	cmd := &cobra.Command{
		Use:   "create NODE_POOL",
		Short: "Create a node pool",
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
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}

			ctx := context.Background()
			client, err := nodePoolClient(ctx, creds)
			if err != nil {
				return err
			}

			req := &CreateNodePoolRequest{
				Name:        args[0],
				MachineType: machineType,
				NumNodes:    numNodes,
			}
			if err := client.CreateNodePool(ctx, project, location, cluster, req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created node pool %s in cluster %s.\n", args[0], cluster)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	cmd.Flags().Int32Var(&numNodes, "num-nodes", 3, "Number of nodes")
	cmd.Flags().StringVar(&machineType, "machine-type", "e2-medium", "Node machine type")

	return cmd
}

func newNodePoolsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location string
		cluster  string
	)

	cmd := &cobra.Command{
		Use:   "delete NODE_POOL",
		Short: "Delete a node pool",
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
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}

			ctx := context.Background()
			client, err := nodePoolClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteNodePool(ctx, project, location, cluster, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted node pool %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")

	return cmd
}

func newNodePoolsUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location string
		cluster  string
		numNodes int32
	)

	cmd := &cobra.Command{
		Use:   "update NODE_POOL",
		Short: "Update a node pool",
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
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}

			ctx := context.Background()
			client, err := nodePoolClient(ctx, creds)
			if err != nil {
				return err
			}

			req := &UpdateNodePoolRequest{NumNodes: numNodes}
			if err := client.UpdateNodePool(ctx, project, location, cluster, args[0], req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated node pool %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	cmd.Flags().Int32Var(&numNodes, "num-nodes", 3, "Desired number of nodes")

	return cmd
}

func newNodePoolsUpgradeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var (
		location    string
		cluster     string
		nodeVersion string
		imageType   string
	)

	cmd := &cobra.Command{
		Use:   "upgrade NODE_POOL",
		Short: "Upgrade a node pool",
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
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}
			if nodeVersion == "" && imageType == "" {
				return fmt.Errorf("at least one of --node-version or --image-type is required")
			}

			ctx := context.Background()
			client, err := nodePoolClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.UpgradeNodePool(ctx, project, location, cluster, args[0], &UpgradeNodePoolRequest{
				NodeVersion: nodeVersion,
				ImageType:   imageType,
			}); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Upgraded node pool %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Cluster location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	cmd.Flags().StringVar(&nodeVersion, "node-version", "", "Desired node version")
	cmd.Flags().StringVar(&imageType, "image-type", "", "Desired node image type")

	return cmd
}
