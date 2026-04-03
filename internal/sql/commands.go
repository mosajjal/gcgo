package sql

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the sql command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "sql",
		Short: "Manage Cloud SQL resources",
	}

	cmd.AddCommand(
		newInstancesCommand(cfg, creds),
		newDatabasesCommand(cfg, creds),
		newUsersCommand(cfg, creds),
		newBackupsCommand(cfg, creds),
		newOperationsCommand(cfg, creds),
	)

	return cmd
}

func newOperationsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Manage Cloud SQL operations",
	}
	cmd.AddCommand(
		newOperationsListCommand(cfg, creds),
		newOperationsDescribeCommand(cfg, creds),
	)
	return cmd
}

func newOperationsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Cloud SQL operations",
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
			operations, err := client.ListOperations(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), operations)
			}

			headers := []string{"NAME", "TYPE", "STATUS", "TARGET", "INSERT_TIME"}
			rows := make([][]string, len(operations))
			for i, op := range operations {
				rows[i] = []string{op.Name, op.Type, op.Status, op.TargetID, op.InsertTime}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newOperationsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe OPERATION",
		Short: "Describe a Cloud SQL operation",
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
			op, err := client.GetOperation(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), op)
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
		Short: "Manage Cloud SQL instances",
	}

	cmd.AddCommand(
		newInstancesListCommand(cfg, creds),
		newInstancesDescribeCommand(cfg, creds),
		newInstancesCreateCommand(cfg, creds),
		newInstancesUpdateCommand(cfg, creds),
		newInstancesDeleteCommand(cfg, creds),
		newInstancesRestartCommand(cfg, creds),
		newInstancesExportCommand(cfg, creds),
		newInstancesImportCommand(cfg, creds),
		newInstancesCloneCommand(cfg, creds),
		newInstancesPromoteReplicaCommand(cfg, creds),
		newInstancesFailoverCommand(cfg, creds),
	)

	return cmd
}

func newInstancesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Cloud SQL instances",
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

			headers := []string{"NAME", "DATABASE_VERSION", "REGION", "TIER", "STATE", "IP_ADDRESS"}
			rows := make([][]string, len(instances))
			for i, inst := range instances {
				rows[i] = []string{inst.Name, inst.DatabaseVersion, inst.Region, inst.Tier, inst.State, inst.IPAddress}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newInstancesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe a Cloud SQL instance",
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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:             %s\n", inst.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Database Version: %s\n", inst.DatabaseVersion)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Region:           %s\n", inst.Region)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Tier:             %s\n", inst.Tier)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:            %s\n", inst.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "IP Address:       %s\n", inst.IPAddress)
			return nil
		},
	}
}

func newInstancesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateInstanceRequest

	cmd := &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a Cloud SQL instance",
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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created Cloud SQL instance %q.\n", req.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DatabaseVersion, "database-version", "POSTGRES_15", "Database version")
	cmd.Flags().StringVar(&req.Tier, "tier", "db-f1-micro", "Machine tier")
	cmd.Flags().StringVar(&req.Region, "region", "", "Region")

	return cmd
}

func newInstancesUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req UpdateInstanceRequest

	cmd := &cobra.Command{
		Use:   "update INSTANCE",
		Short: "Update a Cloud SQL instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DatabaseVersion == "" && req.Tier == "" {
				return fmt.Errorf("at least one of --database-version or --tier is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			opName, err := client.UpdateInstance(ctx, project, args[0], &req)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started update operation %s.\n", opName)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DatabaseVersion, "database-version", "", "Database version")
	cmd.Flags().StringVar(&req.Tier, "tier", "", "Machine tier")

	return cmd
}

func newInstancesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete INSTANCE",
		Short: "Delete a Cloud SQL instance",
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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted Cloud SQL instance %q.\n", args[0])
			return nil
		},
	}
}

func newInstancesRestartCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "restart INSTANCE",
		Short: "Restart a Cloud SQL instance",
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

			if err := client.RestartInstance(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Restarted Cloud SQL instance %q.\n", args[0])
			return nil
		},
	}
}

func newInstancesExportCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req ExportInstanceRequest

	cmd := &cobra.Command{
		Use:   "export INSTANCE",
		Short: "Export data from a Cloud SQL instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if req.URI == "" {
				return fmt.Errorf("--uri is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.ExportInstance(ctx, project, args[0], &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started export operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.URI, "uri", "", "GCS export URI")
	cmd.Flags().StringVar(&req.FileType, "file-type", "sql", "Export file type: sql, csv, bak")
	cmd.Flags().StringSliceVar(&req.Databases, "database", nil, "Database name to export")
	cmd.Flags().BoolVar(&req.Offload, "offload", false, "Use offloaded export")
	return cmd
}

func newInstancesImportCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req ImportInstanceRequest

	cmd := &cobra.Command{
		Use:   "import INSTANCE",
		Short: "Import data into a Cloud SQL instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if req.URI == "" {
				return fmt.Errorf("--uri is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.ImportInstance(ctx, project, args[0], &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started import operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.URI, "uri", "", "GCS import URI")
	cmd.Flags().StringVar(&req.FileType, "file-type", "sql", "Import file type: sql, csv, bak")
	cmd.Flags().StringVar(&req.Database, "database", "", "Database name for import")
	cmd.Flags().StringVar(&req.ImportUser, "import-user", "", "Import user name")
	return cmd
}

func newInstancesCloneCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CloneInstanceRequest

	cmd := &cobra.Command{
		Use:   "clone INSTANCE",
		Short: "Clone a Cloud SQL instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DestinationInstance == "" {
				return fmt.Errorf("--destination-instance is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CloneInstance(ctx, project, args[0], &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started clone operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.DestinationInstance, "destination-instance", "", "New instance name")
	cmd.Flags().StringVar(&req.PointInTime, "point-in-time", "", "RFC3339 point-in-time for cloning")
	return cmd
}

func newInstancesPromoteReplicaCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var failover bool

	cmd := &cobra.Command{
		Use:   "promote-replica INSTANCE",
		Short: "Promote a Cloud SQL read replica",
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
			opName, err := client.PromoteReplica(ctx, project, args[0], failover)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started promote operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().BoolVar(&failover, "failover", false, "Perform replica promotion as failover")
	return cmd
}

func newInstancesFailoverCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var settingsVersion int64

	cmd := &cobra.Command{
		Use:   "failover INSTANCE",
		Short: "Fail over a Cloud SQL replica",
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

			if settingsVersion == 0 {
				inst, err := client.GetInstance(ctx, project, args[0])
				if err != nil {
					return err
				}
				settingsVersion = inst.SettingsVersion
			}

			opName, err := client.FailoverInstance(ctx, project, args[0], settingsVersion)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started failover operation %s.\n", opName)
			return nil
		},
	}

	cmd.Flags().Int64Var(&settingsVersion, "settings-version", 0, "Current settings version")

	return cmd
}

// --- databases subcommands ---

func newDatabasesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "databases",
		Short: "Manage Cloud SQL databases",
	}

	cmd.AddCommand(
		newDatabasesListCommand(cfg, creds),
		newDatabasesDescribeCommand(cfg, creds),
		newDatabasesCreateCommand(cfg, creds),
		newDatabasesDeleteCommand(cfg, creds),
	)

	return cmd
}

func newDatabasesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List databases in a Cloud SQL instance",
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

			headers := []string{"NAME", "CHARSET", "COLLATION"}
			rows := make([][]string, len(databases))
			for i, d := range databases {
				rows[i] = []string{d.Name, d.Charset, d.Collation}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newDatabasesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe DATABASE",
		Short: "Describe a database in a Cloud SQL instance",
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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:      %s\n", db.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Charset:   %s\n", db.Charset)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Collation: %s\n", db.Collation)
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newDatabasesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create DATABASE",
		Short: "Create a database in a Cloud SQL instance",
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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created database %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newDatabasesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete DATABASE",
		Short: "Delete a database from a Cloud SQL instance",
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

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted database %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

// --- users subcommands ---

func newUsersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "users",
		Short: "Manage Cloud SQL users",
	}

	cmd.AddCommand(
		newUsersListCommand(cfg, creds),
		newUsersCreateCommand(cfg, creds),
		newUsersDeleteCommand(cfg, creds),
		newUsersSetPasswordCommand(cfg, creds),
	)

	return cmd
}

func newUsersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List users in a Cloud SQL instance",
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

			users, err := client.ListUsers(ctx, project, instance)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), users)
			}

			headers := []string{"NAME", "HOST"}
			rows := make([][]string, len(users))
			for i, u := range users {
				rows[i] = []string{u.Name, u.Host}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newUsersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "create USER",
		Short: "Create a user in a Cloud SQL instance",
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

			if err := client.CreateUser(ctx, project, instance, args[0], password); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created user %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")
	cmd.Flags().StringVar(&password, "password", "", "User password")

	return cmd
}

func newUsersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete USER",
		Short: "Delete a user from a Cloud SQL instance",
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

			if err := client.DeleteUser(ctx, project, instance, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted user %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newUsersSetPasswordCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var password string

	cmd := &cobra.Command{
		Use:   "set-password USER",
		Short: "Set password for a Cloud SQL user",
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

			if err := client.SetPassword(ctx, project, instance, args[0], password); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated password for user %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")
	cmd.Flags().StringVar(&password, "password", "", "New password")

	return cmd
}

// --- backups subcommands ---

func newBackupsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backups",
		Short: "Manage Cloud SQL backups",
	}

	cmd.AddCommand(
		newBackupsListCommand(cfg, creds),
		newBackupsDescribeCommand(cfg, creds),
		newBackupsCreateCommand(cfg, creds),
		newBackupsDeleteCommand(cfg, creds),
		newBackupsRestoreCommand(cfg, creds),
	)

	return cmd
}

func newBackupsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List backups for a Cloud SQL instance",
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

			backups, err := client.ListBackups(ctx, project, instance)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), backups)
			}

			headers := []string{"ID", "STATUS", "TYPE", "ENQUEUED_AT"}
			rows := make([][]string, len(backups))
			for i, b := range backups {
				rows[i] = []string{b.ID, b.Status, b.Type, b.EnqueuedAt}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newBackupsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe BACKUP_ID",
		Short: "Describe a backup",
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

			backup, err := client.GetBackup(ctx, project, instance, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), backup)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ID:         %s\n", backup.ID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Status:     %s\n", backup.Status)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:       %s\n", backup.Type)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Enqueued:   %s\n", backup.EnqueuedAt)
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newBackupsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a backup for a Cloud SQL instance",
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

			if err := client.CreateBackup(ctx, project, instance); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created backup for instance %q.\n", instance)
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newBackupsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete BACKUP_ID",
		Short: "Delete a backup",
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

			if err := client.DeleteBackup(ctx, project, instance, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted backup %q.\n", args[0])
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}

func newBackupsRestoreCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "restore BACKUP_ID",
		Short: "Restore a backup to its Cloud SQL instance",
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

			if err := client.RestoreBackup(ctx, project, instance, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Restored backup %q to instance %q.\n", args[0], instance)
			return nil
		},
	}

	cmd.Flags().String("instance", "", "Cloud SQL instance name")

	return cmd
}
