package compute

import (
	"context"
	"errors"
	"fmt"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
)

// Images

func (c *gcpClient) ListImages(ctx context.Context, project string) ([]*Image, error) {
	it := c.images.List(ctx, &computepb.ListImagesRequest{Project: project})
	var out []*Image
	for {
		img, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list images: %w", err)
		}
		// skip deprecated images
		if img.GetDeprecated() != nil {
			continue
		}
		out = append(out, imageFromProto(img))
	}
	return out, nil
}

func (c *gcpClient) GetImage(ctx context.Context, project, name string) (*Image, error) {
	img, err := c.images.Get(ctx, &computepb.GetImageRequest{
		Project: project,
		Image:   name,
	})
	if err != nil {
		return nil, fmt.Errorf("get image %s: %w", name, err)
	}
	return imageFromProto(img), nil
}

func (c *gcpClient) CreateImage(ctx context.Context, project string, req *CreateImageRequest) error {
	img := &computepb.Image{
		Name:        &req.Name,
		Description: strPtrOrNil(req.Description),
		Family:      strPtrOrNil(req.Family),
	}
	if req.SourceDisk != "" {
		img.SourceDisk = &req.SourceDisk
	}
	op, err := c.images.Insert(ctx, &computepb.InsertImageRequest{
		Project:       project,
		ImageResource: img,
	})
	if err != nil {
		return fmt.Errorf("create image %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteImage(ctx context.Context, project, name string) error {
	op, err := c.images.Delete(ctx, &computepb.DeleteImageRequest{
		Project: project,
		Image:   name,
	})
	if err != nil {
		return fmt.Errorf("delete image %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func imageFromProto(img *computepb.Image) *Image {
	return &Image{
		Name:        img.GetName(),
		Family:      img.GetFamily(),
		Status:      img.GetStatus(),
		DiskSizeGb:  img.GetDiskSizeGb(),
		Description: img.GetDescription(),
		SelfLink:    img.GetSelfLink(),
	}
}

// VPN Tunnels

func (c *gcpClient) ListVPNTunnels(ctx context.Context, project, region string) ([]*VPNTunnel, error) {
	it := c.vpnTunnels.List(ctx, &computepb.ListVpnTunnelsRequest{
		Project: project,
		Region:  region,
	})
	var out []*VPNTunnel
	for {
		t, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list vpn tunnels: %w", err)
		}
		out = append(out, vpnTunnelFromProto(t))
	}
	return out, nil
}

func (c *gcpClient) GetVPNTunnel(ctx context.Context, project, region, name string) (*VPNTunnel, error) {
	t, err := c.vpnTunnels.Get(ctx, &computepb.GetVpnTunnelRequest{
		Project:   project,
		Region:    region,
		VpnTunnel: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get vpn tunnel %s: %w", name, err)
	}
	return vpnTunnelFromProto(t), nil
}

func (c *gcpClient) CreateVPNTunnel(ctx context.Context, project, region string, req *CreateVPNTunnelRequest) error {
	ikeVersion := req.IKEVersion
	if ikeVersion == 0 {
		ikeVersion = 2
	}
	tunnel := &computepb.VpnTunnel{
		Name:        &req.Name,
		PeerIp:      strPtrOrNil(req.PeerIP),
		SharedSecret: strPtrOrNil(req.SharedSecret),
		IkeVersion:  &ikeVersion,
		Description: strPtrOrNil(req.Description),
	}
	if req.VPNGateway != "" {
		tunnel.VpnGateway = &req.VPNGateway
	}
	op, err := c.vpnTunnels.Insert(ctx, &computepb.InsertVpnTunnelRequest{
		Project:           project,
		Region:            region,
		VpnTunnelResource: tunnel,
	})
	if err != nil {
		return fmt.Errorf("create vpn tunnel %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteVPNTunnel(ctx context.Context, project, region, name string) error {
	op, err := c.vpnTunnels.Delete(ctx, &computepb.DeleteVpnTunnelRequest{
		Project:   project,
		Region:    region,
		VpnTunnel: name,
	})
	if err != nil {
		return fmt.Errorf("delete vpn tunnel %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func vpnTunnelFromProto(t *computepb.VpnTunnel) *VPNTunnel {
	return &VPNTunnel{
		Name:        t.GetName(),
		Region:      t.GetRegion(),
		Status:      t.GetStatus(),
		PeerIP:      t.GetPeerIp(),
		IKEVersion:  t.GetIkeVersion(),
		Description: t.GetDescription(),
	}
}

// Unmanaged Instance Groups

func (c *gcpClient) ListUnmanagedInstanceGroups(ctx context.Context, project, zone string) ([]*UnmanagedInstanceGroup, error) {
	it := c.unmanagedInstanceGroups.List(ctx, &computepb.ListInstanceGroupsRequest{
		Project: project,
		Zone:    zone,
	})
	var out []*UnmanagedInstanceGroup
	for {
		g, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list unmanaged instance groups: %w", err)
		}
		out = append(out, unmanagedInstanceGroupFromProto(g))
	}
	return out, nil
}

func (c *gcpClient) GetUnmanagedInstanceGroup(ctx context.Context, project, zone, name string) (*UnmanagedInstanceGroup, error) {
	g, err := c.unmanagedInstanceGroups.Get(ctx, &computepb.GetInstanceGroupRequest{
		Project:       project,
		Zone:          zone,
		InstanceGroup: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get unmanaged instance group %s: %w", name, err)
	}
	return unmanagedInstanceGroupFromProto(g), nil
}

func (c *gcpClient) CreateUnmanagedInstanceGroup(ctx context.Context, project, zone string, req *CreateUnmanagedInstanceGroupRequest) error {
	network := req.Network
	if network == "" {
		network = fmt.Sprintf("projects/%s/global/networks/default", project)
	}
	op, err := c.unmanagedInstanceGroups.Insert(ctx, &computepb.InsertInstanceGroupRequest{
		Project: project,
		Zone:    zone,
		InstanceGroupResource: &computepb.InstanceGroup{
			Name:        &req.Name,
			Network:     &network,
			Description: strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create unmanaged instance group %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpClient) DeleteUnmanagedInstanceGroup(ctx context.Context, project, zone, name string) error {
	op, err := c.unmanagedInstanceGroups.Delete(ctx, &computepb.DeleteInstanceGroupRequest{
		Project:       project,
		Zone:          zone,
		InstanceGroup: name,
	})
	if err != nil {
		return fmt.Errorf("delete unmanaged instance group %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func unmanagedInstanceGroupFromProto(g *computepb.InstanceGroup) *UnmanagedInstanceGroup {
	return &UnmanagedInstanceGroup{
		Name:        g.GetName(),
		Zone:        g.GetZone(),
		Size:        g.GetSize(),
		Network:     g.GetNetwork(),
		Description: g.GetDescription(),
	}
}
