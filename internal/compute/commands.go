package compute

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

// NewCommand returns the compute command group.
func NewCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "compute",
		Short: "Manage Compute Engine resources",
	}

	cmd.AddCommand(
		newInstancesCommand(cfg, creds),
		newInstanceTemplatesCommand(cfg, creds),
		newInstanceGroupsCommand(cfg, creds),
		newAutoscalersCommand(cfg, creds),
		newDisksCommand(cfg, creds),
		newSnapshotsCommand(cfg, creds),
		newFirewallRulesCommand(cfg, creds),
		newNetworksCommand(cfg, creds),
		newSubnetsCommand(cfg, creds),
		newAddressesCommand(cfg, creds),
		newRoutersCommand(cfg, creds),
		newRoutesCommand(cfg, creds),
		newForwardingRulesCommand(cfg, creds),
		newBackendServicesCommand(cfg, creds),
		newHealthChecksCommand(cfg, creds),
		newUrlMapsCommand(cfg, creds),
		newTargetHttpProxiesCommand(cfg, creds),
		newTargetHttpsProxiesCommand(cfg, creds),
		newTargetTcpProxiesCommand(cfg, creds),
		newTargetSslProxiesCommand(cfg, creds),
		newImagesCommand(cfg, creds),
		newVPNTunnelsCommand(cfg, creds),
		newSSLCertificatesCommand(cfg, creds),
		newSecurityPoliciesCommand(cfg, creds),
		newZonesCommand(cfg, creds),
		newRegionsCommand(cfg, creds),
		newMachineTypesCommand(cfg, creds),
		newDiskTypesCommand(cfg, creds),
		newSSHCommand(cfg, creds),
		newSCPCommand(cfg, creds),
	)

	return cmd
}

func newDisksCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "disks",
		Short: "Manage persistent disks",
	}

	cmd.AddCommand(
		newDisksListCommand(cfg, creds),
		newDisksDescribeCommand(cfg, creds),
		newDisksCreateCommand(cfg, creds),
		newDisksDeleteCommand(cfg, creds),
	)

	return cmd
}

func newDisksListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List persistent disks",
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

			disks, err := client.ListDisks(ctx, project, zone)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), disks)
			}

			headers := []string{"NAME", "ZONE", "SIZE_GB", "TYPE", "STATUS"}
			rows := make([][]string, len(disks))
			for i, disk := range disks {
				rows[i] = []string{disk.Name, disk.Zone, fmt.Sprintf("%d", disk.SizeGb), disk.Type, disk.Status}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	AddZoneFlag(cmd)
	return cmd
}

func newDisksDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe DISK",
		Short: "Describe a persistent disk",
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
			disk, err := client.GetDisk(ctx, project, zone, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), disk)
		},
	}

	AddZoneFlag(cmd)
	return cmd
}

func newDisksCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateDiskRequest

	cmd := &cobra.Command{
		Use:   "create DISK",
		Short: "Create a persistent disk",
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
			if err := client.CreateDisk(ctx, project, zone, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created disk %q.\n", req.Name)
			return nil
		},
	}

	AddZoneFlag(cmd)
	cmd.Flags().Int64Var(&req.SizeGb, "size", 10, "Disk size in GB")
	cmd.Flags().StringVar(&req.Type, "type", "pd-balanced", "Disk type")
	cmd.Flags().StringVar(&req.ImageFamily, "image-family", "", "Optional source image family")
	cmd.Flags().StringVar(&req.ImageProject, "image-project", "", "Optional source image project")
	return cmd
}

func newDisksDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete DISK",
		Short: "Delete a persistent disk",
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
			if err := client.DeleteDisk(ctx, project, zone, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted disk %q.\n", args[0])
			return nil
		},
	}

	AddZoneFlag(cmd)
	return cmd
}

func newSnapshotsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "snapshots",
		Short: "Manage persistent disk snapshots",
	}

	cmd.AddCommand(
		newSnapshotsListCommand(cfg, creds),
		newSnapshotsDescribeCommand(cfg, creds),
		newSnapshotsCreateCommand(cfg, creds),
		newSnapshotsDeleteCommand(cfg, creds),
	)

	return cmd
}

func newSnapshotsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List snapshots",
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
			snapshots, err := client.ListSnapshots(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), snapshots)
			}
			headers := []string{"NAME", "STATUS", "SOURCE_DISK", "STORAGE_BYTES"}
			rows := make([][]string, len(snapshots))
			for i, snapshot := range snapshots {
				rows[i] = []string{snapshot.Name, snapshot.Status, snapshot.SourceDisk, fmt.Sprintf("%d", snapshot.StorageBytes)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newSnapshotsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe SNAPSHOT",
		Short: "Describe a snapshot",
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
			snapshot, err := client.GetSnapshot(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), snapshot)
		},
	}
}

func newSnapshotsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateSnapshotRequest

	cmd := &cobra.Command{
		Use:   "create SNAPSHOT",
		Short: "Create a snapshot from a disk",
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
			if req.SourceDisk == "" {
				return fmt.Errorf("--source-disk is required")
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateSnapshot(ctx, project, zone, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created snapshot %q.\n", req.Name)
			return nil
		},
	}

	AddZoneFlag(cmd)
	cmd.Flags().StringVar(&req.SourceDisk, "source-disk", "", "Source persistent disk")
	cmd.Flags().StringVar(&req.Description, "description", "", "Snapshot description")
	return cmd
}

func newSnapshotsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete SNAPSHOT",
		Short: "Delete a snapshot",
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
			if err := client.DeleteSnapshot(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted snapshot %q.\n", args[0])
			return nil
		},
	}
}

func newInstancesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "instances",
		Short: "Manage VM instances",
	}

	cmd.AddCommand(
		newInstancesListCommand(cfg, creds),
		newInstancesDescribeCommand(cfg, creds),
		newInstancesCreateCommand(cfg, creds),
		newInstancesDeleteCommand(cfg, creds),
		newInstancesStartCommand(cfg, creds),
		newInstancesStopCommand(cfg, creds),
		newInstancesResetCommand(cfg, creds),
		newInstancesAddTagsCommand(cfg, creds),
		newInstancesRemoveTagsCommand(cfg, creds),
		newInstancesSetMachineTypeCommand(cfg, creds),
		newInstancesAttachDiskCommand(cfg, creds),
		newInstancesDetachDiskCommand(cfg, creds),
	)

	return cmd
}

func newInstancesAddTagsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var tags []string

	cmd := &cobra.Command{
		Use:   "add-tags INSTANCE",
		Short: "Add network tags to a VM instance",
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
			existing, err := instanceTags(ctx, client, project, zone, args[0])
			if err != nil {
				return err
			}
			merged := mergeTags(existing, tags)
			if err := client.SetTags(ctx, project, zone, args[0], merged); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated tags on instance %q.\n", args[0])
			return nil
		},
	}
	AddZoneFlag(cmd)
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags to add")
	return cmd
}

func newInstancesRemoveTagsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var tags []string

	cmd := &cobra.Command{
		Use:   "remove-tags INSTANCE",
		Short: "Remove network tags from a VM instance",
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
			existing, err := instanceTags(ctx, client, project, zone, args[0])
			if err != nil {
				return err
			}
			updated := removeTags(existing, tags)
			if err := client.SetTags(ctx, project, zone, args[0], updated); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Updated tags on instance %q.\n", args[0])
			return nil
		},
	}
	AddZoneFlag(cmd)
	cmd.Flags().StringSliceVar(&tags, "tags", nil, "Tags to remove")
	return cmd
}

func newInstancesSetMachineTypeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var machineType string

	cmd := &cobra.Command{
		Use:   "set-machine-type INSTANCE",
		Short: "Change the machine type of a stopped VM instance",
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
			if machineType == "" {
				return fmt.Errorf("--machine-type is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.SetMachineType(ctx, project, zone, args[0], machineType); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Set machine type on instance %q to %q.\n", args[0], machineType)
			return nil
		},
	}
	AddZoneFlag(cmd)
	cmd.Flags().StringVar(&machineType, "machine-type", "", "New machine type (e.g. n1-standard-2)")
	return cmd
}

func newInstancesAttachDiskCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var disk string
	var mode string

	cmd := &cobra.Command{
		Use:   "attach-disk INSTANCE",
		Short: "Attach a persistent disk to a VM instance",
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
			if disk == "" {
				return fmt.Errorf("--disk is required")
			}
			readOnly := strings.EqualFold(mode, "ro")
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.AttachDisk(ctx, project, zone, args[0], disk, readOnly); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Attached disk %q to instance %q.\n", disk, args[0])
			return nil
		},
	}
	AddZoneFlag(cmd)
	cmd.Flags().StringVar(&disk, "disk", "", "Name of the disk to attach")
	cmd.Flags().StringVar(&mode, "mode", "rw", "Disk mode: ro (read-only) or rw (read-write)")
	return cmd
}

func newInstancesDetachDiskCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var deviceName string

	cmd := &cobra.Command{
		Use:   "detach-disk INSTANCE",
		Short: "Detach a persistent disk from a VM instance",
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
			if deviceName == "" {
				return fmt.Errorf("--device-name is required")
			}
			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DetachDisk(ctx, project, zone, args[0], deviceName); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Detached device %q from instance %q.\n", deviceName, args[0])
			return nil
		},
	}
	AddZoneFlag(cmd)
	cmd.Flags().StringVar(&deviceName, "device-name", "", "Device name of the disk to detach")
	return cmd
}

// instanceTags retrieves the current network tags for an instance.
func instanceTags(ctx context.Context, client Client, project, zone, instance string) ([]string, error) {
	inst, err := client.GetInstance(ctx, project, zone, instance)
	if err != nil {
		return nil, err
	}
	return inst.Tags, nil
}

func mergeTags(existing, add []string) []string {
	seen := make(map[string]struct{}, len(existing)+len(add))
	for _, t := range existing {
		seen[t] = struct{}{}
	}
	for _, t := range add {
		seen[t] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for t := range seen {
		out = append(out, t)
	}
	return out
}

func removeTags(existing, remove []string) []string {
	rm := make(map[string]struct{}, len(remove))
	for _, t := range remove {
		rm[t] = struct{}{}
	}
	var out []string
	for _, t := range existing {
		if _, skip := rm[t]; !skip {
			out = append(out, t)
		}
	}
	return out
}

func requireProject(cmd *cobra.Command, cfg *config.Config) (string, error) {
	flagVal, _ := cmd.Flags().GetString("project")
	project := cfg.Project(flagVal)
	if project == "" {
		return "", fmt.Errorf("no project set (use --project or 'gcgo config set project PROJECT_ID')")
	}
	return project, nil
}

func requireZone(cmd *cobra.Command, cfg *config.Config) (string, error) {
	zone, _ := cmd.Flags().GetString("zone")
	if zone == "" {
		zone = cfg.Zone()
	}
	if zone == "" {
		return "", fmt.Errorf("no zone set — use --zone ZONE, run 'gcgo config set zone ZONE' to persist, or 'gcgo compute zones list' to see available zones")
	}
	return zone, nil
}

func makeClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func newInstancesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List VM instances",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}

			zone, _ := cmd.Flags().GetString("zone")
			if zone == "" {
				zone = cfg.Zone()
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			var instances []*Instance
			if zone != "" {
				instances, err = client.ListInstances(ctx, project, zone)
			} else {
				instances, err = client.AggregatedListInstances(ctx, project)
			}
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), instances)
			}

			headers := []string{"NAME", "ZONE", "STATUS", "INTERNAL_IP", "EXTERNAL_IP"}
			rows := make([][]string, len(instances))
			for i, inst := range instances {
				rows[i] = []string{inst.Name, inst.Zone, inst.Status, inst.InternalIP, inst.ExternalIP}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}

	AddZoneFlag(cmd)

	return cmd
}

func newInstancesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe INSTANCE",
		Short: "Describe a VM instance",
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

			inst, err := client.GetInstance(ctx, project, zone, args[0])
			if err != nil {
				return err
			}

			return output.PrintJSON(cmd.OutOrStdout(), inst)
		},
	}

	AddZoneFlag(cmd)

	return cmd
}

func newInstancesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateInstanceRequest

	cmd := &cobra.Command{
		Use:   "create INSTANCE",
		Short: "Create a VM instance",
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

			if err := client.CreateInstance(ctx, project, zone, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created instance %q in %s/%s.\n", req.Name, project, zone)
			return nil
		},
	}

	AddZoneFlag(cmd)
	cmd.Flags().StringVar(&req.MachineType, "machine-type", "e2-medium", "Machine type")
	cmd.Flags().StringVar(&req.ImageFamily, "image-family", "debian-12", "Image family")
	cmd.Flags().StringVar(&req.ImageProject, "image-project", "debian-cloud", "Image project")
	cmd.Flags().Int64Var(&req.DiskSizeGB, "disk-size", 10, "Boot disk size in GB")
	cmd.Flags().StringVar(&req.Network, "network", "", "VPC network")
	cmd.Flags().StringVar(&req.Subnet, "subnet", "", "Subnet")
	cmd.Flags().StringSliceVar(&req.Tags, "tags", nil, "Network tags")

	return cmd
}

func newInstancesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete INSTANCE",
		Short: "Delete a VM instance",
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

			quiet, _ := cmd.Flags().GetBool("quiet")
			if !quiet {
				_, _ = fmt.Fprintf(os.Stderr, "Delete instance %q in %s/%s? (y/N): ", args[0], project, zone)
				reader := bufio.NewReader(os.Stdin)
				answer, _ := reader.ReadString('\n')
				if strings.TrimSpace(strings.ToLower(answer)) != "y" {
					_, _ = fmt.Fprintln(cmd.OutOrStdout(), "Aborted.")
					return nil
				}
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			if err := client.DeleteInstance(ctx, project, zone, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted instance %q.\n", args[0])
			return nil
		},
	}

	AddZoneFlag(cmd)

	return cmd
}

func newInstancesStartCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start INSTANCE",
		Short: "Start a VM instance",
		Args:  cobra.ExactArgs(1),
		RunE:  instanceLifecycleRunner(cfg, creds, "start"),
	}
	AddZoneFlag(cmd)
	return cmd
}

func newInstancesStopCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "stop INSTANCE",
		Short: "Stop a VM instance",
		Args:  cobra.ExactArgs(1),
		RunE:  instanceLifecycleRunner(cfg, creds, "stop"),
	}
	AddZoneFlag(cmd)
	return cmd
}

func newInstancesResetCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reset INSTANCE",
		Short: "Reset a VM instance",
		Args:  cobra.ExactArgs(1),
		RunE:  instanceLifecycleRunner(cfg, creds, "reset"),
	}
	AddZoneFlag(cmd)
	return cmd
}

func instanceLifecycleRunner(cfg *config.Config, creds *auth.Credentials, action string) func(*cobra.Command, []string) error {
	return func(cmd *cobra.Command, args []string) error {
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

		var opErr error
		switch action {
		case "start":
			opErr = client.StartInstance(ctx, project, zone, args[0])
		case "stop":
			opErr = client.StopInstance(ctx, project, zone, args[0])
		case "reset":
			opErr = client.ResetInstance(ctx, project, zone, args[0])
		}
		if opErr != nil {
			return opErr
		}

		_, _ = fmt.Fprintf(cmd.OutOrStdout(), "%sed instance %q.\n",
			strings.ToUpper(action[:1])+action[1:], args[0])
		return nil
	}
}

