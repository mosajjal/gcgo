package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"cloud.google.com/go/iam"
	"cloud.google.com/go/storage"
	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
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
		newIAMCommand(cfg, creds),
		newLifecycleCommand(cfg, creds),
		newRetentionCommand(cfg, creds),
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

// rawStorageClient returns the underlying *storage.Client for operations not
// covered by the Client interface (IAM, lifecycle, retention).
func rawStorageClient(ctx context.Context, creds *auth.Credentials) (*storage.Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	sc, err := storage.NewClient(ctx, opt)
	if err != nil {
		return nil, fmt.Errorf("create storage client: %w", err)
	}
	return sc, nil
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

func newIAMCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage Cloud Storage IAM policies",
	}
	cmd.AddCommand(
		newIAMGetPolicyCommand(creds),
		newIAMSetPolicyCommand(creds),
		newIAMTestPermissionsCommand(creds),
	)
	return cmd
}

func newIAMGetPolicyCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "get-policy gs://BUCKET",
		Short: "Get a bucket IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			policy, err := sc.Bucket(uri.Bucket).IAM().Policy(ctx)
			if err != nil {
				return fmt.Errorf("get bucket iam policy: %w", err)
			}

			// Serialize into a human-readable bindings format.
			var out iamPolicyJSON
			for _, role := range policy.Roles() {
				out.Bindings = append(out.Bindings, iamBindingJSON{
					Role:    string(role),
					Members: policy.Members(role),
				})
			}
			return output.PrintJSON(cmd.OutOrStdout(), out)
		},
	}
}

// iamPolicyJSON is the JSON representation of an IAM policy as used by GCP.
// Format: {"bindings": [{"role": "roles/storage.admin", "members": ["user:foo@example.com"]}]}
type iamPolicyJSON struct {
	Bindings []iamBindingJSON `json:"bindings"`
}

type iamBindingJSON struct {
	Role    string   `json:"role"`
	Members []string `json:"members"`
}

func newIAMSetPolicyCommand(creds *auth.Credentials) *cobra.Command {
	var policyFile string

	cmd := &cobra.Command{
		Use:   "set-policy gs://BUCKET",
		Short: "Set a bucket IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if policyFile == "" {
				return fmt.Errorf("--policy-file is required")
			}
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}

			data, err := os.ReadFile(policyFile) //nolint:gosec // user explicitly provides path
			if err != nil {
				return fmt.Errorf("read policy file: %w", err)
			}

			var pj iamPolicyJSON
			if err := json.Unmarshal(data, &pj); err != nil {
				return fmt.Errorf("parse policy file: %w", err)
			}

			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			handle := sc.Bucket(uri.Bucket).IAM()
			policy, err := handle.Policy(ctx)
			if err != nil {
				return fmt.Errorf("get bucket iam policy: %w", err)
			}

			// Remove all existing bindings then apply from file.
			for _, role := range policy.Roles() {
				for _, member := range policy.Members(role) {
					policy.Remove(member, role)
				}
			}
			for _, b := range pj.Bindings {
				for _, member := range b.Members {
					policy.Add(member, iam.RoleName(b.Role))
				}
			}

			if err := handle.SetPolicy(ctx, policy); err != nil {
				return fmt.Errorf("set bucket iam policy: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "IAM policy updated for %s.\n", uri.Bucket)
			return nil
		},
	}

	cmd.Flags().StringVar(&policyFile, "policy-file", "", "Path to JSON file containing the IAM policy")
	return cmd
}

func newIAMTestPermissionsCommand(creds *auth.Credentials) *cobra.Command {
	var permissionsFlag string

	cmd := &cobra.Command{
		Use:   "test-permissions gs://BUCKET",
		Short: "Test permissions against a bucket",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if permissionsFlag == "" {
				return fmt.Errorf("--permissions is required")
			}
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			permissions := strings.Split(permissionsFlag, ",")

			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			allowed, err := sc.Bucket(uri.Bucket).IAM().TestPermissions(ctx, permissions)
			if err != nil {
				return fmt.Errorf("test bucket iam permissions: %w", err)
			}
			return output.PrintJSON(cmd.OutOrStdout(), allowed)
		},
	}

	cmd.Flags().StringVar(&permissionsFlag, "permissions", "", "Comma-separated list of permissions to test")
	return cmd
}

func newLifecycleCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lifecycle",
		Short: "Manage Cloud Storage lifecycle policies",
	}
	cmd.AddCommand(
		newLifecycleDescribeCommand(creds),
		newLifecycleUpdateCommand(creds),
	)
	return cmd
}

func newLifecycleDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe gs://BUCKET",
		Short: "Describe a bucket lifecycle policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			attrs, err := sc.Bucket(uri.Bucket).Attrs(ctx)
			if err != nil {
				return fmt.Errorf("get bucket attrs: %w", err)
			}
			return output.PrintJSON(cmd.OutOrStdout(), attrs.Lifecycle)
		},
	}
}

// lifecycleJSON is the JSON representation of a GCS lifecycle config.
type lifecycleJSON struct {
	Rule []lifecycleRuleJSON `json:"rule"`
}

type lifecycleRuleJSON struct {
	Action    lifecycleActionJSON    `json:"action"`
	Condition lifecycleConditionJSON `json:"condition"`
}

type lifecycleActionJSON struct {
	Type         string `json:"type"`
	StorageClass string `json:"storageClass,omitempty"`
}

type lifecycleConditionJSON struct {
	Age                int      `json:"age,omitempty"`
	IsLive             *bool    `json:"isLive,omitempty"`
	MatchesStorageClass []string `json:"matchesStorageClass,omitempty"`
	NumNewerVersions   int      `json:"numNewerVersions,omitempty"`
}

func newLifecycleUpdateCommand(creds *auth.Credentials) *cobra.Command {
	var lifecycleFile string

	cmd := &cobra.Command{
		Use:   "update gs://BUCKET",
		Short: "Update a bucket lifecycle policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if lifecycleFile == "" {
				return fmt.Errorf("--lifecycle-file is required")
			}
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}

			data, err := os.ReadFile(lifecycleFile) //nolint:gosec // user explicitly provides path
			if err != nil {
				return fmt.Errorf("read lifecycle file: %w", err)
			}

			var lj lifecycleJSON
			if err := json.Unmarshal(data, &lj); err != nil {
				return fmt.Errorf("parse lifecycle file: %w", err)
			}

			lc := storage.Lifecycle{}
			for _, r := range lj.Rule {
				rule := storage.LifecycleRule{
					Action: storage.LifecycleAction{
						Type:         r.Action.Type,
						StorageClass: r.Action.StorageClass,
					},
					Condition: storage.LifecycleCondition{
						AgeInDays:           int64(r.Condition.Age),
						Liveness:            storage.LiveAndArchived,
						MatchesStorageClasses: r.Condition.MatchesStorageClass,
						NumNewerVersions:    int64(r.Condition.NumNewerVersions),
					},
				}
				if r.Condition.IsLive != nil {
					if *r.Condition.IsLive {
						rule.Condition.Liveness = storage.Live
					} else {
						rule.Condition.Liveness = storage.Archived
					}
				}
				lc.Rules = append(lc.Rules, rule)
			}

			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			if _, err := sc.Bucket(uri.Bucket).Update(ctx, storage.BucketAttrsToUpdate{Lifecycle: &lc}); err != nil {
				return fmt.Errorf("update bucket lifecycle: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Lifecycle policy updated for %s.\n", uri.Bucket)
			return nil
		},
	}

	cmd.Flags().StringVar(&lifecycleFile, "lifecycle-file", "", "Path to JSON file containing lifecycle rules")
	return cmd
}

