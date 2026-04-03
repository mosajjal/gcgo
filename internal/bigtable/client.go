package bigtable

import (
	"context"
	"fmt"

	bigtableadmin "google.golang.org/api/bigtableadmin/v2"
	"google.golang.org/api/option"
)

// Instance holds the fields we display.
type Instance struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	Type        string `json:"type"`
	Edition     string `json:"edition"`
	CreateTime  string `json:"create_time"`
}

// Table holds the fields we display.
type Table struct {
	Name               string `json:"name"`
	Granularity        string `json:"granularity"`
	DeletionProtection bool   `json:"deletion_protection"`
	ColumnFamilyCount  int    `json:"column_family_count"`
	AutomatedBackup    bool   `json:"automated_backup"`
}

// Operation holds Bigtable operation fields.
type Operation struct {
	Name  string `json:"name"`
	Done  bool   `json:"done"`
	Error string `json:"error,omitempty"`
}

// Backup holds Cloud Bigtable backup fields.
type Backup struct {
	Name          string `json:"name"`
	SourceTable   string `json:"source_table"`
	State         string `json:"state"`
	BackupType    string `json:"backup_type"`
	ExpireTime    string `json:"expire_time"`
	StartTime     string `json:"start_time"`
	EndTime       string `json:"end_time"`
	SizeBytes     int64  `json:"size_bytes"`
	SourceBackup  string `json:"source_backup"`
}

// CreateBackupRequest holds backup creation parameters.
type CreateBackupRequest struct {
	BackupID    string
	SourceTable string
	ExpireTime  string
	BackupType  string
}

// CreateInstanceRequest holds Bigtable instance creation parameters.
type CreateInstanceRequest struct {
	InstanceID  string
	DisplayName string
	ClusterID   string
	Zone        string
	ServeNodes  int64
	Type        string
	Edition     string
	StorageType string
}

// Client defines Bigtable admin operations.
type Client interface {
	ListInstances(ctx context.Context, project string) ([]*Instance, error)
	GetInstance(ctx context.Context, project, name string) (*Instance, error)
	CreateInstance(ctx context.Context, project string, req *CreateInstanceRequest) (string, error)
	DeleteInstance(ctx context.Context, project, name string) error
	ListTables(ctx context.Context, project, instance string) ([]*Table, error)
	GetTable(ctx context.Context, project, instance, name string) (*Table, error)
	CreateTable(ctx context.Context, project, instance, name string) (*Table, error)
	DeleteTable(ctx context.Context, project, instance, name string) error
	ListOperations(ctx context.Context, project, filter string) ([]*Operation, error)
	GetOperation(ctx context.Context, name string) (*Operation, error)
	ListBackups(ctx context.Context, project, instance, cluster string) ([]*Backup, error)
	GetBackup(ctx context.Context, name string) (*Backup, error)
	CreateBackup(ctx context.Context, project, instance, cluster string, req *CreateBackupRequest) (string, error)
	DeleteBackup(ctx context.Context, name string) error
}

type gcpClient struct {
	svc *bigtableadmin.Service
}

// NewClient creates a Client backed by the real Bigtable Admin API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := bigtableadmin.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create bigtable admin client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListInstances(ctx context.Context, project string) ([]*Instance, error) {
	parent := fmt.Sprintf("projects/%s", project)
	call := c.svc.Projects.Instances.List(parent).Context(ctx)

	var instances []*Instance
	if err := call.Pages(ctx, func(resp *bigtableadmin.ListInstancesResponse) error {
		for _, instance := range resp.Instances {
			instances = append(instances, instanceFromAPI(instance))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list instances: %w", err)
	}
	return instances, nil
}

func (c *gcpClient) GetInstance(ctx context.Context, project, name string) (*Instance, error) {
	fullName := instanceName(project, name)
	instance, err := c.svc.Projects.Instances.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get instance %s: %w", name, err)
	}
	return instanceFromAPI(instance), nil
}

