package sql

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	instances    []*Instance
	instanceMap  map[string]*Instance
	databases    []*Database
	databaseMap  map[string]*Database
	users        []*User
	backups      []*Backup
	backupMap    map[string]*Backup
	operations   []*Operation
	operationMap map[string]*Operation

	listInstancesErr  error
	getInstanceErr    error
	createInstanceErr error
	updateInstanceErr error
	deleteInstanceErr error
	restartErr        error

	listDatabasesErr  error
	getDatabaseErr    error
	createDatabaseErr error
	deleteDatabaseErr error

	listUsersErr   error
	createUserErr  error
	deleteUserErr  error
	setPasswordErr error

	listBackupsErr   error
	getBackupErr     error
	createBackupErr  error
	deleteBackupErr  error
	restoreBackupErr error
}

func (m *mockClient) ListInstances(_ context.Context, _ string) ([]*Instance, error) {
	return m.instances, m.listInstancesErr
}

func (m *mockClient) GetInstance(_ context.Context, _, name string) (*Instance, error) {
	if m.getInstanceErr != nil {
		return nil, m.getInstanceErr
	}
	inst, ok := m.instanceMap[name]
	if !ok {
		return nil, fmt.Errorf("instance %q not found", name)
	}
	return inst, nil
}

func (m *mockClient) CreateInstance(_ context.Context, _ string, _ *CreateInstanceRequest) error {
	return m.createInstanceErr
}

func (m *mockClient) UpdateInstance(_ context.Context, _, _ string, _ *UpdateInstanceRequest) (string, error) {
	return "", m.updateInstanceErr
}

func (m *mockClient) DeleteInstance(_ context.Context, _, _ string) error {
	return m.deleteInstanceErr
}

func (m *mockClient) RestartInstance(_ context.Context, _, _ string) error {
	return m.restartErr
}

func (m *mockClient) ExportInstance(_ context.Context, _, _ string, _ *ExportInstanceRequest) (string, error) {
	return "", nil
}

func (m *mockClient) ImportInstance(_ context.Context, _, _ string, _ *ImportInstanceRequest) (string, error) {
	return "", nil
}

func (m *mockClient) CloneInstance(_ context.Context, _, _ string, _ *CloneInstanceRequest) (string, error) {
	return "", nil
}

func (m *mockClient) PromoteReplica(_ context.Context, _, _ string, _ bool) (string, error) {
	return "", nil
}

func (m *mockClient) FailoverInstance(_ context.Context, _, _ string, _ int64) (string, error) {
	return "", nil
}

func (m *mockClient) ListDatabases(_ context.Context, _, _ string) ([]*Database, error) {
	return m.databases, m.listDatabasesErr
}

func (m *mockClient) GetDatabase(_ context.Context, _, _, name string) (*Database, error) {
	if m.getDatabaseErr != nil {
		return nil, m.getDatabaseErr
	}
	db, ok := m.databaseMap[name]
	if !ok {
		return nil, fmt.Errorf("database %q not found", name)
	}
	return db, nil
}

func (m *mockClient) CreateDatabase(_ context.Context, _, _, _ string) error {
	return m.createDatabaseErr
}

func (m *mockClient) DeleteDatabase(_ context.Context, _, _, _ string) error {
	return m.deleteDatabaseErr
}

func (m *mockClient) ListUsers(_ context.Context, _, _ string) ([]*User, error) {
	return m.users, m.listUsersErr
}

func (m *mockClient) CreateUser(_ context.Context, _, _, _, _ string) error {
	return m.createUserErr
}

func (m *mockClient) DeleteUser(_ context.Context, _, _, _ string) error {
	return m.deleteUserErr
}

func (m *mockClient) SetPassword(_ context.Context, _, _, _, _ string) error {
	return m.setPasswordErr
}

func (m *mockClient) ListBackups(_ context.Context, _, _ string) ([]*Backup, error) {
	return m.backups, m.listBackupsErr
}

func (m *mockClient) GetBackup(_ context.Context, _, _, id string) (*Backup, error) {
	if m.getBackupErr != nil {
		return nil, m.getBackupErr
	}
	b, ok := m.backupMap[id]
	if !ok {
		return nil, fmt.Errorf("backup %q not found", id)
	}
	return b, nil
}

func (m *mockClient) CreateBackup(_ context.Context, _, _ string) error {
	return m.createBackupErr
}

func (m *mockClient) DeleteBackup(_ context.Context, _, _, _ string) error {
	return m.deleteBackupErr
}

func (m *mockClient) RestoreBackup(_ context.Context, _, _, _ string) error {
	return m.restoreBackupErr
}

func (m *mockClient) ListOperations(_ context.Context, _ string) ([]*Operation, error) {
	return m.operations, nil
}

