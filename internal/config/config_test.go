package config

import (
	"path/filepath"
	"testing"

	"github.com/BurntSushi/toml"
)

func newTestConfig(t *testing.T) *Config {
	t.Helper()
	return &Config{path: filepath.Join(t.TempDir(), "properties.toml")}
}

func TestConfigRoundTrip(t *testing.T) {
	c := newTestConfig(t)

	_ = c.Set("project", "my-project")
	_ = c.Set("region", "us-central1")

	if err := c.Save(); err != nil {
		t.Fatalf("save: %v", err)
	}

	// Load from the same file
	loaded := &Config{path: c.path}
	if _, err := toml.DecodeFile(c.path, &loaded.props); err != nil {
		t.Fatalf("reload: %v", err)
	}

	val, ok := loaded.Get("project")
	if !ok || val != "my-project" {
		t.Errorf("project: got %q, want %q", val, "my-project")
	}
	val, ok = loaded.Get("region")
	if !ok || val != "us-central1" {
		t.Errorf("region: got %q, want %q", val, "us-central1")
	}
}

func TestConfigUnset(t *testing.T) {
	c := newTestConfig(t)
	_ = c.Set("zone", "us-east1-b")

	if err := c.Unset("zone"); err != nil {
		t.Fatalf("unset: %v", err)
	}
	if _, ok := c.Get("zone"); ok {
		t.Error("zone should be unset")
	}
}

func TestConfigSetUnknownKey(t *testing.T) {
	c := newTestConfig(t)
	if err := c.Set("bogus", "value"); err == nil {
		t.Fatal("expected error for unknown key")
	}
}

func TestConfigEnvOverride(t *testing.T) {
	c := newTestConfig(t)
	_ = c.Set("project", "from-config")

	t.Setenv("GCGO_PROJECT", "from-env")
	if got := c.Project(""); got != "from-env" {
		t.Errorf("env override: got %q, want %q", got, "from-env")
	}

	if got := c.Project("from-flag"); got != "from-flag" {
		t.Errorf("flag override: got %q, want %q", got, "from-flag")
	}
}

func TestConfigAll(t *testing.T) {
	c := newTestConfig(t)
	_ = c.Set("project", "p")
	_ = c.Set("zone", "z")

	all := c.All()
	if len(all) != 2 {
		t.Errorf("expected 2 properties, got %d", len(all))
	}
	if all["project"] != "p" {
		t.Errorf("project: got %q", all["project"])
	}
	if all["zone"] != "z" {
		t.Errorf("zone: got %q", all["zone"])
	}
}
