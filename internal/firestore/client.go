package firestore

import (
	"context"
	"fmt"

	firestoreapi "google.golang.org/api/firestore/v1"
	"google.golang.org/api/option"
)

// Database holds the fields we display.
type Database struct {
	Name                  string `json:"name"`
	LocationID            string `json:"location_id"`
	Type                  string `json:"type"`
	ConcurrencyMode       string `json:"concurrency_mode"`
	DatabaseEdition       string `json:"database_edition"`
	DeleteProtectionState string `json:"delete_protection_state"`
	CreateTime            string `json:"create_time"`
}

// Operation holds Firestore admin operation fields.
type Operation struct {
	Name  string `json:"name"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// Index holds Firestore index fields for display.
type Index struct {
	Name       string `json:"name"`
	QueryScope string `json:"query_scope"`
	State      string `json:"state"`
	FieldCount int    `json:"field_count"`
}

// CreateIndexRequest holds Firestore index creation parameters.
type CreateIndexRequest struct {
	Index *firestoreapi.GoogleFirestoreAdminV1Index
}

// ExportRequest holds database export parameters.
type ExportRequest struct {
	OutputURI     string
	CollectionIDs []string
	NamespaceIDs  []string
	SnapshotTime  string
}

// ImportRequest holds database import parameters.
type ImportRequest struct {
	InputURI      string
	CollectionIDs []string
	NamespaceIDs  []string
}

// Client defines Firestore admin operations.
type Client interface {
	ListDatabases(ctx context.Context, project string) ([]*Database, error)
	GetDatabase(ctx context.Context, project, name string) (*Database, error)
	CreateDatabase(ctx context.Context, project, name, location, databaseType string) (string, error)
	DeleteDatabase(ctx context.Context, project, name, etag string) (string, error)
	ExportDocuments(ctx context.Context, project, name string, req *ExportRequest) (string, error)
	ImportDocuments(ctx context.Context, project, name string, req *ImportRequest) (string, error)
	ListOperations(ctx context.Context, project, filter string) ([]*Operation, error)
	GetOperation(ctx context.Context, name string) (*Operation, error)
	ListIndexes(ctx context.Context, project, database, collectionGroup string) ([]*Index, error)
	GetIndex(ctx context.Context, name string) (*Index, error)
	CreateIndex(ctx context.Context, project, database, collectionGroup string, req *CreateIndexRequest) (string, error)
	DeleteIndex(ctx context.Context, name string) error
}

type gcpClient struct {
	svc *firestoreapi.Service
}

// NewClient creates a Client backed by the real Firestore Admin API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := firestoreapi.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create firestore client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListDatabases(ctx context.Context, project string) ([]*Database, error) {
	parent := fmt.Sprintf("projects/%s", project)
	resp, err := c.svc.Projects.Databases.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	var databases []*Database
	for _, database := range resp.Databases {
		databases = append(databases, databaseFromAPI(database))
	}
	return databases, nil
}

func (c *gcpClient) GetDatabase(ctx context.Context, project, name string) (*Database, error) {
	fullName := databaseName(project, name)
	database, err := c.svc.Projects.Databases.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get database %s: %w", name, err)
	}
	return databaseFromAPI(database), nil
}

func (c *gcpClient) CreateDatabase(ctx context.Context, project, name, location, databaseType string) (string, error) {
	op, err := c.svc.Projects.Databases.Create(fmt.Sprintf("projects/%s", project), &firestoreapi.GoogleFirestoreAdminV1Database{
		LocationId: location,
		Type:       firestoreDatabaseType(databaseType),
	}).DatabaseId(name).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create database %s: %w", name, err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeleteDatabase(ctx context.Context, project, name, etag string) (string, error) {
	call := c.svc.Projects.Databases.Delete(databaseName(project, name)).Context(ctx)
	if etag != "" {
		call = call.Etag(etag)
	}
	op, err := call.Do()
	if err != nil {
		return "", fmt.Errorf("delete database %s: %w", name, err)
	}
	return op.Name, nil
}

func (c *gcpClient) ExportDocuments(ctx context.Context, project, name string, req *ExportRequest) (string, error) {
	fullName := databaseName(project, name)
	op, err := c.svc.Projects.Databases.ExportDocuments(fullName, &firestoreapi.GoogleFirestoreAdminV1ExportDocumentsRequest{
		OutputUriPrefix: req.OutputURI,
		CollectionIds:   req.CollectionIDs,
		NamespaceIds:    req.NamespaceIDs,
		SnapshotTime:    req.SnapshotTime,
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("export database %s: %w", name, err)
	}
	return op.Name, nil
}

func (c *gcpClient) ImportDocuments(ctx context.Context, project, name string, req *ImportRequest) (string, error) {
	fullName := databaseName(project, name)
	op, err := c.svc.Projects.Databases.ImportDocuments(fullName, &firestoreapi.GoogleFirestoreAdminV1ImportDocumentsRequest{
		InputUriPrefix: req.InputURI,
		CollectionIds:  req.CollectionIDs,
		NamespaceIds:   req.NamespaceIDs,
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("import database %s: %w", name, err)
	}
	return op.Name, nil
}

func (c *gcpClient) ListOperations(ctx context.Context, project, filter string) ([]*Operation, error) {
	parent := fmt.Sprintf("projects/%s/databases", project)
	call := c.svc.Projects.Databases.Operations.List(parent).Context(ctx)
	if filter != "" {
		call = call.Filter(filter)
	}

	var operations []*Operation
	if err := call.Pages(ctx, func(resp *firestoreapi.GoogleLongrunningListOperationsResponse) error {
		for _, op := range resp.Operations {
			operations = append(operations, operationFromAPI(op))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list firestore operations: %w", err)
	}
	return operations, nil
}

func (c *gcpClient) GetOperation(ctx context.Context, name string) (*Operation, error) {
	op, err := c.svc.Projects.Databases.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get firestore operation %s: %w", name, err)
	}
	return operationFromAPI(op), nil
}

func (c *gcpClient) ListIndexes(ctx context.Context, project, database, collectionGroup string) ([]*Index, error) {
	parent := firestoreIndexParent(project, database, collectionGroup)
	call := c.svc.Projects.Databases.CollectionGroups.Indexes.List(parent).Context(ctx)

	var indexes []*Index
	if err := call.Pages(ctx, func(resp *firestoreapi.GoogleFirestoreAdminV1ListIndexesResponse) error {
		for _, idx := range resp.Indexes {
			indexes = append(indexes, indexFromAPI(idx))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list firestore indexes: %w", err)
	}
	return indexes, nil
}

func (c *gcpClient) GetIndex(ctx context.Context, name string) (*Index, error) {
	idx, err := c.svc.Projects.Databases.CollectionGroups.Indexes.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get firestore index %s: %w", name, err)
	}
	return indexFromAPI(idx), nil
}

func (c *gcpClient) CreateIndex(ctx context.Context, project, database, collectionGroup string, req *CreateIndexRequest) (string, error) {
	if req == nil || req.Index == nil {
		return "", fmt.Errorf("create firestore index: nil index")
	}
	op, err := c.svc.Projects.Databases.CollectionGroups.Indexes.Create(firestoreIndexParent(project, database, collectionGroup), req.Index).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create firestore index: %w", err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeleteIndex(ctx context.Context, name string) error {
	if _, err := c.svc.Projects.Databases.CollectionGroups.Indexes.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete firestore index %s: %w", name, err)
	}
	return nil
}

func databaseFromAPI(database *firestoreapi.GoogleFirestoreAdminV1Database) *Database {
	if database == nil {
		return nil
	}
	return &Database{
		Name:                  database.Name,
		LocationID:            database.LocationId,
		Type:                  database.Type,
		ConcurrencyMode:       database.ConcurrencyMode,
		DatabaseEdition:       database.DatabaseEdition,
		DeleteProtectionState: database.DeleteProtectionState,
		CreateTime:            database.CreateTime,
	}
}

func operationFromAPI(op *firestoreapi.GoogleLongrunningOperation) *Operation {
	if op == nil {
		return nil
	}
	out := &Operation{
		Name: op.Name,
		Done: op.Done,
	}
	if op.Error != nil {
		out.Error = op.Error.Message
	}
	return out
}

func indexFromAPI(idx *firestoreapi.GoogleFirestoreAdminV1Index) *Index {
	if idx == nil {
		return nil
	}
	return &Index{
		Name:       idx.Name,
		QueryScope: idx.QueryScope,
		State:      idx.State,
		FieldCount: len(idx.Fields),
	}
}

func firestoreDatabaseType(databaseType string) string {
	switch databaseType {
	case "datastore":
		return "DATASTORE_MODE"
	default:
		return "FIRESTORE_NATIVE"
	}
}

func databaseName(project, name string) string {
	return fmt.Sprintf("projects/%s/databases/%s", project, name)
}

func firestoreIndexParent(project, database, collectionGroup string) string {
	return fmt.Sprintf("projects/%s/databases/%s/collectionGroups/%s", project, database, collectionGroup)
}
