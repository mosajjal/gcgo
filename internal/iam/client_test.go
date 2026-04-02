package iam

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	accounts     []*ServiceAccount
	keys         []*SAKey
	bindings     []*IAMBinding
	createdKey   []byte
	listErr      error
	createErr    error
	deleteErr    error
	listKeysErr  error
	createKeyErr error
	deleteKeyErr error
	getPolicyErr error
	addBindErr   error
	rmBindErr    error
}

func (m *mockClient) ListServiceAccounts(_ context.Context, _ string) ([]*ServiceAccount, error) {
	return m.accounts, m.listErr
}

func (m *mockClient) CreateServiceAccount(_ context.Context, _, _, _ string) (*ServiceAccount, error) {
	if m.createErr != nil {
		return nil, m.createErr
	}
	return &ServiceAccount{Email: "new@proj.iam.gserviceaccount.com"}, nil
}

func (m *mockClient) DeleteServiceAccount(_ context.Context, _ string) error {
	return m.deleteErr
}

func (m *mockClient) ListKeys(_ context.Context, _ string) ([]*SAKey, error) {
	return m.keys, m.listKeysErr
}

func (m *mockClient) CreateKey(_ context.Context, _ string) ([]byte, error) {
	return m.createdKey, m.createKeyErr
}

func (m *mockClient) DeleteKey(_ context.Context, _ string) error {
	return m.deleteKeyErr
}

func (m *mockClient) GetPolicy(_ context.Context, _ string) ([]*IAMBinding, error) {
	return m.bindings, m.getPolicyErr
}

func (m *mockClient) AddBinding(_ context.Context, _, _, _ string) error {
	return m.addBindErr
}

func (m *mockClient) RemoveBinding(_ context.Context, _, _, _ string) error {
	return m.rmBindErr
}

func TestMockListServiceAccounts(t *testing.T) {
	mock := &mockClient{
		accounts: []*ServiceAccount{
			{Email: "sa1@proj.iam.gserviceaccount.com", DisplayName: "SA 1"},
			{Email: "sa2@proj.iam.gserviceaccount.com", DisplayName: "SA 2"},
		},
	}

	accounts, err := mock.ListServiceAccounts(context.Background(), "proj")
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2, got %d", len(accounts))
	}
}

func TestMockListServiceAccountsError(t *testing.T) {
	mock := &mockClient{listErr: fmt.Errorf("denied")}
	_, err := mock.ListServiceAccounts(context.Background(), "proj")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockCreateServiceAccount(t *testing.T) {
	mock := &mockClient{}
	sa, err := mock.CreateServiceAccount(context.Background(), "proj", "test", "Test SA")
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if sa.Email == "" {
		t.Error("expected non-empty email")
	}
}

func TestMockGetPolicy(t *testing.T) {
	mock := &mockClient{
		bindings: []*IAMBinding{
			{Role: "roles/viewer", Members: []string{"user:a@b.com"}},
		},
	}

	bindings, err := mock.GetPolicy(context.Background(), "proj")
	if err != nil {
		t.Fatalf("get policy: %v", err)
	}
	if len(bindings) != 1 {
		t.Errorf("expected 1 binding, got %d", len(bindings))
	}
}

func TestMockAddRemoveBinding(t *testing.T) {
	mock := &mockClient{}

	if err := mock.AddBinding(context.Background(), "proj", "user:a@b.com", "roles/editor"); err != nil {
		t.Fatalf("add: %v", err)
	}
	if err := mock.RemoveBinding(context.Background(), "proj", "user:a@b.com", "roles/editor"); err != nil {
		t.Fatalf("remove: %v", err)
	}
}

func TestMockCreateKey(t *testing.T) {
	mock := &mockClient{createdKey: []byte(`{"type":"service_account"}`)}

	data, err := mock.CreateKey(context.Background(), "sa@proj.iam.gserviceaccount.com")
	if err != nil {
		t.Fatalf("create key: %v", err)
	}
	if len(data) == 0 {
		t.Error("expected non-empty key data")
	}
}
