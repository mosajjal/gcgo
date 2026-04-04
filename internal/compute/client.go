package compute

import (
	"context"
	"errors"
	"fmt"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"

	compute "cloud.google.com/go/compute/apiv1"
)

// Instance holds the fields we display.
type Instance struct {
	Name        string   `json:"name"`
	Zone        string   `json:"zone"`
	Status      string   `json:"status"`
	MachineType string   `json:"machine_type"`
	InternalIP  string   `json:"internal_ip"`
	ExternalIP  string   `json:"external_ip"`
	Tags        []string `json:"tags,omitempty"`
}

// FirewallRule holds firewall rule fields.
type FirewallRule struct {
	Name         string   `json:"name"`
	Network      string   `json:"network"`
	Direction    string   `json:"direction"`
	Priority     int64    `json:"priority"`
	Allowed      []string `json:"allowed"`
	SourceRanges []string `json:"source_ranges"`
}

// Disk holds persistent disk fields.
type Disk struct {
	Name        string `json:"name"`
	Zone        string `json:"zone"`
	SizeGb      int64  `json:"size_gb"`
	Type        string `json:"type"`
	Status      string `json:"status"`
	SourceImage string `json:"source_image,omitempty"`
}

// Snapshot holds snapshot fields.
type Snapshot struct {
	Name         string `json:"name"`
	Status       string `json:"status"`
	SourceDisk   string `json:"source_disk,omitempty"`
	StorageBytes int64  `json:"storage_bytes"`
}

// Image holds compute image fields.
type Image struct {
	Name        string `json:"name"`
	Family      string `json:"family,omitempty"`
	Status      string `json:"status"`
	DiskSizeGb  int64  `json:"disk_size_gb"`
	Description string `json:"description,omitempty"`
	SelfLink    string `json:"self_link,omitempty"`
}

// VPNTunnel holds VPN tunnel fields.
type VPNTunnel struct {
	Name        string `json:"name"`
	Region      string `json:"region"`
	Status      string `json:"status"`
	PeerIP      string `json:"peer_ip,omitempty"`
	IKEVersion  int32  `json:"ike_version,omitempty"`
	Description string `json:"description,omitempty"`
}

// UnmanagedInstanceGroup holds unmanaged instance group fields.
type UnmanagedInstanceGroup struct {
	Name        string `json:"name"`
	Zone        string `json:"zone"`
	Size        int32  `json:"size"`
	Network     string `json:"network,omitempty"`
	Description string `json:"description,omitempty"`
}

// InstanceTemplate holds instance template fields.
type InstanceTemplate struct {
	Name           string `json:"name"`
	Description    string `json:"description,omitempty"`
	MachineType    string `json:"machine_type,omitempty"`
	Network        string `json:"network,omitempty"`
	Subnetwork     string `json:"subnetwork,omitempty"`
	SourceImage    string `json:"source_image,omitempty"`
	Region         string `json:"region,omitempty"`
	SelfLink       string `json:"self_link,omitempty"`
	SourceInstance string `json:"source_instance,omitempty"`
}

// ManagedInstanceGroup holds managed instance group fields.
type ManagedInstanceGroup struct {
	Name             string `json:"name"`
	Zone             string `json:"zone,omitempty"`
	Description      string `json:"description,omitempty"`
	BaseInstanceName string `json:"base_instance_name,omitempty"`
	InstanceTemplate string `json:"instance_template,omitempty"`
	TargetSize       int32  `json:"target_size"`
	Status           string `json:"status,omitempty"`
	Autoscaler       string `json:"autoscaler,omitempty"`
}

// Autoscaler holds autoscaler fields.
type Autoscaler struct {
	Name            string  `json:"name"`
	Zone            string  `json:"zone,omitempty"`
	Description     string  `json:"description,omitempty"`
	Target          string  `json:"target,omitempty"`
	MinReplicas     int32   `json:"min_replicas"`
	MaxReplicas     int32   `json:"max_replicas"`
	CpuUtilization  float64 `json:"cpu_utilization,omitempty"`
	Status          string  `json:"status,omitempty"`
	RecommendedSize int32   `json:"recommended_size,omitempty"`
}

