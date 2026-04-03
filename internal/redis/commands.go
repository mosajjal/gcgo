package redis

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the redis command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "redis",
		Short: "Manage Memorystore for Redis",
	}

	cmd.AddCommand(newInstancesCommand(cfg, creds))
	return cmd
}

func redisClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
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

func newInstancesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "Manage Redis instances",
	}

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
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Redis instances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				location = "-"
			}

			ctx := context.Background()
			client, err := redisClient(ctx, creds)
			if err != nil {
				return err
			}
			instances, err := client.ListInstances(ctx, project, location)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), instances)
			}
			headers := []string{"NAME", "STATE", "TIER", "MEMORY_GB", "HOST", "LOCATION"}
			rows := make([][]string, len(instances))
			for i, inst := range instances {
				rows[i] = []string{inst.Name, inst.State, inst.Tier, fmt.Sprintf("%d", inst.MemorySizeGB), inst.Host, inst.LocationID}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Redis location (default: all locations)")
	return cmd
}

func newInstancesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe a Redis instance",
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
			client, err := redisClient(ctx, creds)
			if err != nil {
				return err
			}
			inst, err := client.GetInstance(ctx, project, location, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), inst)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:              %s\n", inst.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name:      %s\n", inst.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:             %s\n", inst.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Tier:              %s\n", inst.Tier)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Memory Size (GiB): %d\n", inst.MemorySizeGB)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Host:              %s\n", inst.Host)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Location:          %s\n", inst.LocationID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Network:           %s\n", inst.AuthorizedNetwork)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Redis location")
	return cmd
}

func newInstancesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var tier string
	var memoryGB int64
	var displayName string
	var authorizedNetwork string

	cmd := &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a Redis instance",
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
			if tier == "" {
				tier = "BASIC"
			}
			if memoryGB <= 0 {
				return fmt.Errorf("--memory-size-gb must be greater than zero")
			}

			ctx := context.Background()
			client, err := redisClient(ctx, creds)
			if err != nil {
				return err
			}
			inst, err := client.CreateInstance(ctx, project, location, &CreateInstanceRequest{
				Name:              args[0],
				DisplayName:       displayName,
				Tier:              tier,
				MemorySizeGB:      memoryGB,
				LocationID:        location,
				AuthorizedNetwork: authorizedNetwork,
			})
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created Redis instance %s.\n", inst.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Redis location")
	cmd.Flags().StringVar(&tier, "tier", "BASIC", "Redis tier (BASIC or STANDARD_HA)")
	cmd.Flags().Int64Var(&memoryGB, "memory-size-gb", 1, "Redis memory size in GiB")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&authorizedNetwork, "authorized-network", "", "Authorized network")
	return cmd
}

func newInstancesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	cmd := &cobra.Command{
		Use:   "delete INSTANCE",
		Short: "Delete a Redis instance",
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
			client, err := redisClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteInstance(ctx, project, location, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted Redis instance %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Redis location")
	return cmd
}
