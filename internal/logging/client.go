package logging

import (
	"context"
	"errors"
	"fmt"
	"time"

	clogging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	loggingapi "google.golang.org/api/logging/v2"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
	"google.golang.org/protobuf/encoding/protojson"
)

// Entry holds a log entry for display.
type Entry struct {
	Timestamp string `json:"timestamp"`
	Severity  string `json:"severity"`
	LogName   string `json:"log_name"`
	Payload   string `json:"payload"`
}

// Sink holds logging sink fields for display.
type Sink struct {
	Name              string `json:"name"`
	ResourceName      string `json:"resource_name"`
	Description       string `json:"description"`
	Destination       string `json:"destination"`
	Filter            string `json:"filter"`
	Disabled          bool   `json:"disabled"`
	IncludeChildren   bool   `json:"include_children"`
	InterceptChildren bool   `json:"intercept_children"`
	CreateTime        string `json:"create_time"`
	UpdateTime        string `json:"update_time"`
	WriterIdentity    string `json:"writer_identity"`
}

// Client defines logging operations.
type Client interface {
	ReadLogs(ctx context.Context, project, filter string, limit int) ([]*Entry, error)
	ListSinks(ctx context.Context, parent, filter string) ([]*Sink, error)
	GetSink(ctx context.Context, name string) (*Sink, error)
	CreateSink(ctx context.Context, parent string, sink *loggingapi.LogSink) (*Sink, error)
	UpdateSink(ctx context.Context, name string, sink *loggingapi.LogSink, updateMask string) (*Sink, error)
	DeleteSink(ctx context.Context, name string) error
}

type gcpClient struct {
	lc  *clogging.Client
	api *loggingapi.Service
}

// NewClient creates a Client backed by the real Cloud Logging API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	lc, err := clogging.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create logging client: %w", err)
	}
	api, err := loggingapi.NewService(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create logging sink client: %w", err)
	}
	return &gcpClient{lc: lc, api: api}, nil
}

func (c *gcpClient) ReadLogs(ctx context.Context, project, filter string, limit int) ([]*Entry, error) {
	if limit <= 0 {
		limit = 50
	}

	it := c.lc.ListLogEntries(ctx, &loggingpb.ListLogEntriesRequest{
		ResourceNames: []string{"projects/" + project},
		Filter:        filter,
		OrderBy:       "timestamp desc",
		PageSize:      int32(limit),
	})

	var entries []*Entry
	for i := 0; i < limit; i++ {
		e, err := it.Next()
		if errors.Is(err, iterator.Done) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read logs: %w", err)
		}
		entries = append(entries, entryFromProto(e))
	}
	return entries, nil
}

func (c *gcpClient) ListSinks(ctx context.Context, parent, filter string) ([]*Sink, error) {
	call := c.api.Sinks.List(parent).Context(ctx)
	if filter != "" {
		call = call.Filter(filter)
	}

	var sinks []*Sink
	if err := call.Pages(ctx, func(resp *loggingapi.ListSinksResponse) error {
		for _, sink := range resp.Sinks {
			sinks = append(sinks, sinkFromAPI(sink))
		}
		return nil
	}); err != nil {
		return nil, fmt.Errorf("list logging sinks: %w", err)
	}
	return sinks, nil
}

func (c *gcpClient) GetSink(ctx context.Context, name string) (*Sink, error) {
	sink, err := c.api.Sinks.Get(name).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("get logging sink %s: %w", name, err)
	}
	return sinkFromAPI(sink), nil
}

func (c *gcpClient) CreateSink(ctx context.Context, parent string, sink *loggingapi.LogSink) (*Sink, error) {
	if sink == nil {
		return nil, fmt.Errorf("create logging sink: nil sink")
	}
	created, err := c.api.Sinks.Create(parent, sink).Context(ctx).Do()
	if err != nil {
		return nil, fmt.Errorf("create logging sink %s: %w", sink.Name, err)
	}
	return sinkFromAPI(created), nil
}

func (c *gcpClient) UpdateSink(ctx context.Context, name string, sink *loggingapi.LogSink, updateMask string) (*Sink, error) {
	call := c.api.Sinks.Update(name, sink).Context(ctx)
	if updateMask != "" {
		call = call.UpdateMask(updateMask)
	}
	updated, err := call.Do()
	if err != nil {
		return nil, fmt.Errorf("update logging sink %s: %w", name, err)
	}
	return sinkFromAPI(updated), nil
}

func (c *gcpClient) DeleteSink(ctx context.Context, name string) error {
	if _, err := c.api.Sinks.Delete(name).Context(ctx).Do(); err != nil {
		return fmt.Errorf("delete logging sink %s: %w", name, err)
	}
	return nil
}

func entryFromProto(e *loggingpb.LogEntry) *Entry {
	var ts string
	if e.GetTimestamp() != nil {
		ts = e.GetTimestamp().AsTime().Format(time.RFC3339)
	}

	payload := ""
	if e.GetTextPayload() != "" {
		payload = e.GetTextPayload()
	} else if e.GetJsonPayload() != nil {
		b, err := protojson.Marshal(e.GetJsonPayload())
		if err == nil {
			payload = string(b)
		}
	}

	return &Entry{
		Timestamp: ts,
		Severity:  e.GetSeverity().String(),
		LogName:   e.GetLogName(),
		Payload:   payload,
	}
}

func sinkFromAPI(sink *loggingapi.LogSink) *Sink {
	if sink == nil {
		return nil
	}
	return &Sink{
		Name:              sink.Name,
		ResourceName:      sink.ResourceName,
		Description:       sink.Description,
		Destination:       sink.Destination,
		Filter:            sink.Filter,
		Disabled:          sink.Disabled,
		IncludeChildren:   sink.IncludeChildren,
		InterceptChildren: sink.InterceptChildren,
		CreateTime:        sink.CreateTime,
		UpdateTime:        sink.UpdateTime,
		WriterIdentity:    sink.WriterIdentity,
	}
}