// Client defines compute operations.
type Client interface {
	ListInstances(ctx context.Context, project, zone string) ([]*Instance, error)
	GetInstance(ctx context.Context, project, zone, name string) (*Instance, error)
	CreateInstance(ctx context.Context, project, zone string, req *CreateInstanceRequest) error
	DeleteInstance(ctx context.Context, project, zone, name string) error
	StartInstance(ctx context.Context, project, zone, name string) error
	StopInstance(ctx context.Context, project, zone, name string) error
	ResetInstance(ctx context.Context, project, zone, name string) error
	ListFirewallRules(ctx context.Context, project string) ([]*FirewallRule, error)
	CreateFirewallRule(ctx context.Context, project string, req *CreateFirewallRequest) error
	DeleteFirewallRule(ctx context.Context, project, name string) error
	ListDisks(ctx context.Context, project, zone string) ([]*Disk, error)
	GetDisk(ctx context.Context, project, zone, name string) (*Disk, error)
	CreateDisk(ctx context.Context, project, zone string, req *CreateDiskRequest) error
	DeleteDisk(ctx context.Context, project, zone, name string) error
	ListSnapshots(ctx context.Context, project string) ([]*Snapshot, error)
	GetSnapshot(ctx context.Context, project, name string) (*Snapshot, error)
	CreateSnapshot(ctx context.Context, project, zone string, req *CreateSnapshotRequest) error
	DeleteSnapshot(ctx context.Context, project, name string) error
	ListInstanceTemplates(ctx context.Context, project string) ([]*InstanceTemplate, error)
	GetInstanceTemplate(ctx context.Context, project, name string) (*InstanceTemplate, error)
	CreateInstanceTemplate(ctx context.Context, project string, req *CreateInstanceTemplateRequest) error
	DeleteInstanceTemplate(ctx context.Context, project, name string) error
	ListInstanceGroupManagers(ctx context.Context, project, zone string) ([]*ManagedInstanceGroup, error)
	GetInstanceGroupManager(ctx context.Context, project, zone, name string) (*ManagedInstanceGroup, error)
	CreateInstanceGroupManager(ctx context.Context, project, zone string, req *CreateInstanceGroupManagerRequest) error
	DeleteInstanceGroupManager(ctx context.Context, project, zone, name string) error
	ListAutoscalers(ctx context.Context, project, zone string) ([]*Autoscaler, error)
	GetAutoscaler(ctx context.Context, project, zone, name string) (*Autoscaler, error)
	CreateAutoscaler(ctx context.Context, project, zone string, req *CreateAutoscalerRequest) error
	DeleteAutoscaler(ctx context.Context, project, zone, name string) error
	ListImages(ctx context.Context, project string) ([]*Image, error)
	GetImage(ctx context.Context, project, name string) (*Image, error)
	CreateImage(ctx context.Context, project string, req *CreateImageRequest) error
	DeleteImage(ctx context.Context, project, name string) error
	ListVPNTunnels(ctx context.Context, project, region string) ([]*VPNTunnel, error)
	GetVPNTunnel(ctx context.Context, project, region, name string) (*VPNTunnel, error)
	CreateVPNTunnel(ctx context.Context, project, region string, req *CreateVPNTunnelRequest) error
	DeleteVPNTunnel(ctx context.Context, project, region, name string) error
	ListUnmanagedInstanceGroups(ctx context.Context, project, zone string) ([]*UnmanagedInstanceGroup, error)
	GetUnmanagedInstanceGroup(ctx context.Context, project, zone, name string) (*UnmanagedInstanceGroup, error)
	CreateUnmanagedInstanceGroup(ctx context.Context, project, zone string, req *CreateUnmanagedInstanceGroupRequest) error
	DeleteUnmanagedInstanceGroup(ctx context.Context, project, zone, name string) error
	SetTags(ctx context.Context, project, zone, instance string, tags []string) error
	SetMachineType(ctx context.Context, project, zone, instance, machineType string) error
	AttachDisk(ctx context.Context, project, zone, instance, diskName string, readOnly bool) error
	DetachDisk(ctx context.Context, project, zone, instance, deviceName string) error
	ListSSLCertificates(ctx context.Context, project string) ([]*SSLCertificate, error)
	GetSSLCertificate(ctx context.Context, project, name string) (*SSLCertificate, error)
	CreateSSLCertificate(ctx context.Context, project string, req *CreateSSLCertificateRequest) error
	DeleteSSLCertificate(ctx context.Context, project, name string) error
	ListSecurityPolicies(ctx context.Context, project string) ([]*SecurityPolicy, error)
	GetSecurityPolicy(ctx context.Context, project, name string) (*SecurityPolicy, error)
	CreateSecurityPolicy(ctx context.Context, project string, req *CreateSecurityPolicyRequest) error
	DeleteSecurityPolicy(ctx context.Context, project, name string) error
	AddSecurityPolicyRule(ctx context.Context, project, policy string, rule *SecurityPolicyRuleRequest) error
	RemoveSecurityPolicyRule(ctx context.Context, project, policy string, priority int32) error
	ListZones(ctx context.Context, project, region string) ([]*Zone, error)
	ListRegions(ctx context.Context, project string) ([]*Region, error)
	ListMachineTypes(ctx context.Context, project, zone string) ([]*MachineType, error)
	AggregatedListInstances(ctx context.Context, project string) ([]*Instance, error)
	ListDiskTypes(ctx context.Context, project, zone string) ([]*DiskType, error)
}

