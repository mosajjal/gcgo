package iam

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	iamv2 "google.golang.org/api/iam/v2"
	"google.golang.org/api/option"
)

// DenyPolicy holds deny policy fields.
type DenyPolicy struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Etag        string `json:"etag,omitempty"`
}

// DenyPoliciesClient defines deny policy operations.
type DenyPoliciesClient interface {
	ListDenyPolicies(ctx context.Context, attachmentPoint string) ([]*DenyPolicy, error)
	DescribeDenyPolicy(ctx context.Context, name string) (*DenyPolicy, error)
	CreateDenyPolicy(ctx context.Context, attachmentPoint, displayName string, rulesFile string) (*DenyPolicy, error)
	DeleteDenyPolicy(ctx context.Context, name string) error
}

type gcpDenyClient struct {
	iam *iamv2.Service
}

// NewDenyPoliciesClient creates a DenyPoliciesClient backed by the real GCP IAM v2 API.
func NewDenyPoliciesClient(ctx context.Context, opts ...option.ClientOption) (DenyPoliciesClient, error) {
	svc, err := iamv2.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create iam v2 client: %w", err)
	}
	return &gcpDenyClient{iam: svc}, nil
}

func (c *gcpDenyClient) ListDenyPolicies(ctx context.Context, attachmentPoint string) ([]*DenyPolicy, error) {
	var policies []*DenyPolicy
	call := c.iam.Policies.ListPolicies(attachmentPoint)
	if err := call.Pages(ctx, func(resp *iamv2.GoogleIamV2ListPoliciesResponse) error {
		for _, p := range resp.Policies {
			policies = append(policies, denyPolicyFromAPI(p))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list deny policies: %w", err)
	}
	return policies, nil
}

func (c *gcpDenyClient) DescribeDenyPolicy(ctx context.Context, name string) (*DenyPolicy, error) {
	p, err := c.iam.Policies.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("describe deny policy %s: %w", name, err)
	}
	return denyPolicyFromAPI(p), nil
}

func (c *gcpDenyClient) CreateDenyPolicy(ctx context.Context, attachmentPoint, displayName string, rulesFile string) (*DenyPolicy, error) {
	policy := &iamv2.GoogleIamV2Policy{
		DisplayName: displayName,
	}

	if rulesFile != "" {
		data, err := os.ReadFile(rulesFile) //nolint:gosec // user provides path via --rules flag
		if err != nil {
			return nil, fmt.Errorf("read rules file: %w", err)
		}
		var rules []*iamv2.GoogleIamV2DenyRule
		if err := json.Unmarshal(data, &rules); err != nil {
			return nil, fmt.Errorf("parse rules file: %w", err)
		}
		policy.Rules = make([]*iamv2.GoogleIamV2PolicyRule, len(rules))
		for i, r := range rules {
			policy.Rules[i] = &iamv2.GoogleIamV2PolicyRule{
				DenyRule: r,
			}
		}
	}

	op, err := c.iam.Policies.CreatePolicy(attachmentPoint, policy).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create deny policy: %w", err)
	}
	if op.Done && op.Response != nil {
		var p iamv2.GoogleIamV2Policy
		if err := json.Unmarshal(op.Response, &p); err == nil {
			return denyPolicyFromAPI(&p), nil
		}
	}
	// Return a stub if the operation is still in-flight
	return &DenyPolicy{DisplayName: displayName}, nil
}

func (c *gcpDenyClient) DeleteDenyPolicy(ctx context.Context, name string) error {
	if _, err := c.iam.Policies.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete deny policy %s: %w", name, err)
	}
	return nil
}

func denyPolicyFromAPI(p *iamv2.GoogleIamV2Policy) *DenyPolicy {
	return &DenyPolicy{
		Name:        p.Name,
		DisplayName: p.DisplayName,
		Etag:        p.Etag,
	}
}
