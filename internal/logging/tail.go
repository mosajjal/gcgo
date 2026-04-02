package logging

import (
	"context"
	"fmt"
	"io"
	"time"

	logging "cloud.google.com/go/logging/apiv2"
	"cloud.google.com/go/logging/apiv2/loggingpb"
	"google.golang.org/api/option"
)

// TailLogs streams log entries matching filter and writes them to w.
func TailLogs(ctx context.Context, w io.Writer, project, filter string, opts ...option.ClientOption) error {
	lc, err := logging.NewClient(ctx, opts...)
	if err != nil {
		return fmt.Errorf("create logging client: %w", err)
	}
	defer func() { _ = lc.Close() }()

	stream, err := lc.TailLogEntries(ctx)
	if err != nil {
		return fmt.Errorf("start tail stream: %w", err)
	}

	err = stream.Send(&loggingpb.TailLogEntriesRequest{
		ResourceNames: []string{"projects/" + project},
		Filter:        filter,
	})
	if err != nil {
		return fmt.Errorf("send tail request: %w", err)
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("receive log entry: %w", err)
		}

		for _, e := range resp.GetEntries() {
			entry := entryFromProto(e)
			ts := entry.Timestamp
			if ts == "" {
				ts = time.Now().Format(time.RFC3339)
			}
			_, _ = fmt.Fprintf(w, "%s  %s  %s  %s\n", ts, entry.Severity, entry.LogName, entry.Payload)
		}
	}
}
