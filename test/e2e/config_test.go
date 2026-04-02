//go:build integration

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConfigRoundTrip(t *testing.T) {
	// Use a temp config dir so we don't clobber real config
	dir := t.TempDir()
	configDir := filepath.Join(dir, "gcgo")
	if err := os.MkdirAll(configDir, 0o700); err != nil {
		t.Fatal(err)
	}
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Set
	gcgo(t, "config", "set", "project", "test-project-123")

	// Get
	out := gcgo(t, "config", "get", "project")
	if strings.TrimSpace(out) != "test-project-123" {
		t.Errorf("get project: got %q", strings.TrimSpace(out))
	}

	// List
	out = gcgo(t, "config", "list")
	if !strings.Contains(out, "test-project-123") {
		t.Errorf("list should contain project: %s", out)
	}

	// Unset
	gcgo(t, "config", "unset", "project")

	// Verify gone
	stderr := gcgoFail(t, "config", "get", "project")
	if !strings.Contains(stderr, "not set") {
		t.Errorf("expected 'not set' error, got: %s", stderr)
	}
}