// CreateInstanceRequest holds parameters for instance creation.
type CreateInstanceRequest struct {
	Name         string
	MachineType  string
	ImageFamily  string
	ImageProject string
	DiskSizeGB   int64
	Network      string
	Subnet       string
	Tags         []string
}

// CreateFirewallRequest holds parameters for firewall rule creation.
type CreateFirewallRequest struct {
	Name         string
	Network      string
	Allow        []string // e.g. "tcp:80", "udp:53"
	SourceRanges []string
	TargetTags   []string
}

// CreateDiskRequest holds parameters for disk creation.
type CreateDiskRequest struct {
	Name         string
	SizeGb       int64
	Type         string
	ImageFamily  string
	ImageProject string
}

// CreateSnapshotRequest holds parameters for snapshot creation.
type CreateSnapshotRequest struct {
	Name        string
	SourceDisk  string
	Description string
}

// CreateInstanceTemplateRequest holds parameters for instance template creation.
type CreateInstanceTemplateRequest struct {
	Name         string
	MachineType  string
	ImageFamily  string
	ImageProject string
	Network      string
	Subnet       string
	Description  string
}

// CreateInstanceGroupManagerRequest holds parameters for managed instance group creation.
type CreateInstanceGroupManagerRequest struct {
	Name             string
	Template         string
	BaseInstanceName string
	TargetSize       int32
	Description      string
}

// CreateAutoscalerRequest holds parameters for autoscaler creation.
type CreateAutoscalerRequest struct {
	Name           string
	Target         string
	MinReplicas    int32
	MaxReplicas    int32
	CpuUtilization float64
	Description    string
}

// CreateImageRequest holds parameters for image creation.
type CreateImageRequest struct {
	Name        string
	SourceDisk  string
	Family      string
	Description string
}

// CreateVPNTunnelRequest holds parameters for VPN tunnel creation.
type CreateVPNTunnelRequest struct {
	Name              string
	PeerIP            string
	SharedSecret      string
	VPNGateway        string
	IKEVersion        int32
	Description       string
}

// CreateUnmanagedInstanceGroupRequest holds parameters for unmanaged instance group creation.
type CreateUnmanagedInstanceGroupRequest struct {
	Name        string
	Network     string
	Description string
}

// Zone holds compute zone fields.
type Zone struct {
	Name   string `json:"name"`
	Region string `json:"region"`
	Status string `json:"status"`
}

// Region holds compute region fields.
type Region struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Zones  []string `json:"zones,omitempty"`
}

// MachineType holds machine type fields.
type MachineType struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	VCPUs       int32  `json:"vcpus"`
	MemoryMb    int32  `json:"memory_mb"`
	Zone        string `json:"zone"`
}

// DiskType holds persistent disk type fields.
type DiskType struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Zone        string `json:"zone"`
}

// SSLCertificate holds SSL certificate fields.
type SSLCertificate struct {
	Name        string   `json:"name"`
	Type        string   `json:"type"`
	Status      string   `json:"status,omitempty"`
	Domains     []string `json:"domains,omitempty"`
	Description string   `json:"description,omitempty"`
}

// CreateSSLCertificateRequest holds parameters for SSL certificate creation.
type CreateSSLCertificateRequest struct {
	Name        string
	Domains     []string // managed cert
	CertFile    string   // self-managed: path to PEM cert
	KeyFile     string   // self-managed: path to PEM key
	Description string
}

