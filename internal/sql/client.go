package sql

import (
	"context"
	"fmt"

	sqladmin "google.golang.org/api/option"
	api "google.golang.org/api/sqladmin/v1beta4"
)

// Instance holds Cloud SQL instance fields.
type Instance struct {
	Name            string `json:"name"`
	DatabaseVersion string `json:"database_version"`
	Region          string `json:"region"`
	Tier            string `json:"tier"`
	State           string `json:"state"`
	IPAddress       string `json:"ip_address"`
	SettingsVersion int64  `json:"settings_version"`
}

// Database holds Cloud SQL database fields.
type Database struct {
	Name      string `json:"name"`
	Charset   string `json:"charset"`
	Collation string `json:"collation"`
}

// User holds Cloud SQL user fields.
type User struct {
	Name string `json:"name"`
	Host string `json:"host"`
}

// Backup holds Cloud SQL backup fields.
type Backup struct {
	ID         string `json:"id"`
	Status     string `json:"status"`
	Type       string `json:"type"`
	EnqueuedAt string `json:"enqueued_at"`
}

// Operation holds Cloud SQL operation fields.
type Operation struct {
	Name       string `json:"name"`
	Type       string `json:"type"`
	Status     string `json:"status"`
	TargetID   string `json:"target_id"`
	InsertTime string `json:"insert_time"`
	EndTime    string `json:"end_time"`
	Error      string `json:"error,omitempty"`
}

// ExportInstanceRequest holds Cloud SQL export parameters.
type ExportInstanceRequest struct {
	URI       string
	FileType  string
	Databases []string
	Offload   bool
}

// ImportInstanceRequest holds Cloud SQL import parameters.
type ImportInstanceRequest struct {
	URI        string
	FileType   string
	Database   string
	ImportUser string
}

// CloneInstanceRequest holds Cloud SQL clone parameters.
type CloneInstanceRequest struct {
	DestinationInstance string
	PointInTime         string
}

// UpdateInstanceRequest holds Cloud SQL update parameters.
type UpdateInstanceRequest struct {
	DatabaseVersion string
	Tier            string
}

// Client defines Cloud SQL operations.
type Client interface {
	ListInstances(ctx context.Context, project string) ([]*Instance, error)
	GetInstance(ctx context.Context, project, name string) (*Instance, error)
	CreateInstance(ctx context.Context, project string, req *CreateInstanceRequest) error
	UpdateInstance(ctx context.Context, project, name string, req *UpdateInstanceRequest) (string, error)
	DeleteInstance(ctx context.Context, project, name string) error
	RestartInstance(ctx context.Context, project, name string) error
	ExportInstance(ctx context.Context, project, instance string, req *ExportInstanceRequest) (string, error)
	ImportInstance(ctx context.Context, project, instance string, req *ImportInstanceRequest) (string, error)
	CloneInstance(ctx context.Context, project, instance string, req *CloneInstanceRequest) (string, error)
	PromoteReplica(ctx context.Context, project, instance string, failover bool) (string, error)
	FailoverInstance(ctx context.Context, project, instance string, settingsVersion int64) (string, error)

	ListDatabases(ctx context.Context, project, instance string) ([]*Database, error)
	GetDatabase(ctx context.Context, project, instance, name string) (*Database, error)
	CreateDatabase(ctx context.Context, project, instance, name string) error
	DeleteDatabase(ctx context.Context, project, instance, name string) error

	ListUsers(ctx context.Context, project, instance string) ([]*User, error)
	CreateUser(ctx context.Context, project, instance, name, password string) error
	DeleteUser(ctx context.Context, project, instance, name string) error
	SetPassword(ctx context.Context, project, instance, name, password string) error

	ListBackups(ctx context.Context, project, instance string) ([]*Backup, error)
	GetBackup(ctx context.Context, project, instance, id string) (*Backup, error)
	CreateBackup(ctx context.Context, project, instance string) error
	DeleteBackup(ctx context.Context, project, instance, id string) error
	RestoreBackup(ctx context.Context, project, instance, backupID string) error
	ListOperations(ctx context.Context, project string) ([]*Operation, error)
	GetOperation(ctx context.Context, project, name string) (*Operation, error)
}

// CreateInstanceRequest holds parameters for creating a Cloud SQL instance.
type CreateInstanceRequest struct {
	Name            string
	DatabaseVersion string
	Tier            string
	Region          string
}

type gcpClient struct {
	svc *api.Service
}

// NewClient creates a Client backed by the real Cloud SQL Admin API.
func NewClient(ctx context.Context, opts ...sqladmin.ClientOption) (Client, error) {
	svc, err := api.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create sqladmin client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListInstances(ctx context.Context, project string) ([]*Instance, error) {
	resp, err := c.svc.Instances.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list sql instances: %w", err)
	}
	var out []*Instance
	for _, i := range resp.Items {
		out = append(out, instanceFromAPI(i))
	}
	return out, nil
}

