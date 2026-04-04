package compute

import (
	"context"
	"errors"
	"fmt"
	"path"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

func (c *gcpClient) ListZones(ctx context.Context, project, region string) ([]*Zone, error) {
	req := &computepb.ListZonesRequest{Project: project}
	if region != "" {
		filter := fmt.Sprintf("region eq .*/regions/%s", region)
		req.Filter = &filter
	}
	it := c.zones.List(ctx, req)
	var out []*Zone
	for {
		z, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list zones: %w", err)
		}
		out = append(out, &Zone{
			Name:   z.GetName(),
			Region: path.Base(z.GetRegion()),
			Status: z.GetStatus(),
		})
	}
	return out, nil
}

func (c *gcpClient) ListRegions(ctx context.Context, project string) ([]*Region, error) {
	it := c.regions.List(ctx, &computepb.ListRegionsRequest{Project: project})
	var out []*Region
	for {
		r, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list regions: %w", err)
		}
		var zones []string
		for _, z := range r.GetZones() {
			zones = append(zones, path.Base(z))
		}
		out = append(out, &Region{
			Name:   r.GetName(),
			Status: r.GetStatus(),
			Zones:  zones,
		})
	}
	return out, nil
}

func (c *gcpClient) ListMachineTypes(ctx context.Context, project, zone string) ([]*MachineType, error) {
	it := c.machineTypes.List(ctx, &computepb.ListMachineTypesRequest{
		Project: project,
		Zone:    zone,
	})
	var out []*MachineType
	for {
		mt, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list machine types: %w", err)
		}
		out = append(out, &MachineType{
			Name:        mt.GetName(),
			Description: mt.GetDescription(),
			VCPUs:       mt.GetGuestCpus(),
			MemoryMb:    mt.GetMemoryMb(),
			Zone:        path.Base(mt.GetZone()),
		})
	}
	return out, nil
}
