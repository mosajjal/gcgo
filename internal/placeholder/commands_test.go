package placeholder

import (
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func TestPlaceholderLeafIncludesIssueURLAndDocs(t *testing.T) {
	leaf := NewCommand("list", "List placeholder resources", "https://cloud.google.com/example/docs")
	group := NewGroup("example", "Manage example resources", "https://cloud.google.com/example", leaf)
	root := &cobra.Command{Use: "gcgo"}
	root.AddCommand(group)

	if !strings.Contains(leaf.Long, IssueURL) {
		t.Fatalf("expected issue URL in help text, got %q", leaf.Long)
	}
	if !strings.Contains(leaf.Long, "https://cloud.google.com/example/docs") {
		t.Fatalf("expected docs URL in help text, got %q", leaf.Long)
	}

	err := leaf.RunE(leaf, nil)
	if err == nil {
		t.Fatal("expected placeholder command to fail with not-built message")
	}
	if !strings.Contains(err.Error(), "gcgo example list is not built yet") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), IssueURL) {
		t.Fatalf("expected issue URL in error, got %v", err)
	}
}
