package monitoring

import (
	"context"
	"errors"
	"fmt"

	monitoring "cloud.google.com/go/monitoring/apiv3/v2"
	"cloud.google.com/go/monitoring/apiv3/v2/monitoringpb"
	dashboard "cloud.google.com/go/monitoring/dashboard/apiv1"
	"cloud.google.com/go/monitoring/dashboard/apiv1/dashboardpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
)

// Dashboard holds dashboard fields.
type Dashboard struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Etag        string `json:"etag,omitempty"`
}

// AlertPolicy holds alerting policy fields.
type AlertPolicy struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Enabled     bool   `json:"enabled"`
}

// NotificationChannel holds notification channel fields.
type NotificationChannel struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Type        string `json:"type"`
	Enabled     bool   `json:"enabled"`
}

// Client defines monitoring operations.
type Client interface {
	// Dashboards
	ListDashboards(ctx context.Context, project string) ([]*Dashboard, error)
	GetDashboard(ctx context.Context, name string) (*Dashboard, error)
	CreateDashboard(ctx context.Context, project, displayName string) (*Dashboard, error)
	DeleteDashboard(ctx context.Context, name string) error

	// Alert Policies
	ListAlertPolicies(ctx context.Context, project string) ([]*AlertPolicy, error)
	GetAlertPolicy(ctx context.Context, name string) (*AlertPolicy, error)
	CreateAlertPolicy(ctx context.Context, project, displayName string) (*AlertPolicy, error)
	DeleteAlertPolicy(ctx context.Context, name string) error

	// Notification Channels
	ListNotificationChannels(ctx context.Context, project string) ([]*NotificationChannel, error)
	GetNotificationChannel(ctx context.Context, name string) (*NotificationChannel, error)
	CreateNotificationChannel(ctx context.Context, project, displayName, channelType string) (*NotificationChannel, error)
	DeleteNotificationChannel(ctx context.Context, name string) error
}

type gcpClient struct {
	dash  *dashboard.DashboardsClient
	alert *monitoring.AlertPolicyClient
	notif *monitoring.NotificationChannelClient
}

// NewClient creates a Client backed by the real GCP APIs.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	dc, err := dashboard.NewDashboardsClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create dashboards client: %w", err)
	}

	ac, err := monitoring.NewAlertPolicyClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create alert policy client: %w", err)
	}

	nc, err := monitoring.NewNotificationChannelClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create notification channel client: %w", err)
	}

	return &gcpClient{dash: dc, alert: ac, notif: nc}, nil
}

// Dashboards.

func (c *gcpClient) ListDashboards(ctx context.Context, project string) ([]*Dashboard, error) {
	it := c.dash.ListDashboards(ctx, &dashboardpb.ListDashboardsRequest{
		Parent: "projects/" + project,
	})

	var dashes []*Dashboard
	for {
		d, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list dashboards: %w", err)
		}
		dashes = append(dashes, dashFromProto(d))
	}
	return dashes, nil
}

func (c *gcpClient) GetDashboard(ctx context.Context, name string) (*Dashboard, error) {
	d, err := c.dash.GetDashboard(ctx, &dashboardpb.GetDashboardRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get dashboard %s: %w", name, err)
	}
	return dashFromProto(d), nil
}

func (c *gcpClient) CreateDashboard(ctx context.Context, project, displayName string) (*Dashboard, error) {
	// Minimal dashboard with a blank grid layout.
	dashJSON := fmt.Sprintf(`{"displayName": %q, "gridLayout": {}}`, displayName)
	var pb dashboardpb.Dashboard
	if err := protojson.Unmarshal([]byte(dashJSON), &pb); err != nil {
		return nil, fmt.Errorf("build dashboard proto: %w", err)
	}

	d, err := c.dash.CreateDashboard(ctx, &dashboardpb.CreateDashboardRequest{
		Parent:    "projects/" + project,
		Dashboard: &pb,
	})
	if err != nil {
		return nil, fmt.Errorf("create dashboard: %w", err)
	}
	return dashFromProto(d), nil
}

func (c *gcpClient) DeleteDashboard(ctx context.Context, name string) error {
	if err := c.dash.DeleteDashboard(ctx, &dashboardpb.DeleteDashboardRequest{
		Name: name,
	}); err != nil {
		return fmt.Errorf("delete dashboard %s: %w", name, err)
	}
	return nil
}

