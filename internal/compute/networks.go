package compute

import (
	"context"
	"errors"
	"fmt"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// Network holds VPC network fields.
type Network struct {
	Name                  string   `json:"name"`
	AutoCreateSubnetworks bool     `json:"auto_create_subnetworks"`
	RoutingMode           string   `json:"routing_mode"`
	Subnetworks           []string `json:"subnetworks"`
	Description           string   `json:"description"`
}

// Subnet holds subnetwork fields.
type Subnet struct {
	Name                string `json:"name"`
	Region              string `json:"region"`
	Network             string `json:"network"`
	IPCIDRRange         string `json:"ip_cidr_range"`
	GatewayAddress      string `json:"gateway_address"`
	PrivateIPGoogleAccs string `json:"private_ip_google_access"`
}

// Address holds reserved address fields.
type Address struct {
	Name        string `json:"name"`
	Region      string `json:"region"`
	Address     string `json:"address"`
	Status      string `json:"status"`
	AddressType string `json:"address_type"`
	Purpose     string `json:"purpose"`
}

// Router holds cloud router fields.
type Router struct {
	Name    string `json:"name"`
	Region  string `json:"region"`
	Network string `json:"network"`
	BGPAsn  int64  `json:"bgp_asn"`
}

// Route holds route fields.
type Route struct {
	Name            string `json:"name"`
	Network         string `json:"network"`
	DestRange       string `json:"dest_range"`
	NextHopGateway  string `json:"next_hop_gateway"`
	NextHopIP       string `json:"next_hop_ip"`
	NextHopInstance string `json:"next_hop_instance"`
	Priority        int64  `json:"priority"`
}

// CreateNetworkRequest holds parameters for network creation.
type CreateNetworkRequest struct {
	Name                  string
	AutoCreateSubnetworks bool
	RoutingMode           string // REGIONAL or GLOBAL
	Description           string
}

// CreateSubnetRequest holds parameters for subnet creation.
type CreateSubnetRequest struct {
	Name        string
	Network     string
	IPCIDRRange string
	Region      string
	Description string
}

// CreateAddressRequest holds parameters for address reservation.
type CreateAddressRequest struct {
	Name        string
	Region      string
	AddressType string // INTERNAL or EXTERNAL
	Purpose     string
	Subnetwork  string
}

// CreateRouterRequest holds parameters for router creation.
type CreateRouterRequest struct {
	Name    string
	Network string
	Region  string
	BGPAsn  int64
}

// NetworkClient defines networking operations.
type NetworkClient interface {
	// Networks
	ListNetworks(ctx context.Context, project string) ([]*Network, error)
	GetNetwork(ctx context.Context, project, name string) (*Network, error)
	CreateNetwork(ctx context.Context, project string, req *CreateNetworkRequest) error
	DeleteNetwork(ctx context.Context, project, name string) error

	// Subnets
	ListSubnets(ctx context.Context, project, region string) ([]*Subnet, error)
	GetSubnet(ctx context.Context, project, region, name string) (*Subnet, error)
	CreateSubnet(ctx context.Context, project, region string, req *CreateSubnetRequest) error
	DeleteSubnet(ctx context.Context, project, region, name string) error
	ExpandSubnetIPRange(ctx context.Context, project, region, name, newCIDR string) error

	// Addresses
	ListAddresses(ctx context.Context, project, region string) ([]*Address, error)
	GetAddress(ctx context.Context, project, region, name string) (*Address, error)
	CreateAddress(ctx context.Context, project, region string, req *CreateAddressRequest) error
	DeleteAddress(ctx context.Context, project, region, name string) error

	// Routers
	ListRouters(ctx context.Context, project, region string) ([]*Router, error)
	GetRouter(ctx context.Context, project, region, name string) (*Router, error)
	CreateRouter(ctx context.Context, project, region string, req *CreateRouterRequest) error
	DeleteRouter(ctx context.Context, project, region, name string) error

	// Routes
	ListRoutes(ctx context.Context, project string) ([]*Route, error)
	GetRoute(ctx context.Context, project, name string) (*Route, error)
}

type gcpNetworkClient struct {
	networks   *compute.NetworksClient
	subnets    *compute.SubnetworksClient
	addresses  *compute.AddressesClient
	routers    *compute.RoutersClient
	routes     *compute.RoutesClient
}

// NewNetworkClient creates a NetworkClient backed by the real GCP Compute API.
func NewNetworkClient(ctx context.Context, opts ...option.ClientOption) (NetworkClient, error) {
	nc, err := compute.NewNetworksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create networks client: %w", err)
	}
	sc, err := compute.NewSubnetworksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create subnetworks client: %w", err)
	}
	ac, err := compute.NewAddressesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create addresses client: %w", err)
	}
	rc, err := compute.NewRoutersRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create routers client: %w", err)
	}
	rtc, err := compute.NewRoutesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create routes client: %w", err)
	}
	return &gcpNetworkClient{
		networks:  nc,
		subnets:   sc,
		addresses: ac,
		routers:   rc,
		routes:    rtc,
	}, nil
}

