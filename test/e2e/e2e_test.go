//go:build integration

package e2e

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// gcgo runs the gcgo binary and returns stdout.
func gcgo(t *testing.T, args ...string) string {
	t.Helper()

	bin := gcgoBinary(t)
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(),
		"GCGO_PROJECT="+testProject(t),
	)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		t.Fatalf("gcgo %s failed: %v\nstderr: %s", strings.Join(args, " "), err, stderr.String())
	}

	return stdout.String()
}

// gcgoFail runs gcgo expecting failure and returns stderr.
func gcgoFail(t *testing.T, args ...string) string {
	t.Helper()

	bin := gcgoBinary(t)
	cmd := exec.Command(bin, args...)

	var stderr bytes.Buffer
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err == nil {
		t.Fatalf("expected gcgo %s to fail", strings.Join(args, " "))
	}

	return stderr.String()
}

func gcgoBinary(t *testing.T) string {
	t.Helper()

	// Look for binary in standard locations
	candidates := []string{
		filepath.Join("..", "..", "bin", "gcgo"),
		filepath.Join("bin", "gcgo"),
	}

	if v := os.Getenv("GCGO_BIN"); v != "" {
		candidates = []string{v}
	}

	for _, c := range candidates {
		if _, err := os.Stat(c); err == nil {
			abs, _ := filepath.Abs(c)
			return abs
		}
	}

	t.Fatal("gcgo binary not found — run 'make build' first, or set GCGO_BIN")
	return ""
}

func testProject(t *testing.T) string {
	t.Helper()
	p := os.Getenv("GCGO_TEST_PROJECT")
	if p == "" {
		t.Skip("GCGO_TEST_PROJECT not set — skipping E2E")
	}
	return p
}