func newRetentionCommand(_ *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "retention",
		Short: "Manage Cloud Storage retention policies",
	}
	cmd.AddCommand(
		newRetentionDescribeCommand(creds),
		newRetentionUpdateCommand(creds),
		newRetentionLockCommand(creds),
	)
	return cmd
}

func newRetentionDescribeCommand(creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe gs://BUCKET",
		Short: "Describe a bucket retention policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			attrs, err := sc.Bucket(uri.Bucket).Attrs(ctx)
			if err != nil {
				return fmt.Errorf("get bucket attrs: %w", err)
			}
			type retentionOutput struct {
				RetentionPeriod    string `json:"retentionPeriod"`
				EffectiveTime      string `json:"effectiveTime"`
				IsLocked           bool   `json:"isLocked"`
				Metageneration     int64  `json:"metageneration"`
			}
			var out retentionOutput
			if attrs.RetentionPolicy != nil {
				out.RetentionPeriod = attrs.RetentionPolicy.RetentionPeriod.String()
				out.EffectiveTime = attrs.RetentionPolicy.EffectiveTime.String()
				out.IsLocked = attrs.RetentionPolicy.IsLocked
			}
			out.Metageneration = attrs.MetaGeneration
			return output.PrintJSON(cmd.OutOrStdout(), out)
		},
	}
}

// parseRetentionDuration parses durations like "30d", "1y", "2h", or standard
// Go duration strings (e.g. "720h").
func parseRetentionDuration(s string) (time.Duration, error) {
	s = strings.TrimSpace(s)
	if strings.HasSuffix(s, "d") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "d"))
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return time.Duration(n) * 24 * time.Hour, nil
	}
	if strings.HasSuffix(s, "y") {
		n, err := strconv.Atoi(strings.TrimSuffix(s, "y"))
		if err != nil {
			return 0, fmt.Errorf("invalid duration %q: %w", s, err)
		}
		return time.Duration(n) * 365 * 24 * time.Hour, nil
	}
	return time.ParseDuration(s)
}

func newRetentionUpdateCommand(creds *auth.Credentials) *cobra.Command {
	var period string

	cmd := &cobra.Command{
		Use:   "update gs://BUCKET",
		Short: "Update a bucket retention policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if period == "" {
				return fmt.Errorf("--retention-period is required")
			}
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}
			d, err := parseRetentionDuration(period)
			if err != nil {
				return err
			}

			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			rp := &storage.RetentionPolicy{RetentionPeriod: d}
			if _, err := sc.Bucket(uri.Bucket).Update(ctx, storage.BucketAttrsToUpdate{RetentionPolicy: rp}); err != nil {
				return fmt.Errorf("update bucket retention policy: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Retention policy updated for %s (period: %s).\n", uri.Bucket, d)
			return nil
		},
	}

	cmd.Flags().StringVar(&period, "retention-period", "", "Retention period (e.g. 30d, 1y, 720h)")
	return cmd
}

func newRetentionLockCommand(creds *auth.Credentials) *cobra.Command {
	var metageneration int64

	cmd := &cobra.Command{
		Use:   "lock gs://BUCKET",
		Short: "Lock a bucket retention policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if metageneration == 0 {
				return fmt.Errorf("--metageneration is required (get it from 'retention describe')")
			}
			uri, err := ParseGSURI(args[0])
			if err != nil {
				return err
			}

			ctx := context.Background()
			sc, err := rawStorageClient(ctx, creds)
			if err != nil {
				return err
			}
			defer func() { _ = sc.Close() }()

			handle := sc.Bucket(uri.Bucket).If(storage.BucketConditions{MetagenerationMatch: metageneration})
			if err := handle.LockRetentionPolicy(ctx); err != nil {
				return fmt.Errorf("lock bucket retention policy: %w", err)
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Retention policy locked for %s.\n", uri.Bucket)
			return nil
		},
	}

	cmd.Flags().Int64Var(&metageneration, "metageneration", 0, "Current bucket metageneration (from 'retention describe')")
	return cmd
}