// SecurityPolicy holds security policy fields.
type SecurityPolicy struct {
	Name        string `json:"name"`
	Description string `json:"description,omitempty"`
	Rules       int    `json:"rules"`
	Fingerprint string `json:"fingerprint,omitempty"`
}

// SecurityPolicyRule holds security policy rule fields.
type SecurityPolicyRule struct {
	Priority    int32    `json:"priority"`
	Action      string   `json:"action"`
	Description string   `json:"description,omitempty"`
	SrcIPRanges []string `json:"src_ip_ranges,omitempty"`
	Preview     bool     `json:"preview"`
}

// CreateSecurityPolicyRequest holds parameters for security policy creation.
type CreateSecurityPolicyRequest struct {
	Name        string
	Description string
}

// SecurityPolicyRuleRequest holds parameters for adding a security policy rule.
type SecurityPolicyRuleRequest struct {
	Priority    int32
	Action      string
	Description string
	SrcIPRanges []string
	Preview     bool
}

type gcpClient struct {
	instances               *compute.InstancesClient
	firewalls               *compute.FirewallsClient
	disks                   *compute.DisksClient
	snapshots               *compute.SnapshotsClient
	instanceTemplates       *compute.InstanceTemplatesClient
	instanceGroups          *compute.InstanceGroupManagersClient
	autoscalers             *compute.AutoscalersClient
	images                  *compute.ImagesClient
	vpnTunnels              *compute.VpnTunnelsClient
	unmanagedInstanceGroups *compute.InstanceGroupsClient
	sslCertificates         *compute.SslCertificatesClient
	securityPolicies        *compute.SecurityPoliciesClient
	zones                   *compute.ZonesClient
	regions                 *compute.RegionsClient
	machineTypes            *compute.MachineTypesClient
	diskTypes               *compute.DiskTypesClient
}

// NewClient creates a Client backed by the real GCP Compute API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	ic, err := compute.NewInstancesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create instances client: %w", err)
	}

	fc, err := compute.NewFirewallsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create firewalls client: %w", err)
	}
	dc, err := compute.NewDisksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create disks client: %w", err)
	}
	sc, err := compute.NewSnapshotsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create snapshots client: %w", err)
	}
	itc, err := compute.NewInstanceTemplatesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create instance templates client: %w", err)
	}
	igc, err := compute.NewInstanceGroupManagersRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create instance group managers client: %w", err)
	}
	ac, err := compute.NewAutoscalersRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create autoscalers client: %w", err)
	}
	imgc, err := compute.NewImagesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create images client: %w", err)
	}
	vpnc, err := compute.NewVpnTunnelsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create vpn tunnels client: %w", err)
	}
	uigc, err := compute.NewInstanceGroupsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create instance groups client: %w", err)
	}
	sslc, err := compute.NewSslCertificatesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create ssl certificates client: %w", err)
	}
	spc, err := compute.NewSecurityPoliciesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create security policies client: %w", err)
	}
	zc, err := compute.NewZonesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create zones client: %w", err)
	}
	rc, err := compute.NewRegionsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create regions client: %w", err)
	}
	mtc, err := compute.NewMachineTypesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create machine types client: %w", err)
	}
	dtc, err := compute.NewDiskTypesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create disk types client: %w", err)
	}

	return &gcpClient{
		instances:               ic,
		firewalls:               fc,
		disks:                   dc,
		snapshots:               sc,
		instanceTemplates:       itc,
		instanceGroups:          igc,
		autoscalers:             ac,
		images:                  imgc,
		vpnTunnels:              vpnc,
		unmanagedInstanceGroups: uigc,
		sslCertificates:         sslc,
		securityPolicies:        spc,
		zones:                   zc,
		regions:                 rc,
		machineTypes:            mtc,
		diskTypes:               dtc,
	}, nil
}

func (c *gcpClient) ListInstances(ctx context.Context, project, zone string) ([]*Instance, error) {
	it := c.instances.List(ctx, &computepb.ListInstancesRequest{
		Project: project,
		Zone:    zone,
	})

	var instances []*Instance
	for {
		inst, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list instances: %w", err)
		}
		instances = append(instances, instanceFromProto(inst))
	}
	return instances, nil
}

func (c *gcpClient) GetInstance(ctx context.Context, project, zone, name string) (*Instance, error) {
	inst, err := c.instances.Get(ctx, &computepb.GetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get instance %s: %w", name, err)
	}
	return instanceFromProto(inst), nil
}

