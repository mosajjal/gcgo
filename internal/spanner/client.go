package spanner

import (
	"context"
	"errors"
	"fmt"
	"time"

	spanner "cloud.google.com/go/spanner"
	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	spannerapi "google.golang.org/api/spanner/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Instance holds Spanner instance fields.
type Instance struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Config      string `json:"config"`
	NodeCount   int32  `json:"node_count"`
	State       string `json:"state"`
}

// Database holds Spanner database fields.
type Database struct {
	Name  string `json:"name"`
	State string `json:"state"`
}

// QueryResult holds the output of an execute-sql call.
type QueryResult struct {
	Columns []string   `json:"columns"`
	Rows    [][]string `json:"rows"`
}

// Backup holds Spanner backup fields.
type Backup struct {
	Name       string `json:"name"`
	Instance   string `json:"instance"`
	Database   string `json:"database"`
	State      string `json:"state"`
	SizeBytes  int64  `json:"size_bytes"`
	CreateTime string `json:"create_time,omitempty"`
	ExpireTime string `json:"expire_time,omitempty"`
}

// CreateBackupRequest holds parameters for creating a Spanner backup.
type CreateBackupRequest struct {
	BackupID   string
	Database   string
	ExpireTime time.Time
}

// Operation holds a Spanner long-running operation.
type Operation struct {
	Name  string `json:"name"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// Client defines Spanner operations.
type Client interface {
	ListInstances(ctx context.Context, project string) ([]*Instance, error)
	GetInstance(ctx context.Context, project, name string) (*Instance, error)
	CreateInstance(ctx context.Context, project string, req *CreateInstanceRequest) error
	DeleteInstance(ctx context.Context, project, name string) error

	ListDatabases(ctx context.Context, project, instance string) ([]*Database, error)
	GetDatabase(ctx context.Context, project, instance, name string) (*Database, error)
	CreateDatabase(ctx context.Context, project, instance, name string) error
	DeleteDatabase(ctx context.Context, project, instance, name string) error
	ExecuteSQL(ctx context.Context, project, instance, db, sql string) (*QueryResult, error)

	ListBackups(ctx context.Context, project, instance string) ([]*Backup, error)
	GetBackup(ctx context.Context, project, instance, backup string) (*Backup, error)
	CreateBackup(ctx context.Context, project, instance string, req *CreateBackupRequest) error
	DeleteBackup(ctx context.Context, project, instance, backup string) error

	ListOperations(ctx context.Context, project, instance string) ([]*Operation, error)
	GetOperation(ctx context.Context, operationName string) (*Operation, error)
}

// CreateInstanceRequest holds parameters for creating a Spanner instance.
type CreateInstanceRequest struct {
	Name        string
	DisplayName string
	Config      string
	NodeCount   int32
}

type gcpClient struct {
	instances  *instance.InstanceAdminClient
	databases  *database.DatabaseAdminClient
	clientOpts []option.ClientOption
}

// NewClient creates a Client backed by the real Spanner Admin APIs.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	ic, err := instance.NewInstanceAdminClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create spanner instance admin client: %w", err)
	}

	dc, err := database.NewDatabaseAdminClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create spanner database admin client: %w", err)
	}

	return &gcpClient{instances: ic, databases: dc, clientOpts: opts}, nil
}

func (c *gcpClient) ListInstances(ctx context.Context, project string) ([]*Instance, error) {
	parent := fmt.Sprintf("projects/%s", project)
	it := c.instances.ListInstances(ctx, &instancepb.ListInstancesRequest{
		Parent: parent,
	})

	var out []*Instance
	for {
		i, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list spanner instances: %w", err)
		}
		out = append(out, instanceFromProto(i))
	}
	return out, nil
}

func (c *gcpClient) GetInstance(ctx context.Context, project, name string) (*Instance, error) {
	fullName := fmt.Sprintf("projects/%s/instances/%s", project, name)
	i, err := c.instances.GetInstance(ctx, &instancepb.GetInstanceRequest{
		Name: fullName,
	})
	if err != nil {
		return nil, fmt.Errorf("get spanner instance %s: %w", name, err)
	}
	return instanceFromProto(i), nil
}

func (c *gcpClient) CreateInstance(ctx context.Context, project string, req *CreateInstanceRequest) error {
	parent := fmt.Sprintf("projects/%s", project)
	cfg := req.Config
	if cfg == "" {
		cfg = fmt.Sprintf("projects/%s/instanceConfigs/regional-us-central1", project)
	}
	nodeCount := req.NodeCount
	if nodeCount == 0 {
		nodeCount = 1
	}
	displayName := req.DisplayName
	if displayName == "" {
		displayName = req.Name
	}

	op, err := c.instances.CreateInstance(ctx, &instancepb.CreateInstanceRequest{
		Parent:     parent,
		InstanceId: req.Name,
		Instance: &instancepb.Instance{
			Name:        fmt.Sprintf("%s/instances/%s", parent, req.Name),
			Config:      cfg,
			DisplayName: displayName,
			NodeCount:   nodeCount,
		},
	})
	if err != nil {
		return fmt.Errorf("create spanner instance %s: %w", req.Name, err)
	}
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for spanner instance %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpClient) DeleteInstance(ctx context.Context, project, name string) error {
	fullName := fmt.Sprintf("projects/%s/instances/%s", project, name)
	err := c.instances.DeleteInstance(ctx, &instancepb.DeleteInstanceRequest{
		Name: fullName,
	})
	if err != nil {
		return fmt.Errorf("delete spanner instance %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListDatabases(ctx context.Context, project, inst string) ([]*Database, error) {
	parent := fmt.Sprintf("projects/%s/instances/%s", project, inst)
	it := c.databases.ListDatabases(ctx, &databasepb.ListDatabasesRequest{
		Parent: parent,
	})

	var out []*Database
	for {
		d, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list spanner databases: %w", err)
		}
		out = append(out, databaseFromProto(d))
	}
	return out, nil
}

func (c *gcpClient) GetDatabase(ctx context.Context, project, inst, name string) (*Database, error) {
	fullName := fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, inst, name)
	d, err := c.databases.GetDatabase(ctx, &databasepb.GetDatabaseRequest{
		Name: fullName,
	})
	if err != nil {
		return nil, fmt.Errorf("get spanner database %s: %w", name, err)
	}
	return databaseFromProto(d), nil
}

func (c *gcpClient) CreateDatabase(ctx context.Context, project, inst, name string) error {
	parent := fmt.Sprintf("projects/%s/instances/%s", project, inst)
	op, err := c.databases.CreateDatabase(ctx, &databasepb.CreateDatabaseRequest{
		Parent:          parent,
		CreateStatement: fmt.Sprintf("CREATE DATABASE `%s`", name),
	})
	if err != nil {
		return fmt.Errorf("create spanner database %s: %w", name, err)
	}
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for spanner database %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) DeleteDatabase(ctx context.Context, project, inst, name string) error {
	fullName := fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, inst, name)
	err := c.databases.DropDatabase(ctx, &databasepb.DropDatabaseRequest{
		Database: fullName,
	})
	if err != nil {
		return fmt.Errorf("delete spanner database %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ExecuteSQL(ctx context.Context, project, inst, db, sql string) (*QueryResult, error) {
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, inst, db)

	dataClient, err := spanner.NewClient(ctx, dbPath, c.clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("create spanner data client: %w", err)
	}
	defer dataClient.Close()

	stmt := spanner.Statement{SQL: sql}
	iter := dataClient.Single().Query(ctx, stmt)
	defer iter.Stop()

	var result QueryResult
	for {
		row, err := iter.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("execute sql: %w", err)
		}
		if result.Columns == nil {
			result.Columns = row.ColumnNames()
		}
		rowData := make([]string, row.Size())
		for i := 0; i < row.Size(); i++ {
			var v spanner.GenericColumnValue
			if err := row.Column(i, &v); err != nil {
				rowData[i] = ""
				continue
			}
			rowData[i] = fmt.Sprintf("%v", v.Value.AsInterface())
		}
		result.Rows = append(result.Rows, rowData)
	}
	return &result, nil
}

func (c *gcpClient) ListBackups(ctx context.Context, project, inst string) ([]*Backup, error) {
	parent := fmt.Sprintf("projects/%s/instances/%s", project, inst)
	it := c.databases.ListBackups(ctx, &databasepb.ListBackupsRequest{
		Parent: parent,
	})

	var out []*Backup
	for {
		b, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list spanner backups: %w", err)
		}
		out = append(out, backupFromProto(b, inst))
	}
	return out, nil
}

func (c *gcpClient) GetBackup(ctx context.Context, project, inst, backup string) (*Backup, error) {
	fullName := fmt.Sprintf("projects/%s/instances/%s/backups/%s", project, inst, backup)
	b, err := c.databases.GetBackup(ctx, &databasepb.GetBackupRequest{
		Name: fullName,
	})
	if err != nil {
		return nil, fmt.Errorf("get spanner backup %s: %w", backup, err)
	}
	return backupFromProto(b, inst), nil
}

func (c *gcpClient) CreateBackup(ctx context.Context, project, inst string, req *CreateBackupRequest) error {
	parent := fmt.Sprintf("projects/%s/instances/%s", project, inst)
	dbPath := fmt.Sprintf("projects/%s/instances/%s/databases/%s", project, inst, req.Database)
	op, err := c.databases.CreateBackup(ctx, &databasepb.CreateBackupRequest{
		Parent:   parent,
		BackupId: req.BackupID,
		Backup: &databasepb.Backup{
			Database:   dbPath,
			ExpireTime: timestamppb.New(req.ExpireTime),
		},
	})
	if err != nil {
		return fmt.Errorf("create spanner backup %s: %w", req.BackupID, err)
	}
	_, err = op.Wait(ctx)
	if err != nil {
		return fmt.Errorf("wait for spanner backup %s: %w", req.BackupID, err)
	}
	return nil
}

func (c *gcpClient) DeleteBackup(ctx context.Context, project, inst, backup string) error {
	fullName := fmt.Sprintf("projects/%s/instances/%s/backups/%s", project, inst, backup)
	err := c.databases.DeleteBackup(ctx, &databasepb.DeleteBackupRequest{
		Name: fullName,
	})
	if err != nil {
		return fmt.Errorf("delete spanner backup %s: %w", backup, err)
	}
	return nil
}

func (c *gcpClient) ListOperations(ctx context.Context, project, inst string) ([]*Operation, error) {
	svc, err := spannerapi.NewService(ctx, c.clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("create spanner api client: %w", err)
	}
	parent := fmt.Sprintf("projects/%s/instances/%s", project, inst)
	resp, err := svc.Projects.Instances.Operations.List(parent).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list spanner operations: %w", err)
	}
	out := make([]*Operation, 0, len(resp.Operations))
	for _, o := range resp.Operations {
		out = append(out, operationFromREST(o))
	}
	return out, nil
}

func (c *gcpClient) GetOperation(ctx context.Context, operationName string) (*Operation, error) {
	svc, err := spannerapi.NewService(ctx, c.clientOpts...)
	if err != nil {
		return nil, fmt.Errorf("create spanner api client: %w", err)
	}
	o, err := svc.Projects.Instances.Operations.Get(operationName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get spanner operation %s: %w", operationName, err)
	}
	return operationFromREST(o), nil
}

func instanceFromProto(i *instancepb.Instance) *Instance {
	return &Instance{
		Name:        i.GetName(),
		DisplayName: i.GetDisplayName(),
		Config:      i.GetConfig(),
		NodeCount:   i.GetNodeCount(),
		State:       i.GetState().String(),
	}
}

func databaseFromProto(d *databasepb.Database) *Database {
	return &Database{
		Name:  d.GetName(),
		State: d.GetState().String(),
	}
}

func backupFromProto(b *databasepb.Backup, inst string) *Backup {
	createTime := ""
	if t := b.GetCreateTime(); t != nil {
		createTime = t.AsTime().Format(time.RFC3339)
	}
	expireTime := ""
	if t := b.GetExpireTime(); t != nil {
		expireTime = t.AsTime().Format(time.RFC3339)
	}
	return &Backup{
		Name:       b.GetName(),
		Instance:   inst,
		Database:   b.GetDatabase(),
		State:      b.GetState().String(),
		SizeBytes:  b.GetSizeBytes(),
		CreateTime: createTime,
		ExpireTime: expireTime,
	}
}

func operationFromREST(o *spannerapi.Operation) *Operation {
	op := &Operation{
		Name: o.Name,
		Done: o.Done,
	}
	if o.Error != nil {
		op.Error = o.Error.Message
	}
	return op
}
