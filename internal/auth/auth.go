package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	gcauth "cloud.google.com/go/auth"
	authcredentials "cloud.google.com/go/auth/credentials"
	authidtoken "cloud.google.com/go/auth/credentials/idtoken"
	authimpersonate "cloud.google.com/go/auth/credentials/impersonate"
	"google.golang.org/api/option"
)

const (
	credFileName  = "credentials.json"
	scopePlatform = "https://www.googleapis.com/auth/cloud-platform"
)

// Credentials manages GCP authentication state.
type Credentials struct {
	credDir           string
	impersonateTarget string
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
	creds, err := c.detectDefault(ctx, []string{scopePlatform})
	if err != nil {
		return nil, err
	}
	if target := c.impersonateTarget; target != "" {
		creds, err = newImpersonatedAccessCredentials(creds, target, []string{scopePlatform})
		if err != nil {
			return nil, err
		}
	}
	return option.WithAuthCredentials(creds), nil
}

// AccessToken returns an OAuth2 access token using stored credentials or ADC.
// If targetPrincipal is set, the token is generated via service account impersonation.
func (c *Credentials) AccessToken(ctx context.Context, targetPrincipal string, scopes []string) (string, error) {
	return c.accessTokenWithFactories(ctx, c.effectiveImpersonationTarget(targetPrincipal), scopes, c.detectDefault, newImpersonatedAccessCredentials)
}

// IdentityToken returns an ID token for the given audience. If targetPrincipal
// is set, the token is generated via service account impersonation.
func (c *Credentials) IdentityToken(ctx context.Context, audience, targetPrincipal string, includeEmail bool) (string, error) {
	return c.identityTokenWithFactories(ctx, audience, c.effectiveImpersonationTarget(targetPrincipal), includeEmail, c.newIDTokenCredentials, newImpersonatedIDTokenCredentials)
}

// SetImpersonateTarget configures a service account email to impersonate for
// client-backed commands that rely on ClientOption.
func (c *Credentials) SetImpersonateTarget(target string) {
	c.impersonateTarget = target
}

// ImpersonateTarget returns the currently configured global impersonation target.
func (c *Credentials) ImpersonateTarget() string {
	return c.impersonateTarget
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

func (c *Credentials) effectiveImpersonationTarget(override string) string {
	if override != "" {
		return override
	}
	return c.impersonateTarget
}

func (c *Credentials) detectDefault(ctx context.Context, scopes []string) (*gcauth.Credentials, error) {
	if len(scopes) == 0 {
		scopes = []string{scopePlatform}
	}

	data, err := c.credentialsJSON()
	if err == nil {
		creds, err := authcredentials.DetectDefault(&authcredentials.DetectOptions{
			Scopes:          scopes,
			CredentialsJSON: data,
		})
		if err != nil {
			return nil, fmt.Errorf("parse stored credentials: %w", err)
		}
		return creds, nil
	}

	creds, err := authcredentials.DetectDefault(&authcredentials.DetectOptions{
		Scopes: scopes,
	})
	if err != nil {
		return nil, fmt.Errorf("find credentials: %w (run 'gcgo auth login' to authenticate)", err)
	}
	return creds, nil
}

func (c *Credentials) newIDTokenCredentials(audience string) (*gcauth.Credentials, error) {
	data, err := c.credentialsJSON()
	if err != nil {
		data = nil
	}

	creds, err := authidtoken.NewCredentials(&authidtoken.Options{
		Audience:        audience,
		CredentialsJSON: data,
	})
	if err != nil {
		return nil, fmt.Errorf("create id token credentials: %w", err)
	}
	return creds, nil
}

func (c *Credentials) credentialsJSON() ([]byte, error) {
	data, err := os.ReadFile(c.credPath()) //nolint:gosec // path is from our own config dir
	if err != nil {
		return nil, err
	}
	return data, nil
}

func newImpersonatedAccessCredentials(base *gcauth.Credentials, targetPrincipal string, scopes []string) (*gcauth.Credentials, error) {
	creds, err := authimpersonate.NewCredentials(&authimpersonate.CredentialsOptions{
		TargetPrincipal: targetPrincipal,
		Scopes:          scopes,
		Credentials:     base,
	})
	if err != nil {
		return nil, fmt.Errorf("create impersonated access token credentials: %w", err)
	}
	return creds, nil
}

func newImpersonatedIDTokenCredentials(base *gcauth.Credentials, audience, targetPrincipal string, includeEmail bool) (*gcauth.Credentials, error) {
	creds, err := authimpersonate.NewIDTokenCredentials(&authimpersonate.IDTokenOptions{
		Audience:        audience,
		TargetPrincipal: targetPrincipal,
		IncludeEmail:    includeEmail,
		Credentials:     base,
	})
	if err != nil {
		return nil, fmt.Errorf("create impersonated id token credentials: %w", err)
	}
	return creds, nil
}

func (c *Credentials) accessTokenWithFactories(
	ctx context.Context,
	targetPrincipal string,
	scopes []string,
	detect func(context.Context, []string) (*gcauth.Credentials, error),
	impersonate func(*gcauth.Credentials, string, []string) (*gcauth.Credentials, error),
) (string, error) {
	if len(scopes) == 0 {
		scopes = []string{scopePlatform}
	}

	creds, err := detect(ctx, scopes)
	if err != nil {
		return "", err
	}
	if targetPrincipal != "" {
		creds, err = impersonate(creds, targetPrincipal, scopes)
		if err != nil {
			return "", err
		}
	}

	token, err := creds.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch access token: %w", err)
	}
	return token.Value, nil
}

func (c *Credentials) identityTokenWithFactories(
	ctx context.Context,
	audience, targetPrincipal string,
	includeEmail bool,
	direct func(string) (*gcauth.Credentials, error),
	impersonate func(*gcauth.Credentials, string, string, bool) (*gcauth.Credentials, error),
) (string, error) {
	var (
		creds *gcauth.Credentials
		err   error
	)

	if targetPrincipal != "" {
		baseCreds, err := c.detectDefault(ctx, []string{scopePlatform})
		if err != nil {
			return "", err
		}
		creds, err = impersonate(baseCreds, audience, targetPrincipal, includeEmail)
	} else {
		creds, err = direct(audience)
	}
	if err != nil {
		return "", err
	}

	token, err := creds.Token(ctx)
	if err != nil {
		return "", fmt.Errorf("fetch id token: %w", err)
	}
	return token.Value, nil
}
