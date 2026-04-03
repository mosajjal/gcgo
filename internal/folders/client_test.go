package folders

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	folders   []*Folder
	getMap    map[string]*Folder
	created   *Folder
	moved     *Folder
	listErr   error
	getErr    error
	createErr error
	deleteErr error
	moveErr   error
}

func (m *mockClient) List(_ context.Context, _ string) ([]*Folder, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.folders, nil
}

func (m *mockClient) Get(_ context.Context, id string) (*Folder, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	f, ok := m.getMap[id]
	if !ok {
		return nil, fmt.Errorf("folder %q not found", id)
	}
	return f, nil
}

func (m *mockClient) Create(_ context.Context, parent, displayName string) (*Folder, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	if m.created != nil {
		return m.created, nil
	}
	return &Folder{
		Name:        "folders/999",
		DisplayName: displayName,
		Parent:      parent,
		State:       "ACTIVE",
	}, nil
}

func (m *mockClient) Delete(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockClient) Move(_ context.Context, _ string, destParent string) (*Folder, error) {
	if m.moveErr != nil {
		return nil, m.moveErr
	}
	if m.moved != nil {
		return m.moved, nil
	}
	return &Folder{
		Name:        "folders/123",
		DisplayName: "Moved",
		Parent:      destParent,
		State:       "ACTIVE",
	}, nil
}

func TestMockClientList(t *testing.T) {
	mock := &mockClient{
		folders: []*Folder{
			{Name: "folders/1", DisplayName: "Folder One", Parent: "organizations/123", State: "ACTIVE"},
			{Name: "folders/2", DisplayName: "Folder Two", Parent: "organizations/123", State: "ACTIVE"},
		},
	}

	folders, err := mock.List(context.Background(), "organizations/123")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(folders) != 2 {
		t.Errorf("expected 2 folders, got %d", len(folders))
	}
}

func TestMockClientListError(t *testing.T) {
	mock := &mockClient{listErr: fmt.Errorf("permission denied")}
	_, err := mock.List(context.Background(), "organizations/123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClientGet(t *testing.T) {
	mock := &mockClient{
		getMap: map[string]*Folder{
			"1": {Name: "folders/1", DisplayName: "Folder One", State: "ACTIVE"},
		},
	}

	f, err := mock.Get(context.Background(), "1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if f.DisplayName != "Folder One" {
		t.Errorf("display name: got %q", f.DisplayName)
	}
}

func TestMockClientGetNotFound(t *testing.T) {
	mock := &mockClient{getMap: map[string]*Folder{}}
	_, err := mock.Get(context.Background(), "999")
	if err == nil {
		t.Fatal("expected error for missing folder")
	}
}

func TestMockClientCreate(t *testing.T) {
	mock := &mockClient{}
	f, err := mock.Create(context.Background(), "organizations/123", "New Folder")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if f.DisplayName != "New Folder" {
		t.Errorf("display name: got %q", f.DisplayName)
	}
}

func TestMockClientCreateError(t *testing.T) {
	mock := &mockClient{createErr: fmt.Errorf("quota exceeded")}
	_, err := mock.Create(context.Background(), "organizations/123", "New Folder")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClientDelete(t *testing.T) {
	mock := &mockClient{}
	if err := mock.Delete(context.Background(), "1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
}

func TestMockClientDeleteError(t *testing.T) {
	mock := &mockClient{deleteErr: fmt.Errorf("not empty")}
	if err := mock.Delete(context.Background(), "1"); err == nil {
		t.Fatal("expected error")
	}
}

func TestMockClientMove(t *testing.T) {
	mock := &mockClient{}
	f, err := mock.Move(context.Background(), "1", "organizations/456")
	if err != nil {
		t.Fatalf("move: %v", err)
	}
	if f.Parent != "organizations/456" {
		t.Errorf("parent: got %q", f.Parent)
	}
}

func TestMockClientMoveError(t *testing.T) {
	mock := &mockClient{moveErr: fmt.Errorf("access denied")}
	_, err := mock.Move(context.Background(), "1", "organizations/456")
	if err == nil {
		t.Fatal("expected error")
	}
}
