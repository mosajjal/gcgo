package billing

import (
	"context"
	"fmt"

	cloudbilling "google.golang.org/api/cloudbilling/v1"
	billingbudgets "google.golang.org/api/billingbudgets/v1beta1"
	"google.golang.org/api/option"
)

// Account holds billing account fields.
type Account struct {
	Name            string `json:"name"`
	DisplayName     string `json:"display_name"`
	Open            bool   `json:"open"`
	MasterAccountID string `json:"master_account,omitempty"`
}

// ProjectBillingInfo holds project billing association.
type ProjectBillingInfo struct {
	ProjectID          string `json:"project_id"`
	BillingAccountName string `json:"billing_account_name"`
	BillingEnabled     bool   `json:"billing_enabled"`
}

// Budget holds the fields we display for billing budgets.
type Budget struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	OwnershipScope  string `json:"ownership_scope"`
	Amount         string `json:"amount"`
}

// CreateBudgetRequest holds budget creation parameters.
type CreateBudgetRequest struct {
	BudgetID      string
	DisplayName   string
	Amount        int64
	CurrencyCode  string
	CalendarPeriod string
}

// Client defines billing operations.
type Client interface {
	ListAccounts(ctx context.Context) ([]*Account, error)
	GetAccount(ctx context.Context, accountID string) (*Account, error)
	ListProjectBillingInfo(ctx context.Context, accountID string) ([]*ProjectBillingInfo, error)
	GetProjectBillingInfo(ctx context.Context, projectID string) (*ProjectBillingInfo, error)
	LinkProject(ctx context.Context, projectID, accountID string) (*ProjectBillingInfo, error)
	UnlinkProject(ctx context.Context, projectID string) (*ProjectBillingInfo, error)
	ListBudgets(ctx context.Context, billingAccount string) ([]*Budget, error)
	GetBudget(ctx context.Context, name string) (*Budget, error)
	CreateBudget(ctx context.Context, billingAccount string, req *CreateBudgetRequest) (*Budget, error)
	DeleteBudget(ctx context.Context, name string) error
}

type gcpClient struct {
	cloud   *cloudbilling.APIService
	budgets *billingbudgets.Service
}

// NewClient creates a Client backed by the real GCP API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	c, err := cloudbilling.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create billing client: %w", err)
	}
	b, err := billingbudgets.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create billing budgets client: %w", err)
	}
	return &gcpClient{cloud: c, budgets: b}, nil
}

