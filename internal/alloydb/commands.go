package alloydb

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the AlloyDB command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "alloydb",
		Short: "Manage AlloyDB resources",
	}
	cmd.AddCommand(
		newClustersCommand(cfg, creds),
		newInstancesCommand(cfg, creds),
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

func requireLocation(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("location")
	if flagVal != "" {
		return flagVal, nil
	}
	location := cfg.Region()
	if location == "" {
		return "", fmt.Errorf("--location is required (or set region in config)")
	}
	return location, nil
}

func alloydbClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newClustersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{Use: "clusters", Short: "Manage AlloyDB clusters"}
	cmd.AddCommand(
		newClustersListCommand(cfg, creds),
		newClustersDescribeCommand(cfg, creds),
		newClustersCreateCommand(cfg, creds),
		newClustersDeleteCommand(cfg, creds),
	)
	return cmd
}

func newClustersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List clusters",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
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
			headers := []string{"NAME", "DISPLAY_NAME", "DB_VERSION", "NETWORK"}
			rows := make([][]string, len(clusters))
			for i, cluster := range clusters {
				rows[i] = []string{cluster.Name, cluster.DisplayName, cluster.DatabaseVersion, cluster.Network}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	return cmd
}

func newClustersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "describe CLUSTER",
		Short: "Describe a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
			if err != nil {
				return err
			}
			cluster, err := client.GetCluster(ctx, project, location, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), cluster)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:            %s\n", cluster.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DisplayName:     %s\n", cluster.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DatabaseVersion: %s\n", cluster.DatabaseVersion)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Network:         %s\n", cluster.Network)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	return cmd
}

func newClustersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var displayName string
	var databaseVersion string
	var network string
	var allocatedIPRange string
	var username string
	var password string

	cmd := &cobra.Command{
		Use:   "create CLUSTER",
		Short: "Create a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if network == "" || username == "" || password == "" {
				return fmt.Errorf("--network, --username, and --password are required")
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
			if err != nil {
				return err
			}
			cluster, err := client.CreateCluster(ctx, project, location, &CreateClusterRequest{
				Name:             args[0],
				DisplayName:      displayName,
				DatabaseVersion:  databaseVersion,
				Network:          network,
				AllocatedIPRange: allocatedIPRange,
				Username:         username,
				Password:         password,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created cluster %s.\n", cluster.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&databaseVersion, "database-version", "POSTGRES_15", "Database version")
	cmd.Flags().StringVar(&network, "network", "", "VPC network resource path")
	cmd.Flags().StringVar(&allocatedIPRange, "allocated-ip-range", "", "Allocated private service range")
	cmd.Flags().StringVar(&username, "username", "", "Initial database username")
	cmd.Flags().StringVar(&password, "password", "", "Initial database password")
	return cmd
}

func newClustersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "delete CLUSTER",
		Short: "Delete a cluster",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
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
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	return cmd
}

func newInstancesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{Use: "instances", Short: "Manage AlloyDB instances"}
	cmd.AddCommand(
		newInstancesListCommand(cfg, creds),
		newInstancesDescribeCommand(cfg, creds),
		newInstancesCreateCommand(cfg, creds),
		newInstancesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newInstancesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var cluster string
	cmd := &cobra.Command{
		Use:   "list --cluster=CLUSTER",
		Short: "List instances in a cluster",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
			if err != nil {
				return err
			}
			instances, err := client.ListInstances(ctx, project, location, cluster)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), instances)
			}
			headers := []string{"NAME", "TYPE", "AVAILABILITY", "CPU", "STATE", "IP"}
			rows := make([][]string, len(instances))
			for i, instance := range instances {
				rows[i] = []string{
					instance.Name,
					instance.InstanceType,
					instance.AvailabilityType,
					fmt.Sprintf("%d", instance.CpuCount),
					instance.State,
					instance.IPAddress,
				}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	return cmd
}

func newInstancesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var cluster string
	cmd := &cobra.Command{
		Use:   "describe INSTANCE --cluster=CLUSTER",
		Short: "Describe an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
			if err != nil {
				return err
			}
			instance, err := client.GetInstance(ctx, project, location, cluster, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), instance)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", instance.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DisplayName:  %s\n", instance.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:         %s\n", instance.InstanceType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Availability: %s\n", instance.AvailabilityType)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "CPU:          %d\n", instance.CpuCount)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:        %s\n", instance.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "IP:           %s\n", instance.IPAddress)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	return cmd
}

func newInstancesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var cluster string
	var displayName string
	var instanceType string
	var availabilityType string
	var cpuCount int64
	var nodeCount int64
	var zone string

	cmd := &cobra.Command{
		Use:   "create INSTANCE --cluster=CLUSTER",
		Short: "Create an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
			if err != nil {
				return err
			}
			instance, err := client.CreateInstance(ctx, project, location, cluster, &CreateInstanceRequest{
				Name:             args[0],
				DisplayName:      displayName,
				InstanceType:     instanceType,
				AvailabilityType: availabilityType,
				CPUCount:         cpuCount,
				NodeCount:        nodeCount,
				Zone:             zone,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created instance %s.\n", instance.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&instanceType, "instance-type", "PRIMARY", "Instance type")
	cmd.Flags().StringVar(&availabilityType, "availability-type", "REGIONAL", "Availability type")
	cmd.Flags().Int64Var(&cpuCount, "cpu-count", 2, "CPU count")
	cmd.Flags().Int64Var(&nodeCount, "node-count", 1, "Read pool node count")
	cmd.Flags().StringVar(&zone, "zone", "", "GCE zone for zonal instances")
	return cmd
}

func newInstancesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var cluster string
	cmd := &cobra.Command{
		Use:   "delete INSTANCE --cluster=CLUSTER",
		Short: "Delete an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			location, err = requireLocation(cmd, cfg)
			if err != nil {
				return err
			}
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}
			ctx := context.Background()
			client, err := alloydbClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteInstance(ctx, project, location, cluster, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted instance %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "AlloyDB location")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Cluster name")
	return cmd
}
