package firestore

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	firestoreapi "google.golang.org/api/firestore/v1"
	"github.com/spf13/cobra"
)

// NewCommand returns the Firestore command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firestore",
		Short: "Manage Firestore admin resources",
	}
	cmd.AddCommand(
		newListCommand(cfg, creds),
		newDescribeCommand(cfg, creds),
		newCreateCommand(cfg, creds),
		newDeleteCommand(cfg, creds),
		newExportCommand(cfg, creds),
		newImportCommand(cfg, creds),
		newOperationsCommand(cfg, creds),
		newIndexesCommand(cfg, creds),
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

func firestoreClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List Firestore databases",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			databases, err := client.ListDatabases(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), databases)
			}
			headers := []string{"NAME", "LOCATION", "TYPE", "EDITION", "DELETE_PROTECTION"}
			rows := make([][]string, len(databases))
			for i, database := range databases {
				rows[i] = []string{
					database.Name,
					database.LocationID,
					database.Type,
					database.DatabaseEdition,
					database.DeleteProtectionState,
				}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe DATABASE",
		Short: "Describe a Firestore database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			database, err := client.GetDatabase(ctx, project, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), database)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:             %s\n", database.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Location:         %s\n", database.LocationID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Type:             %s\n", database.Type)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "ConcurrencyMode:  %s\n", database.ConcurrencyMode)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DatabaseEdition:  %s\n", database.DatabaseEdition)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "DeleteProtection: %s\n", database.DeleteProtectionState)
			return nil
		},
	}
}

func newCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string
	var databaseType string

	cmd := &cobra.Command{
		Use:   "create DATABASE",
		Short: "Create a Firestore database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if location == "" {
				return fmt.Errorf("--location is required")
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CreateDatabase(ctx, project, args[0], location, databaseType)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started create operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&location, "location", "", "Database location")
	cmd.Flags().StringVar(&databaseType, "type", "firestore-native", "Database type: firestore-native or datastore")
	return cmd
}

func newDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var etag string

	cmd := &cobra.Command{
		Use:   "delete DATABASE",
		Short: "Delete a Firestore database",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.DeleteDatabase(ctx, project, args[0], etag)
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started delete operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&etag, "etag", "", "Database etag precondition")
	return cmd
}

func newExportCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var outputURI string
	var collectionIDs []string
	var namespaceIDs []string
	var snapshotTime string

	cmd := &cobra.Command{
		Use:   "export DATABASE --output-uri=gs://BUCKET/PREFIX",
		Short: "Export Firestore documents",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if outputURI == "" {
				return fmt.Errorf("--output-uri is required")
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.ExportDocuments(ctx, project, args[0], &ExportRequest{
				OutputURI:     outputURI,
				CollectionIDs: collectionIDs,
				NamespaceIDs:  namespaceIDs,
				SnapshotTime:  snapshotTime,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started export operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&outputURI, "output-uri", "", "GCS destination prefix")
	cmd.Flags().StringArrayVar(&collectionIDs, "collection", nil, "Collection group ID")
	cmd.Flags().StringArrayVar(&namespaceIDs, "namespace", nil, "Namespace ID")
	cmd.Flags().StringVar(&snapshotTime, "snapshot-time", "", "RFC3339 snapshot time")
	return cmd
}

func newImportCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var inputURI string
	var collectionIDs []string
	var namespaceIDs []string

	cmd := &cobra.Command{
		Use:   "import DATABASE --input-uri=gs://BUCKET/PREFIX",
		Short: "Import Firestore documents",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if inputURI == "" {
				return fmt.Errorf("--input-uri is required")
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.ImportDocuments(ctx, project, args[0], &ImportRequest{
				InputURI:      inputURI,
				CollectionIDs: collectionIDs,
				NamespaceIDs:  namespaceIDs,
			})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started import operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&inputURI, "input-uri", "", "GCS source prefix")
	cmd.Flags().StringArrayVar(&collectionIDs, "collection", nil, "Collection group ID")
	cmd.Flags().StringArrayVar(&namespaceIDs, "namespace", nil, "Namespace ID")
	return cmd
}

func newOperationsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "operations",
		Short: "Manage Firestore admin operations",
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
		Short: "List Firestore admin operations",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
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
		Short: "Describe a Firestore admin operation",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			_, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
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

func newIndexesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "indexes",
		Short: "Manage Firestore composite indexes",
	}
	cmd.AddCommand(
		newIndexesListCommand(cfg, creds),
		newIndexesDescribeCommand(creds),
		newIndexesCreateCommand(cfg, creds),
		newIndexesDeleteCommand(creds),
	)
	return cmd
}

func newIndexesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var database string
	var collectionGroup string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Firestore composite indexes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if database == "" {
				database = "(default)"
			}
			if collectionGroup == "" {
				return fmt.Errorf("--collection-group is required")
			}
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			indexes, err := client.ListIndexes(ctx, project, database, collectionGroup)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), indexes)
			}
			headers := []string{"NAME", "QUERY_SCOPE", "STATE", "FIELDS"}
			rows := make([][]string, len(indexes))
			for i, index := range indexes {
				rows[i] = []string{index.Name, index.QueryScope, index.State, fmt.Sprintf("%d", index.FieldCount)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().StringVar(&database, "database", "(default)", "Firestore database ID")
	cmd.Flags().StringVar(&collectionGroup, "collection-group", "", "Collection group ID")
	_ = cmd.MarkFlagRequired("collection-group")
	return cmd
}

func newIndexesDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe INDEX_NAME",
		Short: "Describe a Firestore index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			index, err := client.GetIndex(ctx, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), index)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:        %s\n", index.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "QueryScope:  %s\n", index.QueryScope)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:       %s\n", index.State)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Fields:      %d\n", index.FieldCount)
			return nil
		},
	}
}

func newIndexesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var database string
	var collectionGroup string
	var indexJSON string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a Firestore index",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if database == "" {
				database = "(default)"
			}
			if collectionGroup == "" {
				return fmt.Errorf("--collection-group is required")
			}
			if indexJSON == "" {
				return fmt.Errorf("--index-json is required")
			}

			var index firestoreapi.GoogleFirestoreAdminV1Index
			if err := json.Unmarshal([]byte(indexJSON), &index); err != nil {
				return fmt.Errorf("parse index json: %w", err)
			}

			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			opName, err := client.CreateIndex(ctx, project, database, collectionGroup, &CreateIndexRequest{Index: &index})
			if err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Started create operation %s.\n", opName)
			return nil
		},
	}
	cmd.Flags().StringVar(&database, "database", "(default)", "Firestore database ID")
	cmd.Flags().StringVar(&collectionGroup, "collection-group", "", "Collection group ID")
	cmd.Flags().StringVar(&indexJSON, "index-json", "", "JSON representation of a Firestore index")
	_ = cmd.MarkFlagRequired("collection-group")
	_ = cmd.MarkFlagRequired("index-json")
	return cmd
}

func newIndexesDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete INDEX_NAME",
		Short: "Delete a Firestore index",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := firestoreClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteIndex(ctx, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted index %s.\n", args[0])
			return nil
		},
	}
}