func (c *gcpClient) CreateInstance(ctx context.Context, project, zone string, req *CreateInstanceRequest) error {
	machineType := fmt.Sprintf("zones/%s/machineTypes/%s", zone, req.MachineType)
	sourceImage := fmt.Sprintf("projects/%s/global/images/family/%s", req.ImageProject, req.ImageFamily)

	diskSizeGB := req.DiskSizeGB
	if diskSizeGB == 0 {
		diskSizeGB = 10
	}

	pbReq := &computepb.InsertInstanceRequest{
		Project: project,
		Zone:    zone,
		InstanceResource: &computepb.Instance{
			Name:        &req.Name,
			MachineType: &machineType,
			Disks: []*computepb.AttachedDisk{
				{
					AutoDelete: ptr(true),
					Boot:       ptr(true),
					InitializeParams: &computepb.AttachedDiskInitializeParams{
						SourceImage: &sourceImage,
						DiskSizeGb:  &diskSizeGB,
					},
				},
			},
			NetworkInterfaces: []*computepb.NetworkInterface{
				{
					Network:    strPtrOrNil(req.Network),
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

	if len(req.Tags) > 0 {
		pbReq.InstanceResource.Tags = &computepb.Tags{
			Items: req.Tags,
		}
	}

	op, err := c.instances.Insert(ctx, pbReq)
	if err != nil {
		return fmt.Errorf("create instance %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteInstance(ctx context.Context, project, zone, name string) error {
	op, err := c.instances.Delete(ctx, &computepb.DeleteInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: name,
	})
	if err != nil {
		return fmt.Errorf("delete instance %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) StartInstance(ctx context.Context, project, zone, name string) error {
	op, err := c.instances.Start(ctx, &computepb.StartInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: name,
	})
	if err != nil {
		return fmt.Errorf("start instance %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) StopInstance(ctx context.Context, project, zone, name string) error {
	op, err := c.instances.Stop(ctx, &computepb.StopInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: name,
	})
	if err != nil {
		return fmt.Errorf("stop instance %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) ResetInstance(ctx context.Context, project, zone, name string) error {
	op, err := c.instances.Reset(ctx, &computepb.ResetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: name,
	})
	if err != nil {
		return fmt.Errorf("reset instance %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) ListFirewallRules(ctx context.Context, project string) ([]*FirewallRule, error) {
	it := c.firewalls.List(ctx, &computepb.ListFirewallsRequest{
		Project: project,
	})

	var rules []*FirewallRule
	for {
		fw, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list firewall rules: %w", err)
		}
		rules = append(rules, firewallFromProto(fw))
	}
	return rules, nil
}

func (c *gcpClient) CreateFirewallRule(ctx context.Context, project string, req *CreateFirewallRequest) error {
	network := req.Network
	if network == "" {
		network = "global/networks/default"
	}

	var allowed []*computepb.Allowed
	for _, a := range req.Allow {
		allowed = append(allowed, parseAllowed(a))
	}

	op, err := c.firewalls.Insert(ctx, &computepb.InsertFirewallRequest{
		Project: project,
		FirewallResource: &computepb.Firewall{
			Name:         &req.Name,
			Network:      &network,
			Allowed:      allowed,
			SourceRanges: req.SourceRanges,
			TargetTags:   req.TargetTags,
		},
	})
	if err != nil {
		return fmt.Errorf("create firewall rule %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteFirewallRule(ctx context.Context, project, name string) error {
	op, err := c.firewalls.Delete(ctx, &computepb.DeleteFirewallRequest{
		Project:  project,
		Firewall: name,
	})
	if err != nil {
		return fmt.Errorf("delete firewall rule %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) ListDisks(ctx context.Context, project, zone string) ([]*Disk, error) {
	it := c.disks.List(ctx, &computepb.ListDisksRequest{
		Project: project,
		Zone:    zone,
	})

	var disks []*Disk
	for {
		disk, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list disks: %w", err)
		}
		disks = append(disks, diskFromProto(disk))
	}
	return disks, nil
}

func (c *gcpClient) GetDisk(ctx context.Context, project, zone, name string) (*Disk, error) {
	disk, err := c.disks.Get(ctx, &computepb.GetDiskRequest{
		Project: project,
		Zone:    zone,
		Disk:    name,
	})
	if err != nil {
		return nil, fmt.Errorf("get disk %s: %w", name, err)
	}
	return diskFromProto(disk), nil
}

func (c *gcpClient) CreateDisk(ctx context.Context, project, zone string, req *CreateDiskRequest) error {
	sizeGb := req.SizeGb
	if sizeGb == 0 {
		sizeGb = 10
	}
	diskType := req.Type
	if diskType == "" {
		diskType = "pd-balanced"
	}
	diskTypeURL := fmt.Sprintf("zones/%s/diskTypes/%s", zone, diskType)

	pbReq := &computepb.InsertDiskRequest{
		Project: project,
		Zone:    zone,
		DiskResource: &computepb.Disk{
			Name:   &req.Name,
			SizeGb: &sizeGb,
			Type:   &diskTypeURL,
		},
	}
	if req.ImageFamily != "" && req.ImageProject != "" {
		sourceImage := fmt.Sprintf("projects/%s/global/images/family/%s", req.ImageProject, req.ImageFamily)
		pbReq.SourceImage = &sourceImage
	}

	op, err := c.disks.Insert(ctx, pbReq)
	if err != nil {
		return fmt.Errorf("create disk %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteDisk(ctx context.Context, project, zone, name string) error {
	op, err := c.disks.Delete(ctx, &computepb.DeleteDiskRequest{
		Project: project,
		Zone:    zone,
		Disk:    name,
	})
	if err != nil {
		return fmt.Errorf("delete disk %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) ListSnapshots(ctx context.Context, project string) ([]*Snapshot, error) {
	it := c.snapshots.List(ctx, &computepb.ListSnapshotsRequest{Project: project})

	var snapshots []*Snapshot
	for {
		snapshot, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list snapshots: %w", err)
		}
		snapshots = append(snapshots, snapshotFromProto(snapshot))
	}
	return snapshots, nil
}

func (c *gcpClient) GetSnapshot(ctx context.Context, project, name string) (*Snapshot, error) {
	snapshot, err := c.snapshots.Get(ctx, &computepb.GetSnapshotRequest{
		Project:  project,
		Snapshot: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get snapshot %s: %w", name, err)
	}
	return snapshotFromProto(snapshot), nil
}

func (c *gcpClient) CreateSnapshot(ctx context.Context, project, zone string, req *CreateSnapshotRequest) error {
	op, err := c.disks.CreateSnapshot(ctx, &computepb.CreateSnapshotDiskRequest{
		Project: project,
		Zone:    zone,
		Disk:    req.SourceDisk,
		SnapshotResource: &computepb.Snapshot{
			Name:        &req.Name,
			Description: strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create snapshot %s from disk %s: %w", req.Name, req.SourceDisk, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteSnapshot(ctx context.Context, project, name string) error {
	op, err := c.snapshots.Delete(ctx, &computepb.DeleteSnapshotRequest{
		Project:  project,
		Snapshot: name,
	})
	if err != nil {
		return fmt.Errorf("delete snapshot %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func instanceFromProto(inst *computepb.Instance) *Instance {
	i := &Instance{
		Name:        inst.GetName(),
		Zone:        inst.GetZone(),
		Status:      inst.GetStatus(),
		MachineType: inst.GetMachineType(),
		Tags:        inst.GetTags().GetItems(),
	}

	for _, ni := range inst.GetNetworkInterfaces() {
		if ip := ni.GetNetworkIP(); ip != "" {
			i.InternalIP = ip
		}
		for _, ac := range ni.GetAccessConfigs() {
			if ip := ac.GetNatIP(); ip != "" {
				i.ExternalIP = ip
			}
		}
	}

	return i
}

func firewallFromProto(fw *computepb.Firewall) *FirewallRule {
	var allowed []string
	for _, a := range fw.GetAllowed() {
		proto := a.GetIPProtocol()
		for _, port := range a.GetPorts() {
			allowed = append(allowed, proto+":"+port)
		}
		if len(a.GetPorts()) == 0 {
			allowed = append(allowed, proto)
		}
	}

	return &FirewallRule{
		Name:         fw.GetName(),
		Network:      fw.GetNetwork(),
		Direction:    fw.GetDirection(),
		Priority:     int64(fw.GetPriority()),
		Allowed:      allowed,
		SourceRanges: fw.GetSourceRanges(),
	}
}

func diskFromProto(disk *computepb.Disk) *Disk {
	return &Disk{
		Name:        disk.GetName(),
		Zone:        disk.GetZone(),
		SizeGb:      disk.GetSizeGb(),
		Type:        disk.GetType(),
		Status:      disk.GetStatus(),
		SourceImage: disk.GetSourceImage(),
	}
}

func snapshotFromProto(snapshot *computepb.Snapshot) *Snapshot {
	return &Snapshot{
		Name:         snapshot.GetName(),
		Status:       snapshot.GetStatus(),
		SourceDisk:   snapshot.GetSourceDisk(),
		StorageBytes: snapshot.GetStorageBytes(),
	}
}

// parseAllowed parses "tcp:80" or "icmp" into an Allowed proto.
func parseAllowed(s string) *computepb.Allowed {
	for i, c := range s {
		if c == ':' {
			proto := s[:i]
			port := s[i+1:]
			return &computepb.Allowed{
				IPProtocol: &proto,
				Ports:      []string{port},
			}
		}
	}
	return &computepb.Allowed{
		IPProtocol: &s,
	}
}

func ptr[T any](v T) *T {
	return &v
}

func strPtrOrNil(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (c *gcpClient) SetTags(ctx context.Context, project, zone, instance string, tags []string) error {
	// Fetch the instance first to get the current tag fingerprint.
	inst, err := c.instances.Get(ctx, &computepb.GetInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instance,
	})
	if err != nil {
		return fmt.Errorf("get instance %s for set-tags: %w", instance, err)
	}
	fingerprint := inst.GetTags().GetFingerprint()
	op, err := c.instances.SetTags(ctx, &computepb.SetTagsInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instance,
		TagsResource: &computepb.Tags{
			Items:       tags,
			Fingerprint: &fingerprint,
		},
	})
	if err != nil {
		return fmt.Errorf("set tags on instance %s: %w", instance, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) SetMachineType(ctx context.Context, project, zone, instance, machineType string) error {
	machineTypeURL := fmt.Sprintf("zones/%s/machineTypes/%s", zone, machineType)
	op, err := c.instances.SetMachineType(ctx, &computepb.SetMachineTypeInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instance,
		InstancesSetMachineTypeRequestResource: &computepb.InstancesSetMachineTypeRequest{
			MachineType: &machineTypeURL,
		},
	})
	if err != nil {
		return fmt.Errorf("set machine type on instance %s: %w", instance, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) AttachDisk(ctx context.Context, project, zone, instance, diskName string, readOnly bool) error {
	source := fmt.Sprintf("zones/%s/disks/%s", zone, diskName)
	mode := "READ_WRITE"
	if readOnly {
		mode = "READ_ONLY"
	}
	op, err := c.instances.AttachDisk(ctx, &computepb.AttachDiskInstanceRequest{
		Project:  project,
		Zone:     zone,
		Instance: instance,
		AttachedDiskResource: &computepb.AttachedDisk{
			Source: &source,
			Mode:   &mode,
		},
	})
	if err != nil {
		return fmt.Errorf("attach disk %s to instance %s: %w", diskName, instance, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DetachDisk(ctx context.Context, project, zone, instance, deviceName string) error {
	op, err := c.instances.DetachDisk(ctx, &computepb.DetachDiskInstanceRequest{
		Project:    project,
		Zone:       zone,
		Instance:   instance,
		DeviceName: deviceName,
	})
	if err != nil {
		return fmt.Errorf("detach disk %s from instance %s: %w", deviceName, instance, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) AggregatedListInstances(ctx context.Context, project string) ([]*Instance, error) {
	it := c.instances.AggregatedList(ctx, &computepb.AggregatedListInstancesRequest{
		Project: project,
	})
	var instances []*Instance
	for {
		pair, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("aggregated list instances: %w", err)
		}
		for _, inst := range pair.Value.GetInstances() {
			instances = append(instances, instanceFromProto(inst))
		}
	}
	return instances, nil
}

func (c *gcpClient) ListDiskTypes(ctx context.Context, project, zone string) ([]*DiskType, error) {
	it := c.diskTypes.List(ctx, &computepb.ListDiskTypesRequest{
		Project: project,
		Zone:    zone,
	})
	var out []*DiskType
	for {
		dt, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list disk types: %w", err)
		}
		out = append(out, &DiskType{
			Name:        dt.GetName(),
			Description: dt.GetDescription(),
			Zone:        dt.GetZone(),
		})
	}
	return out, nil
}
