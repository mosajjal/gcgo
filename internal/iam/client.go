package iam

import (
	"context"
	"errors"
	"fmt"

	iamadmin "cloud.google.com/go/iam/admin/apiv1"
	"cloud.google.com/go/iam/admin/apiv1/adminpb"
	"cloud.google.com/go/iam/apiv1/iampb"
	resourcemanager "cloud.google.com/go/resourcemanager/apiv3"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ServiceAccount holds SA fields.
type ServiceAccount struct {
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
	UniqueID    string `json:"unique_id"`
	Disabled    bool   `json:"disabled"`
}

// SAKey holds service account key fields.
type SAKey struct {
	Name         string `json:"name"`
	ValidAfter   string `json:"valid_after"`
	ValidBefore  string `json:"valid_before"`
	KeyAlgorithm string `json:"key_algorithm"`
}

// IAMBinding holds a single IAM policy binding.
type IAMBinding struct {
	Role    string   `json:"role"`
	Members []string `json:"members"`
}

// Client defines IAM operations.
type Client interface {
	ListServiceAccounts(ctx context.Context, project string) ([]*ServiceAccount, error)
	CreateServiceAccount(ctx context.Context, project, accountID, displayName string) (*ServiceAccount, error)
	DeleteServiceAccount(ctx context.Context, email string) error
	ListKeys(ctx context.Context, email string) ([]*SAKey, error)
	CreateKey(ctx context.Context, email string) ([]byte, error)
	DeleteKey(ctx context.Context, keyName string) error
	GetPolicy(ctx context.Context, project string) ([]*IAMBinding, error)
	AddBinding(ctx context.Context, project, member, role string) error
	RemoveBinding(ctx context.Context, project, member, role string) error
}

type gcpClient struct {
	iam *iamadmin.IamClient
	rm  *resourcemanager.ProjectsClient
}

// NewClient creates a Client backed by the real GCP IAM API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	ic, err := iamadmin.NewIamClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create iam client: %w", err)
	}

	rm, err := resourcemanager.NewProjectsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create resource manager client: %w", err)
	}

	return &gcpClient{iam: ic, rm: rm}, nil
}

func (c *gcpClient) ListServiceAccounts(ctx context.Context, project string) ([]*ServiceAccount, error) {
	it := c.iam.ListServiceAccounts(ctx, &adminpb.ListServiceAccountsRequest{
		Name: "projects/" + project,
	})

	var accounts []*ServiceAccount
	for {
		sa, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list service accounts: %w", err)
		}
		accounts = append(accounts, saFromProto(sa))
	}
	return accounts, nil
}

func (c *gcpClient) CreateServiceAccount(ctx context.Context, project, accountID, displayName string) (*ServiceAccount, error) {
	sa, err := c.iam.CreateServiceAccount(ctx, &adminpb.CreateServiceAccountRequest{
		Name:      "projects/" + project,
		AccountId: accountID,
		ServiceAccount: &adminpb.ServiceAccount{
			DisplayName: displayName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create service account: %w", err)
	}
	return saFromProto(sa), nil
}

func (c *gcpClient) DeleteServiceAccount(ctx context.Context, email string) error {
	if err := c.iam.DeleteServiceAccount(ctx, &adminpb.DeleteServiceAccountRequest{
		Name: "projects/-/serviceAccounts/" + email,
	}); err != nil {
		return fmt.Errorf("delete service account %s: %w", email, err)
	}
	return nil
}

func (c *gcpClient) ListKeys(ctx context.Context, email string) ([]*SAKey, error) {
	resp, err := c.iam.ListServiceAccountKeys(ctx, &adminpb.ListServiceAccountKeysRequest{
		Name: "projects/-/serviceAccounts/" + email,
	})
	if err != nil {
		return nil, fmt.Errorf("list keys for %s: %w", email, err)
	}

	var keys []*SAKey
	for _, k := range resp.GetKeys() {
		keys = append(keys, &SAKey{
			Name:         k.GetName(),
			ValidAfter:   k.GetValidAfterTime().AsTime().String(),
			ValidBefore:  k.GetValidBeforeTime().AsTime().String(),
			KeyAlgorithm: k.GetKeyAlgorithm().String(),
		})
	}
	return keys, nil
}

func (c *gcpClient) CreateKey(ctx context.Context, email string) ([]byte, error) {
	key, err := c.iam.CreateServiceAccountKey(ctx, &adminpb.CreateServiceAccountKeyRequest{
		Name: "projects/-/serviceAccounts/" + email,
	})
	if err != nil {
		return nil, fmt.Errorf("create key for %s: %w", email, err)
	}
	return key.GetPrivateKeyData(), nil
}

func (c *gcpClient) DeleteKey(ctx context.Context, keyName string) error {
	if err := c.iam.DeleteServiceAccountKey(ctx, &adminpb.DeleteServiceAccountKeyRequest{
		Name: keyName,
	}); err != nil {
		return fmt.Errorf("delete key %s: %w", keyName, err)
	}
	return nil
}

func (c *gcpClient) GetPolicy(ctx context.Context, project string) ([]*IAMBinding, error) {
	policy, err := c.rm.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: "projects/" + project,
	})
	if err != nil {
		return nil, fmt.Errorf("get iam policy: %w", err)
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

func (c *gcpClient) AddBinding(ctx context.Context, project, member, role string) error {
	policy, err := c.rm.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: "projects/" + project,
	})
	if err != nil {
		return fmt.Errorf("get policy: %w", err)
	}

	// Find existing binding or create new one
	found := false
	for _, b := range policy.GetBindings() {
		if b.GetRole() == role {
			b.Members = append(b.Members, member)
			found = true
			break
		}
	}
	if !found {
		policy.Bindings = append(policy.Bindings, &iampb.Binding{
			Role:    role,
			Members: []string{member},
		})
	}

	_, err = c.rm.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: "projects/" + project,
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("set policy: %w", err)
	}
	return nil
}

func (c *gcpClient) RemoveBinding(ctx context.Context, project, member, role string) error {
	policy, err := c.rm.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: "projects/" + project,
	})
	if err != nil {
		return fmt.Errorf("get policy: %w", err)
	}

	for _, b := range policy.GetBindings() {
		if b.GetRole() == role {
			var filtered []string
			for _, m := range b.GetMembers() {
				if m != member {
					filtered = append(filtered, m)
				}
			}
			b.Members = filtered
			break
		}
	}

	_, err = c.rm.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: "projects/" + project,
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("set policy: %w", err)
	}
	return nil
}

func saFromProto(sa *adminpb.ServiceAccount) *ServiceAccount {
	return &ServiceAccount{
		Email:       sa.GetEmail(),
		DisplayName: sa.GetDisplayName(),
		UniqueID:    sa.GetUniqueId(),
		Disabled:    sa.GetDisabled(),
	}
}
