package compute

import (
	"testing"

	computepb "cloud.google.com/go/compute/apiv1/computepb"
	"github.com/mosajjal/gcgo/internal/auth"
	"github.com/mosajjal/gcgo/internal/config"
)

func TestLoadBalancingCommandsRegistered(t *testing.T) {
	cmd := NewCommand(&config.Config{}, &auth.Credentials{})
	got := map[string]bool{}
	for _, sub := range cmd.Commands() {
		got[sub.Name()] = true
	}

	want := []string{
		"forwarding-rules",
		"backend-services",
		"health-checks",
		"url-maps",
		"target-http-proxies",
		"target-https-proxies",
		"target-tcp-proxies",
		"target-ssl-proxies",
	}
	for _, name := range want {
		if !got[name] {
			t.Fatalf("missing top-level command %q", name)
		}
	}
}

func TestLoadBalancingProtoMapping(t *testing.T) {
	rule := forwardingRuleFromProto(&computepb.ForwardingRule{
		Name:                ptr("fr-1"),
		IPAddress:           ptr("34.1.2.3"),
		IPProtocol:          ptr("TCP"),
		LoadBalancingScheme: ptr("EXTERNAL"),
		Target:              ptr("global/targetHttpProxies/proxy-1"),
		Description:         ptr("forwarding rule"),
		Region:              ptr("global"),
	})
	if rule.Name != "fr-1" || rule.IPAddress != "34.1.2.3" || rule.Target == "" {
		t.Fatalf("unexpected forwarding rule mapping: %+v", rule)
	}

	hc := healthCheckFromProto(&computepb.HealthCheck{
		Name:             ptr("hc-1"),
		Type:             ptr("HTTP"),
		CheckIntervalSec: ptr(int32(5)),
		TimeoutSec:       ptr(int32(5)),
		Description:      ptr("health check"),
		HttpHealthCheck: &computepb.HTTPHealthCheck{
			Port:        ptr(int32(80)),
			RequestPath: ptr("/healthz"),
		},
	})
	if hc.Name != "hc-1" || hc.Port != 80 || hc.RequestPath != "/healthz" {
		t.Fatalf("unexpected health check mapping: %+v", hc)
	}
}
