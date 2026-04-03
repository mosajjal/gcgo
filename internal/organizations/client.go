package organizations

import (
	"context"
	"errors"
	"fmt"

	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"cloud.google.com/go/resourcemanager/apiv3/resourcemanagerpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Organization holds the fields we care about.
type Organization struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	DirectoryID string `json:"directory_customer_id"`
}

// IAMBinding holds a single IAM policy binding.
type IAMBinding struct {
	Role    string   `json:"role"`
	Members []string `json:"members"`
}

// Client defines operations for organizations.
type Client interface {
	List(ctx context.Context) ([]*Organization, error)
	Get(ctx context.Context, orgID string) (*Organization, error)
	GetIAMPolicy(ctx context.Context, orgID string) ([]*IAMBinding, error)
}

type gcpClient struct {
	rm *resourcemanager.OrganizationsClient
}

// NewClient creates a Client backed by the real GCP API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	rm, err := resourcemanager.NewOrganizationsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create organizations client: %w", err)
	}
	return &gcpClient{rm: rm}, nil
}

func (c *gcpClient) List(ctx context.Context) ([]*Organization, error) {
	it := c.rm.SearchOrganizations(ctx, &resourcemanagerpb.SearchOrganizationsRequest{})

	var orgs []*Organization
	for {
		o, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list organizations: %w", err)
		}
		orgs = append(orgs, fromProto(o))
	}
	return orgs, nil
}

func (c *gcpClient) Get(ctx context.Context, orgID string) (*Organization, error) {
	o, err := c.rm.GetOrganization(ctx, &resourcemanagerpb.GetOrganizationRequest{
		Name: "organizations/" + orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get organization %s: %w", orgID, err)
	}
	return fromProto(o), nil
}

func (c *gcpClient) GetIAMPolicy(ctx context.Context, orgID string) ([]*IAMBinding, error) {
	policy, err := c.rm.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: "organizations/" + orgID,
	})
	if err != nil {
		return nil, fmt.Errorf("get iam policy for organization %s: %w", orgID, err)
	}

	var bindings []*IAMBinding
	for _, b := range policy.GetBindings() {
		bindings = append(bindings, &IAMBinding{
			Role:    b.GetRole(),
			Members: b.GetMembers(),
		})
	}
	return bindings, nil
}

func fromProto(o *resourcemanagerpb.Organization) *Organization {
	org := &Organization{
		Name:        o.GetName(),
		DisplayName: o.GetDisplayName(),
		State:       o.GetState().String(),
	}
	if dc := o.GetDirectoryCustomerId(); dc != "" {
		org.DirectoryID = dc
	}
	return org
}
