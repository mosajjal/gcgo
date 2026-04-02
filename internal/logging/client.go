package logging

import (
	"context"
	"errors"
	"fmt"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
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

// Client defines logging operations.
type Client interface {
	ReadLogs(ctx context.Context, project, filter string, limit int) ([]*Entry, error)
}

type gcpClient struct {
	lc *logging.Client
}

// NewClient creates a Client backed by the real Cloud Logging API.
func NewClient(ctx context.Context, opts ...option.ClientOption) (Client, error) {
	lc, err := logging.NewClient(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("create logging client: %w", err)
	}
	return &gcpClient{lc: lc}, nil
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
