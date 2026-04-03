package compute

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func makeNetworkClient(ctx context.Context, creds *auth.Credentials) (NetworkClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewNetworkClient(ctx, opt)
}

func requireRegion(cmd *cobra.Command, cfg *config.Config) (string, error) {
	region, _ := cmd.Flags().GetString("region")
	if region == "" {
		region = cfg.Region()
	}
	if region == "" {
		return "", fmt.Errorf("no region set (use --region or 'gcgo config set region REGION')")
	}
	return region, nil
}

// Networks

func newNetworksCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "networks",
		Short: "Manage VPC networks",
	}
	cmd.AddCommand(
		newNetworksListCommand(cfg, creds),
		newNetworksDescribeCommand(cfg, creds),
		newNetworksCreateCommand(cfg, creds),
		newNetworksDeleteCommand(cfg, creds),
	)
	return cmd
}

func newNetworksListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List VPC networks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			networks, err := client.ListNetworks(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), networks)
			}
			headers := []string{"NAME", "ROUTING_MODE", "AUTO_SUBNETS", "SUBNETS"}
			rows := make([][]string, len(networks))
			for i, n := range networks {
				rows[i] = []string{n.Name, n.RoutingMode, fmt.Sprintf("%v", n.AutoCreateSubnetworks), fmt.Sprintf("%d", len(n.Subnetworks))}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newNetworksDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe NETWORK",
		Short: "Describe a VPC network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			net, err := client.GetNetwork(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), net)
		},
	}
}

func newNetworksCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateNetworkRequest

	cmd := &cobra.Command{
		Use:   "create NETWORK",
		Short: "Create a VPC network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateNetwork(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created network %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().BoolVar(&req.AutoCreateSubnetworks, "auto-create-subnetworks", true, "Auto-create subnetworks")
	cmd.Flags().StringVar(&req.RoutingMode, "routing-mode", "REGIONAL", "Routing mode (REGIONAL or GLOBAL)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Network description")
	return cmd
}

func newNetworksDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete NETWORK",
		Short: "Delete a VPC network",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteNetwork(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted network %q.\n", args[0])
			return nil
		},
	}
}

// Subnets

func newSubnetsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "subnets",
		Short: "Manage subnetworks",
	}
	cmd.AddCommand(
		newSubnetsListCommand(cfg, creds),
		newSubnetsDescribeCommand(cfg, creds),
		newSubnetsCreateCommand(cfg, creds),
		newSubnetsDeleteCommand(cfg, creds),
		newSubnetsExpandIPRangeCommand(cfg, creds),
	)
	return cmd
}

func newSubnetsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List subnetworks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			subnets, err := client.ListSubnets(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), subnets)
			}
			headers := []string{"NAME", "REGION", "NETWORK", "IP_CIDR_RANGE"}
			rows := make([][]string, len(subnets))
			for i, s := range subnets {
				rows[i] = []string{s.Name, s.Region, s.Network, s.IPCIDRRange}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

func newSubnetsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe SUBNET",
		Short: "Describe a subnetwork",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			subnet, err := client.GetSubnet(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), subnet)
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

func newSubnetsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateSubnetRequest

	cmd := &cobra.Command{
		Use:   "create SUBNET",
		Short: "Create a subnetwork",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateSubnet(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created subnet %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	cmd.Flags().StringVar(&req.Network, "network", "", "VPC network")
	cmd.Flags().StringVar(&req.IPCIDRRange, "range", "", "IP CIDR range (e.g. 10.0.0.0/24)")
	cmd.Flags().StringVar(&req.Description, "description", "", "Subnet description")
	return cmd
}

func newSubnetsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete SUBNET",
		Short: "Delete a subnetwork",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteSubnet(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted subnet %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

func newSubnetsExpandIPRangeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var newCIDR string

	cmd := &cobra.Command{
		Use:   "expand-ip-range SUBNET",
		Short: "Expand the IP range of a subnetwork",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.ExpandSubnetIPRange(ctx, project, region, args[0], newCIDR); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Expanded subnet %q IP range to %s.\n", args[0], newCIDR)
			return nil
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	cmd.Flags().StringVar(&newCIDR, "prefix-length", "", "New prefix length or CIDR (e.g. 10.0.0.0/16)")
	return cmd
}

// Addresses

func newAddressesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "addresses",
		Short: "Manage reserved IP addresses",
	}
	cmd.AddCommand(
		newAddressesListCommand(cfg, creds),
		newAddressesDescribeCommand(cfg, creds),
		newAddressesCreateCommand(cfg, creds),
		newAddressesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newAddressesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List reserved addresses",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			addrs, err := client.ListAddresses(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), addrs)
			}
			headers := []string{"NAME", "ADDRESS", "STATUS", "TYPE", "REGION"}
			rows := make([][]string, len(addrs))
			for i, a := range addrs {
				rows[i] = []string{a.Name, a.Address, a.Status, a.AddressType, a.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

func newAddressesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe ADDRESS",
		Short: "Describe a reserved address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			addr, err := client.GetAddress(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), addr)
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

func newAddressesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateAddressRequest

	cmd := &cobra.Command{
		Use:   "create ADDRESS",
		Short: "Reserve an IP address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateAddress(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Reserved address %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	cmd.Flags().StringVar(&req.AddressType, "address-type", "EXTERNAL", "Address type (INTERNAL or EXTERNAL)")
	cmd.Flags().StringVar(&req.Purpose, "purpose", "", "Purpose (e.g. GCE_ENDPOINT)")
	cmd.Flags().StringVar(&req.Subnetwork, "subnet", "", "Subnet (for internal addresses)")
	return cmd
}

func newAddressesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete ADDRESS",
		Short: "Release a reserved address",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteAddress(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted address %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

// Routers

func newRoutersCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "routers",
		Short: "Manage Cloud Routers",
	}
	cmd.AddCommand(
		newRoutersListCommand(cfg, creds),
		newRoutersDescribeCommand(cfg, creds),
		newRoutersCreateCommand(cfg, creds),
		newRoutersDeleteCommand(cfg, creds),
	)
	return cmd
}

func newRoutersListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "List Cloud Routers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			routers, err := client.ListRouters(ctx, project, region)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), routers)
			}
			headers := []string{"NAME", "REGION", "NETWORK", "BGP_ASN"}
			rows := make([][]string, len(routers))
			for i, r := range routers {
				rows[i] = []string{r.Name, r.Region, r.Network, fmt.Sprintf("%d", r.BGPAsn)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

func newRoutersDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "describe ROUTER",
		Short: "Describe a Cloud Router",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			router, err := client.GetRouter(ctx, project, region, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), router)
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

func newRoutersCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateRouterRequest

	cmd := &cobra.Command{
		Use:   "create ROUTER",
		Short: "Create a Cloud Router",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateRouter(ctx, project, region, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created router %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	cmd.Flags().StringVar(&req.Network, "network", "", "VPC network")
	cmd.Flags().Int64Var(&req.BGPAsn, "asn", 64512, "BGP ASN")
	return cmd
}

func newRoutersDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "delete ROUTER",
		Short: "Delete a Cloud Router",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			region, err := requireRegion(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteRouter(ctx, project, region, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted router %q.\n", args[0])
			return nil
		},
	}
	cmd.Flags().String("region", "", "Region (falls back to config)")
	return cmd
}

// Routes

func newRoutesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "routes",
		Short: "Manage routes",
	}
	cmd.AddCommand(
		newRoutesListCommand(cfg, creds),
		newRoutesDescribeCommand(cfg, creds),
	)
	return cmd
}

func newRoutesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List routes",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			routes, err := client.ListRoutes(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), routes)
			}
			headers := []string{"NAME", "NETWORK", "DEST_RANGE", "NEXT_HOP", "PRIORITY"}
			rows := make([][]string, len(routes))
			for i, r := range routes {
				nextHop := r.NextHopGateway
				if nextHop == "" {
					nextHop = r.NextHopIP
				}
				if nextHop == "" {
					nextHop = r.NextHopInstance
				}
				rows[i] = []string{r.Name, r.Network, r.DestRange, nextHop, fmt.Sprintf("%d", r.Priority)}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newRoutesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe ROUTE",
		Short: "Describe a route",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeNetworkClient(ctx, creds)
			if err != nil {
				return err
			}
			route, err := client.GetRoute(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), route)
		},
	}
}