func (c *gcpClient) GetInstance(ctx context.Context, project, name string) (*Instance, error) {
	i, err := c.svc.Instances.Get(project, name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get sql instance %s: %w", name, err)
	}
	return instanceFromAPI(i), nil
}

func (c *gcpClient) CreateInstance(ctx context.Context, project string, req *CreateInstanceRequest) error {
	dbVersion := req.DatabaseVersion
	if dbVersion == "" {
		dbVersion = "POSTGRES_15"
	}
	tier := req.Tier
	if tier == "" {
		tier = "db-f1-micro"
	}

	inst := &api.DatabaseInstance{
		Name:            req.Name,
		DatabaseVersion: dbVersion,
		Region:          req.Region,
		Settings: &api.Settings{
			Tier: tier,
		},
	}

	_, err := c.svc.Instances.Insert(project, inst).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create sql instance %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpClient) UpdateInstance(ctx context.Context, project, name string, req *UpdateInstanceRequest) (string, error) {
	inst := &api.DatabaseInstance{}
	if req.DatabaseVersion != "" {
		inst.DatabaseVersion = req.DatabaseVersion
	}
	if req.Tier != "" {
		inst.Settings = &api.Settings{Tier: req.Tier}
	}
	if inst.DatabaseVersion == "" && inst.Settings == nil {
		return "", fmt.Errorf("no update fields provided")
	}

	op, err := c.svc.Instances.Patch(project, name, inst).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("update sql instance %s: %w", name, err)
	}
	return op.Name, nil
}