func (c *gcpClient) CreateInstance(ctx context.Context, project string, req *CreateInstanceRequest) (string, error) {
	clusterID := req.ClusterID
	if clusterID == "" {
		clusterID = req.InstanceID + "-c1"
	}
	serveNodes := req.ServeNodes
	if serveNodes == 0 {
		serveNodes = 1
	}

	op, err := c.svc.Projects.Instances.Create(fmt.Sprintf("projects/%s", project), &bigtableadmin.CreateInstanceRequest{
		InstanceId: req.InstanceID,
		Instance: &bigtableadmin.Instance{
			DisplayName: req.DisplayName,
			Type:        bigtableInstanceType(req.Type),
			Edition:     bigtableEdition(req.Edition),
		},
		Clusters: map[string]bigtableadmin.Cluster{
			clusterID: {
				Location:           fmt.Sprintf("projects/%s/locations/%s", project, req.Zone),
				ServeNodes:         serveNodes,
				DefaultStorageType: bigtableStorageType(req.StorageType),
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create bigtable instance %s: %w", req.InstanceID, err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeleteInstance(ctx context.Context, project, name string) error {
	if _, err := c.svc.Projects.Instances.Delete(instanceName(project, name)).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete instance %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListTables(ctx context.Context, project, instance string) ([]*Table, error) {
	parent := instanceName(project, instance)
	call := c.svc.Projects.Instances.Tables.List(parent).Context(ctx)

	var tables []*Table
	if err := call.Pages(ctx, func(resp *bigtableadmin.ListTablesResponse) error {
		for _, table := range resp.Tables {
			tables = append(tables, tableFromAPI(table))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list tables: %w", err)
	}
	return tables, nil
}

func (c *gcpClient) GetTable(ctx context.Context, project, instance, name string) (*Table, error) {
	fullName := tableName(project, instance, name)
	table, err := c.svc.Projects.Instances.Tables.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get table %s: %w", name, err)
	}
	return tableFromAPI(table), nil
}

func (c *gcpClient) CreateTable(ctx context.Context, project, instance, name string) (*Table, error) {
	parent := instanceName(project, instance)
	table, err := c.svc.Projects.Instances.Tables.Create(parent, &bigtableadmin.CreateTableRequest{
		TableId: name,
		Table:   &bigtableadmin.Table{Granularity: "MILLIS"},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create table %s: %w", name, err)
	}
	return tableFromAPI(table), nil
}

func (c *gcpClient) DeleteTable(ctx context.Context, project, instance, name string) error {
	fullName := tableName(project, instance, name)
	if _, err := c.svc.Projects.Instances.Tables.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete table %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListOperations(ctx context.Context, project, filter string) ([]*Operation, error) {
	call := c.svc.Operations.Projects.Operations.List(fmt.Sprintf("projects/%s", project)).Context(ctx)
	if filter != "" {
		call = call.Filter(filter)
	}

	var operations []*Operation
	if err := call.Pages(ctx, func(resp *bigtableadmin.ListOperationsResponse) error {
		for _, op := range resp.Operations {
			operations = append(operations, operationFromAPI(op))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list bigtable operations: %w", err)
	}
	return operations, nil
}

func (c *gcpClient) GetOperation(ctx context.Context, name string) (*Operation, error) {
	op, err := c.svc.Operations.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get bigtable operation %s: %w", name, err)
	}
	return operationFromAPI(op), nil
}

func (c *gcpClient) ListBackups(ctx context.Context, project, instance, cluster string) ([]*Backup, error) {
	parent := fmt.Sprintf("projects/%s/instances/%s/clusters/%s", project, instance, cluster)
	call := c.svc.Projects.Instances.Clusters.Backups.List(parent).Context(ctx)

	var backups []*Backup
	if err := call.Pages(ctx, func(resp *bigtableadmin.ListBackupsResponse) error {
		for _, backup := range resp.Backups {
			backups = append(backups, backupFromAPI(backup))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list bigtable backups: %w", err)
	}
	return backups, nil
}

func (c *gcpClient) GetBackup(ctx context.Context, name string) (*Backup, error) {
	backup, err := c.svc.Projects.Instances.Clusters.Backups.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get bigtable backup %s: %w", name, err)
	}
	return backupFromAPI(backup), nil
}

func (c *gcpClient) CreateBackup(ctx context.Context, project, instance, cluster string, req *CreateBackupRequest) (string, error) {
	parent := fmt.Sprintf("projects/%s/instances/%s/clusters/%s", project, instance, cluster)
	op, err := c.svc.Projects.Instances.Clusters.Backups.Create(parent, &bigtableadmin.Backup{
		SourceTable: req.SourceTable,
		ExpireTime:  req.ExpireTime,
		BackupType:  bigtableBackupType(req.BackupType),
	}).BackupId(req.BackupID).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("create bigtable backup %s: %w", req.BackupID, err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeleteBackup(ctx context.Context, name string) error {
	if _, err := c.svc.Projects.Instances.Clusters.Backups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete bigtable backup %s: %w", name, err)
	}
	return nil
}

func instanceFromAPI(instance *bigtableadmin.Instance) *Instance {
	if instance == nil {
		return nil
	}
	return &Instance{
		Name:        instance.Name,
		DisplayName: instance.DisplayName,
		State:       instance.State,
		Type:        instance.Type,
		Edition:     instance.Edition,
		CreateTime:  instance.CreateTime,
	}
}

func tableFromAPI(table *bigtableadmin.Table) *Table {
	if table == nil {
		return nil
	}
	return &Table{
		Name:               table.Name,
		Granularity:        table.Granularity,
		DeletionProtection: table.DeletionProtection,
		ColumnFamilyCount:  len(table.ColumnFamilies),
		AutomatedBackup:    table.AutomatedBackupPolicy != nil,
	}
}

func operationFromAPI(op *bigtableadmin.Operation) *Operation {
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

func backupFromAPI(backup *bigtableadmin.Backup) *Backup {
	if backup == nil {
		return nil
	}
	return &Backup{
		Name:         backup.Name,
		SourceTable:  backup.SourceTable,
		State:        backup.State,
		BackupType:   backup.BackupType,
		ExpireTime:   backup.ExpireTime,
		StartTime:    backup.StartTime,
		EndTime:      backup.EndTime,
		SizeBytes:    backup.SizeBytes,
		SourceBackup: backup.SourceBackup,
	}
}

func bigtableInstanceType(instanceType string) string {
	switch instanceType {
	case "development":
		return "DEVELOPMENT"
	default:
		return "PRODUCTION"
	}
}

func bigtableEdition(edition string) string {
	switch edition {
	case "enterprise-plus":
		return "ENTERPRISE_PLUS"
	default:
		return "ENTERPRISE"
	}
}

func bigtableBackupType(backupType string) string {
	switch backupType {
	case "hot":
		return "HOT"
	default:
		return "STANDARD"
	}
}

func bigtableStorageType(storageType string) string {
	switch storageType {
	case "hdd":
		return "HDD"
	default:
		return "SSD"
	}
}

func instanceName(project, name string) string {
	return fmt.Sprintf("projects/%s/instances/%s", project, name)
}

func tableName(project, instance, name string) string {
	return fmt.Sprintf("%s/tables/%s", instanceName(project, instance), name)
}
