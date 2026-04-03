package logging

import (
	"testing"

	loggingapi "google.golang.org/api/logging/v2"
)

func TestSinkFromAPI(t *testing.T) {
	sink := sinkFromAPI(&loggingapi.LogSink{
		Name:              "projects/p/sinks/s1",
		ResourceName:      "projects/p/sinks/s1",
		Description:       "test sink",
		Destination:       "storage.googleapis.com/bucket",
		Filter:            "severity>=ERROR",
		Disabled:          true,
		IncludeChildren:   true,
		InterceptChildren: true,
		CreateTime:        "2024-01-01T00:00:00Z",
		UpdateTime:        "2024-01-02T00:00:00Z",
		WriterIdentity:    "serviceAccount:test@example.com",
	})

	if sink == nil || sink.Name != "projects/p/sinks/s1" || sink.Destination == "" {
		t.Fatalf("unexpected sink: %+v", sink)
	}
}
