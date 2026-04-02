package auth

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// Google's public OAuth2 client for installed applications.
// These are not secrets — Google documents that installed-app
// client IDs are embedded in binaries and cannot be kept confidential.
// See: https://developers.google.com/identity/protocols/oauth2/native-app
const (
	oauthClientID     = "764086051850-6qr4p6gpi6hn506pt8ejuq83di341hur.apps.googleusercontent.com" //nolint:gosec // public installed-app client ID, not a secret
	oauthClientSecret = "d-FL95Q19q7MQmFpd7hHD0Ty"                                                 //nolint:gosec // public installed-app client secret, not confidential
)

// authorizedUserCred is the ADC-compatible JSON format for user credentials.
// The account field is a gcgo extension (ignored by ADC, used by `gcgo auth list`).
type authorizedUserCred struct {
	Type         string `json:"type"`
	ClientID     string `json:"client_id"`
	ClientSecret string `json:"client_secret"` //nolint:gosec // ADC JSON field name, not a leaked secret
	RefreshToken string `json:"refresh_token"` //nolint:gosec // ADC JSON field name, not a leaked secret
	Account      string `json:"account,omitempty"`
}

// BrowserLogin runs the full OAuth2 authorization code flow with a loopback redirect.
// It opens the user's browser, waits for the callback, exchanges the code for tokens,
// and stores the credentials.
func (c *Credentials) BrowserLogin(ctx context.Context) (string, error) {
	// Pick a free port
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return "", fmt.Errorf("start local server: %w", err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	redirectURL := fmt.Sprintf("http://127.0.0.1:%d", port)

	conf := &oauth2.Config{
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		Scopes: []string{
			"openid",
			"https://www.googleapis.com/auth/userinfo.email",
			scopePlatform,
		},
		Endpoint:    google.Endpoint,
		RedirectURL: redirectURL,
	}

	state, err := randomState()
	if err != nil {
		_ = listener.Close()
		return "", fmt.Errorf("generate state: %w", err)
	}

	// Channel to receive the auth code from the callback handler
	type result struct {
		code string
		err  error
	}
	codeCh := make(chan result, 1)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("state") != state {
			codeCh <- result{err: fmt.Errorf("state mismatch — possible CSRF")}
			http.Error(w, "state mismatch", http.StatusBadRequest)
			return
		}

		if errParam := r.URL.Query().Get("error"); errParam != "" {
			desc := r.URL.Query().Get("error_description")
			codeCh <- result{err: fmt.Errorf("auth error: %s — %s", errParam, desc)}
			_, _ = fmt.Fprintf(w, "<html><body><h2>Authentication failed</h2><p>Error: %s</p></body></html>", errParam) //nolint:gosec // errParam is from Google's OAuth response, not user input
			return
		}

		code := r.URL.Query().Get("code")
		if code == "" {
			codeCh <- result{err: fmt.Errorf("no code in callback")}
			http.Error(w, "no code", http.StatusBadRequest)
			return
		}

		codeCh <- result{code: code}
		w.Header().Set("Content-Type", "text/html")
		_, _ = fmt.Fprint(w, `<html><body><h2>Authenticated</h2><p>You can close this tab and return to the terminal.</p></body></html>`)
	})

	server := &http.Server{
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			codeCh <- result{err: fmt.Errorf("local server: %w", err)}
		}
	}()

	// Open the browser
	authURL := conf.AuthCodeURL(state, oauth2.AccessTypeOffline, oauth2.ApprovalForce)
	if err := openBrowser(authURL); err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Could not open browser automatically.\nOpen this URL in your browser:\n\n  %s\n\n", authURL)
	}

	// Wait for the callback
	var res result
	select {
	case res = <-codeCh:
	case <-ctx.Done():
		_ = server.Close()
		return "", ctx.Err()
	}

	// Shut down the server regardless
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = server.Shutdown(shutdownCtx)

	if res.err != nil {
		return "", res.err
	}

	// Exchange code for token
	token, err := conf.Exchange(ctx, res.code)
	if err != nil {
		return "", fmt.Errorf("exchange auth code: %w", err)
	}

	if token.RefreshToken == "" {
		return "", fmt.Errorf("no refresh token returned — try revoking access at https://myaccount.google.com/permissions and login again")
	}

	// Store as ADC-compatible authorized_user JSON
	// Fetch the user's email using the token (best-effort)
	email, _ := fetchUserEmail(ctx, conf, token)
	if email == "" {
		email = "(authorized user)"
	}

	cred := authorizedUserCred{
		Type:         "authorized_user",
		ClientID:     oauthClientID,
		ClientSecret: oauthClientSecret,
		RefreshToken: token.RefreshToken,
		Account:      email,
	}

	data, err := json.MarshalIndent(cred, "", "  ")
	if err != nil {
		return "", fmt.Errorf("marshal credentials: %w", err)
	}

	if err := os.MkdirAll(c.credDir, 0o700); err != nil {
		return "", fmt.Errorf("create credential dir: %w", err)
	}
	if err := os.WriteFile(c.credPath(), data, 0o600); err != nil {
		return "", fmt.Errorf("store credentials: %w", err)
	}

	return email, nil
}

func fetchUserEmail(ctx context.Context, conf *oauth2.Config, token *oauth2.Token) (string, error) {
	client := conf.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()

	var info struct {
		Email string `json:"email"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return "", err
	}
	return info.Email, nil
}

func randomState() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "linux":
		cmd = "xdg-open"
		args = []string{url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	default:
		return fmt.Errorf("unsupported platform %s", runtime.GOOS)
	}

	p := filepath.Clean(cmd)
	return exec.Command(p, args...).Start() //nolint:gosec // cmd is a fixed string per platform, not user input
}
