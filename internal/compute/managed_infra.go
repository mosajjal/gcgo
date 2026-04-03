package compute

import (
	"context"
	"errors"
	"fmt"
	"strings"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

func (c *gcpClient) ListInstanceTemplates(ctx context.Context, project string) ([]*InstanceTemplate, error) {
	it := c.instanceTemplates.List(ctx, &computepb.ListInstanceTemplatesRequest{Project: project})

	var templates []*InstanceTemplate
	for {
		tpl, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list instance templates: %w", err)
		}
		templates = append(templates, instanceTemplateFromProto(tpl))
	}
	return templates, nil
}

func (c *gcpClient) GetInstanceTemplate(ctx context.Context, project, name string) (*InstanceTemplate, error) {
	tpl, err := c.instanceTemplates.Get(ctx, &computepb.GetInstanceTemplateRequest{
		Project:          project,
		InstanceTemplate: resourceName(name),
	})
	if err != nil {
		return nil, fmt.Errorf("get instance template %s: %w", name, err)
	}
	return instanceTemplateFromProto(tpl), nil
}

func (c *gcpClient) CreateInstanceTemplate(ctx context.Context, project string, req *CreateInstanceTemplateRequest) error {
	machineType := req.MachineType
	if machineType == "" {
		machineType = "e2-medium"
	}
	network := req.Network
	if network == "" {
		network = "global/networks/default"
	}
	sourceImage := ""
	if req.ImageFamily != "" && req.ImageProject != "" {
		sourceImage = fmt.Sprintf("projects/%s/global/images/family/%s", req.ImageProject, req.ImageFamily)
	}
	diskSizeGb := int64(10)
	disk := &computepb.AttachedDisk{
		AutoDelete: ptr(true),
		Boot:       ptr(true),
		Type:       ptr("PERSISTENT"),
		InitializeParams: &computepb.AttachedDiskInitializeParams{
			DiskSizeGb: &diskSizeGb,
		},
	}
	if sourceImage != "" {
		disk.InitializeParams.SourceImage = &sourceImage
	}

	template := &computepb.InstanceTemplate{
		Name:        &req.Name,
		Description: strPtrOrNil(req.Description),
		Properties: &computepb.InstanceProperties{
			MachineType: &machineType,
			Disks:       []*computepb.AttachedDisk{disk},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Network:    &network,
					Subnetwork: strPtrOrNil(req.Subnet),
					AccessConfigs: []*computepb.AccessConfig{
						{
							Name: ptr("External NAT"),
							Type: ptr("ONE_TO_ONE_NAT"),
						},
					},
				},
			},
		},
	}

	op, err := c.instanceTemplates.Insert(ctx, &computepb.InsertInstanceTemplateRequest{
		Project:                  project,
		InstanceTemplateResource: template,
	})
	if err != nil {
		return fmt.Errorf("create instance template %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteInstanceTemplate(ctx context.Context, project, name string) error {
	op, err := c.instanceTemplates.Delete(ctx, &computepb.DeleteInstanceTemplateRequest{
		Project:          project,
		InstanceTemplate: resourceName(name),
	})
	if err != nil {
		return fmt.Errorf("delete instance template %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) ListInstanceGroupManagers(ctx context.Context, project, zone string) ([]*ManagedInstanceGroup, error) {
	it := c.instanceGroups.List(ctx, &computepb.ListInstanceGroupManagersRequest{
		Project: project,
		Zone:    zone,
	})

	var groups []*ManagedInstanceGroup
	for {
		mig, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list managed instance groups: %w", err)
		}
		groups = append(groups, instanceGroupManagerFromProto(mig))
	}
	return groups, nil
}

func (c *gcpClient) GetInstanceGroupManager(ctx context.Context, project, zone, name string) (*ManagedInstanceGroup, error) {
	mig, err := c.instanceGroups.Get(ctx, &computepb.GetInstanceGroupManagerRequest{
		Project:              project,
		Zone:                 zone,
		InstanceGroupManager: resourceName(name),
	})
	if err != nil {
		return nil, fmt.Errorf("get managed instance group %s: %w", name, err)
	}
	return instanceGroupManagerFromProto(mig), nil
}

func (c *gcpClient) CreateInstanceGroupManager(ctx context.Context, project, zone string, req *CreateInstanceGroupManagerRequest) error {
	targetSize := req.TargetSize
	if targetSize <= 0 {
		targetSize = 1
	}
	group := &computepb.InstanceGroupManager{
		Name:             &req.Name,
		Description:      strPtrOrNil(req.Description),
		BaseInstanceName: strPtrOrNil(req.BaseInstanceName),
		InstanceTemplate: ptr(normalizeGlobalRef(project, "instanceTemplates", req.Template)),
		TargetSize:       &targetSize,
	}
	op, err := c.instanceGroups.Insert(ctx, &computepb.InsertInstanceGroupManagerRequest{
		Project:                      project,
		Zone:                         zone,
		InstanceGroupManagerResource: group,
	})
	if err != nil {
		return fmt.Errorf("create managed instance group %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteInstanceGroupManager(ctx context.Context, project, zone, name string) error {
	op, err := c.instanceGroups.Delete(ctx, &computepb.DeleteInstanceGroupManagerRequest{
		Project:              project,
		Zone:                 zone,
		InstanceGroupManager: resourceName(name),
	})
	if err != nil {
		return fmt.Errorf("delete managed instance group %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) ListAutoscalers(ctx context.Context, project, zone string) ([]*Autoscaler, error) {
	it := c.autoscalers.List(ctx, &computepb.ListAutoscalersRequest{
		Project: project,
		Zone:    zone,
	})

	var autoscalers []*Autoscaler
	for {
		as, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list autoscalers: %w", err)
		}
		autoscalers = append(autoscalers, autoscalerFromProto(as))
	}
	return autoscalers, nil
}

func (c *gcpClient) GetAutoscaler(ctx context.Context, project, zone, name string) (*Autoscaler, error) {
	as, err := c.autoscalers.Get(ctx, &computepb.GetAutoscalerRequest{
		Project:    project,
		Zone:       zone,
		Autoscaler: resourceName(name),
	})
	if err != nil {
		return nil, fmt.Errorf("get autoscaler %s: %w", name, err)
	}
	return autoscalerFromProto(as), nil
}

func (c *gcpClient) CreateAutoscaler(ctx context.Context, project, zone string, req *CreateAutoscalerRequest) error {
	cpuTarget := req.CpuUtilization
	if cpuTarget <= 0 {
		cpuTarget = 0.6
	}
	minReplicas := req.MinReplicas
	if minReplicas < 0 {
		minReplicas = 0
	}
	maxReplicas := req.MaxReplicas
	if maxReplicas < minReplicas {
		maxReplicas = minReplicas
	}
	as := &computepb.Autoscaler{
		Name:        &req.Name,
		Description: strPtrOrNil(req.Description),
		Target:      ptr(normalizeZonalRef(project, zone, "instanceGroupManagers", req.Target)),
		AutoscalingPolicy: &computepb.AutoscalingPolicy{
			MinNumReplicas: &minReplicas,
			MaxNumReplicas: &maxReplicas,
			CpuUtilization: &computepb.AutoscalingPolicyCpuUtilization{
				UtilizationTarget: &cpuTarget,
			},
		},
	}
	op, err := c.autoscalers.Insert(ctx, &computepb.InsertAutoscalerRequest{
		Project:            project,
		Zone:               zone,
		AutoscalerResource: as,
	})
	if err != nil {
		return fmt.Errorf("create autoscaler %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteAutoscaler(ctx context.Context, project, zone, name string) error {
	op, err := c.autoscalers.Delete(ctx, &computepb.DeleteAutoscalerRequest{
		Project:    project,
		Zone:       zone,
		Autoscaler: resourceName(name),
	})
	if err != nil {
		return fmt.Errorf("delete autoscaler %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func instanceTemplateFromProto(tpl *computepb.InstanceTemplate) *InstanceTemplate {
	out := &InstanceTemplate{
		Name:           tpl.GetName(),
		Description:    tpl.GetDescription(),
		Region:         tpl.GetRegion(),
		SelfLink:       tpl.GetSelfLink(),
		SourceInstance: tpl.GetSourceInstance(),
	}
	if props := tpl.GetProperties(); props != nil {
		out.MachineType = props.GetMachineType()
		for _, disk := range props.GetDisks() {
			if disk.GetBoot() {
				if init := disk.GetInitializeParams(); init != nil {
					out.SourceImage = init.GetSourceImage()
				}
				break
			}
		}
		for _, ni := range props.GetNetworkInterfaces() {
			if out.Network == "" {
				out.Network = ni.GetNetwork()
			}
			if out.Subnetwork == "" {
				out.Subnetwork = ni.GetSubnetwork()
			}
		}
	}
	return out
}

func instanceGroupManagerFromProto(mig *computepb.InstanceGroupManager) *ManagedInstanceGroup {
	out := &ManagedInstanceGroup{
		Name:             mig.GetName(),
		Zone:             mig.GetZone(),
		Description:      mig.GetDescription(),
		BaseInstanceName: mig.GetBaseInstanceName(),
		InstanceTemplate: mig.GetInstanceTemplate(),
		TargetSize:       mig.GetTargetSize(),
	}
	if status := mig.GetStatus(); status != nil {
		if status.GetIsStable() {
			out.Status = "STABLE"
		} else {
			out.Status = "UPDATING"
		}
		out.Autoscaler = status.GetAutoscaler()
	}
	return out
}

func autoscalerFromProto(as *computepb.Autoscaler) *Autoscaler {
	out := &Autoscaler{
		Name:            as.GetName(),
		Zone:            as.GetZone(),
		Description:     as.GetDescription(),
		Target:          as.GetTarget(),
		Status:          as.GetStatus(),
		RecommendedSize: as.GetRecommendedSize(),
	}
	if policy := as.GetAutoscalingPolicy(); policy != nil {
		out.MinReplicas = policy.GetMinNumReplicas()
		out.MaxReplicas = policy.GetMaxNumReplicas()
		if cpu := policy.GetCpuUtilization(); cpu != nil {
			out.CpuUtilization = cpu.GetUtilizationTarget()
		}
	}
	return out
}

func normalizeGlobalRef(project, resource, ref string) string {
	if strings.Contains(ref, "/") {
		return ref
	}
	return fmt.Sprintf("projects/%s/global/%s/%s", project, resource, ref)
}

func normalizeZonalRef(project, zone, resource, ref string) string {
	if strings.Contains(ref, "/") {
		return ref
	}
	return fmt.Sprintf("projects/%s/zones/%s/%s/%s", project, zone, resource, ref)
}

func resourceName(ref string) string {
	ref = strings.TrimRight(ref, "/")
	if idx := strings.LastIndex(ref, "/"); idx >= 0 {
		return ref[idx+1:]
	}
	return ref
}
