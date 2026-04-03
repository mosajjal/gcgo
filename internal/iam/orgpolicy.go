package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	orgpolicyv2 "google.golang.org/api/orgpolicy/v2"
	"google.golang.org/api/option"
)

// OrgPolicy holds organization policy fields.
type OrgPolicy struct {
	Name       string `json:"name"`
	Constraint string `json:"constraint,omitempty"`
	Etag       string `json:"etag,omitempty"`
}

// OrgPoliciesClient defines org policy operations.
type OrgPoliciesClient interface {
	ListOrgPolicies(ctx context.Context, resource string) ([]*OrgPolicy, error)
	DescribeOrgPolicy(ctx context.Context, resource, constraint string) (*OrgPolicy, error)
	SetOrgPolicy(ctx context.Context, resource, constraint string, policyFile string) (*OrgPolicy, error)
	// ResetOrgPolicy deletes the policy so it inherits from the parent (effectively a reset).
	ResetOrgPolicy(ctx context.Context, resource, constraint string) error
}

type gcpOrgPolicyClient struct {
	svc *orgpolicyv2.Service
}

// NewOrgPoliciesClient creates an OrgPoliciesClient backed by the real GCP Org Policy v2 API.
func NewOrgPoliciesClient(ctx context.Context, opts ...option.ClientOption) (OrgPoliciesClient, error) {
	svc, err := orgpolicyv2.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create org policy client: %w", err)
	}
	return &gcpOrgPolicyClient{svc: svc}, nil
}

func (c *gcpOrgPolicyClient) ListOrgPolicies(ctx context.Context, resource string) ([]*OrgPolicy, error) {
	var policies []*OrgPolicy
	addPolicy := func(p *orgpolicyv2.GoogleCloudOrgpolicyV2Policy) {
		policies = append(policies, orgPolicyFromAPI(p))
	}

	switch {
	case strings.HasPrefix(resource, "organizations/"):
		err := c.svc.Organizations.Policies.List(resource).Pages(ctx,
			func(resp *orgpolicyv2.GoogleCloudOrgpolicyV2ListPoliciesResponse) error {
				for _, p := range resp.Policies {
					addPolicy(p)
				}
				return nil
			})
		if err != nil {
			return nil, fmt.Errorf("list org policies: %w", err)
		}
	case strings.HasPrefix(resource, "folders/"):
		err := c.svc.Folders.Policies.List(resource).Pages(ctx,
			func(resp *orgpolicyv2.GoogleCloudOrgpolicyV2ListPoliciesResponse) error {
				for _, p := range resp.Policies {
					addPolicy(p)
				}
				return nil
			})
		if err != nil {
			return nil, fmt.Errorf("list org policies: %w", err)
		}
	default:
		err := c.svc.Projects.Policies.List(resource).Pages(ctx,
			func(resp *orgpolicyv2.GoogleCloudOrgpolicyV2ListPoliciesResponse) error {
				for _, p := range resp.Policies {
					addPolicy(p)
				}
				return nil
			})
		if err != nil {
			return nil, fmt.Errorf("list org policies: %w", err)
		}
	}
	return policies, nil
}

func (c *gcpOrgPolicyClient) DescribeOrgPolicy(ctx context.Context, resource, constraint string) (*OrgPolicy, error) {
	name := orgPolicyName(resource, constraint)
	p, err := c.svc.Organizations.Policies.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("describe org policy %s: %w", name, err)
	}
	return orgPolicyFromAPI(p), nil
}

func (c *gcpOrgPolicyClient) SetOrgPolicy(ctx context.Context, resource, constraint string, policyFile string) (*OrgPolicy, error) {
	data, err := os.ReadFile(policyFile) //nolint:gosec // user provides path via --policy-file flag
	if err != nil {
		return nil, fmt.Errorf("read policy file: %w", err)
	}
	var policy orgpolicyv2.GoogleCloudOrgpolicyV2Policy
	if err := json.Unmarshal(data, &policy); err != nil {
		return nil, fmt.Errorf("parse policy file: %w", err)
	}
	policy.Name = orgPolicyName(resource, constraint)

	// Try patch first; if not found, create.
	p, patchErr := c.svc.Organizations.Policies.Patch(policy.Name, &policy).Context(ctx).Do()
	if patchErr != nil {
		p, err = c.svc.Organizations.Policies.Create(resource, &policy).Context(ctx).Do()
		if err != nil {
			return nil, fmt.Errorf("set org policy %s: %w", constraint, err)
		}
	}
	return orgPolicyFromAPI(p), nil
}

// ResetOrgPolicy deletes the policy so the resource inherits from its parent.
// This is the v2 equivalent of "reset to default".
func (c *gcpOrgPolicyClient) ResetOrgPolicy(ctx context.Context, resource, constraint string) error {
	name := orgPolicyName(resource, constraint)
	if _, err := c.svc.Organizations.Policies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("reset org policy %s: %w", constraint, err)
	}
	return nil
}

func orgPolicyName(resource, constraint string) string {
	return resource + "/policies/" + constraint
}

// constraintFromPolicyName extracts the constraint from a full policy name like
// "organizations/123/policies/constraints/compute.requireOsLogin".
func constraintFromPolicyName(name string) string {
	idx := strings.LastIndex(name, "/policies/")
	if idx < 0 {
		return ""
	}
	return name[idx+len("/policies/"):]
}

func orgPolicyFromAPI(p *orgpolicyv2.GoogleCloudOrgpolicyV2Policy) *OrgPolicy {
	return &OrgPolicy{
		Name:       p.Name,
		Constraint: constraintFromPolicyName(p.Name),
		Etag:       p.Etag,
	}
}
