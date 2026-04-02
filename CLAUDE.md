# gcgo — Google Cloud CLI in Go

A fast, single-binary alternative to `gcloud`. Wraps Google Cloud Go client libraries with a cobra CLI layer.

## Project Rules

### Architecture
- Entrypoint: `cmd/gcgo/main.go`
- All application code lives under `internal/` — nothing in `pkg/`
- Each GCP service gets its own package: `internal/compute/`, `internal/storage/`, etc.
- Shared CLI plumbing (output formatting, table printer, error display) lives in `internal/cli/`
- Auth and config are foundational — `internal/auth/` and `internal/config/`
- Every cobra command lives in a `commands.go` file inside its service package

### Code Style
- Go 1.22+, use standard library where possible
- cobra for CLI, no other CLI frameworks
- google-cloud-go client libraries for API calls — never raw HTTP
- `context.Context` on every function that does I/O
- Errors wrap with `fmt.Errorf("verb noun: %w", err)` — lowercase, no punctuation
- No `panic`, no `os.Exit` outside `main.go`
- No global state. Config and clients passed explicitly via dependency injection
- Interfaces defined where consumed, not where implemented
- Run `go vet ./...` and `golangci-lint run` before considering code done

### Testing
- Table-driven tests with subtests, always `-race`
- Integration tests use build tag `//go:build integration`
- E2E tests in `test/e2e/` — these hit real GCP APIs (require auth + test project)
- Unit tests mock GCP clients via interfaces — never hit the network
- Test helpers use `t.Helper()`
- Golden files in `testdata/` directories where output formatting matters

### Guardrails
- NEVER commit credentials, tokens, or project IDs to the repo
- NEVER use `os.Exit` outside of `main.go`
- NEVER store secrets in config files — delegate to ADC/gcloud auth
- NEVER skip error handling — every error must be checked
- NEVER add a dependency without justification — keep the binary small
- NEVER break existing command signatures once released — flags and args are API
- Check `TASKS.md` before starting work — pick up where you left off
- After completing a task, mark it done in `TASKS.md` with `[x]`
- Run `make test` after every change
- Run `make lint` before marking any task complete

### Output Conventions
- Default output: human-readable tables
- `--format json` flag on every list/describe command
- `--project` flag on every command that needs a project (falls back to config)
- `--quiet` flag suppresses non-essential output
- Errors go to stderr, data goes to stdout

### Config
- Config dir: `~/.config/gcgo/`
- Properties file: `~/.config/gcgo/properties.toml`
- Mimics gcloud's config structure but simpler — project, account, region, zone
- Auth delegates to Application Default Credentials (ADC)

### Dependency Policy
- `github.com/spf13/cobra` — CLI framework
- `cloud.google.com/go/*` — GCP client libraries (add per-service as needed)
- `google.golang.org/api/*` — GCP API clients where cloud.google.com/go doesn't cover
- `github.com/BurntSushi/toml` — config file parsing
- `github.com/olekukonez/tablewriter` or similar — table output
- That's it. Think hard before adding anything else.
