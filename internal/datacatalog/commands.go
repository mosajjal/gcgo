package datacatalog

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the datacatalog command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "datacatalog",
		Short: "Manage Data Catalog resources",
	}

	cmd.AddCommand(
		newEntryGroupsCommand(cfg, creds),
		newEntriesCommand(cfg, creds),
	)

	return cmd
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func requireRegion(cmd *cobra.Command, cfg *config.Config) (string, error) {
	region, _ := cmd.Flags().GetString("region")
	if region == "" {
		region = cfg.Region()
	}
	if region == "" {
		return "", fmt.Errorf("no region set (use --region or 'gcgo config set region REGION')")
	}
	return region, nil
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newEntryGroupsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entry-groups",
		Short: "Manage Data Catalog entry groups",
	}

	cmd.AddCommand(
		newEntryGroupsListCommand(cfg, creds),
		newEntryGroupsDescribeCommand(cfg, creds),
		newEntryGroupsCreateCommand(cfg, creds),
		newEntryGroupsDeleteCommand(cfg, creds),
	)

	return cmd
}

func newEntryGroupsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List entry groups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			groups, err := client.ListEntryGroups(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), groups)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "DESCRIPTION"}
			rows := make([][]string, len(groups))
			for i, g := range groups {
				rows[i] = []string{g.Name, g.DisplayName, g.Description}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newEntryGroupsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ENTRY_GROUP",
		Short: "Describe an entry group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			group, err := client.GetEntryGroup(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), group)
		},
	}
}

func newEntryGroupsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateEntryGroupRequest

	cmd := &cobra.Command{
		Use:   "create ENTRY_GROUP",
		Short: "Create an entry group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			req.EntryGroupID = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateEntryGroup(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created entry group %q.\n", req.EntryGroupID)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.DisplayName, "display-name", "", "Display name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Description")

	return cmd
}

func newEntryGroupsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete ENTRY_GROUP",
		Short: "Delete an entry group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteEntryGroup(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted entry group %q.\n", args[0])
			return nil
		},
	}
}

func newEntriesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "entries",
		Short: "Manage Data Catalog entries",
	}

	cmd.AddCommand(
		newEntriesListCommand(cfg, creds),
		newEntriesDescribeCommand(cfg, creds),
	)

	return cmd
}

func newEntriesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var entryGroup string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List entries in an entry group",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if entryGroup == "" {
				return fmt.Errorf("--entry-group is required")
			}
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			entries, err := client.ListEntries(ctx, project, region, entryGroup)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), entries)
			}
			headers := []string{"NAME", "DISPLAY_NAME", "TYPE", "LINKED_RESOURCE"}
			rows := make([][]string, len(entries))
			for i, e := range entries {
				rows[i] = []string{e.Name, e.DisplayName, e.Type, e.LinkedResource}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&entryGroup, "entry-group", "", "Entry group ID")

	return cmd
}

func newEntriesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var entryGroup string

	cmd := &cobra.Command{
		Use:   "describe ENTRY",
		Short: "Describe an entry",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if entryGroup == "" {
				return fmt.Errorf("--entry-group is required")
			}
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			entry, err := client.GetEntry(ctx, project, region, entryGroup, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), entry)
		},
	}

	cmd.Flags().StringVar(&entryGroup, "entry-group", "", "Entry group ID")

	return cmd
}
