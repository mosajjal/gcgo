//go:build integration

package e2e

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestComputeInstanceLifecycle(t *testing.T) {
	zone := requireTestEnv(t, "GCGO_TEST_ZONE")
	name := uniqueName("gcgo-e2e-vm")

	gcgo(t, "compute", "instances", "create", name, "--zone", zone, "--machine-type", "e2-micro")
	t.Cleanup(func() {
		_, _ = gcgoMaybe(t, "compute", "instances", "delete", name, "--zone", zone, "--quiet")
	})

	out := gcgo(t, "compute", "instances", "list", "--zone", zone)
	if !strings.Contains(out, name) {
		t.Fatalf("instance %s not found in list output: %s", name, out)
	}

	out = gcgo(t, "compute", "instances", "describe", name, "--zone", zone)
	if !strings.Contains(out, name) {
		t.Fatalf("instance %s not found in describe output: %s", name, out)
	}

	gcgo(t, "compute", "instances", "stop", name, "--zone", zone)
	gcgo(t, "compute", "instances", "start", name, "--zone", zone)
	gcgo(t, "compute", "instances", "delete", name, "--zone", zone, "--quiet")
}

func TestIAMServiceAccountLifecycle(t *testing.T) {
	project := testProject(t)
	accountID := strings.ReplaceAll(uniqueName("gcgoe2e"), "-", "")
	email := accountID + "@" + project + ".iam.gserviceaccount.com"
	member := "serviceAccount:" + email

	gcgo(t, "iam", "service-accounts", "create", accountID, "--display-name", "gcgo e2e")
	t.Cleanup(func() {
		_, _ = gcgoMaybe(t, "iam", "service-accounts", "delete", email)
	})

	out := gcgo(t, "iam", "service-accounts", "list")
	if !strings.Contains(out, email) {
		t.Fatalf("service account %s not found in list output: %s", email, out)
	}

	gcgo(t, "iam", "policy", "add-binding", "--member", member, "--role", "roles/viewer")
	gcgo(t, "iam", "policy", "remove-binding", "--member", member, "--role", "roles/viewer")
	gcgo(t, "iam", "service-accounts", "delete", email)
}

func TestGKEGetCredentials(t *testing.T) {
	cluster := requireTestEnv(t, "GCGO_TEST_GKE_CLUSTER")
	location := requireTestEnv(t, "GCGO_TEST_GKE_LOCATION")
	kubeconfig := filepath.Join(t.TempDir(), "config")

	out := gcgoWithEnv(t, []string{"KUBECONFIG=" + kubeconfig}, "container", "clusters", "get-credentials", cluster, "--location", location)
	if !strings.Contains(out, cluster) {
		t.Fatalf("expected get-credentials output to mention cluster %s: %s", cluster, out)
	}

	data := readFile(t, kubeconfig)
	if !strings.Contains(data, cluster) {
		t.Fatalf("expected kubeconfig to contain cluster %s: %s", cluster, data)
	}
}

func TestRunDeployDescribeDelete(t *testing.T) {
	region := requireTestEnv(t, "GCGO_TEST_RUN_REGION")
	image := requireTestEnv(t, "GCGO_TEST_RUN_IMAGE")
	name := uniqueName("gcgo-e2e-run")

	gcgo(t, "run", "deploy", name, "--region", region, "--image", image)
	t.Cleanup(func() {
		_, _ = gcgoMaybe(t, "run", "services", "delete", name, "--region", region)
	})

	out := gcgo(t, "run", "services", "describe", name, "--region", region)
	if !strings.Contains(out, name) {
		t.Fatalf("service %s not found in describe output: %s", name, out)
	}

	gcgo(t, "run", "services", "delete", name, "--region", region)
}

func TestLoggingRead(t *testing.T) {
	filter := requireTestEnv(t, "GCGO_TEST_LOG_FILTER")

	out := gcgo(t, "logging", "read", filter, "--limit", "1")
	if strings.TrimSpace(out) == "" {
		t.Fatal("expected logging read output")
	}
}

func readFile(t *testing.T, path string) string {
	t.Helper()

	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}
