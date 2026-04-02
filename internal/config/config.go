package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/BurntSushi/toml"
)

const (
	dirName  = "gcgo"
	fileName = "properties.toml"
)

// Properties holds all configuration values.
type Properties struct {
	Project string `toml:"project,omitempty"`
	Account string `toml:"account,omitempty"`
	Region  string `toml:"region,omitempty"`
	Zone    string `toml:"zone,omitempty"`
}

// Config provides access to gcgo configuration.
type Config struct {
	props Properties
	path  string
}

// Load reads the config file from disk. Returns empty config if file doesn't exist.
func Load() (*Config, error) {
	path, err := configPath()
	if err != nil {
		return nil, fmt.Errorf("resolve config path: %w", err)
	}

	c := &Config{path: path}

	if _, err := os.Stat(path); os.IsNotExist(err) {
		return c, nil
	}

	if _, err := toml.DecodeFile(path, &c.props); err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	return c, nil
}

// Save writes the current config to disk.
func (c *Config) Save() error {
	dir := filepath.Dir(c.path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}

	f, err := os.OpenFile(c.path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("open config file: %w", err)
	}
	defer func() { _ = f.Close() }()

	if err := toml.NewEncoder(f).Encode(c.props); err != nil {
		return fmt.Errorf("write config: %w", err)
	}

	return nil
}

// Project returns the active project, checking flag override > env var > config file.
func (c *Config) Project(flagOverride string) string {
	if flagOverride != "" {
		return flagOverride
	}
	if v := os.Getenv("GCGO_PROJECT"); v != "" {
		return v
	}
	return c.props.Project
}

// Region returns the active region.
func (c *Config) Region() string {
	if v := os.Getenv("GCGO_REGION"); v != "" {
		return v
	}
	return c.props.Region
}

// Zone returns the active zone.
func (c *Config) Zone() string {
	if v := os.Getenv("GCGO_ZONE"); v != "" {
		return v
	}
	return c.props.Zone
}

// Get returns a property by key name.
func (c *Config) Get(key string) (string, bool) {
	switch key {
	case "project":
		return c.props.Project, c.props.Project != ""
	case "account":
		return c.props.Account, c.props.Account != ""
	case "region":
		return c.props.Region, c.props.Region != ""
	case "zone":
		return c.props.Zone, c.props.Zone != ""
	default:
		return "", false
	}
}

// Set sets a property by key name.
func (c *Config) Set(key, value string) error {
	switch key {
	case "project":
		c.props.Project = value
	case "account":
		c.props.Account = value
	case "region":
		c.props.Region = value
	case "zone":
		c.props.Zone = value
	default:
		return fmt.Errorf("unknown property: %s (valid: project, account, region, zone)", key)
	}
	return nil
}

// Unset removes a property by key name.
func (c *Config) Unset(key string) error {
	return c.Set(key, "")
}

// All returns all set properties as a map.
func (c *Config) All() map[string]string {
	m := make(map[string]string)
	if c.props.Project != "" {
		m["project"] = c.props.Project
	}
	if c.props.Account != "" {
		m["account"] = c.props.Account
	}
	if c.props.Region != "" {
		m["region"] = c.props.Region
	}
	if c.props.Zone != "" {
		m["zone"] = c.props.Zone
	}
	return m
}

func configPath() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(cfgDir, dirName, fileName), nil
}
