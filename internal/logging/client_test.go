package logging

import (
	"context"
	"fmt"
	"testing"
)

type mockClient struct {
	entries []*Entry
	readErr error
}

func (m *mockClient) ReadLogs(_ context.Context, _, _ string, _ int) ([]*Entry, error) {
	return m.entries, m.readErr
}

func TestMockReadLogs(t *testing.T) {
	mock := &mockClient{
		entries: []*Entry{
			{Timestamp: "2026-04-03T00:00:00Z", Severity: "ERROR", LogName: "syslog", Payload: "disk full"},
			{Timestamp: "2026-04-03T00:01:00Z", Severity: "INFO", LogName: "app", Payload: "started"},
		},
	}

	entries, err := mock.ReadLogs(context.Background(), "proj", "severity=ERROR", 10)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(entries) != 2 {
		t.Errorf("expected 2 entries, got %d", len(entries))
	}
}

func TestMockReadLogsError(t *testing.T) {
	mock := &mockClient{readErr: fmt.Errorf("denied")}
	_, err := mock.ReadLogs(context.Background(), "proj", "", 10)
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestEntryPayloadTruncation(t *testing.T) {
	long := ""
	for i := 0; i < 200; i++ {
		long += "x"
	}
	e := &Entry{Payload: long}
	if len(e.Payload) != 200 {
		t.Errorf("payload len: got %d", len(e.Payload))
	}
	// Truncation happens at display time in commands.go, not in the struct
}