func (c *gcpClient) DeleteInstance(ctx context.Context, project, name string) error {
	_, err := c.svc.Instances.Delete(project, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete sql instance %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) RestartInstance(ctx context.Context, project, name string) error {
	_, err := c.svc.Instances.Restart(project, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restart sql instance %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ExportInstance(ctx context.Context, project, instance string, req *ExportInstanceRequest) (string, error) {
	op, err := c.svc.Instances.Export(project, instance, &api.InstancesExportRequest{
		ExportContext: &api.ExportContext{
			Uri:       req.URI,
			FileType:  sqlFileType(req.FileType),
			Databases: req.Databases,
			Offload:   req.Offload,
		},
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("export sql instance %s: %w", instance, err)
	}
	return op.Name, nil
}

func (c *gcpClient) ImportInstance(ctx context.Context, project, instance string, req *ImportInstanceRequest) (string, error) {
	op, err := c.svc.Instances.Import(project, instance, &api.InstancesImportRequest{
		ImportContext: &api.ImportContext{
			Uri:        req.URI,
			FileType:   sqlFileType(req.FileType),
			Database:   req.Database,
			ImportUser: req.ImportUser,
		},
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("import into sql instance %s: %w", instance, err)
	}
	return op.Name, nil
}

func (c *gcpClient) CloneInstance(ctx context.Context, project, instance string, req *CloneInstanceRequest) (string, error) {
	op, err := c.svc.Instances.Clone(project, instance, &api.InstancesCloneRequest{
		CloneContext: &api.CloneContext{
			DestinationInstanceName: req.DestinationInstance,
			PointInTime:             req.PointInTime,
		},
	}).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("clone sql instance %s: %w", instance, err)
	}
	return op.Name, nil
}

func (c *gcpClient) PromoteReplica(ctx context.Context, project, instance string, failover bool) (string, error) {
	op, err := c.svc.Instances.PromoteReplica(project, instance).Failover(failover).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("promote replica %s: %w", instance, err)
	}
	return op.Name, nil
}

func (c *gcpClient) FailoverInstance(ctx context.Context, project, instance string, settingsVersion int64) (string, error) {
	req := &api.InstancesFailoverRequest{}
	if settingsVersion > 0 {
		req.FailoverContext = &api.FailoverContext{SettingsVersion: settingsVersion}
	}

	op, err := c.svc.Instances.Failover(project, instance, req).Context(ctx).Do()
	if err != nil {
		return "", fmt.Errorf("failover sql instance %s: %w", instance, err)
	}
	return op.Name, nil
}

func (c *gcpClient) ListDatabases(ctx context.Context, project, instance string) ([]*Database, error) {
	resp, err := c.svc.Databases.List(project, instance).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list databases: %w", err)
	}
	var out []*Database
	for _, d := range resp.Items {
		out = append(out, &Database{
			Name:      d.Name,
			Charset:   d.Charset,
			Collation: d.Collation,
		})
	}
	return out, nil
}

func (c *gcpClient) GetDatabase(ctx context.Context, project, instance, name string) (*Database, error) {
	d, err := c.svc.Databases.Get(project, instance, name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get database %s: %w", name, err)
	}
	return &Database{
		Name:      d.Name,
		Charset:   d.Charset,
		Collation: d.Collation,
	}, nil
}

func (c *gcpClient) CreateDatabase(ctx context.Context, project, instance, name string) error {
	db := &api.Database{Name: name}
	_, err := c.svc.Databases.Insert(project, instance, db).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create database %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) DeleteDatabase(ctx context.Context, project, instance, name string) error {
	_, err := c.svc.Databases.Delete(project, instance, name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete database %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListUsers(ctx context.Context, project, instance string) ([]*User, error) {
	resp, err := c.svc.Users.List(project, instance).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	var out []*User
	for _, u := range resp.Items {
		out = append(out, &User{Name: u.Name, Host: u.Host})
	}
	return out, nil
}

func (c *gcpClient) CreateUser(ctx context.Context, project, instance, name, password string) error {
	u := &api.User{Name: name, Password: password}
	_, err := c.svc.Users.Insert(project, instance, u).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create user %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) DeleteUser(ctx context.Context, project, instance, name string) error {
	_, err := c.svc.Users.Delete(project, instance).Name(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete user %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) SetPassword(ctx context.Context, project, instance, name, password string) error {
	u := &api.User{Name: name, Password: password}
	_, err := c.svc.Users.Update(project, instance, u).Name(name).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("set password for user %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListBackups(ctx context.Context, project, instance string) ([]*Backup, error) {
	resp, err := c.svc.BackupRuns.List(project, instance).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list backups: %w", err)
	}
	var out []*Backup
	for _, b := range resp.Items {
		out = append(out, backupFromAPI(b))
	}
	return out, nil
}

func (c *gcpClient) GetBackup(ctx context.Context, project, instance, id string) (*Backup, error) {
	bid, err := parseBackupID(id)
	if err != nil {
		return nil, err
	}
	b, err := c.svc.BackupRuns.Get(project, instance, bid).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get backup %s: %w", id, err)
	}
	return backupFromAPI(b), nil
}

func (c *gcpClient) CreateBackup(ctx context.Context, project, instance string) error {
	_, err := c.svc.Backups.CreateBackup(project, &api.Backup{
		Instance: instance,
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create backup: %w", err)
	}
	return nil
}

func (c *gcpClient) DeleteBackup(ctx context.Context, project, instance, id string) error {
	bid, err := parseBackupID(id)
	if err != nil {
		return err
	}
	_, err = c.svc.BackupRuns.Delete(project, instance, bid).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete backup %s: %w", id, err)
	}
	return nil
}

func (c *gcpClient) RestoreBackup(ctx context.Context, project, instance, backupID string) error {
	bid, err := parseBackupID(backupID)
	if err != nil {
		return err
	}
	req := &api.InstancesRestoreBackupRequest{
		RestoreBackupContext: &api.RestoreBackupContext{
			BackupRunId: bid,
			InstanceId:  instance,
			Project:     project,
		},
	}
	_, err = c.svc.Instances.RestoreBackup(project, instance, req).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("restore backup %s: %w", backupID, err)
	}
	return nil
}

func (c *gcpClient) ListOperations(ctx context.Context, project string) ([]*Operation, error) {
	resp, err := c.svc.Operations.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list sql operations: %w", err)
	}
	var out []*Operation
	for _, op := range resp.Items {
		out = append(out, operationFromAPI(op))
	}
	return out, nil
}

func (c *gcpClient) GetOperation(ctx context.Context, project, name string) (*Operation, error) {
	op, err := c.svc.Operations.Get(project, name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get sql operation %s: %w", name, err)
	}
	return operationFromAPI(op), nil
}

func instanceFromAPI(i *api.DatabaseInstance) *Instance {
	ip := ""
	for _, addr := range i.IpAddresses {
		if addr.Type == "PRIMARY" {
			ip = addr.IpAddress
			break
		}
	}
	tier := ""
	if i.Settings != nil {
		tier = i.Settings.Tier
	}
	return &Instance{
		Name:            i.Name,
		DatabaseVersion: i.DatabaseVersion,
		Region:          i.Region,
		Tier:            tier,
		State:           i.State,
		IPAddress:       ip,
		SettingsVersion: func() int64 {
			if i.Settings != nil {
				return i.Settings.SettingsVersion
			}
			return 0
		}(),
	}
}

func backupFromAPI(b *api.BackupRun) *Backup {
	return &Backup{
		ID:         fmt.Sprintf("%d", b.Id),
		Status:     b.Status,
		Type:       b.Type,
		EnqueuedAt: b.EnqueuedTime,
	}
}

func operationFromAPI(op *api.Operation) *Operation {
	if op == nil {
		return nil
	}
	out := &Operation{
		Name:       op.Name,
		Type:       op.OperationType,
		Status:     op.Status,
		TargetID:   op.TargetId,
		InsertTime: op.InsertTime,
		EndTime:    op.EndTime,
	}
	if op.Error != nil && len(op.Error.Errors) > 0 {
		out.Error = op.Error.Errors[0].Message
	}
	return out
}

func sqlFileType(fileType string) string {
	switch fileType {
	case "csv":
		return "CSV"
	case "bak":
		return "BAK"
	default:
		return "SQL"
	}
}

func parseBackupID(id string) (int64, error) {
	var bid int64
	_, err := fmt.Sscanf(id, "%d", &bid)
	if err != nil {
		return 0, fmt.Errorf("parse backup id %q: %w", id, err)
	}
	return bid, nil
}
