//go:build integration

package e2e

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestStorageCopyRoundTrip(t *testing.T) {
	project := testProject(t)
	bucket := fmt.Sprintf("gcgo-e2e-%d", time.Now().UnixNano())
	workDir := t.TempDir()

	localSrc := filepath.Join(workDir, "source.bin")
	localDst := filepath.Join(workDir, "roundtrip.bin")
	contents := bytes.Repeat([]byte("storage-e2e"), 600000)
	if err := os.WriteFile(localSrc, contents, 0o600); err != nil {
		t.Fatal(err)
	}

	gcgo(t, "storage", "mb", "gs://"+bucket, "--location", "US")
	t.Cleanup(func() {
		_, _ = gcgoMaybe(t, "storage", "rm", "gs://"+bucket+"/source.bin")
		_, _ = gcgoMaybe(t, "storage", "rm", "gs://"+bucket+"/copy.bin")
		_, _ = gcgoMaybe(t, "storage", "rb", "gs://"+bucket)
	})

	gcgo(t, "storage", "cp", localSrc, "gs://"+bucket+"/source.bin")
	gcgo(t, "storage", "cp", "gs://"+bucket+"/source.bin", "gs://"+bucket+"/copy.bin")

	out := gcgo(t, "storage", "ls", "gs://"+bucket)
	if !strings.Contains(out, "source.bin") || !strings.Contains(out, "copy.bin") {
		t.Fatalf("list output missing copied objects: %s", out)
	}

	gcgo(t, "storage", "cp", "gs://"+bucket+"/copy.bin", localDst)

	got, err := os.ReadFile(localDst)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, contents) {
		t.Fatalf("downloaded contents do not match source for project %s", project)
	}

	gcgo(t, "storage", "rm", "gs://"+bucket+"/source.bin")
	gcgo(t, "storage", "rm", "gs://"+bucket+"/copy.bin")

	out = gcgo(t, "storage", "ls", "gs://"+bucket)
	if strings.Contains(out, "source.bin") || strings.Contains(out, "copy.bin") {
		t.Fatalf("expected objects to be removed, got: %s", out)
	}
}

func gcgoMaybe(t *testing.T, args ...string) (string, error) {
	t.Helper()

	bin := gcgoBinary(t)
	cmd := exec.Command(bin, args...)
	cmd.Env = append(os.Environ(), "GCGO_PROJECT="+testProject(t))

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		return stderr.String(), err
	}
	return stdout.String(), nil
}
