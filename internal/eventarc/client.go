package eventarc

import (
	"context"
	"fmt"
	"strings"

	eventarcapi "google.golang.org/api/eventarc/v1"
	"google.golang.org/api/option"
)

// Trigger holds the trigger fields we display.
type Trigger struct {
	Name                 string         `json:"name"`
	Location             string         `json:"location"`
	Destination          string         `json:"destination"`
	EventType            string         `json:"event_type"`
	EventFilters         []*EventFilter `json:"event_filters,omitempty"`
	ServiceAccount       string         `json:"service_account"`
	Channel              string         `json:"channel"`
	EventDataContentType string         `json:"event_data_content_type"`
	CreateTime           string         `json:"create_time"`
	UpdateTime           string         `json:"update_time"`
}

// EventFilter represents a single trigger filter.
type EventFilter struct {
	Attribute string `json:"attribute"`
	Operator  string `json:"operator,omitempty"`
	Value     string `json:"value"`
}

// CreateTriggerRequest holds parameters for trigger creation.
type CreateTriggerRequest struct {
	Name                 string
	EventType            string
	EventFilters         []*EventFilter
	CloudRunService      string
	CloudRunRegion       string
	CloudRunPath         string
	Workflow             string
	HttpEndpoint         string
	ServiceAccount       string
	Channel              string
	EventDataContentType string
}

// Client defines Eventarc operations.
type Client interface {
	ListTriggers(ctx context.Context, project, location string) ([]*Trigger, error)
	GetTrigger(ctx context.Context, project, location, name string) (*Trigger, error)
	CreateTrigger(ctx context.Context, project, location string, req *CreateTriggerRequest) (*Trigger, error)
	DeleteTrigger(ctx context.Context, project, location, name string) error
}

type gcpClient struct {
	svc *eventarcapi.Service
}

// NewClient creates a Client backed by the real Eventarc API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	svc, err := eventarcapi.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create eventarc client: %w", err)
	}
	return &gcpClient{svc: svc}, nil
}

func (c *gcpClient) ListTriggers(ctx context.Context, project, location string) ([]*Trigger, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	call := c.svc.Projects.Locations.Triggers.List(parent).Context(ctx)

	var triggers []*Trigger
	if err := call.Pages(ctx, func(resp *eventarcapi.ListTriggersResponse) error {
		for _, trigger := range resp.Triggers {
			triggers = append(triggers, triggerFromAPI(trigger))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list triggers: %w", err)
	}
	return triggers, nil
}

func (c *gcpClient) GetTrigger(ctx context.Context, project, location, name string) (*Trigger, error) {
	fullName := triggerName(project, location, name)
	trigger, err := c.svc.Projects.Locations.Triggers.Get(fullName).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get trigger %s: %w", name, err)
	}
	return triggerFromAPI(trigger), nil
}

func (c *gcpClient) CreateTrigger(ctx context.Context, project, location string, req *CreateTriggerRequest) (*Trigger, error) {
	parent := fmt.Sprintf("projects/%s/locations/%s", project, location)
	destination, err := buildDestination(req)
	if err != nil {
		return nil, err
	}

	trigger := &eventarcapi.Trigger{
		Destination:          destination,
		EventDataContentType: req.EventDataContentType,
		EventFilters:         buildEventFilters(req.EventType, req.EventFilters),
		ServiceAccount:       req.ServiceAccount,
		Channel:              req.Channel,
	}

	if _, err := c.svc.Projects.Locations.Triggers.Create(parent, trigger).TriggerId(req.Name).Context(ctx).Do(); err != nil {
		return nil, fmt.Errorf("create trigger %s: %w", req.Name, err)
	}

	return &Trigger{
		Name:                 triggerName(project, location, req.Name),
		Location:             location,
		Destination:          destinationString(destination),
		EventType:            req.EventType,
		EventFilters:         req.EventFilters,
		ServiceAccount:       req.ServiceAccount,
		Channel:              req.Channel,
		EventDataContentType: req.EventDataContentType,
	}, nil
}

func (c *gcpClient) DeleteTrigger(ctx context.Context, project, location, name string) error {
	fullName := triggerName(project, location, name)
	if _, err := c.svc.Projects.Locations.Triggers.Delete(fullName).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete trigger %s: %w", name, err)
	}
	return nil
}