func (m *mockClient) GetOperation(_ context.Context, _, name string) (*Operation, error) {
	if op, ok := m.operationMap[name]; ok {
		return op, nil
	}
	return nil, fmt.Errorf("operation %q not found", name)
}

func TestMockListInstances(t *testing.T) {
	mock := &mockClient{
		instances: []*Instance{
			{Name: "db-1", DatabaseVersion: "POSTGRES_15", Region: "us-central1", State: "RUNNABLE"},
			{Name: "db-2", DatabaseVersion: "MYSQL_8_0", Region: "us-east1", State: "STOPPED"},
		},
	}

	instances, err := mock.ListInstances(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(instances) != 2 {
		t.Errorf("expected 2 instances, got %d", len(instances))
	}
}

func TestMockListInstancesError(t *testing.T) {
	mock := &mockClient{listInstancesErr: fmt.Errorf("permission denied")}

	_, err := mock.ListInstances(context.Background(), "proj")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetInstance(t *testing.T) {
	mock := &mockClient{
		instanceMap: map[string]*Instance{
			"db-1": {Name: "db-1", DatabaseVersion: "POSTGRES_15", State: "RUNNABLE"},
		},
	}

	inst, err := mock.GetInstance(context.Background(), "proj", "db-1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if inst.Name != "db-1" {
		t.Errorf("name: got %q", inst.Name)
	}

	_, err = mock.GetInstance(context.Background(), "proj", "nope")
	if err == nil {
		t.Fatal("expected error for missing instance")
	}
}

func TestMockListDatabases(t *testing.T) {
	mock := &mockClient{
		databases: []*Database{
			{Name: "mydb", Charset: "UTF8", Collation: "en_US.UTF8"},
		},
	}

	dbs, err := mock.ListDatabases(context.Background(), "proj", "db-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(dbs) != 1 {
		t.Errorf("expected 1 database, got %d", len(dbs))
	}
}

func TestMockListUsers(t *testing.T) {
	mock := &mockClient{
		users: []*User{
			{Name: "admin", Host: "%"},
			{Name: "reader", Host: "10.0.0.0/8"},
		},
	}

	users, err := mock.ListUsers(context.Background(), "proj", "db-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(users) != 2 {
		t.Errorf("expected 2 users, got %d", len(users))
	}
}

func TestMockListBackups(t *testing.T) {
	mock := &mockClient{
		backups: []*Backup{
			{ID: "123", Status: "SUCCESSFUL", Type: "ON_DEMAND"},
		},
	}

	backups, err := mock.ListBackups(context.Background(), "proj", "db-1")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(backups) != 1 {
		t.Errorf("expected 1 backup, got %d", len(backups))
	}
}

func TestMockMutationErrors(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name string
		fn   func(Client) error
		err  error
	}{
		{
			name: "create instance error",
			fn:   func(c Client) error { return c.CreateInstance(ctx, "p", &CreateInstanceRequest{Name: "x"}) },
			err:  fmt.Errorf("quota exceeded"),
		},
		{
			name: "update instance error",
			fn: func(c Client) error {
				_, err := c.UpdateInstance(ctx, "p", "x", &UpdateInstanceRequest{Tier: "db-custom-1-3840"})
				return err
			},
			err: fmt.Errorf("permission denied"),
		},
		{
			name: "delete instance error",
			fn:   func(c Client) error { return c.DeleteInstance(ctx, "p", "x") },
			err:  fmt.Errorf("not found"),
		},
		{
			name: "restart success",
			fn:   func(c Client) error { return c.RestartInstance(ctx, "p", "x") },
		},
		{
			name: "create user error",
			fn:   func(c Client) error { return c.CreateUser(ctx, "p", "i", "u", "pw") },
			err:  fmt.Errorf("already exists"),
		},
		{
			name: "set password success",
			fn:   func(c Client) error { return c.SetPassword(ctx, "p", "i", "u", "pw") },
		},
		{
			name: "restore backup error",
			fn:   func(c Client) error { return c.RestoreBackup(ctx, "p", "i", "123") },
			err:  fmt.Errorf("backup corrupt"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mock := &mockClient{
				createInstanceErr: tt.err,
				updateInstanceErr: tt.err,
				deleteInstanceErr: tt.err,
				restartErr:        tt.err,
				createUserErr:     tt.err,
				setPasswordErr:    tt.err,
				restoreBackupErr:  tt.err,
			}
			err := tt.fn(mock)
			if tt.err != nil && err == nil {
				t.Fatal("expected error")
			}
			if tt.err == nil && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseBackupID(t *testing.T) {
	tests := []struct {
		input   string
		want    int64
		wantErr bool
	}{
		{"123", 123, false},
		{"0", 0, false},
		{"abc", 0, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, err := parseBackupID(tt.input)
			if tt.wantErr && err == nil {
				t.Fatal("expected error")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.want {
				t.Errorf("got %d, want %d", got, tt.want)
			}
		})
	}
}