// Alert Policies.

func (c *gcpClient) ListAlertPolicies(ctx context.Context, project string) ([]*AlertPolicy, error) {
	it := c.alert.ListAlertPolicies(ctx, &monitoringpb.ListAlertPoliciesRequest{
		Name: "projects/" + project,
	})

	var policies []*AlertPolicy
	for {
		p, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list alert policies: %w", err)
		}
		policies = append(policies, alertFromProto(p))
	}
	return policies, nil
}

func (c *gcpClient) GetAlertPolicy(ctx context.Context, name string) (*AlertPolicy, error) {
	p, err := c.alert.GetAlertPolicy(ctx, &monitoringpb.GetAlertPolicyRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get alert policy %s: %w", name, err)
	}
	return alertFromProto(p), nil
}

func (c *gcpClient) CreateAlertPolicy(ctx context.Context, project, displayName string) (*AlertPolicy, error) {
	p, err := c.alert.CreateAlertPolicy(ctx, &monitoringpb.CreateAlertPolicyRequest{
		Name: "projects/" + project,
		AlertPolicy: &monitoringpb.AlertPolicy{
			DisplayName: displayName,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create alert policy: %w", err)
	}
	return alertFromProto(p), nil
}

func (c *gcpClient) DeleteAlertPolicy(ctx context.Context, name string) error {
	if err := c.alert.DeleteAlertPolicy(ctx, &monitoringpb.DeleteAlertPolicyRequest{
		Name: name,
	}); err != nil {
		return fmt.Errorf("delete alert policy %s: %w", name, err)
	}
	return nil
}

// Notification Channels.

func (c *gcpClient) ListNotificationChannels(ctx context.Context, project string) ([]*NotificationChannel, error) {
	it := c.notif.ListNotificationChannels(ctx, &monitoringpb.ListNotificationChannelsRequest{
		Name: "projects/" + project,
	})

	var channels []*NotificationChannel
	for {
		ch, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("list notification channels: %w", err)
		}
		channels = append(channels, channelFromProto(ch))
	}
	return channels, nil
}

func (c *gcpClient) GetNotificationChannel(ctx context.Context, name string) (*NotificationChannel, error) {
	ch, err := c.notif.GetNotificationChannel(ctx, &monitoringpb.GetNotificationChannelRequest{
		Name: name,
	})
	if err != nil {
		return nil, fmt.Errorf("get notification channel %s: %w", name, err)
	}
	return channelFromProto(ch), nil
}

func (c *gcpClient) CreateNotificationChannel(ctx context.Context, project, displayName, channelType string) (*NotificationChannel, error) {
	ch, err := c.notif.CreateNotificationChannel(ctx, &monitoringpb.CreateNotificationChannelRequest{
		Name: "projects/" + project,
		NotificationChannel: &monitoringpb.NotificationChannel{
			DisplayName: displayName,
			Type:        channelType,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("create notification channel: %w", err)
	}
	return channelFromProto(ch), nil
}

func (c *gcpClient) DeleteNotificationChannel(ctx context.Context, name string) error {
	if err := c.notif.DeleteNotificationChannel(ctx, &monitoringpb.DeleteNotificationChannelRequest{
		Name: name,
	}); err != nil {
		return fmt.Errorf("delete notification channel %s: %w", name, err)
	}
	return nil
}

func dashFromProto(d *dashboardpb.Dashboard) *Dashboard {
	return &Dashboard{
		Name:        d.GetName(),
		DisplayName: d.GetDisplayName(),
		Etag:        d.GetEtag(),
	}
}

func alertFromProto(p *monitoringpb.AlertPolicy) *AlertPolicy {
	enabled := true
	if p.GetEnabled() != nil {
		enabled = p.GetEnabled().GetValue()
	}
	return &AlertPolicy{
		Name:        p.GetName(),
		DisplayName: p.GetDisplayName(),
		Enabled:     enabled,
	}
}

func channelFromProto(ch *monitoringpb.NotificationChannel) *NotificationChannel {
	enabled := true
	if ch.GetEnabled() != nil {
		enabled = ch.GetEnabled().GetValue()
	}
	return &NotificationChannel{
		Name:        ch.GetName(),
		DisplayName: ch.GetDisplayName(),
		Type:        ch.GetType(),
		Enabled:     enabled,
	}
}
