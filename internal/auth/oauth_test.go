package auth

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

func TestRandomState(t *testing.T) {
	s1, err := randomState()
	if err != nil {
		t.Fatal(err)
	}
	s2, err := randomState()
	if err != nil {
		t.Fatal(err)
	}

	if len(s1) != 32 { // 16 bytes = 32 hex chars
		t.Errorf("state length: got %d, want 32", len(s1))
	}
	if s1 == s2 {
		t.Error("two random states should differ")
	}
}

func TestAuthorizedUserCredRoundTrip(t *testing.T) {
	cred := authorizedUserCred{
		Type:         "authorized_user",
		ClientID:     "test-client-id",
		ClientSecret: "test-secret",
		RefreshToken: "test-refresh",
		Account:      "user@example.com",
	}

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		t.Fatal(err)
	}

	// Verify it's ADC-compatible (has required fields)
	var check map[string]string
	if err := json.Unmarshal(data, &check); err != nil {
		t.Fatal(err)
	}

	if check["type"] != "authorized_user" {
		t.Errorf("type: got %q", check["type"])
	}
	if check["refresh_token"] != "test-refresh" {
		t.Errorf("refresh_token: got %q", check["refresh_token"])
	}
	if check["account"] != "user@example.com" {
		t.Errorf("account: got %q", check["account"])
	}
}

func TestActiveAccountReadsAuthorizedUser(t *testing.T) {
	dir := t.TempDir()
	creds := New(dir)

	// Write an authorized_user credential with account field
	cred := authorizedUserCred{
		Type:         "authorized_user",
		ClientID:     "test-id",
		ClientSecret: "test-secret",
		RefreshToken: "test-refresh",
		Account:      "alice@example.com",
	}
	data, _ := json.Marshal(cred)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, credFileName), data, 0o600); err != nil {
		t.Fatal(err)
	}

	account, err := creds.ActiveAccount()
	if err != nil {
		t.Fatalf("active account: %v", err)
	}
	if account != "alice@example.com" {
		t.Errorf("account: got %q, want %q", account, "alice@example.com")
	}
}

func TestActiveAccountFallsBackToType(t *testing.T) {
	dir := t.TempDir()
	creds := New(dir)

	// Authorized user without account field (e.g. from gcloud)
	data := []byte(`{"type":"authorized_user","client_id":"x","client_secret":"y","refresh_token":"z"}`)
	if err := os.WriteFile(filepath.Join(dir, credFileName), data, 0o600); err != nil {
		t.Fatal(err)
	}

	account, err := creds.ActiveAccount()
	if err != nil {
		t.Fatalf("active account: %v", err)
	}
	if account != "(authorized user via ADC)" {
		t.Errorf("account: got %q", account)
	}
}
