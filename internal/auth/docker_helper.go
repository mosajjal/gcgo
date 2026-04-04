package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"
)

// RunDockerCredentialHelper handles the Docker credential helper protocol.
// Docker invokes it as: docker-credential-gcgo <verb>
// verbs: get, store, erase, list, version
//
// This is called from main() when os.Args[0] basename is "docker-credential-gcgo".
func RunDockerCredentialHelper(creds *Credentials, args []string) error {
	if len(args) == 0 {
		return fmt.Errorf("usage: docker-credential-gcgo <get|store|erase|list|version>")
	}
	verb := args[0]

	switch verb {
	case "get":
		// Docker writes the registry host to stdin; we return credentials JSON.
		raw, err := io.ReadAll(os.Stdin)
		if err != nil {
			return fmt.Errorf("read registry host: %w", err)
		}
		host := strings.TrimSpace(string(raw))
		if !isSupportedRegistry(host) {
			// Docker expects a specific error message when no credentials are found.
			_, _ = fmt.Fprintln(os.Stderr, "credentials not found in native keychain")
			os.Exit(1)
		}

		ctx := context.Background()
		token, err := creds.AccessToken(ctx, "", nil)
		if err != nil {
			return fmt.Errorf("get access token: %w", err)
		}
		return json.NewEncoder(os.Stdout).Encode(map[string]string{
			"Username": "oauth2accesstoken",
			"Secret":   token,
		})

	case "store":
		// Docker wants us to persist credentials — we don't store Docker creds,
		// our token is always fetched fresh from ADC.
		_, _ = io.ReadAll(os.Stdin)
		return nil

	case "erase":
		// No persistent Docker credentials to remove.
		_, _ = io.ReadAll(os.Stdin)
		return nil

	case "list":
		// Return all registries we handle with their username.
		out := make(map[string]string)
		for _, r := range dockerRegistries {
			out[r] = "oauth2accesstoken"
		}
		return json.NewEncoder(os.Stdout).Encode(out)

	case "version":
		_, _ = fmt.Fprintln(os.Stdout, "gcgo Docker credential helper")
		return nil

	default:
		return fmt.Errorf("unknown verb %q — expected get, store, erase, list, or version", verb)
	}
}

var dockerRegistries = []string{
	"gcr.io",
	"us.gcr.io",
	"eu.gcr.io",
	"asia.gcr.io",
	"us-docker.pkg.dev",
	"europe-docker.pkg.dev",
	"asia-docker.pkg.dev",
}

func isSupportedRegistry(host string) bool {
	for _, r := range dockerRegistries {
		if host == r || strings.HasSuffix(host, "."+r) {
			return true
		}
	}
	return false
}
