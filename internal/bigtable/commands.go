package bigtable

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the Bigtable command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "bigtable",
		Short: "Manage Cloud Bigtable resources",
	}
	cmd.AddCommand(
		newInstancesCommand(cfg, creds),
		newTablesCommand(cfg, creds),
		newOperationsCommand(cfg, creds),
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

func bigtableClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newInstancesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{Use: "instances", Short: "Manage Bigtable instances"}
	cmd.AddCommand(
		newInstancesListCommand(cfg, creds),
		newInstancesDescribeCommand(cfg, creds),
		newInstancesCreateCommand(cfg, creds),
		newInstancesDeleteCommand(cfg, creds),
		newBackupsCommand(cfg, creds),
	)
	return cmd
}

func newInstancesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List instances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
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
			headers := []string{"NAME", "DISPLAY_NAME", "STATE", "TYPE", "EDITION"}
			rows := make([][]string, len(instances))
			for i, instance := range instances {
				rows[i] = []string{instance.Name, instance.DisplayName, instance.State, instance.Type, instance.Edition}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newInstancesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe an instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			instance, err := client.GetInstance(ctx, project, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), instance)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", instance.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DisplayName: %s\n", instance.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", instance.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", instance.Type)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Edition:     %s\n", instance.Edition)
			return nil
		},
	}
}

func newInstancesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateInstanceRequest

	cmd := &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a Bigtable instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if req.DisplayName == "" {
				req.DisplayName = args[0]
			}
			if req.Zone == "" {
				return fmt.Errorf("--zone is required")
			}
			req.InstanceID = args[0]

			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CreateInstance(ctx, project, &req)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started create operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Instance display name")
	cmd.Flags().StringVar(&req.ClusterID, "cluster", "", "Cluster ID to create")
	cmd.Flags().StringVar(&req.Zone, "zone", "", "Cluster zone, for example us-central1-b")
	cmd.Flags().Int64Var(&req.ServeNodes, "serve-nodes", 1, "Cluster serve nodes")
	cmd.Flags().StringVar(&req.Type, "type", "production", "Instance type: production or development")
	cmd.Flags().StringVar(&req.Edition, "edition", "enterprise", "Instance edition: enterprise or enterprise-plus")
	cmd.Flags().StringVar(&req.StorageType, "storage-type", "ssd", "Storage type: ssd or hdd")
	return cmd
}

func newInstancesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete INSTANCE",
		Short: "Delete a Bigtable instance",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteInstance(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted instance %s.\n", args[0])
			return nil
		},
	}
}

func newTablesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{Use: "tables", Short: "Manage Bigtable tables"}
	cmd.AddCommand(
		newTablesListCommand(cfg, creds),
		newTablesDescribeCommand(cfg, creds),
		newTablesCreateCommand(cfg, creds),
		newTablesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newTablesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var instance string
	cmd := &cobra.Command{
		Use:   "list --instance=INSTANCE",
		Short: "List tables in an instance",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			tables, err := client.ListTables(ctx, project, instance)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), tables)
			}
			headers := []string{"NAME", "GRANULARITY", "COLUMN_FAMILIES", "PROTECTED"}
			rows := make([][]string, len(tables))
			for i, table := range tables {
				rows[i] = []string{
					table.Name,
					table.Granularity,
					fmt.Sprintf("%d", table.ColumnFamilyCount),
					fmt.Sprintf("%t", table.DeletionProtection),
				}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "Bigtable instance name")
	return cmd
}

func newTablesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var instance string
	cmd := &cobra.Command{
		Use:   "describe TABLE --instance=INSTANCE",
		Short: "Describe a table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			table, err := client.GetTable(ctx, project, instance, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), table)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:               %s\n", table.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Granularity:        %s\n", table.Granularity)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DeletionProtection: %t\n", table.DeletionProtection)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ColumnFamilies:     %d\n", table.ColumnFamilyCount)
			return nil
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "Bigtable instance name")
	return cmd
}

func newTablesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var instance string
	cmd := &cobra.Command{
		Use:   "create TABLE --instance=INSTANCE",
		Short: "Create a table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			table, err := client.CreateTable(ctx, project, instance, args[0])
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created table %s.\n", table.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "Bigtable instance name")
	return cmd
}

func newTablesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var instance string
	cmd := &cobra.Command{
		Use:   "delete TABLE --instance=INSTANCE",
		Short: "Delete a table",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteTable(ctx, project, instance, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted table %s.\n", args[0])
			return nil
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "Bigtable instance name")
	return cmd
}

func newOperationsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{Use: "operations", Short: "Manage Bigtable operations"}
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
		Short: "List Bigtable operations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			operations, err := client.ListOperations(ctx, project, filter)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), operations)
			}
			headers := []string{"NAME", "DONE", "ERROR"}
			rows := make([][]string, len(operations))
			for i, op := range operations {
				rows[i] = []string{op.Name, fmt.Sprintf("%t", op.Done), op.Error}
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
		Short: "Describe a Bigtable operation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
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

func newBackupsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backups",
		Short: "Manage Bigtable backups",
	}
	cmd.AddCommand(
		newBackupsListCommand(cfg, creds),
		newBackupsDescribeCommand(cfg, creds),
		newBackupsCreateCommand(cfg, creds),
		newBackupsDeleteCommand(creds),
	)
	return cmd
}

func newBackupsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var instance string
	var cluster string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Bigtable backups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			backups, err := client.ListBackups(ctx, project, instance, cluster)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), backups)
			}
			headers := []string{"NAME", "SOURCE_TABLE", "STATE", "EXPIRE_TIME", "TYPE"}
			rows := make([][]string, len(backups))
			for i, backup := range backups {
				rows[i] = []string{backup.Name, backup.SourceTable, backup.State, backup.ExpireTime, backup.BackupType}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "Bigtable instance name")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Bigtable cluster name")
	return cmd
}

func newBackupsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe BACKUP",
		Short: "Describe a Bigtable backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			backup, err := client.GetBackup(ctx, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), backup)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", backup.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "SourceTable: %s\n", backup.SourceTable)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", backup.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ExpireTime:  %s\n", backup.ExpireTime)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:        %s\n", backup.BackupType)
			return nil
		},
	}
}

func newBackupsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var instance string
	var cluster string
	var sourceTable string
	var expireTime string
	var backupType string

	cmd := &cobra.Command{
		Use:   "create BACKUP",
		Short: "Create a Bigtable backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if instance == "" {
				return fmt.Errorf("--instance is required")
			}
			if cluster == "" {
				return fmt.Errorf("--cluster is required")
			}
			if sourceTable == "" {
				return fmt.Errorf("--source-table is required")
			}
			if expireTime == "" {
				return fmt.Errorf("--expire-time is required")
			}
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CreateBackup(ctx, project, instance, cluster, &CreateBackupRequest{
				BackupID:    args[0],
				SourceTable: sourceTable,
				ExpireTime:  expireTime,
				BackupType:  backupType,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started create operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&instance, "instance", "", "Bigtable instance name")
	cmd.Flags().StringVar(&cluster, "cluster", "", "Bigtable cluster name")
	cmd.Flags().StringVar(&sourceTable, "source-table", "", "Source table resource name")
	cmd.Flags().StringVar(&expireTime, "expire-time", "", "RFC3339 expiration time")
	cmd.Flags().StringVar(&backupType, "backup-type", "standard", "Backup type: standard or hot")
	return cmd
}

func newBackupsDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete BACKUP",
		Short: "Delete a Bigtable backup",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := bigtableClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteBackup(ctx, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted backup %s.\n", args[0])
			return nil
		},
	}
}
