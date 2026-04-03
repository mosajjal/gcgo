package version

import (
	"runtime/debug"
	"testing"
)

func TestResolvedMetadataPrefersInjectedValues(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "deadbeef"},
			{Key: "vcs.time", Value: "2026-04-03T00:00:00Z"},
		},
	}

	Version = "custom"
	GitCommit = "abc123"
	BuildTime = "2026-04-04T00:00:00Z"
	t.Cleanup(func() {
		Version = "dev"
		GitCommit = "none"
		BuildTime = "unknown"
	})

	version, gitCommit, buildTime := resolvedMetadata(bi, true)
	if version != "custom" || gitCommit != "abc123" || buildTime != "2026-04-04T00:00:00Z" {
		t.Fatalf("unexpected metadata: version=%q gitCommit=%q buildTime=%q", version, gitCommit, buildTime)
	}
}

func TestResolvedMetadataFallsBackToBuildInfo(t *testing.T) {
	bi := &debug.BuildInfo{
		Main: debug.Module{Version: "v1.2.3"},
		Settings: []debug.BuildSetting{
			{Key: "vcs.revision", Value: "deadbeef"},
			{Key: "vcs.time", Value: "2026-04-03T00:00:00Z"},
		},
	}

	Version = "dev"
	GitCommit = "none"
	BuildTime = "unknown"

	version, gitCommit, buildTime := resolvedMetadata(bi, true)
	if version != "v1.2.3" || gitCommit != "deadbeef" || buildTime != "2026-04-03T00:00:00Z" {
		t.Fatalf("unexpected metadata: version=%q gitCommit=%q buildTime=%q", version, gitCommit, buildTime)
	}
}
