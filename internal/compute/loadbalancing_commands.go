package compute

import (
	"context"
	"fmt"

	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
	"github.com/mosajjal/gcgo/internal/output"
	"github.com/spf13/cobra"
)

func makeLoadBalancingClient(ctx context.Context, creds *auth.Credentials) (LoadBalancingClient, error) {
	opt, err := creds.ClientOption(ctx)
	if err != nil {
		return nil, err
	}
	return NewLoadBalancingClient(ctx, opt)
}

func newForwardingRulesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "forwarding-rules",
		Short: "Manage global forwarding rules",
	}
	cmd.AddCommand(
		newForwardingRulesListCommand(cfg, creds),
		newForwardingRulesDescribeCommand(cfg, creds),
		newForwardingRulesCreateCommand(cfg, creds),
		newForwardingRulesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newForwardingRulesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List forwarding rules",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			rules, err := client.ListForwardingRules(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), rules)
			}
			headers := []string{"NAME", "IP_ADDRESS", "IP_PROTOCOL", "SCHEME", "TARGET", "BACKEND_SERVICE"}
			rows := make([][]string, len(rules))
			for i, rule := range rules {
				rows[i] = []string{rule.Name, rule.IPAddress, rule.IPProtocol, rule.LoadBalancingScheme, rule.Target, rule.BackendService}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newForwardingRulesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe FORWARDING_RULE",
		Short: "Describe a forwarding rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			rule, err := client.GetForwardingRule(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), rule)
		},
	}
}

func newForwardingRulesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateForwardingRuleRequest

	cmd := &cobra.Command{
		Use:   "create FORWARDING_RULE",
		Short: "Create a forwarding rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateForwardingRule(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created forwarding rule %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.IPAddress, "ip-address", "", "Static IP address")
	cmd.Flags().StringVar(&req.IPProtocol, "ip-protocol", "TCP", "IP protocol")
	cmd.Flags().StringVar(&req.LoadBalancingScheme, "load-balancing-scheme", "EXTERNAL", "Load balancing scheme")
	cmd.Flags().StringVar(&req.BackendService, "backend-service", "", "Backend service")
	cmd.Flags().StringVar(&req.Target, "target", "", "Target proxy")
	cmd.Flags().StringVar(&req.Description, "description", "", "Forwarding rule description")
	return cmd
}

func newForwardingRulesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete FORWARDING_RULE",
		Short: "Delete a forwarding rule",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteForwardingRule(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted forwarding rule %q.\n", args[0])
			return nil
		},
	}
}

func newBackendServicesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "backend-services",
		Short: "Manage backend services",
	}
	cmd.AddCommand(
		newBackendServicesListCommand(cfg, creds),
		newBackendServicesDescribeCommand(cfg, creds),
		newBackendServicesCreateCommand(cfg, creds),
		newBackendServicesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newBackendServicesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List backend services",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			services, err := client.ListBackendServices(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), services)
			}
			headers := []string{"NAME", "PROTOCOL", "SCHEME", "PORT_NAME", "HEALTH_CHECKS"}
			rows := make([][]string, len(services))
			for i, svc := range services {
				rows[i] = []string{svc.Name, svc.Protocol, svc.LoadBalancingScheme, svc.PortName, fmt.Sprintf("%d", len(svc.HealthChecks))}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newBackendServicesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe BACKEND_SERVICE",
		Short: "Describe a backend service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			service, err := client.GetBackendService(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), service)
		},
	}
}

func newBackendServicesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateBackendServiceRequest

	cmd := &cobra.Command{
		Use:   "create BACKEND_SERVICE",
		Short: "Create a backend service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateBackendService(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created backend service %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Protocol, "protocol", "HTTP", "Backend protocol")
	cmd.Flags().StringVar(&req.LoadBalancingScheme, "load-balancing-scheme", "EXTERNAL", "Load balancing scheme")
	cmd.Flags().StringVar(&req.PortName, "port-name", "", "Named backend port")
	cmd.Flags().StringArrayVar(&req.HealthChecks, "health-check", nil, "Health check self link or name")
	cmd.Flags().StringVar(&req.Description, "description", "", "Backend service description")
	return cmd
}

func newBackendServicesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete BACKEND_SERVICE",
		Short: "Delete a backend service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteBackendService(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted backend service %q.\n", args[0])
			return nil
		},
	}
}

func newHealthChecksCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "health-checks",
		Short: "Manage health checks",
	}
	cmd.AddCommand(
		newHealthChecksListCommand(cfg, creds),
		newHealthChecksDescribeCommand(cfg, creds),
		newHealthChecksCreateCommand(cfg, creds),
		newHealthChecksDeleteCommand(cfg, creds),
	)
	return cmd
}

func newHealthChecksListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List health checks",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			checks, err := client.ListHealthChecks(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), checks)
			}
			headers := []string{"NAME", "TYPE", "PORT", "REQUEST_PATH", "REGION"}
			rows := make([][]string, len(checks))
			for i, hc := range checks {
				rows[i] = []string{hc.Name, hc.Type, fmt.Sprintf("%d", hc.Port), hc.RequestPath, hc.Region}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newHealthChecksDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe HEALTH_CHECK",
		Short: "Describe a health check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			check, err := client.GetHealthCheck(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), check)
		},
	}
}

func newHealthChecksCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateHealthCheckRequest

	cmd := &cobra.Command{
		Use:   "create HEALTH_CHECK",
		Short: "Create a health check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateHealthCheck(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created health check %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Type, "type", "HTTP", "Health check type (HTTP or TCP)")
	cmd.Flags().Int32Var(&req.Port, "port", 80, "Port")
	cmd.Flags().StringVar(&req.RequestPath, "request-path", "/", "HTTP request path")
	cmd.Flags().Int32Var(&req.CheckIntervalSec, "check-interval-sec", 5, "Check interval")
	cmd.Flags().Int32Var(&req.TimeoutSec, "timeout-sec", 5, "Timeout")
	cmd.Flags().StringVar(&req.Description, "description", "", "Health check description")
	return cmd
}

func newHealthChecksDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete HEALTH_CHECK",
		Short: "Delete a health check",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteHealthCheck(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted health check %q.\n", args[0])
			return nil
		},
	}
}

func newUrlMapsCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "url-maps",
		Short: "Manage URL maps",
	}
	cmd.AddCommand(
		newUrlMapsListCommand(cfg, creds),
		newUrlMapsDescribeCommand(cfg, creds),
		newUrlMapsCreateCommand(cfg, creds),
		newUrlMapsDeleteCommand(cfg, creds),
	)
	return cmd
}

func newUrlMapsListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List URL maps",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			urlMaps, err := client.ListUrlMaps(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), urlMaps)
			}
			headers := []string{"NAME", "DEFAULT_SERVICE"}
			rows := make([][]string, len(urlMaps))
			for i, um := range urlMaps {
				rows[i] = []string{um.Name, um.DefaultService}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newUrlMapsDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe URL_MAP",
		Short: "Describe a URL map",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			um, err := client.GetUrlMap(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), um)
		},
	}
}

func newUrlMapsCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateUrlMapRequest

	cmd := &cobra.Command{
		Use:   "create URL_MAP",
		Short: "Create a URL map",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateUrlMap(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created URL map %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.DefaultService, "default-service", "", "Default backend service")
	cmd.Flags().StringVar(&req.Description, "description", "", "URL map description")
	return cmd
}

func newUrlMapsDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete URL_MAP",
		Short: "Delete a URL map",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteUrlMap(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted URL map %q.\n", args[0])
			return nil
		},
	}
}

func newTargetHttpProxiesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target-http-proxies",
		Short: "Manage target HTTP proxies",
	}
	cmd.AddCommand(
		newTargetHttpProxiesListCommand(cfg, creds),
		newTargetHttpProxiesDescribeCommand(cfg, creds),
		newTargetHttpProxiesCreateCommand(cfg, creds),
		newTargetHttpProxiesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newTargetHttpProxiesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List target HTTP proxies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			items, err := client.ListTargetHttpProxies(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), items)
			}
			headers := []string{"NAME", "URL_MAP"}
			rows := make([][]string, len(items))
			for i, item := range items {
				rows[i] = []string{item.Name, item.UrlMap}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newTargetHttpProxiesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe TARGET_HTTP_PROXY",
		Short: "Describe a target HTTP proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			item, err := client.GetTargetHttpProxy(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), item)
		},
	}
}

func newTargetHttpProxiesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateTargetHttpProxyRequest

	cmd := &cobra.Command{
		Use:   "create TARGET_HTTP_PROXY",
		Short: "Create a target HTTP proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateTargetHttpProxy(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created target HTTP proxy %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.UrlMap, "url-map", "", "URL map")
	cmd.Flags().StringVar(&req.Description, "description", "", "Proxy description")
	return cmd
}

func newTargetHttpProxiesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete TARGET_HTTP_PROXY",
		Short: "Delete a target HTTP proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteTargetHttpProxy(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted target HTTP proxy %q.\n", args[0])
			return nil
		},
	}
}

func newTargetHttpsProxiesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target-https-proxies",
		Short: "Manage target HTTPS proxies",
	}
	cmd.AddCommand(
		newTargetHttpsProxiesListCommand(cfg, creds),
		newTargetHttpsProxiesDescribeCommand(cfg, creds),
		newTargetHttpsProxiesCreateCommand(cfg, creds),
		newTargetHttpsProxiesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newTargetHttpsProxiesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List target HTTPS proxies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			items, err := client.ListTargetHttpsProxies(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), items)
			}
			headers := []string{"NAME", "URL_MAP", "CERTIFICATES"}
			rows := make([][]string, len(items))
			for i, item := range items {
				rows[i] = []string{item.Name, item.UrlMap, fmt.Sprintf("%d", len(item.SslCertificates))}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newTargetHttpsProxiesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe TARGET_HTTPS_PROXY",
		Short: "Describe a target HTTPS proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			item, err := client.GetTargetHttpsProxy(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), item)
		},
	}
}

func newTargetHttpsProxiesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateTargetHttpsProxyRequest

	cmd := &cobra.Command{
		Use:   "create TARGET_HTTPS_PROXY",
		Short: "Create a target HTTPS proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateTargetHttpsProxy(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created target HTTPS proxy %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.UrlMap, "url-map", "", "URL map")
	cmd.Flags().StringArrayVar(&req.SslCertificates, "ssl-certificate", nil, "SSL certificate self link or name")
	cmd.Flags().StringVar(&req.CertificateMap, "certificate-map", "", "Certificate map")
	cmd.Flags().StringVar(&req.Description, "description", "", "Proxy description")
	return cmd
}

func newTargetHttpsProxiesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete TARGET_HTTPS_PROXY",
		Short: "Delete a target HTTPS proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteTargetHttpsProxy(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted target HTTPS proxy %q.\n", args[0])
			return nil
		},
	}
}

func newTargetTcpProxiesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target-tcp-proxies",
		Short: "Manage target TCP proxies",
	}
	cmd.AddCommand(
		newTargetTcpProxiesListCommand(cfg, creds),
		newTargetTcpProxiesDescribeCommand(cfg, creds),
		newTargetTcpProxiesCreateCommand(cfg, creds),
		newTargetTcpProxiesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newTargetTcpProxiesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List target TCP proxies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			items, err := client.ListTargetTcpProxies(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), items)
			}
			headers := []string{"NAME", "SERVICE", "PROXY_HEADER"}
			rows := make([][]string, len(items))
			for i, item := range items {
				rows[i] = []string{item.Name, item.Service, item.ProxyHeader}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newTargetTcpProxiesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe TARGET_TCP_PROXY",
		Short: "Describe a target TCP proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			item, err := client.GetTargetTcpProxy(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), item)
		},
	}
}

func newTargetTcpProxiesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateTargetTcpProxyRequest

	cmd := &cobra.Command{
		Use:   "create TARGET_TCP_PROXY",
		Short: "Create a target TCP proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateTargetTcpProxy(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created target TCP proxy %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Service, "service", "", "Backend service")
	cmd.Flags().StringVar(&req.ProxyHeader, "proxy-header", "", "Proxy header")
	cmd.Flags().StringVar(&req.Description, "description", "", "Proxy description")
	return cmd
}

func newTargetTcpProxiesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete TARGET_TCP_PROXY",
		Short: "Delete a target TCP proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteTargetTcpProxy(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted target TCP proxy %q.\n", args[0])
			return nil
		},
	}
}

func newTargetSslProxiesCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "target-ssl-proxies",
		Short: "Manage target SSL proxies",
	}
	cmd.AddCommand(
		newTargetSslProxiesListCommand(cfg, creds),
		newTargetSslProxiesDescribeCommand(cfg, creds),
		newTargetSslProxiesCreateCommand(cfg, creds),
		newTargetSslProxiesDeleteCommand(cfg, creds),
	)
	return cmd
}

func newTargetSslProxiesListCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "list",
		Short: "List target SSL proxies",
		RunE: func(cmd *cobra.Command, _ []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			items, err := client.ListTargetSslProxies(ctx, project)
			if err != nil {
				return err
			}
			format, _ := cmd.Flags().GetString("format")
			if format == "json" {
				return output.PrintJSON(cmd.OutOrStdout(), items)
			}
			headers := []string{"NAME", "SERVICE", "CERTIFICATES"}
			rows := make([][]string, len(items))
			for i, item := range items {
				rows[i] = []string{item.Name, item.Service, fmt.Sprintf("%d", len(item.SslCertificates))}
			}
			return output.PrintTable(cmd.OutOrStdout(), headers, rows)
		},
	}
}

func newTargetSslProxiesDescribeCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "describe TARGET_SSL_PROXY",
		Short: "Describe a target SSL proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			item, err := client.GetTargetSslProxy(ctx, project, args[0])
			if err != nil {
				return err
			}
			return output.PrintJSON(cmd.OutOrStdout(), item)
		},
	}
}

func newTargetSslProxiesCreateCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	var req CreateTargetSslProxyRequest

	cmd := &cobra.Command{
		Use:   "create TARGET_SSL_PROXY",
		Short: "Create a target SSL proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			req.Name = args[0]
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.CreateTargetSslProxy(ctx, project, &req); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Created target SSL proxy %q.\n", req.Name)
			return nil
		},
	}
	cmd.Flags().StringVar(&req.Service, "service", "", "Backend service")
	cmd.Flags().StringArrayVar(&req.SslCertificates, "ssl-certificate", nil, "SSL certificate self link or name")
	cmd.Flags().StringVar(&req.CertificateMap, "certificate-map", "", "Certificate map")
	cmd.Flags().StringVar(&req.Description, "description", "", "Proxy description")
	return cmd
}

func newTargetSslProxiesDeleteCommand(cfg *config.Config, creds *auth.Credentials) *cobra.Command {
	return &cobra.Command{
		Use:   "delete TARGET_SSL_PROXY",
		Short: "Delete a target SSL proxy",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			project, err := requireProject(cmd, cfg)
			if err != nil {
				return err
			}
			ctx := context.Background()
			client, err := makeLoadBalancingClient(ctx, creds)
			if err != nil {
				return err
			}
			if err := client.DeleteTargetSslProxy(ctx, project, args[0]); err != nil {
				return err
			}
			_, _ = fmt.Fprintf(cmd.OutOrStdout(), "Deleted target SSL proxy %q.\n", args[0])
			return nil
		},
	}
}