func (c *gcpClient) ListAccounts(ctx context.Context) ([]*Account, error) {
	call := c.cloud.BillingAccounts.List()
	var accounts []*Account
	if err := call.Pages(ctx, func(resp *cloudbilling.ListBillingAccountsResponse) error {
		for _, a := range resp.BillingAccounts {
			accounts = append(accounts, accountFromProto(a))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list billing accounts: %w", err)
	}
	return accounts, nil
}

func (c *gcpClient) GetAccount(ctx context.Context, accountID string) (*Account, error) {
	a, err := c.cloud.BillingAccounts.Get("billingAccounts/" + accountID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get billing account %s: %w", accountID, err)
	}
	return accountFromProto(a), nil
}

func (c *gcpClient) ListProjectBillingInfo(ctx context.Context, accountID string) ([]*ProjectBillingInfo, error) {
	call := c.cloud.BillingAccounts.Projects.List("billingAccounts/" + accountID)
	var infos []*ProjectBillingInfo
	if err := call.Pages(ctx, func(resp *cloudbilling.ListProjectBillingInfoResponse) error {
		for _, info := range resp.ProjectBillingInfo {
			infos = append(infos, billingInfoFromProto(info))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list project billing info: %w", err)
	}
	return infos, nil
}

func (c *gcpClient) GetProjectBillingInfo(ctx context.Context, projectID string) (*ProjectBillingInfo, error) {
	info, err := c.cloud.Projects.GetBillingInfo("projects/" + projectID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get billing info for project %s: %w", projectID, err)
	}
	return billingInfoFromProto(info), nil
}

func (c *gcpClient) LinkProject(ctx context.Context, projectID, accountID string) (*ProjectBillingInfo, error) {
	info, err := c.cloud.Projects.UpdateBillingInfo("projects/"+projectID, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: "billingAccounts/" + accountID,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("link project %s to billing account %s: %w", projectID, accountID, err)
	}
	return billingInfoFromProto(info), nil
}

func (c *gcpClient) UnlinkProject(ctx context.Context, projectID string) (*ProjectBillingInfo, error) {
	info, err := c.cloud.Projects.UpdateBillingInfo("projects/"+projectID, &cloudbilling.ProjectBillingInfo{
		BillingAccountName: "",
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("unlink project %s from billing: %w", projectID, err)
	}
	return billingInfoFromProto(info), nil
}

func (c *gcpClient) ListBudgets(ctx context.Context, billingAccount string) ([]*Budget, error) {
	parent := "billingAccounts/" + billingAccount
	call := c.budgets.BillingAccounts.Budgets.List(parent).Context(ctx)

	var budgets []*Budget
	if err := call.Pages(ctx, func(resp *billingbudgets.GoogleCloudBillingBudgetsV1beta1ListBudgetsResponse) error {
		for _, b := range resp.Budgets {
			budgets = append(budgets, budgetFromProto(b))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list budgets: %w", err)
	}
	return budgets, nil
}

func (c *gcpClient) GetBudget(ctx context.Context, name string) (*Budget, error) {
	budget, err := c.budgets.BillingAccounts.Budgets.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get budget %s: %w", name, err)
	}
	return budgetFromProto(budget), nil
}

func (c *gcpClient) CreateBudget(ctx context.Context, billingAccount string, req *CreateBudgetRequest) (*Budget, error) {
	if req == nil {
		return nil, fmt.Errorf("create budget: nil request")
	}
	budget := &billingbudgets.GoogleCloudBillingBudgetsV1beta1Budget{
		DisplayName: req.DisplayName,
		Amount: &billingbudgets.GoogleCloudBillingBudgetsV1beta1BudgetAmount{
			SpecifiedAmount: &billingbudgets.GoogleTypeMoney{
				Units:        req.Amount,
				CurrencyCode: req.CurrencyCode,
			},
		},
		BudgetFilter: &billingbudgets.GoogleCloudBillingBudgetsV1beta1Filter{
			CalendarPeriod: budgetCalendarPeriod(req.CalendarPeriod),
		},
	}
	created, err := c.budgets.BillingAccounts.Budgets.Create("billingAccounts/"+billingAccount, &billingbudgets.GoogleCloudBillingBudgetsV1beta1CreateBudgetRequest{
		Budget: budget,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create budget: %w", err)
	}
	return budgetFromProto(created), nil
}

func (c *gcpClient) DeleteBudget(ctx context.Context, name string) error {
	if _, err := c.budgets.BillingAccounts.Budgets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete budget %s: %w", name, err)
	}
	return nil
}

func accountFromProto(a *cloudbilling.BillingAccount) *Account {
	return &Account{
		Name:            a.Name,
		DisplayName:     a.DisplayName,
		Open:            a.Open,
		MasterAccountID: a.MasterBillingAccount,
	}
}

func billingInfoFromProto(info *cloudbilling.ProjectBillingInfo) *ProjectBillingInfo {
	return &ProjectBillingInfo{
		ProjectID:          info.ProjectId,
		BillingAccountName: info.BillingAccountName,
		BillingEnabled:     info.BillingEnabled,
	}
}

func budgetFromProto(budget *billingbudgets.GoogleCloudBillingBudgetsV1beta1Budget) *Budget {
	if budget == nil {
		return nil
	}
	amount := ""
	if budget.Amount != nil && budget.Amount.SpecifiedAmount != nil {
		amount = fmt.Sprintf("%d %s", budget.Amount.SpecifiedAmount.Units, budget.Amount.SpecifiedAmount.CurrencyCode)
	}
	return &Budget{
		Name:          budget.Name,
		DisplayName:   budget.DisplayName,
		OwnershipScope: budget.OwnershipScope,
		Amount:        amount,
	}
}

func budgetCalendarPeriod(period string) string {
	switch period {
	case "quarter":
		return "QUARTER"
	case "year":
		return "YEAR"
	default:
		return "MONTH"
	}
}
