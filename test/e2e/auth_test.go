//go:build integration

package e2e

import (
	"strings"
	"testing"
)

func TestAuthList(t *testing.T) {
	out := gcgo(t, "auth", "list")
	// Should either show active account or say no credentials
	if !strings.Contains(out, "Active account") && !strings.Contains(out, "No active credentials") {
		t.Errorf("unexpected auth list output: %s", out)
	}
}

func TestAuthRevokeIdempotent(t *testing.T) {
	// Revoking when nothing stored should not error
	out := gcgo(t, "auth", "revoke")
	if !strings.Contains(out, "removed") {
		t.Errorf("unexpected revoke output: %s", out)
	}
}
