package storage

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/mosajjal/gcgo/internal/placeholder"
	"github.com/spf13/cobra"
)

// NewCommand returns the storage command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "storage",
		Short: "Manage Cloud Storage",
	}

	cmd.AddCommand(
		newLsCommand(cfg, creds),
		newCpCommand(cfg, creds),
		newMvCommand(creds),
		newRsyncCommand(cfg, creds),
		newCatCommand(creds),
		newSignURLCommand(creds),
		newRmCommand(creds),
		newMbCommand(cfg, creds),
		newRbCommand(creds),
		newIAMCommand(),
		newLifecycleCommand(),
		newRetentionCommand(),
	)

	return cmd
}

func storageClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newLsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "ls [gs://BUCKET[/PREFIX]]",
		Short: "List buckets or objects",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")

			if len(args) == 0 {
				// List buckets
				flagVal, _ := cmd.Flags().GetString("project")
				project := cfg.Project(flagVal)
				if project == "" {
					return fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
				}

				buckets, err := client.ListBuckets(ctx, project)
				if err != nil {
					return err
				}

				if format == "json" {
					return output.PrintJSON(cmd.OutOrStdout(), buckets)
				}
				headers := []string{"NAME", "LOCATION", "CREATED"}
				rows := make([][]string, len(buckets))
				for i, b := range buckets {
					rows[i] = []string{b.Name, b.Location, b.Created}
				}
				return output.PrintTable(cmd.OutOrStdout(), headers, rows)
			}

			// List objects
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}

			objects, err := client.ListObjects(ctx, uri.Bucket, uri.Prefix)
			if err != nil {
				return err
			}

			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), objects)
			}
			headers := []string{"NAME", "SIZE", "UPDATED"}
			rows := make([][]string, len(objects))
			for i, o := range objects {
				rows[i] = []string{o.Name, fmt.Sprintf("%d", o.Size), o.Updated}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newCpCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	_ = cfg // may be needed for project context later
	return &cobra.Command{
		Use:   "cp SRC DST",
		Short: "Copy files between local filesystem and GCS",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcIsGCS, srcBucket, srcPath, err := CopyPath(args[0])
			if err != nil {
				return err
			}
			dstIsGCS, dstBucket, dstPath, err := CopyPath(args[1])
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}

			switch {
			case !srcIsGCS && dstIsGCS:
				// Local → GCS
				f, err := OpenLocalFile(srcPath)
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				if err := client.Upload(ctx, dstBucket, dstPath, f); err != nil {
					return err
				}
			case srcIsGCS && !dstIsGCS:
				// GCS → Local
				f, err := CreateLocalFile(dstPath)
				if err != nil {
					return err
				}
				defer func() { _ = f.Close() }()
				if err := client.Download(ctx, srcBucket, srcPath, f); err != nil {
					return err
				}
			case srcIsGCS && dstIsGCS:
				if err := client.Copy(ctx, srcBucket, srcPath, dstBucket, dstPath); err != nil {
					return err
				}
			default:
				return fmt.Errorf("at least one path must be a gs:// URI")
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Copied %s → %s\n", args[0], args[1])
			return nil
		},
	}
}

func newRmCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "rm gs://BUCKET/OBJECT",
		Short: "Delete an object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.Delete(ctx, uri.Bucket, uri.Prefix); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted %s.\n", args[0])
			return nil
		},
	}
}

func newMbCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var location string

	cmd := &cobra.Command{
		Use:   "mb gs://BUCKET",
		Short: "Create a bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}

			flagVal, _ := cmd.Flags().GetString("project")
			project := cfg.Project(flagVal)
			if project == "" {
				return fmt.Errorf("no project set")
			}

			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateBucket(ctx, project, uri.Bucket, location); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created bucket %s.\n", uri.Bucket)
			return nil
		},
	}

	cmd.Flags().StringVar(&location, "location", "", "Bucket location")

	return cmd
}

func newRbCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "rb gs://BUCKET",
		Short: "Remove a bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteBucket(ctx, uri.Bucket); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Removed bucket %s.\n", uri.Bucket)
			return nil
		},
	}
}

func newIAMCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/storage/docs/access-control/iam"
	return placeholder.NewGroup(
		"iam",
		"Manage Cloud Storage IAM policies",
		docsURL,
		placeholder.NewCommand("get-policy", "Get a bucket IAM policy", docsURL),
		placeholder.NewCommand("set-policy", "Set a bucket IAM policy", docsURL),
		placeholder.NewCommand("test-permissions", "Test permissions against a bucket", docsURL),
	)
}

func newLifecycleCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/storage/docs/lifecycle"
	return placeholder.NewGroup(
		"lifecycle",
		"Manage Cloud Storage lifecycle policies",
		docsURL,
		placeholder.NewCommand("describe", "Describe a bucket lifecycle policy", docsURL),
		placeholder.NewCommand("update", "Update a bucket lifecycle policy", docsURL),
	)
}

func newRetentionCommand() *cobra.Command {
	const docsURL = "https://cloud.google.com/storage/docs/using-bucket-lock"
	return placeholder.NewGroup(
		"retention",
		"Manage Cloud Storage retention policies",
		docsURL,
		placeholder.NewCommand("describe", "Describe a bucket retention policy", docsURL),
		placeholder.NewCommand("update", "Update a bucket retention policy", docsURL),
		placeholder.NewCommand("lock", "Lock a bucket retention policy", docsURL),
	)
}