// Firewall commands

func newFirewallRulesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "firewall-rules",
		Short: "Manage firewall rules",
	}

	cmd.AddCommand(
		newFirewallListCommand(cfg, creds),
		newFirewallCreateCommand(cfg, creds),
		newFirewallDeleteCommand(cfg, creds),
	)

	return cmd
}

func newFirewallListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List firewall rules",
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

			rules, err := client.ListFirewallRules(ctx, project)
			if err != nil {
				return err
			}

			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), rules)
			}

			headers := []string{"NAME", "NETWORK", "DIRECTION", "PRIORITY", "ALLOW"}
			rows := make([][]string, len(rules))
			for i, r := range rules {
				rows[i] = []string{r.Name, r.Network, r.Direction, fmt.Sprintf("%d", r.Priority), strings.Join(r.Allowed, ",")}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newFirewallCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateFirewallRequest

	cmd := &cobra.Command{
		Use:   "create RULE",
		Short: "Create a firewall rule",
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

			if err := client.CreateFirewallRule(ctx, project, &req); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created firewall rule %q.\n", req.Name)
			return nil
		},
	}

	cmd.Flags().StringSliceVar(&req.Allow, "allow", nil, "Allowed protocols and ports (e.g. tcp:80)")
	cmd.Flags().StringSliceVar(&req.SourceRanges, "source-ranges", nil, "Source CIDR ranges")
	cmd.Flags().StringSliceVar(&req.TargetTags, "target-tags", nil, "Target tags")
	cmd.Flags().StringVar(&req.Network, "network", "", "VPC network")

	return cmd
}

func newFirewallDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete RULE",
		Short: "Delete a firewall rule",
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

			if err := client.DeleteFirewallRule(ctx, project, args[0]); err != nil {
				return err
			}

			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted firewall rule %q.\n", args[0])
			return nil
		},
	}
}

// SSH & SCP commands

func newSSHCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var user string

	cmd := &cobra.Command{
		Use:   "ssh INSTANCE [-- EXTRA_ARGS...]",
		Short: "SSH into a VM instance",
		Args:  cobra.MinimumNArgs(1),
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

			ip, err := ResolveInstanceIP(ctx, client, project, zone, args[0])
			if err != nil {
				return err
			}

			sshArgs := SSHArgs(user, ip, args[1:])
			return ExecSSH(sshArgs)
		},
	}

	AddZoneFlag(cmd)
	cmd.Flags().StringVar(&user, "user", "", "SSH username")

	return cmd
}

func newSCPCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var user string

	cmd := &cobra.Command{
		Use:   "scp SRC DST",
		Short: "Copy files to/from a VM instance",
		Long:  "Use INSTANCE:/path for remote paths. The instance name in the path is used to resolve the IP.",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			zone, err := requireZone(cmd, cfg)
			if err != nil {
				return err
			}

			// Find which arg references a remote instance
			instanceName := ""
			for _, a := range args {
				if idx := strings.Index(a, ":"); idx > 0 {
					instanceName = a[:idx]
					break
				}
			}
			if instanceName == "" {
				return fmt.Errorf("one of SRC or DST must be INSTANCE:/path")
			}

			ctx := context.Background()
			client, err := makeClient(ctx, creds)
			if err != nil {
				return err
			}

			ip, err := ResolveInstanceIP(ctx, client, project, zone, instanceName)
			if err != nil {
				return err
			}

			scpArgs := SCPArgs(user, ip, args[0], args[1])
			return ExecSCP(scpArgs)
		},
	}

	AddZoneFlag(cmd)
	cmd.Flags().StringVar(&user, "user", "", "SSH username")

	return cmd
}
