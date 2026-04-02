package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestStoreAndRevokeServiceAccountKey(t *testing.T) {
	dir := t.TempDir()
	creds := New(dir)

	// Create a fake SA key file
	keyFile := filepath.Join(t.TempDir(), "key.json")
	saJSON := `{"type":"service_account","client_email":"test@project.iam.gserviceaccount.com"}`
	if err := os.WriteFile(keyFile, []byte(saJSON), 0o600); err != nil {
		t.Fatal(err)
	}

	if err := creds.StoreServiceAccountKey(keyFile); err != nil {
		t.Fatalf("store: %v", err)
	}

	if !creds.HasStoredCredentials() {
		t.Fatal("expected stored credentials")
	}

	account, err := creds.ActiveAccount()
	if err != nil {
		t.Fatalf("active account: %v", err)
	}
	if account != "test@project.iam.gserviceaccount.com" {
		t.Errorf("account: got %q", account)
	}

	if err := creds.Revoke(); err != nil {
		t.Fatalf("revoke: %v", err)
	}
	if creds.HasStoredCredentials() {
		t.Fatal("credentials should be removed after revoke")
	}
}

func TestStoreRejectsNonServiceAccount(t *testing.T) {
	dir := t.TempDir()
	creds := New(dir)

	keyFile := filepath.Join(t.TempDir(), "key.json")
	if err := os.WriteFile(keyFile, []byte(`{"type":"authorized_user"}`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := creds.StoreServiceAccountKey(keyFile)
	if err == nil {
		t.Fatal("expected error for non-service-account key")
	}
}

func TestStoreRejectsInvalidJSON(t *testing.T) {
	dir := t.TempDir()
	creds := New(dir)

	keyFile := filepath.Join(t.TempDir(), "key.json")
	if err := os.WriteFile(keyFile, []byte(`not json`), 0o600); err != nil {
		t.Fatal(err)
	}

	err := creds.StoreServiceAccountKey(keyFile)
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestRevokeNoCredentials(t *testing.T) {
	dir := t.TempDir()
	creds := New(dir)

	// Should not error when nothing to revoke
	if err := creds.Revoke(); err != nil {
		t.Fatalf("revoke empty: %v", err)
	}
}
