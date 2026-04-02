# gcgo — Agent Coding Guide

Read this before writing any code. These are the patterns and conventions for this project.

## Go Style

### Package Layout
```
cmd/gcgo/main.go           — entrypoint, wires everything together
internal/
  cli/                     — root command, output formatting, shared flags
    root.go                — root cobra command + global flags
    output.go              — table printer, JSON formatter
    errors.go              — user-facing error formatting
  auth/                    — authentication
    auth.go                — token sources, credential loading
    commands.go            — cobra commands (login, list, revoke)
  config/                  — configuration management
    config.go              — read/write properties.toml
    commands.go            — cobra commands (set, get, list, unset)
  compute/                 — Compute Engine
    client.go              — GCP client wrapper + interface
    instances.go           — instance operations
    firewall.go            — firewall operations
    ssh.go                 — SSH/SCP helpers
    commands.go            — cobra commands
  iam/                     — IAM
    client.go
    commands.go
  storage/                 — Cloud Storage
    client.go
    commands.go
  container/               — GKE
    client.go
    commands.go
  run/                     — Cloud Run
    client.go
    commands.go
  logging/                 — Cloud Logging
    client.go
    commands.go
  projects/                — Project management
    client.go
    commands.go
test/
  e2e/                     — end-to-end tests (real GCP)
  testutil/                — shared test helpers
```

### Naming
- Files: lowercase, underscores only if needed (`firewall_rules.go`)
- Packages: single word, lowercase (`compute`, not `computeEngine`)
- Commands: match gcloud naming (`instances list`, not `list-instances`)
- Interfaces: verb-er pattern where possible (`Lister`, `Creator`), or describe capability
- Errors: `fmt.Errorf("list instances: %w", err)` — verb noun, lowercase, no period

### Command Structure
Every service package has a `commands.go` that exports a single function:

```go
// NewCommand returns the top-level cobra command for this service.
// It receives dependencies (config, clients) — no global state.
func NewCommand(cfg *config.Config) *cobra.Command {
    cmd := &cobra.Command{
        Use:   "compute",
        Short: "Manage Compute Engine resources",
    }
    cmd.AddCommand(
        newInstancesCommand(cfg),
        newFirewallCommand(cfg),
    )
    return cmd
}
```

### Client Pattern
Each service wraps the GCP client behind an interface for testing:

```go
// Client defines the operations we use from the GCP Compute API.
type Client interface {
    ListInstances(ctx context.Context, project, zone string) ([]*Instance, error)
    GetInstance(ctx context.Context, project, zone, name string) (*Instance, error)
    // ...
}

// gcpClient implements Client using the real GCP SDK.
type gcpClient struct {
    instances *compute.InstancesClient
}
```

### Error Handling
```go
// Wrap with context at every layer
result, err := client.ListInstances(ctx, project, zone)
if err != nil {
    return fmt.Errorf("list instances in %s/%s: %w", project, zone, err)
}

// At the CLI layer, format for humans
func handleError(err error) {
    // Check for known error types and suggest fixes
    // e.g., auth errors → "run gcgo auth login"
}
```

### Output Formatting
```go
// Every list/describe command supports --format
func printInstances(w io.Writer, instances []*Instance, format string) error {
    switch format {
    case "json":
        return json.NewEncoder(w).Encode(instances)
    default:
        return printTable(w, instances)
    }
}
```

### Testing
```go
// Table-driven, always
func TestListInstances(t *testing.T) {
    tests := []struct {
        name      string
        mock      func(*MockClient)
        wantErr   bool
        wantCount int
    }{
        {
            name: "returns instances",
            mock: func(m *MockClient) {
                m.ListResult = []*Instance{{Name: "vm-1"}}
            },
            wantCount: 1,
        },
        {
            name: "handles API error",
            mock: func(m *MockClient) {
                m.ListErr = fmt.Errorf("permission denied")
            },
            wantErr: true,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            mock := &MockClient{}
            tt.mock(mock)
            // test the actual logic
        })
    }
}
```

### What NOT To Do
- No `init()` functions
- No package-level `var client = ...` globals
- No `interface{}` / `any` when a concrete type works
- No `log.Fatal` or `os.Exit` outside `main.go`
- No hand-rolled HTTP requests to GCP — always use client libraries
- No third-party logging library — `log/slog` from stdlib is fine
- No `panic` for recoverable errors
- No mocking frameworks — plain struct mocks implementing interfaces
- No `cobra.MarkFlagRequired` where a positional arg makes more sense
