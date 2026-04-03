package compute

import (
	"testing"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/spf13/cobra"
)

func TestManagedInfraCommandTree(t *testing.T) {
	cmd := NewCommand(&config.Config{}, auth.New(""))

	want := map[string]bool{
		"instance-templates": false,
		"instance-groups":    false,
		"autoscalers":        false,
		"images":             false,
		"vpn-tunnels":        false,
	}
	for _, sub := range cmd.Commands() {
		switch sub.Name() {
		case "instance-templates":
			want["instance-templates"] = hasSubcommands(sub, "list", "describe", "create", "delete")
		case "instance-groups":
			var managedFound, unmanagedFound bool
			for _, nested := range sub.Commands() {
				if nested.Name() == "managed" {
					managedFound = hasSubcommands(nested, "list", "describe", "create", "delete")
				}
				if nested.Name() == "unmanaged" {
					unmanagedFound = hasSubcommands(nested, "list", "describe", "create", "delete")
				}
			}
			want["instance-groups"] = managedFound && unmanagedFound
		case "autoscalers":
			want["autoscalers"] = hasSubcommands(sub, "list", "describe", "create", "delete")
		case "images":
			want["images"] = hasSubcommands(sub, "list", "describe", "create", "delete")
		case "vpn-tunnels":
			want["vpn-tunnels"] = hasSubcommands(sub, "list", "describe", "create", "delete")
		}
	}

	for name, ok := range want {
		if !ok {
			t.Fatalf("missing command tree for %s", name)
		}
	}
}

func TestManagedInfraProtoConversions(t *testing.T) {
	tpl := instanceTemplateFromProto(&computepb.InstanceTemplate{
		Name:           ptr("tpl-1"),
		Description:    ptr("template"),
		Region:         ptr("global"),
		SelfLink:       ptr("https://example.invalid/templates/tpl-1"),
		SourceInstance: ptr("projects/p/zones/z/instances/base"),
		Properties: &computepb.InstanceProperties{
			MachineType: ptr("e2-medium"),
			Disks: []*computepb.AttachedDisk{
				{
					Boot: ptr(true),
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						SourceImage: ptr("projects/debian-cloud/global/images/family/debian-12"),
					},
				},
			},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Network:    ptr("global/networks/default"),
					Subnetwork: ptr("regions/us-central1/subnetworks/default"),
				},
			},
		},
	})
	if tpl.Name != "tpl-1" || tpl.MachineType != "e2-medium" || tpl.Network != "global/networks/default" || tpl.Subnetwork != "regions/us-central1/subnetworks/default" || tpl.SourceImage != "projects/debian-cloud/global/images/family/debian-12" {
		t.Fatalf("unexpected template conversion: %#v", tpl)
	}

	mig := instanceGroupManagerFromProto(&computepb.InstanceGroupManager{
		Name:             ptr("mig-1"),
		Zone:             ptr("us-central1-a"),
		Description:      ptr("group"),
		BaseInstanceName: ptr("web"),
		InstanceTemplate: ptr("projects/p/global/instanceTemplates/tpl-1"),
		TargetSize:       ptr(int32(3)),
		Status: &computepb.InstanceGroupManagerStatus{
			Autoscaler: ptr("projects/p/zones/us-central1-a/autoscalers/as-1"),
			IsStable:   ptr(true),
		},
	})
	if mig.Name != "mig-1" || mig.Status != "STABLE" || mig.Autoscaler == "" || mig.TargetSize != 3 {
		t.Fatalf("unexpected MIG conversion: %#v", mig)
	}

	as := autoscalerFromProto(&computepb.Autoscaler{
		Name:            ptr("as-1"),
		Zone:            ptr("us-central1-a"),
		Description:     ptr("autoscaler"),
		Target:          ptr("projects/p/zones/us-central1-a/instanceGroupManagers/mig-1"),
		Status:          ptr("ACTIVE"),
		RecommendedSize: ptr(int32(5)),
		AutoscalingPolicy: &computepb.AutoscalingPolicy{
			MinNumReplicas: ptr(int32(1)),
			MaxNumReplicas: ptr(int32(6)),
			CpuUtilization: &computepb.AutoscalingPolicyCpuUtilization{
				UtilizationTarget: ptr(0.75),
			},
		},
	})
	if as.Name != "as-1" || as.MinReplicas != 1 || as.MaxReplicas != 6 || as.CpuUtilization != 0.75 || as.RecommendedSize != 5 {
		t.Fatalf("unexpected autoscaler conversion: %#v", as)
	}
}

func hasSubcommands(cmd *cobra.Command, names ...string) bool {
	wanted := make(map[string]bool, len(names))
	for _, name := range names {
		wanted[name] = false
	}
	for _, sub := range cmd.Commands() {
		if _, ok := wanted[sub.Name()]; ok {
			wanted[sub.Name()] = true
		}
	}
	for _, ok := range wanted {
		if !ok {
			return false
		}
	}
	return true
}
