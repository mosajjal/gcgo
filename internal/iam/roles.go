package iam

import (
	"context"
	"fmt"

	iamv1 "google.golang.org/api/iam/v1"
	"google.golang.org/api/option"
)

// Role holds IAM role fields.
type Role struct {
	Name        string `json:"name"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Stage       string `json:"stage"`
	Deleted     bool   `json:"deleted,omitempty"`
}

// RolesClient defines IAM role operations.
type RolesClient interface {
	ListRoles(ctx context.Context, project string) ([]*Role, error)
	CreateRole(ctx context.Context, project, roleID, title, description string, permissions []string) (*Role, error)
	DeleteRole(ctx context.Context, name string) error
	DescribeRole(ctx context.Context, name string) (*Role, error)
	UpdateRole(ctx context.Context, name, title, description string, permissions []string) (*Role, error)
}

type gcpRolesClient struct {
	iam *iamv1.Service
}

// NewRolesClient creates a RolesClient backed by the real GCP IAM API.
func NewRolesClient(ctx context.Context, opts ...option.ClientOption) (RolesClient, error) {
	svc, err := iamv1.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create iam roles client: %w", err)
	}
	return &gcpRolesClient{iam: svc}, nil
}

func (c *gcpRolesClient) ListRoles(ctx context.Context, project string) ([]*Role, error) {
	call := c.iam.Projects.Roles.List("projects/" + project)

	var roles []*Role
	if err := call.Pages(ctx, func(resp *iamv1.ListRolesResponse) error {
		for _, role := range resp.Roles {
			roles = append(roles, roleFromAPI(role))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list roles: %w", err)
	}
	return roles, nil
}

func (c *gcpRolesClient) CreateRole(ctx context.Context, project, roleID, title, description string, permissions []string) (*Role, error) {
	role, err := c.iam.Projects.Roles.Create("projects/"+project, &iamv1.CreateRoleRequest{
		RoleId: roleID,
		Role: &iamv1.Role{
			Title:               title,
			Description:         description,
			IncludedPermissions: permissions,
			Stage:               "GA",
		},
	}).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create role %s: %w", roleID, err)
	}
	return roleFromAPI(role), nil
}

func (c *gcpRolesClient) DeleteRole(ctx context.Context, name string) error {
	if _, err := c.iam.Projects.Roles.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete role %s: %w", name, err)
	}
	return nil
}

func (c *gcpRolesClient) DescribeRole(ctx context.Context, name string) (*Role, error) {
	r, err := c.iam.Projects.Roles.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get role %s: %w", name, err)
	}
	return roleFromAPI(r), nil
}

func (c *gcpRolesClient) UpdateRole(ctx context.Context, name, title, description string, permissions []string) (*Role, error) {
	role := &iamv1.Role{}
	if title != "" {
		role.Title = title
	}
	if description != "" {
		role.Description = description
	}
	if len(permissions) > 0 {
		role.IncludedPermissions = permissions
	}

	r, err := c.iam.Projects.Roles.Patch(name, role).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("update role %s: %w", name, err)
	}
	return roleFromAPI(r), nil
}

func roleFromAPI(r *iamv1.Role) *Role {
	return &Role{
		Name:        r.Name,
		Title:       r.Title,
		Description: r.Description,
		Stage:       r.Stage,
		Deleted:     r.Deleted,
	}
}
