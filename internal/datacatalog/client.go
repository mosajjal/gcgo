package datacatalog

import (
	"context"
	"errors"
	"fmt"

	"google.golang.org/api/datacatalog/v1"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Entry holds Data Catalog entry fields.
type Entry struct {
	Name           string `json:"name"`
	DisplayName    string `json:"display_name"`
	Type           string `json:"type"`
	LinkedResource string `json:"linked_resource"`
}

// EntryGroup holds Data Catalog entry group fields.
type EntryGroup struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// Client defines Data Catalog operations.
type Client interface {
	ListEntries(ctx context.Context, project, region, entryGroupID string) ([]*Entry, error)
	GetEntry(ctx context.Context, project, region, entryGroupID, entryID string) (*Entry, error)

	ListEntryGroups(ctx context.Context, project, region string) ([]*EntryGroup, error)
	GetEntryGroup(ctx context.Context, project, region, entryGroupID string) (*EntryGroup, error)
	CreateEntryGroup(ctx context.Context, project, region string, req *CreateEntryGroupRequest) error
	DeleteEntryGroup(ctx context.Context, project, region, entryGroupID string) error
}

// CreateEntryGroupRequest holds parameters for entry group creation.
type CreateEntryGroupRequest struct {
	EntryGroupID string
	DisplayName  string
	Description  string
}

type gcpClient struct {
	dc *datacatalog.Service
}

// NewClient creates a Client backed by the real Data Catalog API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := datacatalog.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create datacatalog client: %w", err)
	}
	return &gcpClient{dc: svc}, nil
}

func (c *gcpClient) ListEntries(ctx context.Context, project, region, entryGroupID string) ([]*Entry, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/entryGroups/%s", project, region, entryGroupID)
	it := c.dc.Projects.Locations.EntryGroups.Entries.List(parent).Context(ctx)

	var entries []*Entry
	for {
		resp, err := it.Do()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list entries: %w", err)
		}
		for _, e := range resp.Entries {
			entries = append(entries, entryFromAPI(e))
		}
		if resp.NextPageToken == "" {
			break
		}
		it.PageToken(resp.NextPageToken)
	}
	return entries, nil
}

func (c *gcpClient) GetEntry(ctx context.Context, project, region, entryGroupID, entryID string) (*Entry, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/entryGroups/%s/entries/%s", project, region, entryGroupID, entryID)
	e, err := c.dc.Projects.Locations.EntryGroups.Entries.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get entry %s: %w", entryID, err)
	}
	return entryFromAPI(e), nil
}

func (c *gcpClient) ListEntryGroups(ctx context.Context, project, region string) ([]*EntryGroup, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	it := c.dc.Projects.Locations.EntryGroups.List(parent).Context(ctx)

	var groups []*EntryGroup
	for {
		resp, err := it.Do()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list entry groups: %w", err)
		}
		for _, g := range resp.EntryGroups {
			groups = append(groups, entryGroupFromAPI(g))
		}
		if resp.NextPageToken == "" {
			break
		}
		it.PageToken(resp.NextPageToken)
	}
	return groups, nil
}

func (c *gcpClient) GetEntryGroup(ctx context.Context, project, region, entryGroupID string) (*EntryGroup, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/entryGroups/%s", project, region, entryGroupID)
	g, err := c.dc.Projects.Locations.EntryGroups.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get entry group %s: %w", entryGroupID, err)
	}
	return entryGroupFromAPI(g), nil
}

func (c *gcpClient) CreateEntryGroup(ctx context.Context, project, region string, req *CreateEntryGroupRequest) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	_, err := c.dc.Projects.Locations.EntryGroups.Create(parent, &datacatalog.GoogleCloudDatacatalogV1EntryGroup{
		DisplayName: req.DisplayName,
		Description: req.Description,
	}).EntryGroupId(req.EntryGroupID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create entry group %s: %w", req.EntryGroupID, err)
	}
	return nil
}

func (c *gcpClient) DeleteEntryGroup(ctx context.Context, project, region, entryGroupID string) error {
	name := fmt.Sprintf("projects/%s/locations/%s/entryGroups/%s", project, region, entryGroupID)
	if _, err := c.dc.Projects.Locations.EntryGroups.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete entry group %s: %w", entryGroupID, err)
	}
	return nil
}

func entryFromAPI(e *datacatalog.GoogleCloudDatacatalogV1Entry) *Entry {
	entryType := e.Type
	if entryType == "" {
		entryType = e.UserSpecifiedType
	}
	return &Entry{
		Name:           e.Name,
		DisplayName:    e.DisplayName,
		Type:           entryType,
		LinkedResource: e.LinkedResource,
	}
}

func entryGroupFromAPI(g *datacatalog.GoogleCloudDatacatalogV1EntryGroup) *EntryGroup {
	return &EntryGroup{
		Name:        g.Name,
		DisplayName: g.DisplayName,
		Description: g.Description,
	}
}
