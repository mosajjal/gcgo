package dns

import (
	"context"
	"fmt"

	dnsapi "google.golang.org/api/dns/v1"
	"google.golang.org/api/option"

	"github.com/mosajjal/gcgo/internal/auth"
)

// ManagedZone holds Cloud DNS managed zone fields.
type ManagedZone struct {
	Name        string `json:"name"`
	DNSName     string `json:"dns_name"`
	Description string `json:"description,omitempty"`
	Visibility  string `json:"visibility"`
	ID          uint64 `json:"id"`
}

// RecordSet holds a DNS resource record set.
type RecordSet struct {
	Name    string   `json:"name"`
	Type    string   `json:"type"`
	TTL     int64    `json:"ttl"`
	RRDatas []string `json:"rrdatas"`
}

// Policy holds a Cloud DNS policy.
type Policy struct {
	Name                    string `json:"name"`
	Description             string `json:"description,omitempty"`
	EnableInboundForwarding bool   `json:"enable_inbound_forwarding"`
	EnableLogging           bool   `json:"enable_logging"`
}

// CreateZoneRequest holds parameters for managed zone creation.
type CreateZoneRequest struct {
	Name        string
	DNSName     string
	Description string
	Visibility  string // "public" or "private"
}

// CreateRecordSetRequest holds parameters for record set creation.
type CreateRecordSetRequest struct {
	Name    string
	Type    string
	TTL     int64
	RRDatas []string
}

// CreatePolicyRequest holds parameters for policy creation.
type CreatePolicyRequest struct {
	Name                    string
	Description             string
	EnableInboundForwarding bool
	EnableLogging           bool
}

// Client defines Cloud DNS operations.
type Client interface {
	ListManagedZones(ctx context.Context, project string) ([]*ManagedZone, error)
	GetManagedZone(ctx context.Context, project, zone string) (*ManagedZone, error)
	CreateManagedZone(ctx context.Context, project string, req *CreateZoneRequest) (*ManagedZone, error)
	DeleteManagedZone(ctx context.Context, project, zone string) error

	ListRecordSets(ctx context.Context, project, zone string) ([]*RecordSet, error)
	CreateRecordSet(ctx context.Context, project, zone string, req *CreateRecordSetRequest) error
	DeleteRecordSet(ctx context.Context, project, zone, name, rrtype string) error

	ListPolicies(ctx context.Context, project string) ([]*Policy, error)
	GetPolicy(ctx context.Context, project, policy string) (*Policy, error)
	CreatePolicy(ctx context.Context, project string, req *CreatePolicyRequest) (*Policy, error)
	DeletePolicy(ctx context.Context, project, policy string) error
}

type gcpClient struct {
	svc *dnsapi.Service
}

// NewClient creates a Client backed by the real Cloud DNS API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := dnsapi.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create dns service: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func newClient(ctx context.Context, creds *auth.Credentials) (Client, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewClient(ctx, opt)
}

func (c *gcpClient) ListManagedZones(ctx context.Context, project string) ([]*ManagedZone, error) {
	resp, err := c.svc.ManagedZones.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list managed zones: %w", err)
	}
	zones := make([]*ManagedZone, len(resp.ManagedZones))
	for i, z := range resp.ManagedZones {
		zones[i] = managedZoneFromAPI(z)
	}
	return zones, nil
}

func (c *gcpClient) GetManagedZone(ctx context.Context, project, zone string) (*ManagedZone, error) {
	z, err := c.svc.ManagedZones.Get(project, zone).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get managed zone %s: %w", zone, err)
	}
	return managedZoneFromAPI(z), nil
}

func (c *gcpClient) CreateManagedZone(ctx context.Context, project string, req *CreateZoneRequest) (*ManagedZone, error) {
	visibility := req.Visibility
	if visibility == "" {
		visibility = "public"
	}
	z, err := c.svc.ManagedZones.Create(project, &dnsapi.ManagedZone{
		Name:        req.Name,
		DnsName:     req.DNSName,
		Description: req.Description,
		Visibility:  visibility,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create managed zone %s: %w", req.Name, err)
	}
	return managedZoneFromAPI(z), nil
}

func (c *gcpClient) DeleteManagedZone(ctx context.Context, project, zone string) error {
	if err := c.svc.ManagedZones.Delete(project, zone).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete managed zone %s: %w", zone, err)
	}
	return nil
}

