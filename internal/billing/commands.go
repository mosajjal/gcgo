package billing

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the billing command group.
func NewCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "billing",
		Short: "Manage Cloud Billing",
	}

	cmd.AddCommand(
		newAccountsCommand(creds),
		newProjectsCommand(creds),
		newBudgetsCommand(creds),
	)

	return cmd
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

// Accounts subcommands

func newAccountsCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "accounts",
		Short: "Manage billing accounts",
	}

	cmd.AddCommand(
		newAccountsListCommand(creds),
		newAccountsDescribeCommand(creds),
	)

	return cmd
}

func newAccountsListCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List billing accounts",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			accounts, err := client.ListAccounts(ctx)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), accounts)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "OPEN"}
			rows := make([][]string, len(accounts))
			for i, a := range accounts {
				rows[i] = []string{a.Name, a.DisplayName, fmt.Sprintf("%v", a.Open)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newAccountsDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ACCOUNT_ID",
		Short: "Describe a billing account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			account, err := client.GetAccount(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), account)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:           %s\n", account.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name:   %s\n", account.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Open:           %v\n", account.Open)
			if account.MasterAccountID != "" {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Master Account: %s\n", account.MasterAccountID)
			}
			return nil
		},
	}
}

// Projects subcommands

func newProjectsCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "projects",
		Short: "Manage project billing associations",
	}

	cmd.AddCommand(
		newProjectsListCommand(creds),
		newProjectsDescribeCommand(creds),
		newProjectsLinkCommand(creds),
		newProjectsUnlinkCommand(creds),
	)

	return cmd
}

func newProjectsListCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list ACCOUNT_ID",
		Short: "List projects linked to a billing account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			infos, err := client.ListProjectBillingInfo(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), infos)
			}

			headers := []string{"PROJECT_ID", "BILLING_ACCOUNT", "BILLING_ENABLED"}
			rows := make([][]string, len(infos))
			for i, info := range infos {
				rows[i] = []string{info.ProjectID, info.BillingAccountName, fmt.Sprintf("%v", info.BillingEnabled)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newProjectsDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe PROJECT_ID",
		Short: "Describe a project's billing association",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			info, err := client.GetProjectBillingInfo(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), info)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Project ID:       %s\n", info.ProjectID)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Billing Account:  %s\n", info.BillingAccountName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Billing Enabled:  %v\n", info.BillingEnabled)
			return nil
		},
	}
}

func newProjectsLinkCommand(creds *auth.Credentials) *cobra.Command {
	var accountID string

	cmd := &cobra.Command{
		Use:   "link PROJECT_ID",
		Short: "Link a project to a billing account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			info, err := client.LinkProject(ctx, args[0], accountID)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Linked project %s to %s.\n", info.ProjectID, info.BillingAccountName)
			return nil
		},
	}

	cmd.Flags().StringVar(&accountID, "billing-account", "", "Billing account ID to link")
	_ = cmd.MarkFlagRequired("billing-account")

	return cmd
}

func newProjectsUnlinkCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "unlink PROJECT_ID",
		Short: "Unlink a project from its billing account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if _, err := client.UnlinkProject(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Unlinked project %s from billing.\n", args[0])
			return nil
		},
	}
}

// Budgets subcommands

func newBudgetsCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "budgets",
		Short: "Manage billing budgets",
	}
	cmd.AddCommand(
		newBudgetsListCommand(creds),
		newBudgetsDescribeCommand(creds),
		newBudgetsCreateCommand(creds),
		newBudgetsDeleteCommand(creds),
	)
	return cmd
}

func newBudgetsListCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list BILLING_ACCOUNT_ID",
		Short: "List budgets in a billing account",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			budgets, err := client.ListBudgets(ctx, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), budgets)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "OWNERSHIP_SCOPE", "AMOUNT"}
			rows := make([][]string, len(budgets))
			for i, budget := range budgets {
				rows[i] = []string{budget.Name, budget.DisplayName, budget.OwnershipScope, budget.Amount}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newBudgetsDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe BUDGET_NAME",
		Short: "Describe a budget",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			budget, err := client.GetBudget(ctx, args[0])
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), budget)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:            %s\n", budget.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name:    %s\n", budget.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Ownership Scope: %s\n", budget.OwnershipScope)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Amount:          %s\n", budget.Amount)
			return nil
		},
	}
}

func newBudgetsCreateCommand(creds *auth.Credentials) *cobra.Command {
	var displayName string
	var amount int64
	var currencyCode string
	var calendarPeriod string

	cmd := &cobra.Command{
		Use:   "create BILLING_ACCOUNT_ID",
		Short: "Create a budget",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if amount <= 0 {
				return fmt.Errorf("--amount must be greater than zero")
			}
			if currencyCode == "" {
				return fmt.Errorf("--currency-code is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			budget, err := client.CreateBudget(ctx, args[0], &CreateBudgetRequest{
				DisplayName:    displayName,
				Amount:         amount,
				CurrencyCode:   currencyCode,
				CalendarPeriod: calendarPeriod,
			})
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), budget)
		},
	}
	cmd.Flags().StringVar(&displayName, "display-name", "", "Budget display name")
	cmd.Flags().Int64Var(&amount, "amount", 0, "Budget amount in whole currency units")
	cmd.Flags().StringVar(&currencyCode, "currency-code", "", "Currency code, for example USD")
	cmd.Flags().StringVar(&calendarPeriod, "calendar-period", "month", "Budget calendar period: month, quarter, or year")
	_ = cmd.MarkFlagRequired("display-name")
	return cmd
}

func newBudgetsDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete BUDGET_NAME",
		Short: "Delete a budget",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteBudget(ctx, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted budget %s.\n", args[0])
			return nil
		},
	}
}
