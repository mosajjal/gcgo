package spanner

import (
	"context"
	"errors"
	"fmt"

	database "cloud.google.com/go/spanner/admin/database/apiv1"
	"cloud.google.com/go/spanner/admin/database/apiv1/databasepb"
	instance "cloud.google.com/go/spanner/admin/instance/apiv1"
	"cloud.google.com/go/spanner/admin/instance/apiv1/instancepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
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
	Rows    [][]string  `json:"rows"`
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
}

// CreateInstanceRequest holds parameters for creating a Spanner instance.
type CreateInstanceRequest struct {
	Name        string
	DisplayName string
	Config      string
	NodeCount   int32
}

type gcpClient struct {
	instances *instance.InstanceAdminClient
	databases *database.DatabaseAdminClient
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

	return &gcpClient{instances: ic, databases: dc}, nil
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

func (c *gcpClient) ExecuteSQL(_ context.Context, _, _, _, _ string) (*QueryResult, error) {
	// Spanner admin client doesn't directly execute SQL queries.
	// A full implementation would use the spanner data client.
	// This is a placeholder that returns an error directing users
	// to use the spanner data client for queries.
	return nil, fmt.Errorf("execute-sql requires a spanner data client (not yet implemented)")
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
