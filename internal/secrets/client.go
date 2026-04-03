package secrets

import (
	"context"
	"encoding/base64"
	"fmt"

	"google.golang.org/api/option"
	secretmanager "google.golang.org/api/secretmanager/v1"
)

// Secret holds secret metadata.
type Secret struct {
	Name       string            `json:"name"`
	CreateTime string            `json:"create_time"`
	Labels     map[string]string `json:"labels,omitempty"`
}

// SecretVersion holds version metadata.
type SecretVersion struct {
	Name       string `json:"name"`
	State      string `json:"state"`
	CreateTime string `json:"create_time"`
}

// Client defines secret manager operations.
type Client interface {
	List(ctx context.Context, project string) ([]*Secret, error)
	Create(ctx context.Context, project, secretID string) (*Secret, error)
	Delete(ctx context.Context, name string) error
	Describe(ctx context.Context, name string) (*Secret, error)
	UpdateLabels(ctx context.Context, name string, labels map[string]string) (*Secret, error)
	GetPolicy(ctx context.Context, name string) (*secretmanager.Policy, error)
	SetPolicy(ctx context.Context, name string, policy *secretmanager.Policy) (*secretmanager.Policy, error)
	TestPermissions(ctx context.Context, name string, permissions []string) ([]string, error)
	ListVersions(ctx context.Context, secretName string) ([]*SecretVersion, error)
	DescribeVersion(ctx context.Context, versionName string) (*SecretVersion, error)
	AddVersion(ctx context.Context, secretName string, payload []byte) (*SecretVersion, error)
	AccessVersion(ctx context.Context, versionName string) ([]byte, error)
	DestroyVersion(ctx context.Context, versionName string) error
	DisableVersion(ctx context.Context, versionName string) error
	EnableVersion(ctx context.Context, versionName string) error
}

type gcpClient struct {
	sm *secretmanager.Service
}

// NewClient creates a Client backed by the real GCP Secret Manager API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := secretmanager.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create secret manager client: %w", err)
	}
	return &gcpClient{sm: svc}, nil
}

func (c *gcpClient) List(ctx context.Context, project string) ([]*Secret, error) {
	call := c.sm.Projects.Secrets.List("projects/" + project)

	var secrets []*Secret
	if err := call.Pages(ctx, func(resp *secretmanager.ListSecretsResponse) error {
		for _, s := range resp.Secrets {
			secrets = append(secrets, secretFromAPI(s))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list secrets: %w", err)
	}
	return secrets, nil
}

func (c *gcpClient) Create(ctx context.Context, project, secretID string) (*Secret, error) {
	s, err := c.sm.Projects.Secrets.Create("projects/"+project, &secretmanager.Secret{
		Replication: &secretmanager.Replication{
			Automatic: &secretmanager.Automatic{},
		},
	}).SecretId(secretID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create secret %s: %w", secretID, err)
	}
	return secretFromAPI(s), nil
}

func (c *gcpClient) Delete(ctx context.Context, name string) error {
	if _, err := c.sm.Projects.Secrets.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete secret %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) Describe(ctx context.Context, name string) (*Secret, error) {
	s, err := c.sm.Projects.Secrets.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get secret %s: %w", name, err)
	}
	return secretFromAPI(s), nil
}

func (c *gcpClient) UpdateLabels(ctx context.Context, name string, labels map[string]string) (*Secret, error) {
	s, err := c.sm.Projects.Secrets.Patch(name, &secretmanager.Secret{
		Labels: labels,
	}).UpdateMask("labels").Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("update labels for secret %s: %w", name, err)
	}
	return secretFromAPI(s), nil
}

func (c *gcpClient) GetPolicy(ctx context.Context, name string) (*secretmanager.Policy, error) {
	policy, err := c.sm.Projects.Secrets.GetIamPolicy(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get secret policy %s: %w", name, err)
	}
	return policy, nil
}

func (c *gcpClient) SetPolicy(ctx context.Context, name string, policy *secretmanager.Policy) (*secretmanager.Policy, error) {
	updated, err := c.sm.Projects.Secrets.SetIamPolicy(name, &secretmanager.SetIamPolicyRequest{Policy: policy}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("set secret policy %s: %w", name, err)
	}
	return updated, nil
}

func (c *gcpClient) TestPermissions(ctx context.Context, name string, permissions []string) ([]string, error) {
	resp, err := c.sm.Projects.Secrets.TestIamPermissions(name, &secretmanager.TestIamPermissionsRequest{Permissions: permissions}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("test secret permissions %s: %w", name, err)
	}
	return resp.Permissions, nil
}

func (c *gcpClient) ListVersions(ctx context.Context, secretName string) ([]*SecretVersion, error) {
	call := c.sm.Projects.Secrets.Versions.List(secretName)

	var versions []*SecretVersion
	if err := call.Pages(ctx, func(resp *secretmanager.ListSecretVersionsResponse) error {
		for _, v := range resp.Versions {
			versions = append(versions, versionFromAPI(v))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list secret versions: %w", err)
	}
	return versions, nil
}

func (c *gcpClient) DescribeVersion(ctx context.Context, versionName string) (*SecretVersion, error) {
	v, err := c.sm.Projects.Secrets.Versions.Get(versionName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get secret version %s: %w", versionName, err)
	}
	return versionFromAPI(v), nil
}

func (c *gcpClient) AddVersion(ctx context.Context, secretName string, payload []byte) (*SecretVersion, error) {
	v, err := c.sm.Projects.Secrets.AddVersion(secretName, &secretmanager.AddSecretVersionRequest{
		Payload: &secretmanager.SecretPayload{Data: base64.StdEncoding.EncodeToString(payload)},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("add secret version: %w", err)
	}
	return versionFromAPI(v), nil
}

func (c *gcpClient) AccessVersion(ctx context.Context, versionName string) ([]byte, error) {
	resp, err := c.sm.Projects.Secrets.Versions.Access(versionName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("access secret version %s: %w", versionName, err)
	}
	data, err := base64.StdEncoding.DecodeString(resp.Payload.Data)
	if err != nil {
		return nil, fmt.Errorf("decode secret version %s: %w", versionName, err)
	}
	return data, nil
}

func (c *gcpClient) DestroyVersion(ctx context.Context, versionName string) error {
	if _, err := c.sm.Projects.Secrets.Versions.Destroy(versionName, &secretmanager.DestroySecretVersionRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("destroy secret version %s: %w", versionName, err)
	}
	return nil
}

func (c *gcpClient) DisableVersion(ctx context.Context, versionName string) error {
	if _, err := c.sm.Projects.Secrets.Versions.Disable(versionName, &secretmanager.DisableSecretVersionRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("disable secret version %s: %w", versionName, err)
	}
	return nil
}

func (c *gcpClient) EnableVersion(ctx context.Context, versionName string) error {
	if _, err := c.sm.Projects.Secrets.Versions.Enable(versionName, &secretmanager.EnableSecretVersionRequest{}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("enable secret version %s: %w", versionName, err)
	}
	return nil
}

func secretFromAPI(s *secretmanager.Secret) *Secret {
	return &Secret{
		Name:       s.Name,
		CreateTime: s.CreateTime,
		Labels:     s.Labels,
	}
}

func versionFromAPI(v *secretmanager.SecretVersion) *SecretVersion {
	return &SecretVersion{
		Name:       v.Name,
		State:      v.State,
		CreateTime: v.CreateTime,
	}
}
