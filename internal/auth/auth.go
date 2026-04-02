package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"cloud.google.com/go/auth/credentials"
	"google.golang.org/api/option"
)

const (
	credFileName  = "credentials.json"
	scopePlatform = "https://www.googleapis.com/auth/cloud-platform"
)

// Credentials manages GCP authentication state.
type Credentials struct {
	credDir string
}

// New creates a Credentials manager using the given config directory.
func New(credDir string) *Credentials {
	return &Credentials{credDir: credDir}
}

// DefaultCredDir returns the default credential storage directory.
func DefaultCredDir() (string, error) {
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("resolve config dir: %w", err)
	}
	return filepath.Join(cfgDir, "gcgo"), nil
}

// ClientOption returns a google API client option for the active credentials.
// Priority: stored service account key > ADC.
func (c *Credentials) ClientOption(ctx context.Context) (option.ClientOption, error) {
	keyPath := c.credPath()

	if data, err := os.ReadFile(keyPath); err == nil { //nolint:gosec // path is from our own config dir, not user input
		creds, err := credentials.DetectDefault(&credentials.DetectOptions{
			Scopes:          []string{scopePlatform},
			CredentialsJSON: data,
		})
		if err != nil {
			return nil, fmt.Errorf("parse stored credentials: %w", err)
		}
		return option.WithAuthCredentials(creds), nil
	}

	creds, err := credentials.DetectDefault(&credentials.DetectOptions{
		Scopes: []string{scopePlatform},
	})
	if err != nil {
		return nil, fmt.Errorf("find credentials: %w (run 'gcgo auth login' to authenticate)", err)
	}
	return option.WithAuthCredentials(creds), nil
}

// ActiveAccount returns the email of the active credential, if available.
func (c *Credentials) ActiveAccount() (string, error) {
	data, err := os.ReadFile(c.credPath()) //nolint:gosec // path is from our own config dir
	if err != nil {
		adcPath := adcFilePath()
		if adcPath == "" {
			return "", fmt.Errorf("no credentials found")
		}
		data, err = os.ReadFile(adcPath) //nolint:gosec // path is from env or well-known location
		if err != nil {
			return "", fmt.Errorf("no credentials found")
		}
	}

	var info struct {
		ClientEmail string `json:"client_email"`
		Account     string `json:"account"`
		Type        string `json:"type"`
	}
	if err := json.Unmarshal(data, &info); err != nil {
		return "", fmt.Errorf("parse credentials: %w", err)
	}

	if info.ClientEmail != "" {
		return info.ClientEmail, nil
	}
	if info.Account != "" {
		return info.Account, nil
	}
	if info.Type == "authorized_user" {
		return "(authorized user via ADC)", nil
	}
	return "(unknown credential type)", nil
}

// StoreServiceAccountKey copies a service account JSON key file into the cred dir.
func (c *Credentials) StoreServiceAccountKey(keyFile string) error {
	// Check file permissions — warn if too open
	info, err := os.Stat(keyFile)
	if err != nil {
		return fmt.Errorf("stat key file: %w", err)
	}
	if info.Mode().Perm()&0o077 != 0 {
		_, _ = fmt.Fprintf(os.Stderr, "warning: key file %s has broad permissions (%s), consider chmod 600\n",
			keyFile, info.Mode().Perm())
	}

	data, err := os.ReadFile(keyFile) //nolint:gosec // user provides the path explicitly via --service-account-key flag
	if err != nil {
		return fmt.Errorf("read key file: %w", err)
	}

	var check struct {
		Type string `json:"type"`
	}
	if err := json.Unmarshal(data, &check); err != nil {
		return fmt.Errorf("invalid JSON in key file: %w", err)
	}
	if check.Type != "service_account" {
		return fmt.Errorf("key file type is %q, expected \"service_account\"", check.Type)
	}

	if err := os.MkdirAll(c.credDir, 0o700); err != nil {
		return fmt.Errorf("create credential dir: %w", err)
	}

	dst := c.credPath()
	if err := os.WriteFile(dst, data, 0o600); err != nil {
		return fmt.Errorf("store credentials: %w", err)
	}

	return nil
}

// Revoke removes stored credentials.
func (c *Credentials) Revoke() error {
	dst := c.credPath()
	if err := os.Remove(dst); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove credentials: %w", err)
	}
	return nil
}

// HasStoredCredentials reports whether there are stored credentials.
func (c *Credentials) HasStoredCredentials() bool {
	_, err := os.Stat(c.credPath())
	return err == nil
}

func (c *Credentials) credPath() string {
	return filepath.Join(c.credDir, credFileName)
}

func adcFilePath() string {
	if v := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); v != "" {
		return v
	}
	cfgDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	return filepath.Join(cfgDir, "gcloud", "application_default_credentials.json")
}
