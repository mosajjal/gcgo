package compute

import (
	"context"
	"errors"
	"fmt"
	"strings"

	compute "cloud.google.com/go/compute/apiv1"
	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// ForwardingRule holds global forwarding rule fields.
type ForwardingRule struct {
	Name                string `json:"name"`
	Region              string `json:"region"`
	IPAddress           string `json:"ip_address"`
	IPProtocol          string `json:"ip_protocol"`
	LoadBalancingScheme string `json:"load_balancing_scheme"`
	BackendService      string `json:"backend_service,omitempty"`
	Target              string `json:"target,omitempty"`
	Description         string `json:"description"`
	SelfLink            string `json:"self_link"`
}

// BackendService holds backend service fields.
type BackendService struct {
	Name                string   `json:"name"`
	Protocol            string   `json:"protocol"`
	LoadBalancingScheme string   `json:"load_balancing_scheme"`
	PortName            string   `json:"port_name"`
	HealthChecks        []string `json:"health_checks,omitempty"`
	Description         string   `json:"description"`
	SelfLink            string   `json:"self_link"`
}

// HealthCheck holds health check fields.
type HealthCheck struct {
	Name             string `json:"name"`
	Type             string `json:"type"`
	Region           string `json:"region"`
	CheckIntervalSec int64  `json:"check_interval_sec"`
	TimeoutSec       int64  `json:"timeout_sec"`
	Port             int64  `json:"port"`
	RequestPath      string `json:"request_path,omitempty"`
	Request          string `json:"request,omitempty"`
	Description      string `json:"description"`
	SelfLink         string `json:"self_link"`
}

// UrlMap holds URL map fields.
type UrlMap struct {
	Name           string `json:"name"`
	DefaultService string `json:"default_service"`
	Description    string `json:"description"`
	SelfLink       string `json:"self_link"`
}

// TargetHttpProxy holds target HTTP proxy fields.
type TargetHttpProxy struct {
	Name        string `json:"name"`
	UrlMap      string `json:"url_map"`
	Description string `json:"description"`
	SelfLink    string `json:"self_link"`
}

// TargetHttpsProxy holds target HTTPS proxy fields.
type TargetHttpsProxy struct {
	Name            string   `json:"name"`
	UrlMap          string   `json:"url_map"`
	SslCertificates []string `json:"ssl_certificates,omitempty"`
	CertificateMap  string   `json:"certificate_map"`
	Description     string   `json:"description"`
	SelfLink        string   `json:"self_link"`
}

// TargetTcpProxy holds target TCP proxy fields.
type TargetTcpProxy struct {
	Name        string `json:"name"`
	Service     string `json:"service"`
	ProxyHeader string `json:"proxy_header"`
	Description string `json:"description"`
	SelfLink    string `json:"self_link"`
}

// TargetSslProxy holds target SSL proxy fields.
type TargetSslProxy struct {
	Name            string   `json:"name"`
	Service         string   `json:"service"`
	SslCertificates []string `json:"ssl_certificates,omitempty"`
	CertificateMap  string   `json:"certificate_map"`
	Description     string   `json:"description"`
	SelfLink        string   `json:"self_link"`
}

// CreateForwardingRuleRequest holds parameters for forwarding rule creation.
type CreateForwardingRuleRequest struct {
	Name                string
	IPAddress           string
	IPProtocol          string
	LoadBalancingScheme string
	BackendService      string
	Target              string
	Description         string
}

// CreateBackendServiceRequest holds parameters for backend service creation.
type CreateBackendServiceRequest struct {
	Name                string
	Protocol            string
	LoadBalancingScheme string
	PortName            string
	HealthChecks        []string
	Description         string
}

// CreateHealthCheckRequest holds parameters for health check creation.
type CreateHealthCheckRequest struct {
	Name             string
	Type             string
	Port             int32
	RequestPath      string
	CheckIntervalSec int32
	TimeoutSec       int32
	Description      string
}

// CreateUrlMapRequest holds parameters for URL map creation.
type CreateUrlMapRequest struct {
	Name           string
	DefaultService string
	Description    string
}

// CreateTargetHttpProxyRequest holds parameters for HTTP proxy creation.
type CreateTargetHttpProxyRequest struct {
	Name        string
	UrlMap      string
	Description string
}

// CreateTargetHttpsProxyRequest holds parameters for HTTPS proxy creation.
type CreateTargetHttpsProxyRequest struct {
	Name            string
	UrlMap          string
	SslCertificates []string
	CertificateMap  string
	Description     string
}

// CreateTargetTcpProxyRequest holds parameters for TCP proxy creation.
type CreateTargetTcpProxyRequest struct {
	Name        string
	Service     string
	ProxyHeader string
	Description string
}

// CreateTargetSslProxyRequest holds parameters for SSL proxy creation.
type CreateTargetSslProxyRequest struct {
	Name            string
	Service         string
	SslCertificates []string
	CertificateMap  string
	Description     string
}

// LoadBalancingClient defines load balancing operations.
type LoadBalancingClient interface {
	// Forwarding rules
	ListForwardingRules(ctx context.Context, project string) ([]*ForwardingRule, error)
	GetForwardingRule(ctx context.Context, project, name string) (*ForwardingRule, error)
	CreateForwardingRule(ctx context.Context, project string, req *CreateForwardingRuleRequest) error
	DeleteForwardingRule(ctx context.Context, project, name string) error

	// Backend services
	ListBackendServices(ctx context.Context, project string) ([]*BackendService, error)
	GetBackendService(ctx context.Context, project, name string) (*BackendService, error)
	CreateBackendService(ctx context.Context, project string, req *CreateBackendServiceRequest) error
	DeleteBackendService(ctx context.Context, project, name string) error

	// Health checks
	ListHealthChecks(ctx context.Context, project string) ([]*HealthCheck, error)
	GetHealthCheck(ctx context.Context, project, name string) (*HealthCheck, error)
	CreateHealthCheck(ctx context.Context, project string, req *CreateHealthCheckRequest) error
	DeleteHealthCheck(ctx context.Context, project, name string) error

	// URL maps
	ListUrlMaps(ctx context.Context, project string) ([]*UrlMap, error)
	GetUrlMap(ctx context.Context, project, name string) (*UrlMap, error)
	CreateUrlMap(ctx context.Context, project string, req *CreateUrlMapRequest) error
	DeleteUrlMap(ctx context.Context, project, name string) error

	// Target HTTP proxies
	ListTargetHttpProxies(ctx context.Context, project string) ([]*TargetHttpProxy, error)
	GetTargetHttpProxy(ctx context.Context, project, name string) (*TargetHttpProxy, error)
	CreateTargetHttpProxy(ctx context.Context, project string, req *CreateTargetHttpProxyRequest) error
	DeleteTargetHttpProxy(ctx context.Context, project, name string) error

	// Target HTTPS proxies
	ListTargetHttpsProxies(ctx context.Context, project string) ([]*TargetHttpsProxy, error)
	GetTargetHttpsProxy(ctx context.Context, project, name string) (*TargetHttpsProxy, error)
	CreateTargetHttpsProxy(ctx context.Context, project string, req *CreateTargetHttpsProxyRequest) error
	DeleteTargetHttpsProxy(ctx context.Context, project, name string) error

	// Target TCP proxies
	ListTargetTcpProxies(ctx context.Context, project string) ([]*TargetTcpProxy, error)
	GetTargetTcpProxy(ctx context.Context, project, name string) (*TargetTcpProxy, error)
	CreateTargetTcpProxy(ctx context.Context, project string, req *CreateTargetTcpProxyRequest) error
	DeleteTargetTcpProxy(ctx context.Context, project, name string) error

	// Target SSL proxies
	ListTargetSslProxies(ctx context.Context, project string) ([]*TargetSslProxy, error)
	GetTargetSslProxy(ctx context.Context, project, name string) (*TargetSslProxy, error)
	CreateTargetSslProxy(ctx context.Context, project string, req *CreateTargetSslProxyRequest) error
	DeleteTargetSslProxy(ctx context.Context, project, name string) error
}

type gcpLoadBalancingClient struct {
	forwardingRules *compute.GlobalForwardingRulesClient
	backendServices *compute.BackendServicesClient
	healthChecks    *compute.HealthChecksClient
	urlMaps         *compute.UrlMapsClient
	httpProxies     *compute.TargetHttpProxiesClient
	httpsProxies    *compute.TargetHttpsProxiesClient
	tcpProxies      *compute.TargetTcpProxiesClient
	sslProxies      *compute.TargetSslProxiesClient
}

// NewLoadBalancingClient creates a LoadBalancingClient backed by the real GCP Compute API.
func NewLoadBalancingClient(ctx context.Context, opts ...option.ClientOption) (LoadBalancingClient, error) {
	frc, err := compute.NewGlobalForwardingRulesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create global forwarding rules client: %w", err)
	}
	bsc, err := compute.NewBackendServicesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create backend services client: %w", err)
	}
	hc, err := compute.NewHealthChecksRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create health checks client: %w", err)
	}
	um, err := compute.NewUrlMapsRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create url maps client: %w", err)
	}
	hp, err := compute.NewTargetHttpProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create target http proxies client: %w", err)
	}
	hip, err := compute.NewTargetHttpsProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create target https proxies client: %w", err)
	}
	tcp, err := compute.NewTargetTcpProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create target tcp proxies client: %w", err)
	}
	ssl, err := compute.NewTargetSslProxiesRESTClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create target ssl proxies client: %w", err)
	}

	return &gcpLoadBalancingClient{
		forwardingRules: frc,
		backendServices: bsc,
		healthChecks:    hc,
		urlMaps:         um,
		httpProxies:     hp,
		httpsProxies:    hip,
		tcpProxies:      tcp,
		sslProxies:      ssl,
	}, nil
}

