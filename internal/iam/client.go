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
	GetFolderPolicy(ctx context.Context, folder string) ([]*IAMBinding, error)
	AddFolderBinding(ctx context.Context, folder, member, role string) error
	RemoveFolderBinding(ctx context.Context, folder, member, role string) error
	TestFolderPermissions(ctx context.Context, folder string, permissions []string) ([]string, error)
	GetOrganizationPolicy(ctx context.Context, organization string) ([]*IAMBinding, error)
	AddOrganizationBinding(ctx context.Context, organization, member, role string) error
	RemoveOrganizationBinding(ctx context.Context, organization, member, role string) error
	TestOrganizationPermissions(ctx context.Context, organization string, permissions []string) ([]string, error)
	TestProjectPermissions(ctx context.Context, project string, permissions []string) ([]string, error)
}

type gcpClient struct {
	iam *iamadmin.IamClient
	rm  *resourcemanager.ProjectsClient
	fd  *resourcemanager.FoldersClient
	org *resourcemanager.OrganizationsClient
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

	fd, err := resourcemanager.NewFoldersRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create folders client: %w", err)
	}

	org, err := resourcemanager.NewOrganizationsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create organizations client: %w", err)
	}

	return &gcpClient{iam: ic, rm: rm, fd: fd, org: org}, nil
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
	policy, err := c.getProjectPolicy(ctx, project)
	if err != nil {
		return nil, fmt.Errorf("get iam policy: %w", err)
	}
	return policyBindings(policy), nil
}

func (c *gcpClient) AddBinding(ctx context.Context, project, member, role string) error {
	policy, err := c.getProjectPolicy(ctx, project)
	if err != nil {
		return fmt.Errorf("get policy: %w", err)
	}
	return c.setProjectPolicy(ctx, project, addBinding(policy, member, role))
}

func (c *gcpClient) RemoveBinding(ctx context.Context, project, member, role string) error {
	policy, err := c.getProjectPolicy(ctx, project)
	if err != nil {
		return fmt.Errorf("get policy: %w", err)
	}
	return c.setProjectPolicy(ctx, project, removeBinding(policy, member, role))
}

func (c *gcpClient) GetFolderPolicy(ctx context.Context, folder string) ([]*IAMBinding, error) {
	policy, err := c.getFolderPolicy(ctx, folder)
	if err != nil {
		return nil, fmt.Errorf("get folder iam policy: %w", err)
	}
	return policyBindings(policy), nil
}

func (c *gcpClient) AddFolderBinding(ctx context.Context, folder, member, role string) error {
	policy, err := c.getFolderPolicy(ctx, folder)
	if err != nil {
		return fmt.Errorf("get folder policy: %w", err)
	}
	return c.setFolderPolicy(ctx, folder, addBinding(policy, member, role))
}

func (c *gcpClient) RemoveFolderBinding(ctx context.Context, folder, member, role string) error {
	policy, err := c.getFolderPolicy(ctx, folder)
	if err != nil {
		return fmt.Errorf("get folder policy: %w", err)
	}
	return c.setFolderPolicy(ctx, folder, removeBinding(policy, member, role))
}

func (c *gcpClient) TestFolderPermissions(ctx context.Context, folder string, permissions []string) ([]string, error) {
	resp, err := c.fd.TestIamPermissions(ctx, &iampb.TestIamPermissionsRequest{
		Resource:    normalizeResource("folders", folder),
		Permissions: permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("test folder permissions: %w", err)
	}
	return resp.GetPermissions(), nil
}

func (c *gcpClient) GetOrganizationPolicy(ctx context.Context, organization string) ([]*IAMBinding, error) {
	policy, err := c.getOrganizationPolicy(ctx, organization)
	if err != nil {
		return nil, fmt.Errorf("get organization iam policy: %w", err)
	}
	return policyBindings(policy), nil
}

func (c *gcpClient) AddOrganizationBinding(ctx context.Context, organization, member, role string) error {
	policy, err := c.getOrganizationPolicy(ctx, organization)
	if err != nil {
		return fmt.Errorf("get organization policy: %w", err)
	}
	return c.setOrganizationPolicy(ctx, organization, addBinding(policy, member, role))
}

func (c *gcpClient) RemoveOrganizationBinding(ctx context.Context, organization, member, role string) error {
	policy, err := c.getOrganizationPolicy(ctx, organization)
	if err != nil {
		return fmt.Errorf("get organization policy: %w", err)
	}
	return c.setOrganizationPolicy(ctx, organization, removeBinding(policy, member, role))
}

func (c *gcpClient) TestOrganizationPermissions(ctx context.Context, organization string, permissions []string) ([]string, error) {
	resp, err := c.org.TestIamPermissions(ctx, &iampb.TestIamPermissionsRequest{
		Resource:    normalizeResource("organizations", organization),
		Permissions: permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("test organization permissions: %w", err)
	}
	return resp.GetPermissions(), nil
}

func (c *gcpClient) TestProjectPermissions(ctx context.Context, project string, permissions []string) ([]string, error) {
	resp, err := c.rm.TestIamPermissions(ctx, &iampb.TestIamPermissionsRequest{
		Resource:    normalizeResource("projects", project),
		Permissions: permissions,
	})
	if err != nil {
		return nil, fmt.Errorf("test project permissions: %w", err)
	}
	return resp.GetPermissions(), nil
}

func (c *gcpClient) getProjectPolicy(ctx context.Context, project string) (*iampb.Policy, error) {
	return c.rm.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: normalizeResource("projects", project),
	})
}

func (c *gcpClient) setProjectPolicy(ctx context.Context, project string, policy *iampb.Policy) error {
	_, err := c.rm.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: normalizeResource("projects", project),
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("set project policy: %w", err)
	}
	return nil
}

func (c *gcpClient) getFolderPolicy(ctx context.Context, folder string) (*iampb.Policy, error) {
	return c.fd.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: normalizeResource("folders", folder),
	})
}

func (c *gcpClient) setFolderPolicy(ctx context.Context, folder string, policy *iampb.Policy) error {
	_, err := c.fd.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: normalizeResource("folders", folder),
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("set folder policy: %w", err)
	}
	return nil
}

func (c *gcpClient) getOrganizationPolicy(ctx context.Context, organization string) (*iampb.Policy, error) {
	return c.org.GetIamPolicy(ctx, &iampb.GetIamPolicyRequest{
		Resource: normalizeResource("organizations", organization),
	})
}

func (c *gcpClient) setOrganizationPolicy(ctx context.Context, organization string, policy *iampb.Policy) error {
	_, err := c.org.SetIamPolicy(ctx, &iampb.SetIamPolicyRequest{
		Resource: normalizeResource("organizations", organization),
		Policy:   policy,
	})
	if err != nil {
		return fmt.Errorf("set organization policy: %w", err)
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

func normalizeResource(kind, value string) string {
	if len(value) >= len(kind)+1 && value[:len(kind)+1] == kind+"/" {
		return value
	}
	return kind + "/" + value
}

func policyBindings(policy *iampb.Policy) []*IAMBinding {
	var bindings []*IAMBinding
	for _, b := range policy.GetBindings() {
		bindings = append(bindings, &IAMBinding{
			Role:    b.GetRole(),
			Members: b.GetMembers(),
		})
	}
	return bindings
}

func addBinding(policy *iampb.Policy, member, role string) *iampb.Policy {
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
	return policy
}

func removeBinding(policy *iampb.Policy, member, role string) *iampb.Policy {
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
	return policy
}
