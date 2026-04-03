package spanner

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the spanner command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "spanner",
		Short: "Manage Cloud Spanner resources",
	}

	cmd.AddCommand(
		newInstancesCommand(cfg, creds),
		newDatabasesCommand(cfg, creds),
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

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

// --- instances subcommands ---

func newInstancesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "Manage Spanner instances",
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
	return &cobra.Command{
		Use:   "list",
		Short: "List Spanner instances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			instances, err := client.ListInstances(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), instances)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "CONFIG", "NODE_COUNT", "STATE"}
			rows := make([][]string, len(instances))
			for i, inst := range instances {
				rows[i] = []string{inst.Name, inst.DisplayName, inst.Config, fmt.Sprintf("%d", inst.NodeCount), inst.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newInstancesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe a Spanner instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			inst, err := client.GetInstance(ctx, project, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), inst)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", inst.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", inst.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Config:       %s\n", inst.Config)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Node Count:   %d\n", inst.NodeCount)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:        %s\n", inst.State)
			return nil
		},
	}
}

func newInstancesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateInstanceRequest

	cmd := &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a Spanner instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			req.Name = args[0]

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.CreateInstance(ctx, project, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created Spanner instance %q.\n", req.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&req.Config, "config", "", "Instance config (e.g. regional-us-central1)")
	cmd.Flags().Int32Var(&req.NodeCount, "nodes", 1, "Number of nodes")

	return cmd
}

func newInstancesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete INSTANCE",
		Short: "Delete a Spanner instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteInstance(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted Spanner instance %q.\n", args[0])
			return nil
		},
	}
}

// --- databases subcommands ---

func newDatabasesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "databases",
		Short: "Manage Spanner databases",
	}

	cmd.AddCommand(
		newDatabasesListCommand(cfg, creds),
		newDatabasesDescribeCommand(cfg, creds),
		newDatabasesCreateCommand(cfg, creds),
		newDatabasesDeleteCommand(cfg, creds),
		newDatabasesExecuteSQLCommand(cfg, creds),
	)

	return cmd
}

func newDatabasesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List databases in a Spanner instance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			instance, _ := cmd.Flags().GetString("instance")
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			databases, err := client.ListDatabases(ctx, project, instance)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), databases)
			}

			headers := []string{"NAME", "STATE"}
			rows := make([][]string, len(databases))
			for i, d := range databases {
				rows[i] = []string{d.Name, d.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().String("instance", "", "Spanner instance name")

	return cmd
}

func newDatabasesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe DATABASE",
		Short: "Describe a Spanner database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			instance, _ := cmd.Flags().GetString("instance")
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			db, err := client.GetDatabase(ctx, project, instance, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), db)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:  %s\n", db.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State: %s\n", db.State)
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Spanner instance name")

	return cmd
}

func newDatabasesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create DATABASE",
		Short: "Create a Spanner database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			instance, _ := cmd.Flags().GetString("instance")
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.CreateDatabase(ctx, project, instance, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created Spanner database %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Spanner instance name")

	return cmd
}

func newDatabasesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete DATABASE",
		Short: "Delete a Spanner database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			instance, _ := cmd.Flags().GetString("instance")
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteDatabase(ctx, project, instance, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted Spanner database %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Spanner instance name")

	return cmd
}

func newDatabasesExecuteSQLCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "execute-sql DATABASE",
		Short: "Execute a SQL statement against a Spanner database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			inst, _ := cmd.Flags().GetString("instance")
			if inst == "" {
				return fmt.Errorf("--instance is required")
			}
			sql, _ := cmd.Flags().GetString("sql")
			if sql == "" {
				return fmt.Errorf("--sql is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			result, err := client.ExecuteSQL(ctx, project, inst, args[0], sql)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), result)
			}

			return output.PrintTable(cmd.OutOrStdout(), result.Columns, result.Rows)
		},
	}

	cmd.Flags().String("instance", "", "Spanner instance name")
	cmd.Flags().String("sql", "", "SQL statement to execute")

	return cmd
}
