package compute

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func newInstanceTemplatesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance-templates",
		Short: "Manage instance templates",
	}
	cmd.AddCommand(
		newInstanceTemplatesListCommand(cfg, creds),
		newInstanceTemplatesDescribeCommand(cfg, creds),
		newInstanceTemplatesCreateCommand(cfg, creds),
		newInstanceTemplatesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newInstanceTemplatesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List instance templates",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			templates, err := client.ListInstanceTemplates(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), templates)
			}
			headers := []string{"NAME", "MACHINE_TYPE", "NETWORK", "DESCRIPTION"}
			rows := make([][]string, len(templates))
			for i, tpl := range templates {
				rows[i] = []string{tpl.Name, tpl.MachineType, tpl.Network, tpl.Description}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newInstanceTemplatesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe TEMPLATE",
		Short: "Describe an instance template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			tpl, err := client.GetInstanceTemplate(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), tpl)
		},
	}
}

func newInstanceTemplatesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateInstanceTemplateRequest

	cmd := &cobra.Command{
		Use:   "create TEMPLATE",
		Short: "Create an instance template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateInstanceTemplate(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created instance template %q.\n", req.Name)
			return nil
		},
	}

	cmd.Flags().StringVar(&req.MachineType, "machine-type", "e2-medium", "Machine type")
	cmd.Flags().StringVar(&req.ImageFamily, "image-family", "debian-12", "Source image family")
	cmd.Flags().StringVar(&req.ImageProject, "image-project", "debian-cloud", "Source image project")
	cmd.Flags().StringVar(&req.Network, "network", "", "VPC network")
	cmd.Flags().StringVar(&req.Subnet, "subnet", "", "Subnet")
	cmd.Flags().StringVar(&req.Description, "description", "", "Template description")
	return cmd
}

func newInstanceTemplatesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete TEMPLATE",
		Short: "Delete an instance template",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteInstanceTemplate(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted instance template %q.\n", args[0])
			return nil
		},
	}
}

func newInstanceGroupsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instance-groups",
		Short: "Manage instance groups",
	}
	cmd.AddCommand(
		newManagedInstanceGroupsCommand(cfg, creds),
		newUnmanagedInstanceGroupsCommand(cfg, creds),
	)
	return cmd
}

func newUnmanagedInstanceGroupsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "unmanaged",
		Short: "Manage unmanaged instance groups",
	}
	cmd.AddCommand(
		newUnmanagedInstanceGroupsListCommand(cfg, creds),
		newUnmanagedInstanceGroupsDescribeCommand(cfg, creds),
		newUnmanagedInstanceGroupsCreateCommand(cfg, creds),
		newUnmanagedInstanceGroupsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newUnmanagedInstanceGroupsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List unmanaged instance groups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			groups, err := client.ListUnmanagedInstanceGroups(ctx, project, zone)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), groups)
			}
			headers := []string{"NAME", "ZONE", "SIZE", "NETWORK"}
			rows := make([][]string, len(groups))
			for i, g := range groups {
				rows[i] = []string{g.Name, g.Zone, fmt.Sprintf("%d", g.Size), g.Network}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newUnmanagedInstanceGroupsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe GROUP",
		Short: "Describe an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			g, err := client.GetUnmanagedInstanceGroup(ctx, project, zone, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), g)
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newUnmanagedInstanceGroupsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateUnmanagedInstanceGroupRequest

	cmd := &cobra.Command{
		Use:   "create GROUP",
		Short: "Create an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateUnmanagedInstanceGroup(ctx, project, zone, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created unmanaged instance group %q in %s/%s.\n", req.Name, project, zone)
			return nil
		},
	}
	addZoneFlag(cmd)
	cmd.Flags().StringVar(&req.Network, "network", "", "VPC network")
	cmd.Flags().StringVar(&req.Description, "description", "", "Group description")
	return cmd
}

func newUnmanagedInstanceGroupsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete GROUP",
		Short: "Delete an unmanaged instance group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteUnmanagedInstanceGroup(ctx, project, zone, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted unmanaged instance group %q.\n", args[0])
			return nil
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newManagedInstanceGroupsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "managed",
		Short: "Manage managed instance groups",
	}
	cmd.AddCommand(
		newManagedInstanceGroupsListCommand(cfg, creds),
		newManagedInstanceGroupsDescribeCommand(cfg, creds),
		newManagedInstanceGroupsCreateCommand(cfg, creds),
		newManagedInstanceGroupsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newManagedInstanceGroupsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List managed instance groups",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			groups, err := client.ListInstanceGroupManagers(ctx, project, zone)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), groups)
			}
			headers := []string{"NAME", "ZONE", "INSTANCE_TEMPLATE", "TARGET_SIZE", "STATUS"}
			rows := make([][]string, len(groups))
			for i, mig := range groups {
				rows[i] = []string{mig.Name, mig.Zone, mig.InstanceTemplate, fmt.Sprintf("%d", mig.TargetSize), mig.Status}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newManagedInstanceGroupsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe GROUP",
		Short: "Describe a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			mig, err := client.GetInstanceGroupManager(ctx, project, zone, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), mig)
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newManagedInstanceGroupsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateInstanceGroupManagerRequest

	cmd := &cobra.Command{
		Use:   "create GROUP",
		Short: "Create a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			if req.Template == "" {
				return fmt.Errorf("--template is required")
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateInstanceGroupManager(ctx, project, zone, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created managed instance group %q in %s/%s.\n", req.Name, project, zone)
			return nil
		},
	}
	addZoneFlag(cmd)
	cmd.Flags().StringVar(&req.Template, "template", "", "Instance template")
	cmd.Flags().Int32Var(&req.TargetSize, "size", 1, "Target size")
	cmd.Flags().StringVar(&req.BaseInstanceName, "base-instance-name", "", "Base instance name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Group description")
	return cmd
}

func newManagedInstanceGroupsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete GROUP",
		Short: "Delete a managed instance group",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteInstanceGroupManager(ctx, project, zone, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted managed instance group %q.\n", args[0])
			return nil
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newAutoscalersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "autoscalers",
		Short: "Manage autoscalers",
	}
	cmd.AddCommand(
		newAutoscalersListCommand(cfg, creds),
		newAutoscalersDescribeCommand(cfg, creds),
		newAutoscalersCreateCommand(cfg, creds),
		newAutoscalersDeleteCommand(cfg, creds),
	)
	return cmd
}

func newAutoscalersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List autoscalers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			autoscalers, err := client.ListAutoscalers(ctx, project, zone)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), autoscalers)
			}
			headers := []string{"NAME", "ZONE", "TARGET", "MIN", "MAX", "STATUS"}
			rows := make([][]string, len(autoscalers))
			for i, as := range autoscalers {
				rows[i] = []string{as.Name, as.Zone, as.Target, fmt.Sprintf("%d", as.MinReplicas), fmt.Sprintf("%d", as.MaxReplicas), as.Status}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newAutoscalersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe AUTOSCALER",
		Short: "Describe an autoscaler",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			as, err := client.GetAutoscaler(ctx, project, zone, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), as)
		},
	}
	addZoneFlag(cmd)
	return cmd
}

func newAutoscalersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateAutoscalerRequest

	cmd := &cobra.Command{
		Use:   "create AUTOSCALER",
		Short: "Create an autoscaler",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			if req.Target == "" {
				return fmt.Errorf("--target is required")
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateAutoscaler(ctx, project, zone, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created autoscaler %q in %s/%s.\n", req.Name, project, zone)
			return nil
		},
	}
	addZoneFlag(cmd)
	cmd.Flags().StringVar(&req.Target, "target", "", "Target managed instance group")
	cmd.Flags().Int32Var(&req.MinReplicas, "min-replicas", 1, "Minimum replicas")
	cmd.Flags().Int32Var(&req.MaxReplicas, "max-replicas", 3, "Maximum replicas")
	cmd.Flags().Float64Var(&req.CpuUtilization, "cpu-utilization", 0.6, "Target CPU utilization")
	cmd.Flags().StringVar(&req.Description, "description", "", "Autoscaler description")
	return cmd
}

func newAutoscalersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete AUTOSCALER",
		Short: "Delete an autoscaler",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteAutoscaler(ctx, project, zone, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted autoscaler %q.\n", args[0])
			return nil
		},
	}
	addZoneFlag(cmd)
	return cmd
}