// Forwarding rules.
func (c *gcpLoadBalancingClient) ListForwardingRules(ctx context.Context, project string) ([]*ForwardingRule, error) {
	it := c.forwardingRules.List(ctx, &computepb.ListGlobalForwardingRulesRequest{Project: project})
	var out []*ForwardingRule
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list forwarding rules: %w", err)
		}
		out = append(out, forwardingRuleFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetForwardingRule(ctx context.Context, project, name string) (*ForwardingRule, error) {
	item, err := c.forwardingRules.Get(ctx, &computepb.GetGlobalForwardingRuleRequest{
		Project:        project,
		ForwardingRule: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get forwarding rule %s: %w", name, err)
	}
	return forwardingRuleFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateForwardingRule(ctx context.Context, project string, req *CreateForwardingRuleRequest) error {
	if req.Target != "" && req.BackendService != "" {
		return fmt.Errorf("create forwarding rule %s: target and backend-service are mutually exclusive", req.Name)
	}
	ipProtocol := strings.ToUpper(req.IPProtocol)
	if ipProtocol == "" {
		ipProtocol = "TCP"
	}
	scheme := strings.ToUpper(req.LoadBalancingScheme)
	if scheme == "" {
		scheme = "EXTERNAL"
	}
	if req.IPProtocol == "" {
		req.IPProtocol = ipProtocol
	}

	resource := &computepb.ForwardingRule{
		Name:                &req.Name,
		IPAddress:           strPtrOrNil(req.IPAddress),
		IPProtocol:          strPtrOrNil(ipProtocol),
		LoadBalancingScheme: strPtrOrNil(scheme),
		Description:         strPtrOrNil(req.Description),
		BackendService:      strPtrOrNil(req.BackendService),
		Target:              strPtrOrNil(req.Target),
	}

	op, err := c.forwardingRules.Insert(ctx, &computepb.InsertGlobalForwardingRuleRequest{
		Project:                project,
		ForwardingRuleResource: resource,
	})
	if err != nil {
		return fmt.Errorf("create forwarding rule %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteForwardingRule(ctx context.Context, project, name string) error {
	op, err := c.forwardingRules.Delete(ctx, &computepb.DeleteGlobalForwardingRuleRequest{
		Project:        project,
		ForwardingRule: name,
	})
	if err != nil {
		return fmt.Errorf("delete forwarding rule %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Backend services.
func (c *gcpLoadBalancingClient) ListBackendServices(ctx context.Context, project string) ([]*BackendService, error) {
	it := c.backendServices.List(ctx, &computepb.ListBackendServicesRequest{Project: project})
	var out []*BackendService
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list backend services: %w", err)
		}
		out = append(out, backendServiceFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetBackendService(ctx context.Context, project, name string) (*BackendService, error) {
	item, err := c.backendServices.Get(ctx, &computepb.GetBackendServiceRequest{
		Project:        project,
		BackendService: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get backend service %s: %w", name, err)
	}
	return backendServiceFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateBackendService(ctx context.Context, project string, req *CreateBackendServiceRequest) error {
	protocol := strings.ToUpper(req.Protocol)
	if protocol == "" {
		protocol = "HTTP"
	}
	scheme := strings.ToUpper(req.LoadBalancingScheme)
	if scheme == "" {
		scheme = "EXTERNAL"
	}
	resource := &computepb.BackendService{
		Name:                &req.Name,
		Protocol:            strPtrOrNil(protocol),
		LoadBalancingScheme: strPtrOrNil(scheme),
		PortName:            strPtrOrNil(req.PortName),
		HealthChecks:        req.HealthChecks,
		Description:         strPtrOrNil(req.Description),
	}

	op, err := c.backendServices.Insert(ctx, &computepb.InsertBackendServiceRequest{
		Project:                project,
		BackendServiceResource: resource,
	})
	if err != nil {
		return fmt.Errorf("create backend service %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteBackendService(ctx context.Context, project, name string) error {
	op, err := c.backendServices.Delete(ctx, &computepb.DeleteBackendServiceRequest{
		Project:        project,
		BackendService: name,
	})
	if err != nil {
		return fmt.Errorf("delete backend service %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Health checks.
func (c *gcpLoadBalancingClient) ListHealthChecks(ctx context.Context, project string) ([]*HealthCheck, error) {
	it := c.healthChecks.List(ctx, &computepb.ListHealthChecksRequest{Project: project})
	var out []*HealthCheck
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list health checks: %w", err)
		}
		out = append(out, healthCheckFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetHealthCheck(ctx context.Context, project, name string) (*HealthCheck, error) {
	item, err := c.healthChecks.Get(ctx, &computepb.GetHealthCheckRequest{
		Project:     project,
		HealthCheck: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get health check %s: %w", name, err)
	}
	return healthCheckFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateHealthCheck(ctx context.Context, project string, req *CreateHealthCheckRequest) error {
	hcType := strings.ToUpper(req.Type)
	if hcType == "" {
		hcType = "HTTP"
	}
	resource := &computepb.HealthCheck{
		Name:             &req.Name,
		Type:             strPtrOrNil(hcType),
		CheckIntervalSec: ptr(req.CheckIntervalSec),
		TimeoutSec:       ptr(req.TimeoutSec),
		Description:      strPtrOrNil(req.Description),
	}

	switch hcType {
	case "TCP":
		resource.TcpHealthCheck = &computepb.TCPHealthCheck{
			Port: ptr(req.Port),
		}
		if req.Port > 0 {
			resource.TcpHealthCheck.Port = ptr(req.Port)
		}
		if req.RequestPath != "" {
			// Intentionally ignored for TCP; the field does not exist in the proto.
		}
	default:
		resource.HttpHealthCheck = &computepb.HTTPHealthCheck{
			RequestPath: strPtrOrNil(req.RequestPath),
		}
		if req.Port > 0 {
			resource.HttpHealthCheck.Port = ptr(req.Port)
		}
	}

	op, err := c.healthChecks.Insert(ctx, &computepb.InsertHealthCheckRequest{
		Project:             project,
		HealthCheckResource: resource,
	})
	if err != nil {
		return fmt.Errorf("create health check %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteHealthCheck(ctx context.Context, project, name string) error {
	op, err := c.healthChecks.Delete(ctx, &computepb.DeleteHealthCheckRequest{
		Project:     project,
		HealthCheck: name,
	})
	if err != nil {
		return fmt.Errorf("delete health check %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// URL maps.
func (c *gcpLoadBalancingClient) ListUrlMaps(ctx context.Context, project string) ([]*UrlMap, error) {
	it := c.urlMaps.List(ctx, &computepb.ListUrlMapsRequest{Project: project})
	var out []*UrlMap
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list url maps: %w", err)
		}
		out = append(out, urlMapFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetUrlMap(ctx context.Context, project, name string) (*UrlMap, error) {
	item, err := c.urlMaps.Get(ctx, &computepb.GetUrlMapRequest{
		Project: project,
		UrlMap:  name,
	})
	if err != nil {
		return nil, fmt.Errorf("get url map %s: %w", name, err)
	}
	return urlMapFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateUrlMap(ctx context.Context, project string, req *CreateUrlMapRequest) error {
	op, err := c.urlMaps.Insert(ctx, &computepb.InsertUrlMapRequest{
		Project: project,
		UrlMapResource: &computepb.UrlMap{
			Name:           &req.Name,
			DefaultService: strPtrOrNil(req.DefaultService),
			Description:    strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create url map %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteUrlMap(ctx context.Context, project, name string) error {
	op, err := c.urlMaps.Delete(ctx, &computepb.DeleteUrlMapRequest{
		Project: project,
		UrlMap:  name,
	})
	if err != nil {
		return fmt.Errorf("delete url map %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Target HTTP proxies.
func (c *gcpLoadBalancingClient) ListTargetHttpProxies(ctx context.Context, project string) ([]*TargetHttpProxy, error) {
	it := c.httpProxies.List(ctx, &computepb.ListTargetHttpProxiesRequest{Project: project})
	var out []*TargetHttpProxy
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list target http proxies: %w", err)
		}
		out = append(out, targetHttpProxyFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetTargetHttpProxy(ctx context.Context, project, name string) (*TargetHttpProxy, error) {
	item, err := c.httpProxies.Get(ctx, &computepb.GetTargetHttpProxyRequest{
		Project:         project,
		TargetHttpProxy: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get target http proxy %s: %w", name, err)
	}
	return targetHttpProxyFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateTargetHttpProxy(ctx context.Context, project string, req *CreateTargetHttpProxyRequest) error {
	op, err := c.httpProxies.Insert(ctx, &computepb.InsertTargetHttpProxyRequest{
		Project: project,
		TargetHttpProxyResource: &computepb.TargetHttpProxy{
			Name:        &req.Name,
			UrlMap:      strPtrOrNil(req.UrlMap),
			Description: strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create target http proxy %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteTargetHttpProxy(ctx context.Context, project, name string) error {
	op, err := c.httpProxies.Delete(ctx, &computepb.DeleteTargetHttpProxyRequest{
		Project:         project,
		TargetHttpProxy: name,
	})
	if err != nil {
		return fmt.Errorf("delete target http proxy %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Target HTTPS proxies.
func (c *gcpLoadBalancingClient) ListTargetHttpsProxies(ctx context.Context, project string) ([]*TargetHttpsProxy, error) {
	it := c.httpsProxies.List(ctx, &computepb.ListTargetHttpsProxiesRequest{Project: project})
	var out []*TargetHttpsProxy
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list target https proxies: %w", err)
		}
		out = append(out, targetHttpsProxyFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetTargetHttpsProxy(ctx context.Context, project, name string) (*TargetHttpsProxy, error) {
	item, err := c.httpsProxies.Get(ctx, &computepb.GetTargetHttpsProxyRequest{
		Project:          project,
		TargetHttpsProxy: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get target https proxy %s: %w", name, err)
	}
	return targetHttpsProxyFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateTargetHttpsProxy(ctx context.Context, project string, req *CreateTargetHttpsProxyRequest) error {
	op, err := c.httpsProxies.Insert(ctx, &computepb.InsertTargetHttpsProxyRequest{
		Project: project,
		TargetHttpsProxyResource: &computepb.TargetHttpsProxy{
			Name:            &req.Name,
			UrlMap:          strPtrOrNil(req.UrlMap),
			SslCertificates: req.SslCertificates,
			CertificateMap:  strPtrOrNil(req.CertificateMap),
			Description:     strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create target https proxy %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteTargetHttpsProxy(ctx context.Context, project, name string) error {
	op, err := c.httpsProxies.Delete(ctx, &computepb.DeleteTargetHttpsProxyRequest{
		Project:          project,
		TargetHttpsProxy: name,
	})
	if err != nil {
		return fmt.Errorf("delete target https proxy %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Target TCP proxies.
func (c *gcpLoadBalancingClient) ListTargetTcpProxies(ctx context.Context, project string) ([]*TargetTcpProxy, error) {
	it := c.tcpProxies.List(ctx, &computepb.ListTargetTcpProxiesRequest{Project: project})
	var out []*TargetTcpProxy
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list target tcp proxies: %w", err)
		}
		out = append(out, targetTcpProxyFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetTargetTcpProxy(ctx context.Context, project, name string) (*TargetTcpProxy, error) {
	item, err := c.tcpProxies.Get(ctx, &computepb.GetTargetTcpProxyRequest{
		Project:        project,
		TargetTcpProxy: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get target tcp proxy %s: %w", name, err)
	}
	return targetTcpProxyFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateTargetTcpProxy(ctx context.Context, project string, req *CreateTargetTcpProxyRequest) error {
	op, err := c.tcpProxies.Insert(ctx, &computepb.InsertTargetTcpProxyRequest{
		Project: project,
		TargetTcpProxyResource: &computepb.TargetTcpProxy{
			Name:        &req.Name,
			Service:     strPtrOrNil(req.Service),
			ProxyHeader: strPtrOrNil(req.ProxyHeader),
			Description: strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create target tcp proxy %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteTargetTcpProxy(ctx context.Context, project, name string) error {
	op, err := c.tcpProxies.Delete(ctx, &computepb.DeleteTargetTcpProxyRequest{
		Project:        project,
		TargetTcpProxy: name,
	})
	if err != nil {
		return fmt.Errorf("delete target tcp proxy %s: %w", name, err)
	}
	return op.Wait(ctx)
}

// Target SSL proxies.
func (c *gcpLoadBalancingClient) ListTargetSslProxies(ctx context.Context, project string) ([]*TargetSslProxy, error) {
	it := c.sslProxies.List(ctx, &computepb.ListTargetSslProxiesRequest{Project: project})
	var out []*TargetSslProxy
	for {
		item, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list target ssl proxies: %w", err)
		}
		out = append(out, targetSslProxyFromProto(item))
	}
	return out, nil
}

func (c *gcpLoadBalancingClient) GetTargetSslProxy(ctx context.Context, project, name string) (*TargetSslProxy, error) {
	item, err := c.sslProxies.Get(ctx, &computepb.GetTargetSslProxyRequest{
		Project:        project,
		TargetSslProxy: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get target ssl proxy %s: %w", name, err)
	}
	return targetSslProxyFromProto(item), nil
}

func (c *gcpLoadBalancingClient) CreateTargetSslProxy(ctx context.Context, project string, req *CreateTargetSslProxyRequest) error {
	op, err := c.sslProxies.Insert(ctx, &computepb.InsertTargetSslProxyRequest{
		Project: project,
		TargetSslProxyResource: &computepb.TargetSslProxy{
			Name:            &req.Name,
			Service:         strPtrOrNil(req.Service),
			SslCertificates: req.SslCertificates,
			CertificateMap:  strPtrOrNil(req.CertificateMap),
			Description:     strPtrOrNil(req.Description),
		},
	})
	if err != nil {
		return fmt.Errorf("create target ssl proxy %s: %w", req.Name, err)
	}
	return op.Wait(ctx)
}

func (c *gcpLoadBalancingClient) DeleteTargetSslProxy(ctx context.Context, project, name string) error {
	op, err := c.sslProxies.Delete(ctx, &computepb.DeleteTargetSslProxyRequest{
		Project:        project,
		TargetSslProxy: name,
	})
	if err != nil {
		return fmt.Errorf("delete target ssl proxy %s: %w", name, err)
	}
	return op.Wait(ctx)
}

func forwardingRuleFromProto(rule *computepb.ForwardingRule) *ForwardingRule {
	return &ForwardingRule{
		Name:                rule.GetName(),
		Region:              rule.GetRegion(),
		IPAddress:           rule.GetIPAddress(),
		IPProtocol:          rule.GetIPProtocol(),
		LoadBalancingScheme: rule.GetLoadBalancingScheme(),
		BackendService:      rule.GetBackendService(),
		Target:              rule.GetTarget(),
		Description:         rule.GetDescription(),
		SelfLink:            rule.GetSelfLink(),
	}
}

func backendServiceFromProto(bs *computepb.BackendService) *BackendService {
	return &BackendService{
		Name:                bs.GetName(),
		Protocol:            bs.GetProtocol(),
		LoadBalancingScheme: bs.GetLoadBalancingScheme(),
		PortName:            bs.GetPortName(),
		HealthChecks:        bs.GetHealthChecks(),
		Description:         bs.GetDescription(),
		SelfLink:            bs.GetSelfLink(),
	}
}

func healthCheckFromProto(hc *computepb.HealthCheck) *HealthCheck {
	out := &HealthCheck{
		Name:             hc.GetName(),
		Type:             hc.GetType(),
		Region:           hc.GetRegion(),
		CheckIntervalSec: int64(hc.GetCheckIntervalSec()),
		TimeoutSec:       int64(hc.GetTimeoutSec()),
		Description:      hc.GetDescription(),
		SelfLink:         hc.GetSelfLink(),
	}
	if httpHC := hc.GetHttpHealthCheck(); httpHC != nil {
		out.Port = int64(httpHC.GetPort())
		out.RequestPath = httpHC.GetRequestPath()
	}
	if tcpHC := hc.GetTcpHealthCheck(); tcpHC != nil {
		out.Port = int64(tcpHC.GetPort())
		out.Request = tcpHC.GetRequest()
	}
	return out
}

func urlMapFromProto(um *computepb.UrlMap) *UrlMap {
	return &UrlMap{
		Name:           um.GetName(),
		DefaultService: um.GetDefaultService(),
		Description:    um.GetDescription(),
		SelfLink:       um.GetSelfLink(),
	}
}

func targetHttpProxyFromProto(proxy *computepb.TargetHttpProxy) *TargetHttpProxy {
	return &TargetHttpProxy{
		Name:        proxy.GetName(),
		UrlMap:      proxy.GetUrlMap(),
		Description: proxy.GetDescription(),
		SelfLink:    proxy.GetSelfLink(),
	}
}

func targetHttpsProxyFromProto(proxy *computepb.TargetHttpsProxy) *TargetHttpsProxy {
	return &TargetHttpsProxy{
		Name:            proxy.GetName(),
		UrlMap:          proxy.GetUrlMap(),
		SslCertificates: proxy.GetSslCertificates(),
		CertificateMap:  proxy.GetCertificateMap(),
		Description:     proxy.GetDescription(),
		SelfLink:        proxy.GetSelfLink(),
	}
}

func targetTcpProxyFromProto(proxy *computepb.TargetTcpProxy) *TargetTcpProxy {
	return &TargetTcpProxy{
		Name:        proxy.GetName(),
		Service:     proxy.GetService(),
		ProxyHeader: proxy.GetProxyHeader(),
		Description: proxy.GetDescription(),
		SelfLink:    proxy.GetSelfLink(),
	}
}

func targetSslProxyFromProto(proxy *computepb.TargetSslProxy) *TargetSslProxy {
	return &TargetSslProxy{
		Name:            proxy.GetName(),
		Service:         proxy.GetService(),
		SslCertificates: proxy.GetSslCertificates(),
		CertificateMap:  proxy.GetCertificateMap(),
		Description:     proxy.GetDescription(),
		SelfLink:        proxy.GetSelfLink(),
	}
}
