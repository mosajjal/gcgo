package scc

import (
	"context"
	"fmt"

	"google.golang.org/api/option"
	securitycenter "google.golang.org/api/securitycenter/v1"
)

// Finding holds SCC finding fields.
type Finding struct {
	Name         string `json:"name"`
	Category     string `json:"category"`
	State        string `json:"state"`
	Severity     string `json:"severity"`
	ResourceName string `json:"resource_name"`
	Mute         string `json:"mute"`
	CreateTime   string `json:"create_time"`
}

// Source holds SCC source fields.
type Source struct {
	Name        string `json:"name"`
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
}

// NotificationConfig holds notification config fields.
type NotificationConfig struct {
	Name           string `json:"name"`
	Description    string `json:"description"`
	PubsubTopic    string `json:"pubsub_topic"`
	Filter         string `json:"filter,omitempty"`
	ServiceAccount string `json:"service_account,omitempty"`
}

// Client defines SCC operations.
type Client interface {
	ListFindings(ctx context.Context, sourceName, filter string) ([]*Finding, error)
	UpdateFindingState(ctx context.Context, findingName, state string) error
	SetFindingMute(ctx context.Context, findingName, mute string) error
	ListSources(ctx context.Context, orgID string) ([]*Source, error)
	ListNotifications(ctx context.Context, orgID string) ([]*NotificationConfig, error)
	CreateNotification(ctx context.Context, orgID, configID, pubsubTopic, filter string) (*NotificationConfig, error)
	DeleteNotification(ctx context.Context, name string) error
	DescribeNotification(ctx context.Context, name string) (*NotificationConfig, error)
}

type gcpClient struct {
	scc *securitycenter.Service
}

// NewClient creates a Client backed by the real GCP SCC API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := securitycenter.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create security center client: %w", err)
	}
	return &gcpClient{scc: svc}, nil
}

func (c *gcpClient) ListFindings(ctx context.Context, sourceName, filter string) ([]*Finding, error) {
	call := c.scc.Organizations.Sources.Findings.List(sourceName)
	if filter != "" {
		call = call.Filter(filter)
	}

	var findings []*Finding
	if err := call.Pages(ctx, func(resp *securitycenter.ListFindingsResponse) error {
		for _, result := range resp.ListFindingsResults {
			findings = append(findings, findingFromAPI(result.Finding))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list findings: %w", err)
	}
	return findings, nil
}

func (c *gcpClient) UpdateFindingState(ctx context.Context, findingName, state string) error {
	s := "ACTIVE"
	if state == "INACTIVE" {
		s = "INACTIVE"
	}
	if _, err := c.scc.Organizations.Sources.Findings.SetState(findingName, &securitycenter.SetFindingStateRequest{State: s}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("update finding state %s: %w", findingName, err)
	}
	return nil
}

func (c *gcpClient) SetFindingMute(ctx context.Context, findingName, mute string) error {
	m := "MUTED"
	if mute == "UNMUTED" {
		m = "UNMUTED"
	}
	if _, err := c.scc.Organizations.Sources.Findings.SetMute(findingName, &securitycenter.SetMuteRequest{Mute: m}).Context(ctx).Do(); err != nil {
		return fmt.Errorf("set finding mute %s: %w", findingName, err)
	}
	return nil
}

func (c *gcpClient) ListSources(ctx context.Context, orgID string) ([]*Source, error) {
	call := c.scc.Organizations.Sources.List("organizations/" + orgID)
	var sources []*Source
	if err := call.Pages(ctx, func(resp *securitycenter.ListSourcesResponse) error {
		for _, s := range resp.Sources {
			sources = append(sources, sourceFromAPI(s))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list sources: %w", err)
	}
	return sources, nil
}

func (c *gcpClient) ListNotifications(ctx context.Context, orgID string) ([]*NotificationConfig, error) {
	call := c.scc.Organizations.NotificationConfigs.List("organizations/" + orgID)
	var configs []*NotificationConfig
	if err := call.Pages(ctx, func(resp *securitycenter.ListNotificationConfigsResponse) error {
		for _, nc := range resp.NotificationConfigs {
			configs = append(configs, notificationFromAPI(nc))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list notification configs: %w", err)
	}
	return configs, nil
}

func (c *gcpClient) CreateNotification(ctx context.Context, orgID, configID, pubsubTopic, filter string) (*NotificationConfig, error) {
	nc, err := c.scc.Organizations.NotificationConfigs.Create("organizations/"+orgID, &securitycenter.NotificationConfig{
		Description: configID,
		PubsubTopic: pubsubTopic,
		StreamingConfig: &securitycenter.StreamingConfig{
			Filter: filter,
		},
	}).ConfigId(configID).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create notification config %s: %w", configID, err)
	}
	return notificationFromAPI(nc), nil
}

func (c *gcpClient) DeleteNotification(ctx context.Context, name string) error {
	if _, err := c.scc.Organizations.NotificationConfigs.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete notification config %s: %w", name, err)
	}
	return nil
}

func (c *gcpClient) DescribeNotification(ctx context.Context, name string) (*NotificationConfig, error) {
	nc, err := c.scc.Organizations.NotificationConfigs.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get notification config %s: %w", name, err)
	}
	return notificationFromAPI(nc), nil
}

func findingFromAPI(f *securitycenter.Finding) *Finding {
	return &Finding{
		Name:         f.Name,
		Category:     f.Category,
		State:        f.State,
		Severity:     f.Severity,
		ResourceName: f.ResourceName,
		Mute:         f.Mute,
		CreateTime:   f.CreateTime,
	}
}

func sourceFromAPI(s *securitycenter.Source) *Source {
	return &Source{
		Name:        s.Name,
		DisplayName: s.DisplayName,
		Description: s.Description,
	}
}

func notificationFromAPI(nc *securitycenter.NotificationConfig) *NotificationConfig {
	cfg := &NotificationConfig{
		Name:           nc.Name,
		Description:    nc.Description,
		PubsubTopic:    nc.PubsubTopic,
		ServiceAccount: nc.ServiceAccount,
	}
	if sc := nc.StreamingConfig; sc != nil {
		cfg.Filter = sc.Filter
	}
	return cfg
}
