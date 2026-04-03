package folders

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the resource-manager folders command group.
func NewCommand(creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "folders",
		Short: "Manage resource manager folders",
	}

	cmd.AddCommand(
		newListCommand(creds),
		newDescribeCommand(creds),
		newCreateCommand(creds),
		newDeleteCommand(creds),
		newMoveCommand(creds),
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

func newListCommand(creds *auth.Credentials) *cobra.Command {
	var parent string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List folders under a parent",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if parent == "" {
				return fmt.Errorf("--parent is required (e.g. organizations/123 or folders/456)")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			folders, err := client.List(ctx, parent)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), folders)
			}

			headers := []string{"NAME", "DISPLAY_NAME", "PARENT", "STATE"}
			rows := make([][]string, len(folders))
			for i, f := range folders {
				rows[i] = []string{f.Name, f.DisplayName, f.Parent, f.State}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "Parent resource (e.g. organizations/123)")
	_ = cmd.MarkFlagRequired("parent")

	return cmd
}

func newDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe FOLDER_ID",
		Short: "Describe a folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			f, err := client.Get(ctx, args[0])
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), f)
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Name:         %s\n", f.Name)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Display Name: %s\n", f.DisplayName)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Parent:       %s\n", f.Parent)
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "State:        %s\n", f.State)
			return nil
		},
	}
}

func newCreateCommand(creds *auth.Credentials) *cobra.Command {
	var parent, displayName string

	cmd := &cobra.Command{
		Use:   "create",
		Short: "Create a folder",
		RunE: func(cmd *cobra.Command, _ []string) error {
			if parent == "" {
				return fmt.Errorf("--parent is required (e.g. organizations/123 or folders/456)")
			}
			if displayName == "" {
				return fmt.Errorf("--display-name is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			f, err := client.Create(ctx, parent, displayName)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created folder %s.\n", f.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&parent, "parent", "", "Parent resource (e.g. organizations/123)")
	cmd.Flags().StringVar(&displayName, "display-name", "", "Display name for the folder")
	_ = cmd.MarkFlagRequired("parent")
	_ = cmd.MarkFlagRequired("display-name")

	return cmd
}

func newDeleteCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete FOLDER_ID",
		Short: "Delete a folder",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.Delete(ctx, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted folder %s.\n", args[0])
			return nil
		},
	}
}

func newMoveCommand(creds *auth.Credentials) *cobra.Command {
	var destParent string

	cmd := &cobra.Command{
		Use:   "move FOLDER_ID",
		Short: "Move a folder to a new parent",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if destParent == "" {
				return fmt.Errorf("--destination-parent is required")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			f, err := client.Move(ctx, args[0], destParent)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Moved folder %s to %s.\n", f.Name, f.Parent)
			return nil
		},
	}

	cmd.Flags().StringVar(&destParent, "destination-parent", "", "Destination parent (e.g. organizations/123)")
	_ = cmd.MarkFlagRequired("destination-parent")

	return cmd
}
