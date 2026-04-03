package dataplex

import (
	"context"
	"fmt"

	"google.golang.org/api/dataplex/v1"
	"google.golang.org/api/option"
)

// Lake holds Dataplex lake fields.
type Lake struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	Region      string `json:"region"`
}

// Zone holds Dataplex zone fields.
type Zone struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	State       string `json:"state"`
	Type        string `json:"type"`
}

// Client defines Dataplex operations.
type Client interface {
	ListLakes(ctx context.Context, project, region string) ([]*Lake, error)
	GetLake(ctx context.Context, project, region, lakeID string) (*Lake, error)
	CreateLake(ctx context.Context, project, region string, req *CreateLakeRequest) error
	DeleteLake(ctx context.Context, project, region, lakeID string) error

	ListZones(ctx context.Context, project, region, lakeID string) ([]*Zone, error)
	GetZone(ctx context.Context, project, region, lakeID, zoneID string) (*Zone, error)
	CreateZone(ctx context.Context, project, region, lakeID string, req *CreateZoneRequest) error
	DeleteZone(ctx context.Context, project, region, lakeID, zoneID string) error
}

// CreateLakeRequest holds parameters for lake creation.
type CreateLakeRequest struct {
	LakeID      string
	DisplayName string
}

// CreateZoneRequest holds parameters for zone creation.
type CreateZoneRequest struct {
	ZoneID      string
	DisplayName string
	Type        string
}

type gcpClient struct {
	svc *dataplex.Service
}

// NewClient creates a Client backed by the real Dataplex API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := dataplex.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create dataplex client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListLakes(ctx context.Context, project, region string) ([]*Lake, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	var lakes []*Lake
	err := c.svc.Projects.Locations.Lakes.List(parent).Context(ctx).Pages(ctx, func(resp *dataplex.GoogleCloudDataplexV1ListLakesResponse) error {
		for _, l := range resp.Lakes {
			lakes = append(lakes, lakeFromAPI(l, region))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list lakes: %w", err)
	}
	return lakes, nil
}

func (c *gcpClient) GetLake(ctx context.Context, project, region, lakeID string) (*Lake, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/lakes/%s", project, region, lakeID)
	l, err := c.svc.Projects.Locations.Lakes.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get lake %s: %w", lakeID, err)
	}
	return lakeFromAPI(l, region), nil
}

func (c *gcpClient) CreateLake(ctx context.Context, project, region string, req *CreateLakeRequest) error {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, region)
	_, err := c.svc.Projects.Locations.Lakes.Create(parent, &dataplex.GoogleCloudDataplexV1Lake{
		DisplayName: req.DisplayName,
	}).LakeId(req.LakeID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create lake %s: %w", req.LakeID, err)
	}
	return nil
}

func (c *gcpClient) DeleteLake(ctx context.Context, project, region, lakeID string) error {
	name := fmt.Sprintf("projects/%s/locations/%s/lakes/%s", project, region, lakeID)
	if _, err := c.svc.Projects.Locations.Lakes.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete lake %s: %w", lakeID, err)
	}
	return nil
}

func (c *gcpClient) ListZones(ctx context.Context, project, region, lakeID string) ([]*Zone, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s/lakes/%s", project, region, lakeID)
	var zones []*Zone
	err := c.svc.Projects.Locations.Lakes.Zones.List(parent).Context(ctx).Pages(ctx, func(resp *dataplex.GoogleCloudDataplexV1ListZonesResponse) error {
		for _, z := range resp.Zones {
			zones = append(zones, zoneFromAPI(z))
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("list zones: %w", err)
	}
	return zones, nil
}

func (c *gcpClient) GetZone(ctx context.Context, project, region, lakeID, zoneID string) (*Zone, error) {
	name := fmt.Sprintf("projects/%s/locations/%s/lakes/%s/zones/%s", project, region, lakeID, zoneID)
	z, err := c.svc.Projects.Locations.Lakes.Zones.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get zone %s: %w", zoneID, err)
	}
	return zoneFromAPI(z), nil
}

func (c *gcpClient) CreateZone(ctx context.Context, project, region, lakeID string, req *CreateZoneRequest) error {
	parent := fmt.Sprintf("projects/%s/locations/%s/lakes/%s", project, region, lakeID)
	zoneType := "RAW"
	if req.Type == "CURATED_DATA" {
		zoneType = "CURATED"
	}
	_, err := c.svc.Projects.Locations.Lakes.Zones.Create(parent, &dataplex.GoogleCloudDataplexV1Zone{
		DisplayName: req.DisplayName,
		Type:        zoneType,
		DiscoverySpec: &dataplex.GoogleCloudDataplexV1ZoneDiscoverySpec{
			CsvOptions: &dataplex.GoogleCloudDataplexV1ZoneDiscoverySpecCsvOptions{},
		},
	}).ZoneId(req.ZoneID).Context(ctx).Do()
	if err != nil {
		return fmt.Errorf("create zone %s: %w", req.ZoneID, err)
	}
	return nil
}

func (c *gcpClient) DeleteZone(ctx context.Context, project, region, lakeID, zoneID string) error {
	name := fmt.Sprintf("projects/%s/locations/%s/lakes/%s/zones/%s", project, region, lakeID, zoneID)
	if _, err := c.svc.Projects.Locations.Lakes.Zones.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete zone %s: %w", zoneID, err)
	}
	return nil
}

func lakeFromAPI(l *dataplex.GoogleCloudDataplexV1Lake, region string) *Lake {
	return &Lake{
		Name:        l.Name,
		DisplayName: l.DisplayName,
		State:       l.State,
		Region:      region,
	}
}

func zoneFromAPI(z *dataplex.GoogleCloudDataplexV1Zone) *Zone {
	return &Zone{
		Name:        z.Name,
		DisplayName: z.DisplayName,
		State:       z.State,
		Type:        z.Type,
	}
}