// Networks

func (c *gcpNetworkClient) ListNetworks(ctx context.Context, project string) ([]*Network, error) {
	it := c.networks.List(ctx, &computepb.ListNetworksRequest{Project: project})
	var out []*Network
	for {
		n, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list networks: %w", err)
		}
		out = append(out, networkFromProto(n))
	}
	return out, nil
}

func (c *gcpNetworkClient) GetNetwork(ctx context.Context, project, name string) (*Network, error) {
	n, err := c.networks.Get(ctx, &computepb.GetNetworkRequest{Project: project, Network: name})
	if err != nil {
		return nil, fmt.Errorf("get network %s: %w", name, err)
	}
	return networkFromProto(n), nil
}

func (c *gcpNetworkClient) CreateNetwork(ctx context.Context, project string, req *CreateNetworkRequest) error {
	routingMode := req.RoutingMode
	if routingMode == "" {
		routingMode = "REGIONAL"
	}
	op, err := c.networks.Insert(ctx, &computepb.InsertNetworkRequest{
		Project: project,
		NetworkResource: &computepb.Network{
			Name:                  &req.Name,
			AutoCreateSubnetworks: &req.AutoCreateSubnetworks,
			RoutingConfig: &computepb.NetworkRoutingConfig{
				RoutingMode: &routingMode,
			},
			Description: strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create network %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpNetworkClient) DeleteNetwork(ctx context.Context, project, name string) error {
	op, err := c.networks.Delete(ctx, &computepb.DeleteNetworkRequest{Project: project, Network: name})
	if err != nil {
		return fmt.Errorf("delete network %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Subnets

func (c *gcpNetworkClient) ListSubnets(ctx context.Context, project, region string) ([]*Subnet, error) {
	it := c.subnets.List(ctx, &computepb.ListSubnetworksRequest{Project: project, Region: region})
	var out []*Subnet
	for {
		s, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list subnets: %w", err)
		}
		out = append(out, subnetFromProto(s))
	}
	return out, nil
}

func (c *gcpNetworkClient) GetSubnet(ctx context.Context, project, region, name string) (*Subnet, error) {
	s, err := c.subnets.Get(ctx, &computepb.GetSubnetworkRequest{
		Project:    project,
		Region:     region,
		Subnetwork: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get subnet %s: %w", name, err)
	}
	return subnetFromProto(s), nil
}

func (c *gcpNetworkClient) CreateSubnet(ctx context.Context, project, region string, req *CreateSubnetRequest) error {
	op, err := c.subnets.Insert(ctx, &computepb.InsertSubnetworkRequest{
		Project: project,
		Region:  region,
		SubnetworkResource: &computepb.Subnetwork{
			Name:        &req.Name,
			Network:     &req.Network,
			IpCidrRange: &req.IPCIDRRange,
			Description: strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create subnet %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpNetworkClient) DeleteSubnet(ctx context.Context, project, region, name string) error {
	op, err := c.subnets.Delete(ctx, &computepb.DeleteSubnetworkRequest{
		Project:    project,
		Region:     region,
		Subnetwork: name,
	})
	if err != nil {
		return fmt.Errorf("delete subnet %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpNetworkClient) ExpandSubnetIPRange(ctx context.Context, project, region, name, newCIDR string) error {
	op, err := c.subnets.ExpandIpCidrRange(ctx, &computepb.ExpandIpCidrRangeSubnetworkRequest{
		Project:    project,
		Region:     region,
		Subnetwork: name,
		SubnetworksExpandIpCidrRangeRequestResource: &computepb.SubnetworksExpandIpCidrRangeRequest{
			IpCidrRange: &newCIDR,
		},
	})
	if err != nil {
		return fmt.Errorf("expand subnet %s ip range: %w", name, err)
	}
	return op.Wait(ctx)
}

// Addresses

func (c *gcpNetworkClient) ListAddresses(ctx context.Context, project, region string) ([]*Address, error) {
	it := c.addresses.List(ctx, &computepb.ListAddressesRequest{Project: project, Region: region})
	var out []*Address
	for {
		a, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list addresses: %w", err)
		}
		out = append(out, addressFromProto(a))
	}
	return out, nil
}

func (c *gcpNetworkClient) GetAddress(ctx context.Context, project, region, name string) (*Address, error) {
	a, err := c.addresses.Get(ctx, &computepb.GetAddressRequest{
		Project: project,
		Region:  region,
		Address: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get address %s: %w", name, err)
	}
	return addressFromProto(a), nil
}

func (c *gcpNetworkClient) CreateAddress(ctx context.Context, project, region string, req *CreateAddressRequest) error {
	op, err := c.addresses.Insert(ctx, &computepb.InsertAddressRequest{
		Project: project,
		Region:  region,
		AddressResource: &computepb.Address{
			Name:        &req.Name,
			AddressType: strPtrOrNil(req.AddressType),
			Purpose:     strPtrOrNil(req.Purpose),
			Subnetwork:  strPtrOrNil(req.Subnetwork),
		},
	})
	if err != nil {
		return fmt.Errorf("create address %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpNetworkClient) DeleteAddress(ctx context.Context, project, region, name string) error {
	op, err := c.addresses.Delete(ctx, &computepb.DeleteAddressRequest{
		Project: project,
		Region:  region,
		Address: name,
	})
	if err != nil {
		return fmt.Errorf("delete address %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Routers

func (c *gcpNetworkClient) ListRouters(ctx context.Context, project, region string) ([]*Router, error) {
	it := c.routers.List(ctx, &computepb.ListRoutersRequest{Project: project, Region: region})
	var out []*Router
	for {
		r, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list routers: %w", err)
		}
		out = append(out, routerFromProto(r))
	}
	return out, nil
}

func (c *gcpNetworkClient) GetRouter(ctx context.Context, project, region, name string) (*Router, error) {
	r, err := c.routers.Get(ctx, &computepb.GetRouterRequest{
		Project: project,
		Region:  region,
		Router:  name,
	})
	if err != nil {
		return nil, fmt.Errorf("get router %s: %w", name, err)
	}
	return routerFromProto(r), nil
}

func (c *gcpNetworkClient) CreateRouter(ctx context.Context, project, region string, req *CreateRouterRequest) error {
	asn := uint32(req.BGPAsn)
	op, err := c.routers.Insert(ctx, &computepb.InsertRouterRequest{
		Project: project,
		Region:  region,
		RouterResource: &computepb.Router{
			Name:    &req.Name,
			Network: &req.Network,
			Bgp: &computepb.RouterBgp{
				Asn: &asn,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("create router %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpNetworkClient) DeleteRouter(ctx context.Context, project, region, name string) error {
	op, err := c.routers.Delete(ctx, &computepb.DeleteRouterRequest{
		Project: project,
		Region:  region,
		Router:  name,
	})
	if err != nil {
		return fmt.Errorf("delete router %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Routes

func (c *gcpNetworkClient) ListRoutes(ctx context.Context, project string) ([]*Route, error) {
	it := c.routes.List(ctx, &computepb.ListRoutesRequest{Project: project})
	var out []*Route
	for {
		r, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list routes: %w", err)
		}
		out = append(out, routeFromProto(r))
	}
	return out, nil
}

func (c *gcpNetworkClient) GetRoute(ctx context.Context, project, name string) (*Route, error) {
	r, err := c.routes.Get(ctx, &computepb.GetRouteRequest{Project: project, Route: name})
	if err != nil {
		return nil, fmt.Errorf("get route %s: %w", name, err)
	}
	return routeFromProto(r), nil
}

// Proto conversion helpers

func networkFromProto(n *computepb.Network) *Network {
	net := &Network{
		Name:                  n.GetName(),
		AutoCreateSubnetworks: n.GetAutoCreateSubnetworks(),
		Subnetworks:           n.GetSubnetworks(),
		Description:           n.GetDescription(),
	}
	if rc := n.GetRoutingConfig(); rc != nil {
		net.RoutingMode = rc.GetRoutingMode()
	}
	return net
}

func subnetFromProto(s *computepb.Subnetwork) *Subnet {
	return &Subnet{
		Name:                s.GetName(),
		Region:              s.GetRegion(),
		Network:             s.GetNetwork(),
		IPCIDRRange:         s.GetIpCidrRange(),
		GatewayAddress:      s.GetGatewayAddress(),
		PrivateIPGoogleAccs: fmt.Sprintf("%v", s.GetPrivateIpGoogleAccess()),
	}
}

func addressFromProto(a *computepb.Address) *Address {
	return &Address{
		Name:        a.GetName(),
		Region:      a.GetRegion(),
		Address:     a.GetAddress(),
		Status:      a.GetStatus(),
		AddressType: a.GetAddressType(),
		Purpose:     a.GetPurpose(),
	}
}

func routerFromProto(r *computepb.Router) *Router {
	rtr := &Router{
		Name:    r.GetName(),
		Region:  r.GetRegion(),
		Network: r.GetNetwork(),
	}
	if bgp := r.GetBgp(); bgp != nil {
		rtr.BGPAsn = int64(bgp.GetAsn())
	}
	return rtr
}

func routeFromProto(r *computepb.Route) *Route {
	return &Route{
		Name:            r.GetName(),
		Network:         r.GetNetwork(),
		DestRange:       r.GetDestRange(),
		NextHopGateway:  r.GetNextHopGateway(),
		NextHopIP:       r.GetNextHopIp(),
		NextHopInstance: r.GetNextHopInstance(),
		Priority:        int64(r.GetPriority()),
	}
}
