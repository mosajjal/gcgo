package run

import (
	"context"
	"fmt"
	"strings"

	"cloud.google.com/go/iam/apiv1/iampb"
	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/flags"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the run command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "run",
		Short: "Manage Cloud Run services",
	}

	cmd.AddCommand(
		newServicesCommand(cfg, creds),
		newRevisionsCommand(cfg, creds),
		newJobsCommand(cfg, creds),
		newDeployCommand(cfg, creds),
		newDomainMappingsCommand(cfg, creds),
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

// addRegionCompletion registers region tab-completion on cmd. Call this after
// any cmd.Flags().StringVar(..., "region", ...) to keep the StringVar binding
// while still providing completions.
func addRegionCompletion(cmd *cobra.Command) {
	_ = cmd.RegisterFlagCompletionFunc("region", func(_ *cobra.Command, _ []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		var matches []string
		for _, r := range flags.CommonRegions {
			if strings.HasPrefix(r, toComplete) {
				matches = append(matches, r)
			}
		}
		return matches, cobra.ShellCompDirectiveNoFileComp
	})
}

func runClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newServicesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "Manage Cloud Run services",
	}

	cmd.AddCommand(
		newServicesListCommand(cfg, creds),
		newServicesDescribeCommand(cfg, creds),
		newServicesUpdateCommand(cfg, creds),
		newServicesUpdateTrafficCommand(cfg, creds),
		newServicesRollbackCommand(cfg, creds),
		newServicesIAMCommand(cfg, creds),
		newServicesDeleteCommand(cfg, creds),
	)

	return cmd
}

func newServicesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloud Run services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			services, err := client.ListServices(ctx, project, region)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), services)
			}

			headers := []string{"NAME", "REGION", "URL"}
			rows := make([][]string, len(services))
			for i, s := range services {
				rows[i] = []string{s.Name, s.Region, s.URI}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newServicesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "describe SERVICE",
		Short: "Describe a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			svc, err := client.GetService(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), svc)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newServicesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "delete SERVICE",
		Short: "Delete a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteService(ctx, project, region, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted service %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newServicesUpdateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req DeployRequest
	var region string
	var envSlice []string

	cmd := &cobra.Command{
		Use:   "update SERVICE",
		Short: "Update a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			req.Name = args[0]
			req.Env = make(map[string]string)
			for _, e := range envSlice {
				k, v, ok := strings.Cut(e, "=")
				if !ok {
					return fmt.Errorf("invalid env format %q (expected KEY=VALUE)", e)
				}
				req.Env[k] = v
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.Deploy(ctx, project, region, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated service %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Image, "image", "", "Container image")
	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&req.Memory, "memory", "", "Memory limit")
	cmd.Flags().StringVar(&req.CPU, "cpu", "", "CPU limit")
	cmd.Flags().Int32Var(&req.Port, "port", 0, "Container port")
	cmd.Flags().StringSliceVar(&envSlice, "env", nil, "Environment variables (KEY=VALUE)")
	cmd.Flags().BoolVar(&req.AllowUnauthenticated, "allow-unauthenticated", false, "Allow unauthenticated access")

	return cmd
}

func newServicesUpdateTrafficCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var req UpdateTrafficRequest

	cmd := &cobra.Command{
		Use:   "update-traffic SERVICE",
		Short: "Update traffic routing for a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}
			if !req.ToLatest && req.Revision == "" {
				return fmt.Errorf("either --to-latest or --to-revision is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			service, err := client.UpdateTraffic(ctx, project, region, args[0], &req)
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), service)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().BoolVar(&req.ToLatest, "to-latest", false, "Send traffic to the latest ready revision")
	cmd.Flags().StringVar(&req.Revision, "to-revision", "", "Revision name to receive traffic")
	cmd.Flags().Int32Var(&req.Percent, "percent", 100, "Traffic percent for the target")
	cmd.Flags().StringVar(&req.Tag, "tag", "", "Optional traffic tag")

	return cmd
}

func newServicesRollbackCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var toRevision string
	var percent int32
	var tag string

	cmd := &cobra.Command{
		Use:   "rollback SERVICE",
		Short: "Rollback traffic for a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			target := toRevision
			if target == "" {
				service, err := client.GetService(ctx, project, region, args[0])
				if err != nil {
					return err
				}
				revisions, err := client.ListRevisions(ctx, project, region, args[0])
				if err != nil {
					return err
				}
				for _, rev := range revisions {
					if rev.Name != service.LatestReadyRevision {
						target = rev.Name
						break
					}
				}
				if target == "" {
					target = service.LatestReadyRevision
				}
			}
			if target == "" {
				return fmt.Errorf("no revision available to roll back to")
			}

			updated, err := client.UpdateTraffic(ctx, project, region, args[0], &UpdateTrafficRequest{
				Revision: target,
				Percent:  percent,
				Tag:      tag,
			})
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), updated)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&toRevision, "to-revision", "", "Revision to receive traffic")
	cmd.Flags().Int32Var(&percent, "percent", 100, "Traffic percent for the target revision")
	cmd.Flags().StringVar(&tag, "tag", "", "Optional traffic tag")

	return cmd
}

func newServicesIAMCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "iam",
		Short: "Manage IAM policy for Cloud Run services",
	}

	cmd.AddCommand(
		newServicesIAMGetCommand(cfg, creds),
		newServicesIAMSetCommand(cfg, creds),
		newServicesIAMTestCommand(cfg, creds),
	)

	return cmd
}

func newServicesIAMGetCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "get-policy SERVICE",
		Short: "Get the IAM policy for a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}
			policy, err := client.GetServicePolicy(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), policy)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newServicesIAMSetCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var member string
	var role string
	var remove bool

	cmd := &cobra.Command{
		Use:   "set-policy SERVICE",
		Short: "Add or remove a binding in a Cloud Run service IAM policy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}
			if member == "" || role == "" {
				return fmt.Errorf("--member and --role are required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}
			policy, err := client.GetServicePolicy(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			policy = mutatePolicyBinding(policy, member, role, remove)
			updated, err := client.SetServicePolicy(ctx, project, region, args[0], policy)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), updated)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&member, "member", "", "IAM member")
	cmd.Flags().StringVar(&role, "role", "", "IAM role")
	cmd.Flags().BoolVar(&remove, "remove", false, "Remove the binding instead of adding it")

	return cmd
}

func newServicesIAMTestCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var permissions []string

	cmd := &cobra.Command{
		Use:   "test-permissions SERVICE",
		Short: "Test permissions on a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}
			if len(permissions) == 0 {
				return fmt.Errorf("--permission is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}
			allowed, err := client.TestServicePermissions(ctx, project, region, args[0], permissions)
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), map[string]any{"permissions": allowed})
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringSliceVar(&permissions, "permission", nil, "Permission to test (repeatable)")

	return cmd
}

func newDeployCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req DeployRequest
	var region string
	var envSlice []string

	cmd := &cobra.Command{
		Use:   "deploy SERVICE --image=IMAGE",
		Short: "Deploy a Cloud Run service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			req.Name = args[0]
			req.Env = make(map[string]string)
			for _, e := range envSlice {
				k, v, ok := strings.Cut(e, "=")
				if !ok {
					return fmt.Errorf("invalid env format %q (expected KEY=VALUE)", e)
				}
				req.Env[k] = v
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.Deploy(ctx, project, region, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deployed service %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&req.Image, "image", "", "Container image")
	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&req.Memory, "memory", "", "Memory limit")
	cmd.Flags().StringVar(&req.CPU, "cpu", "", "CPU limit")
	cmd.Flags().Int32Var(&req.Port, "port", 8080, "Container port")
	cmd.Flags().StringSliceVar(&envSlice, "env", nil, "Environment variables (KEY=VALUE)")
	cmd.Flags().BoolVar(&req.AllowUnauthenticated, "allow-unauthenticated", false, "Allow unauthenticated access")
	_ = cmd.MarkFlagRequired("image")

	return cmd
}

func newRevisionsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "revisions",
		Short: "Manage Cloud Run service revisions",
	}

	cmd.AddCommand(
		newRevisionsListCommand(cfg, creds),
		newRevisionsDescribeCommand(cfg, creds),
		newRevisionsDeleteCommand(cfg, creds),
	)

	return cmd
}

func newRevisionsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string
	var service string

	cmd := &cobra.Command{
		Use:   "list",
		Short: "List revisions for a Cloud Run service",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}
			if service == "" {
				return fmt.Errorf("--service is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}
			revisions, err := client.ListRevisions(ctx, project, region, service)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), revisions)
			}

			headers := []string{"NAME", "SERVICE", "IMAGE", "CREATE_TIME", "GENERATION"}
			rows := make([][]string, len(revisions))
			for i, rev := range revisions {
				rows[i] = []string{rev.Name, rev.Service, rev.Image, rev.CreateTime, fmt.Sprintf("%d", rev.Generation)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	cmd.Flags().StringVar(&service, "service", "", "Cloud Run service name")

	return cmd
}

func newRevisionsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "describe REVISION",
		Short: "Describe a Cloud Run revision",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}
			revision, err := client.GetRevision(ctx, project, region, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), revision)
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)

	return cmd
}

func newRevisionsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var region string

	cmd := &cobra.Command{
		Use:   "delete REVISION",
		Short: "Delete a Cloud Run revision",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			if region == "" {
				region = cfg.Region()
			}
			if region == "" {
				return fmt.Errorf("--region is required (or set region in config)")
			}

			ctx := context.Background()
			client, err := runClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteRevision(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted revision %s.\n", args[0])
			return nil
		},
	}

	cmd.Flags().StringVar(&region, "region", "", "Region")
	addRegionCompletion(cmd)
	return cmd
}

func mutatePolicyBinding(policy *iampb.Policy, member, role string, remove bool) *iampb.Policy {
	if policy == nil {
		policy = &iampb.Policy{}
	}

	found := false
	for i, binding := range policy.GetBindings() {
		if binding.GetRole() != role {
			continue
		}
		found = true
		if remove {
			var members []string
			for _, m := range binding.GetMembers() {
				if m != member {
					members = append(members, m)
				}
			}
			if len(members) == 0 {
				policy.Bindings = append(policy.Bindings[:i], policy.Bindings[i+1:]...)
			} else {
				binding.Members = members
			}
		} else {
			for _, m := range binding.GetMembers() {
				if m == member {
					return policy
				}
			}
			binding.Members = append(binding.Members, member)
		}
		break
	}

	if !found && !remove {
		policy.Bindings = append(policy.Bindings, &iampb.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	return policy
}
