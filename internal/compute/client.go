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
	Name        string `json:"name"`
	Zone        string `json:"zone"`
	Status      string `json:"status"`
	MachineType string `json:"machine_type"`
	InternalIP  string `json:"internal_ip"`
	ExternalIP  string `json:"external_ip"`
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

type gcpClient struct {
	instances *compute.InstancesClient
	firewalls *compute.FirewallsClient
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

	return &gcpClient{instances: ic, firewalls: fc}, nil
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

func instanceFromProto(inst *computepb.Instance) *Instance {
	i := &Instance{
		Name:        inst.GetName(),
		Zone:        inst.GetZone(),
		Status:      inst.GetStatus(),
		MachineType: inst.GetMachineType(),
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
