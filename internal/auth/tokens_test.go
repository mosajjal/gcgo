package auth

import (
	"context"
	"fmt"
	"testing"

	gcauth "cloud.google.com/go/auth"
)

type staticTokenProvider struct {
	token string
	err   error
}

func (p staticTokenProvider) Token(context.Context) (*gcauth.Token, error) {
	if p.err != nil {
		return nil, p.err
	}
	return &gcauth.Token{Value: p.token}, nil
}

func TestAccessTokenUsesDetectedCredentials(t *testing.T) {
	creds := New("")

	token, err := creds.accessTokenWithFactories(
		context.Background(),
		"",
		nil,
		func(context.Context, []string) (*gcauth.Credentials, error) {
			return gcauth.NewCredentials(&gcauth.CredentialsOptions{
				TokenProvider: staticTokenProvider{token: "direct-token"},
			}), nil
		},
		func(*gcauth.Credentials, string, []string) (*gcauth.Credentials, error) {
			t.Fatal("impersonation factory should not be called")
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("access token: %v", err)
	}
	if token != "direct-token" {
		t.Fatalf("token = %q, want %q", token, "direct-token")
	}
}

func TestAccessTokenUsesImpersonationWhenRequested(t *testing.T) {
	creds := New("")

	token, err := creds.accessTokenWithFactories(
		context.Background(),
		"sa@example.iam.gserviceaccount.com",
		[]string{"scope-a"},
		func(context.Context, []string) (*gcauth.Credentials, error) {
			return gcauth.NewCredentials(&gcauth.CredentialsOptions{
				TokenProvider: staticTokenProvider{token: "base-token"},
			}), nil
		},
		func(_ *gcauth.Credentials, target string, scopes []string) (*gcauth.Credentials, error) {
			if target != "sa@example.iam.gserviceaccount.com" {
				t.Fatalf("target = %q", target)
			}
			if len(scopes) != 1 || scopes[0] != "scope-a" {
				t.Fatalf("scopes = %v", scopes)
			}
			return gcauth.NewCredentials(&gcauth.CredentialsOptions{
				TokenProvider: staticTokenProvider{token: "impersonated-token"},
			}), nil
		},
	)
	if err != nil {
		t.Fatalf("access token: %v", err)
	}
	if token != "impersonated-token" {
		t.Fatalf("token = %q, want %q", token, "impersonated-token")
	}
}

func TestIdentityTokenUsesDirectCredentials(t *testing.T) {
	creds := New("")

	token, err := creds.identityTokenWithFactories(
		context.Background(),
		"https://example.com",
		"",
		false,
		func(audience string) (*gcauth.Credentials, error) {
			if audience != "https://example.com" {
				t.Fatalf("audience = %q", audience)
			}
			return gcauth.NewCredentials(&gcauth.CredentialsOptions{
				TokenProvider: staticTokenProvider{token: "direct-id-token"},
			}), nil
		},
		func(*gcauth.Credentials, string, string, bool) (*gcauth.Credentials, error) {
			t.Fatal("impersonation factory should not be called")
			return nil, nil
		},
	)
	if err != nil {
		t.Fatalf("identity token: %v", err)
	}
	if token != "direct-id-token" {
		t.Fatalf("token = %q, want %q", token, "direct-id-token")
	}
}

func TestIdentityTokenUsesImpersonationWhenRequested(t *testing.T) {
	creds := New("")

	token, err := creds.identityTokenWithFactories(
		context.Background(),
		"https://example.com",
		"sa@example.iam.gserviceaccount.com",
		true,
		func(string) (*gcauth.Credentials, error) {
			t.Fatal("direct ID token factory should not be called")
			return nil, nil
		},
		func(_ *gcauth.Credentials, audience, target string, includeEmail bool) (*gcauth.Credentials, error) {
			if audience != "https://example.com" {
				t.Fatalf("audience = %q", audience)
			}
			if target != "sa@example.iam.gserviceaccount.com" {
				t.Fatalf("target = %q", target)
			}
			if !includeEmail {
				t.Fatal("expected includeEmail to be true")
			}
			return gcauth.NewCredentials(&gcauth.CredentialsOptions{
				TokenProvider: staticTokenProvider{token: "impersonated-id-token"},
			}), nil
		},
	)
	if err != nil {
		t.Fatalf("identity token: %v", err)
	}
	if token != "impersonated-id-token" {
		t.Fatalf("token = %q, want %q", token, "impersonated-id-token")
	}
}

func TestAccessTokenPropagatesTokenErrors(t *testing.T) {
	creds := New("")

	_, err := creds.accessTokenWithFactories(
		context.Background(),
		"",
		nil,
		func(context.Context, []string) (*gcauth.Credentials, error) {
			return gcauth.NewCredentials(&gcauth.CredentialsOptions{
				TokenProvider: staticTokenProvider{err: fmt.Errorf("boom")},
			}), nil
		},
		func(*gcauth.Credentials, string, []string) (*gcauth.Credentials, error) {
			return nil, nil
		},
	)
	if err == nil {
		t.Fatal("expected error")
	}
}
