package storage

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func newMvCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "mv SRC DST",
		Short: "Move an object in GCS",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			srcURI, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			dstURI, err := ParseGSURI(args[1])
			if err != nil {
				return err
			}

			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.Move(ctx, srcURI.Bucket, srcURI.Prefix, dstURI.Bucket, dstURI.Prefix); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Moved %s -> %s\n", args[0], args[1])
			return nil
		},
	}
}

func newRsyncCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var dryRun bool
	_ = cfg

	cmd := &cobra.Command{
		Use:   "rsync SRC DST",
		Short: "Sync files between local directory and GCS",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}

			srcIsGCS := strings.HasPrefix(args[0], "gs://")
			dstIsGCS := strings.HasPrefix(args[1], "gs://")
			if srcIsGCS == dstIsGCS {
				return fmt.Errorf("exactly one of SRC/DST must be a gs:// URI")
			}

			var actions []RsyncAction
			if !srcIsGCS && dstIsGCS {
				uri, err := ParseGSURI(args[1])
				if err != nil {
					return err
				}
				actions, err = client.Rsync(ctx, true, args[0], uri.Bucket, uri.Prefix, dryRun)
				if err != nil {
					return err
				}
			} else {
				uri, err := ParseGSURI(args[0])
				if err != nil {
					return err
				}
				actions, err = client.Rsync(ctx, false, args[1], uri.Bucket, uri.Prefix, dryRun)
				if err != nil {
					return err
				}
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), actions)
			}
			if len(actions) == 0 {
				_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Already in sync.")
				return nil
			}
			prefix := ""
			if dryRun {
				prefix = "[dry-run] "
			}
			for _, a := range actions {
				_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%s%s: %s\n", prefix, a.Action, a.Path)
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&dryRun, "dry-run", false, "Show what would be synced without making changes")
	return cmd
}

func newCatCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "cat gs://BUCKET/OBJECT",
		Short: "Print object contents to stdout",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			if uri.Prefix == "" {
				return fmt.Errorf("object path required (got bucket only)")
			}

			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}
			return client.Cat(ctx, uri.Bucket, uri.Prefix, cmd.OutOrStdout())
		},
	}
}

func newSignURLCommand(creds *auth.Credentials) *cobra.Command {
	var duration time.Duration

	cmd := &cobra.Command{
		Use:   "signurl gs://BUCKET/OBJECT",
		Short: "Generate a signed URL for an object",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			if uri.Prefix == "" {
				return fmt.Errorf("object path required (got bucket only)")
			}

			ctx := context.Background()
			client, err := storageClient(ctx, creds)
			if err != nil {
				return err
			}
			url, err := client.SignURL(ctx, uri.Bucket, uri.Prefix, duration)
			if err != nil {
				return err
			}

			_, _ = fmt.Fprintln(cmd.OutOrStdout(), url)
			return nil
		},
	}

	cmd.Flags().DurationVar(&duration, "duration", 1*time.Hour, "URL validity duration (e.g. 30m, 2h)")
	return cmd
}
