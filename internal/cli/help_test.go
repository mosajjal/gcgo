package cli

import (
	"bytes"
	"strings"
	"testing"
)

func TestRootHelpIncludesLLMContext(t *testing.T) {
	cmd := NewRootCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "LLM context:") {
		t.Fatal("expected LLM context in root help")
	}
	if !strings.Contains(out, "Last updated: "+helpLastUpdated) {
		t.Fatal("expected last updated date in root help")
	}
	if !strings.Contains(out, "--impersonate-service-account") {
		t.Fatal("expected impersonation flag in root help")
	}
}

func TestCommandHelpIncludesDocsLinks(t *testing.T) {
	cmd := NewRootCommand()
	buf := &bytes.Buffer{}
	cmd.SetOut(buf)
	cmd.SetErr(buf)
	cmd.SetArgs([]string{"compute", "instances", "list", "--help"})

	if err := cmd.Execute(); err != nil {
		t.Fatalf("execute help: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "https://cloud.google.com/compute/docs") {
		t.Fatal("expected official docs link in compute help")
	}
	if !strings.Contains(out, "Last updated: "+helpLastUpdated) {
		t.Fatal("expected last updated date in compute help")
	}
}