func (c *gcpClient) ListRecordSets(ctx context.Context, project, zone string) ([]*RecordSet, error) {
	resp, err := c.svc.ResourceRecordSets.List(project, zone).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list record sets: %w", err)
	}
	sets := make([]*RecordSet, len(resp.Rrsets))
	for i, r := range resp.Rrsets {
		sets[i] = recordSetFromAPI(r)
	}
	return sets, nil
}

func (c *gcpClient) CreateRecordSet(ctx context.Context, project, zone string, req *CreateRecordSetRequest) error {
	_, err := c.svc.Changes.Create(project, zone, &dnsapi.Change{
		Additions: []*dnsapi.ResourceRecordSet{
			{
				Name:    req.Name,
				Type:    req.Type,
				Ttl:     req.TTL,
				Rrdatas: req.RRDatas,
			},
		},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create record set %s: %w", req.Name, err)
	}
	return nil
}

func (c *gcpClient) DeleteRecordSet(ctx context.Context, project, zone, name, rrtype string) error {
	// Fetch current TTL and rrdatas so the deletion change is accurate.
	existing, err := c.svc.ResourceRecordSets.List(project, zone).Context(ctx).Name(name).Type(rrtype).Do()
	if err != nil {
		return fmt.Errorf("fetch record set %s %s: %w", name, rrtype, err)
	}
	if len(existing.Rrsets) == 0 {
		return fmt.Errorf("record set %s %s not found", name, rrtype)
	}
	_, err = c.svc.Changes.Create(project, zone, &dnsapi.Change{
		Deletions: []*dnsapi.ResourceRecordSet{existing.Rrsets[0]},
	}).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("delete record set %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) ListPolicies(ctx context.Context, project string) ([]*Policy, error) {
	resp, err := c.svc.Policies.List(project).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("list policies: %w", err)
	}
	policies := make([]*Policy, len(resp.Policies))
	for i, p := range resp.Policies {
		policies[i] = policyFromAPI(p)
	}
	return policies, nil
}

func (c *gcpClient) GetPolicy(ctx context.Context, project, policy string) (*Policy, error) {
	p, err := c.svc.Policies.Get(project, policy).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get policy %s: %w", policy, err)
	}
	return policyFromAPI(p), nil
}

func (c *gcpClient) CreatePolicy(ctx context.Context, project string, req *CreatePolicyRequest) (*Policy, error) {
	p, err := c.svc.Policies.Create(project, &dnsapi.Policy{
		Name:        req.Name,
		Description: req.Description,
		EnableInboundForwarding: req.EnableInboundForwarding,
		EnableLogging:           req.EnableLogging,
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create policy %s: %w", req.Name, err)
	}
	return policyFromAPI(p), nil
}

func (c *gcpClient) DeletePolicy(ctx context.Context, project, policy string) error {
	if err := c.svc.Policies.Delete(project, policy).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete policy %s: %w", policy, err)
	}
	return nil
}

func managedZoneFromAPI(z *dnsapi.ManagedZone) *ManagedZone {
	return &ManagedZone{
		Name:        z.Name,
		DNSName:     z.DnsName,
		Description: z.Description,
		Visibility:  z.Visibility,
		ID:          z.Id,
	}
}

func recordSetFromAPI(r *dnsapi.ResourceRecordSet) *RecordSet {
	return &RecordSet{
		Name:    r.Name,
		Type:    r.Type,
		TTL:     r.Ttl,
		RRDatas: r.Rrdatas,
	}
}

func policyFromAPI(p *dnsapi.Policy) *Policy {
	return &Policy{
		Name:                    p.Name,
		Description:             p.Description,
		EnableInboundForwarding: p.EnableInboundForwarding,
		EnableLogging:           p.EnableLogging,
	}
}
