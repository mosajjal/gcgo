package billing

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	accounts     []*Account
	accountMap   map[string]*Account
	projectInfos []*ProjectBillingInfo
	budgets      []*Budget
	budgetMap    map[string]*Budget
	createdBudget *Budget
	linked       *ProjectBillingInfo
	unlinked     *ProjectBillingInfo
	listAccErr   error
	getAccErr    error
	listProjErr  error
	listBudgetErr error
	getBudgetErr  error
	createBudgetErr error
	deleteBudgetErr error
	linkErr      error
	unlinkErr    error
}

func (m *mockClient) ListAccounts(_ context.Context) ([]*Account, error) {
	if m.listAccErr != nil {
		return nil, m.listAccErr
	}
	return m.accounts, nil
}

func (m *mockClient) GetAccount(_ context.Context, id string) (*Account, error) {
	if m.getAccErr != nil {
		return nil, m.getAccErr
	}
	a, ok := m.accountMap[id]
	if !ok {
		return nil, fmt.Errorf("billing account %q not found", id)
	}
	return a, nil
}

func (m *mockClient) ListProjectBillingInfo(_ context.Context, _ string) ([]*ProjectBillingInfo, error) {
	if m.listProjErr != nil {
		return nil, m.listProjErr
	}
	return m.projectInfos, nil
}

func (m *mockClient) LinkProject(_ context.Context, projectID, accountID string) (*ProjectBillingInfo, error) {
	if m.linkErr != nil {
		return nil, m.linkErr
	}
	if m.linked != nil {
		return m.linked, nil
	}
	return &ProjectBillingInfo{
		ProjectID:          projectID,
		BillingAccountName: "billingAccounts/" + accountID,
		BillingEnabled:     true,
	}, nil
}

func (m *mockClient) UnlinkProject(_ context.Context, projectID string) (*ProjectBillingInfo, error) {
	if m.unlinkErr != nil {
		return nil, m.unlinkErr
	}
	if m.unlinked != nil {
		return m.unlinked, nil
	}
	return &ProjectBillingInfo{
		ProjectID:      projectID,
		BillingEnabled: false,
	}, nil
}

func (m *mockClient) ListBudgets(_ context.Context, _ string) ([]*Budget, error) {
	if m.listBudgetErr != nil {
		return nil, m.listBudgetErr
	}
	return m.budgets, nil
}

func (m *mockClient) GetBudget(_ context.Context, name string) (*Budget, error) {
	if m.getBudgetErr != nil {
		return nil, m.getBudgetErr
	}
	if m.budgetMap == nil {
		m.budgetMap = map[string]*Budget{}
	}
	budget, ok := m.budgetMap[name]
	if !ok {
		return nil, fmt.Errorf("budget %q not found", name)
	}
	return budget, nil
}

func (m *mockClient) CreateBudget(_ context.Context, billingAccount string, req *CreateBudgetRequest) (*Budget, error) {
	if m.createBudgetErr != nil {
		return nil, m.createBudgetErr
	}
	if m.createdBudget != nil {
		return m.createdBudget, nil
	}
	return &Budget{
		Name:           "billingAccounts/" + billingAccount + "/budgets/demo",
		DisplayName:    req.DisplayName,
		OwnershipScope:  "ALL_USERS",
		Amount:         fmt.Sprintf("%d %s", req.Amount, req.CurrencyCode),
	}, nil
}

func (m *mockClient) DeleteBudget(_ context.Context, _ string) error {
	return m.deleteBudgetErr
}

func TestMockListAccounts(t *testing.T) {
	mock := &mockClient{
		accounts: []*Account{
			{Name: "billingAccounts/ABC-123", DisplayName: "Main Account", Open: true},
			{Name: "billingAccounts/DEF-456", DisplayName: "Dev Account", Open: true},
		},
	}

	accounts, err := mock.ListAccounts(context.Background())
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(accounts) != 2 {
		t.Errorf("expected 2 accounts, got %d", len(accounts))
	}
}

func TestMockListAccountsError(t *testing.T) {
	mock := &mockClient{listAccErr: fmt.Errorf("permission denied")}
	_, err := mock.ListAccounts(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockGetAccount(t *testing.T) {
	mock := &mockClient{
		accountMap: map[string]*Account{
			"ABC-123": {Name: "billingAccounts/ABC-123", DisplayName: "Main Account", Open: true},
		},
	}

	a, err := mock.GetAccount(context.Background(), "ABC-123")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if a.DisplayName != "Main Account" {
		t.Errorf("display name: got %q", a.DisplayName)
	}
}

func TestMockGetAccountNotFound(t *testing.T) {
	mock := &mockClient{accountMap: map[string]*Account{}}
	_, err := mock.GetAccount(context.Background(), "NOPE")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockListProjectBillingInfo(t *testing.T) {
	mock := &mockClient{
		projectInfos: []*ProjectBillingInfo{
			{ProjectID: "proj-1", BillingAccountName: "billingAccounts/ABC-123", BillingEnabled: true},
		},
	}

	infos, err := mock.ListProjectBillingInfo(context.Background(), "ABC-123")
	if err != nil {
		t.Fatalf("list project billing: %v", err)
	}
	if len(infos) != 1 {
		t.Errorf("expected 1 info, got %d", len(infos))
	}
}

func TestMockLinkProject(t *testing.T) {
	mock := &mockClient{}
	info, err := mock.LinkProject(context.Background(), "proj-1", "ABC-123")
	if err != nil {
		t.Fatalf("link: %v", err)
	}
	if !info.BillingEnabled {
		t.Error("expected billing enabled after link")
	}
}

func TestMockLinkProjectError(t *testing.T) {
	mock := &mockClient{linkErr: fmt.Errorf("access denied")}
	_, err := mock.LinkProject(context.Background(), "proj-1", "ABC-123")
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestMockUnlinkProject(t *testing.T) {
	mock := &mockClient{}
	info, err := mock.UnlinkProject(context.Background(), "proj-1")
	if err != nil {
		t.Fatalf("unlink: %v", err)
	}
	if info.BillingEnabled {
		t.Error("expected billing disabled after unlink")
	}
}

func TestMockUnlinkProjectError(t *testing.T) {
	mock := &mockClient{unlinkErr: fmt.Errorf("not allowed")}
	_, err := mock.UnlinkProject(context.Background(), "proj-1")
	if err == nil {
		t.Fatal("expected error")
	}
}
