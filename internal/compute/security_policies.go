package compute

import (
	"context"
	"errors"
	"fmt"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

func (c *gcpClient) ListSecurityPolicies(ctx context.Context, project string) ([]*SecurityPolicy, error) {
	it := c.securityPolicies.List(ctx, &computepb.ListSecurityPoliciesRequest{
		Project: project,
	})
	var out []*SecurityPolicy
	for {
		pol, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list security policies: %w", err)
		}
		out = append(out, securityPolicyFromProto(pol))
	}
	return out, nil
}

func (c *gcpClient) GetSecurityPolicy(ctx context.Context, project, name string) (*SecurityPolicy, error) {
	pol, err := c.securityPolicies.Get(ctx, &computepb.GetSecurityPolicyRequest{
		Project:        project,
		SecurityPolicy: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get security policy %s: %w", name, err)
	}
	return securityPolicyFromProto(pol), nil
}

func (c *gcpClient) CreateSecurityPolicy(ctx context.Context, project string, req *CreateSecurityPolicyRequest) error {
	op, err := c.securityPolicies.Insert(ctx, &computepb.InsertSecurityPolicyRequest{
		Project: project,
		SecurityPolicyResource: &computepb.SecurityPolicy{
			Name:        &req.Name,
			Description: strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create security policy %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteSecurityPolicy(ctx context.Context, project, name string) error {
	op, err := c.securityPolicies.Delete(ctx, &computepb.DeleteSecurityPolicyRequest{
		Project:        project,
		SecurityPolicy: name,
	})
	if err != nil {
		return fmt.Errorf("delete security policy %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) AddSecurityPolicyRule(ctx context.Context, project, policy string, rule *SecurityPolicyRuleRequest) error {
	priority := rule.Priority
	op, err := c.securityPolicies.AddRule(ctx, &computepb.AddRuleSecurityPolicyRequest{
		Project:        project,
		SecurityPolicy: policy,
		SecurityPolicyRuleResource: &computepb.SecurityPolicyRule{
			Priority:    &priority,
			Action:      &rule.Action,
			Description: strPtrOrNil(rule.Description),
			Preview:     &rule.Preview,
			Match: &computepb.SecurityPolicyRuleMatcher{
				Config: &computepb.SecurityPolicyRuleMatcherConfig{
					SrcIpRanges: rule.SrcIPRanges,
				},
				VersionedExpr: ptr("SRC_IPS_V1"),
			},
		},
	})
	if err != nil {
		return fmt.Errorf("add rule to security policy %s: %w", policy, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) RemoveSecurityPolicyRule(ctx context.Context, project, policy string, priority int32) error {
	op, err := c.securityPolicies.RemoveRule(ctx, &computepb.RemoveRuleSecurityPolicyRequest{
		Project:        project,
		SecurityPolicy: policy,
		Priority:       ptr(int32(priority)),
	})
	if err != nil {
		return fmt.Errorf("remove rule from security policy %s: %w", policy, err)
	}
	return op.Wait(ctx)
}

func securityPolicyFromProto(pol *computepb.SecurityPolicy) *SecurityPolicy {
	return &SecurityPolicy{
		Name:        pol.GetName(),
		Description: pol.GetDescription(),
		Rules:       len(pol.GetRules()),
		Fingerprint: pol.GetFingerprint(),
	}
}