func triggerName(project, location, name string) string {
	if strings.HasPrefix(name, "projects/") {
		return name
	}
	return fmt.Sprintf("projects/%s/locations/%s/triggers/%s", project, location, name)
}

func triggerFromAPI(trigger *eventarcapi.Trigger) *Trigger {
	if trigger == nil {
		return nil
	}

	location := ""
	if parts := strings.Split(trigger.Name, "/"); len(parts) >= 4 {
		location = parts[3]
	}

	return &Trigger{
		Name:                 trigger.Name,
		Location:             location,
		Destination:          destinationString(trigger.Destination),
		EventType:            firstEventType(trigger.EventFilters),
		EventFilters:         eventFiltersFromAPI(trigger.EventFilters),
		ServiceAccount:       trigger.ServiceAccount,
		Channel:              trigger.Channel,
		EventDataContentType: trigger.EventDataContentType,
		CreateTime:           trigger.CreateTime,
		UpdateTime:           trigger.UpdateTime,
	}
}

func firstEventType(filters []*eventarcapi.EventFilter) string {
	for _, filter := range filters {
		if filter.Attribute == "type" {
			return filter.Value
		}
	}
	return ""
}

func buildEventFilters(eventType string, extra []*EventFilter) []*eventarcapi.EventFilter {
	filters := make([]*eventarcapi.EventFilter, 0, len(extra)+1)
	if eventType != "" {
		filters = append(filters, &eventarcapi.EventFilter{Attribute: "type", Value: eventType})
	}
	for _, filter := range extra {
		if filter == nil || filter.Attribute == "" {
			continue
		}
		filters = append(filters, &eventarcapi.EventFilter{
			Attribute: filter.Attribute,
			Operator:  filter.Operator,
			Value:     filter.Value,
		})
	}
	return filters
}

func eventFiltersFromAPI(filters []*eventarcapi.EventFilter) []*EventFilter {
	items := make([]*EventFilter, 0, len(filters))
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		items = append(items, &EventFilter{
			Attribute: filter.Attribute,
			Operator:  filter.Operator,
			Value:     filter.Value,
		})
	}
	return items
}

func buildDestination(req *CreateTriggerRequest) (*eventarcapi.Destination, error) {
	switch {
	case req.Workflow != "":
		return &eventarcapi.Destination{Workflow: req.Workflow}, nil
	case req.HttpEndpoint != "":
		return &eventarcapi.Destination{
			HttpEndpoint: &eventarcapi.HttpEndpoint{Uri: req.HttpEndpoint},
		}, nil
	case req.CloudRunService != "":
		if req.CloudRunRegion == "" {
			return nil, fmt.Errorf("--cloud-run-region is required when using --cloud-run-service")
		}
		return &eventarcapi.Destination{
			CloudRun: &eventarcapi.CloudRun{
				Service: req.CloudRunService,
				Region:  req.CloudRunRegion,
				Path:    req.CloudRunPath,
			},
		}, nil
	default:
		return nil, fmt.Errorf("one of --cloud-run-service, --workflow, or --http-endpoint is required")
	}
}

func destinationString(dest *eventarcapi.Destination) string {
	if dest == nil {
		return ""
	}
	switch {
	case dest.CloudRun != nil:
		return fmt.Sprintf("cloud-run:%s/%s", dest.CloudRun.Region, dest.CloudRun.Service)
	case dest.Workflow != "":
		return dest.Workflow
	case dest.HttpEndpoint != nil:
		return dest.HttpEndpoint.Uri
	case dest.CloudFunction != "":
		return dest.CloudFunction
	case dest.Gke != nil:
		return dest.Gke.Service
	default:
		return ""
	}
}
