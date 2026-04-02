# gcgo Task List

Check off tasks as completed. Work top to bottom — later tasks depend on earlier ones.

## Phase 1: Foundation

### 1.1 Project Skeleton
- [x] Create `cmd/gcgo/main.go` with root command wiring
- [x] Create `internal/cli/root.go` — root cobra command with global flags (`--project`, `--format`, `--quiet`)
- [x] Create `internal/output/output.go` — table printer + JSON output helpers
- [x] Create `internal/cli/errors.go` — user-facing error formatting with fix suggestions
- [x] Create `Makefile` (build, test, lint, fmt, clean, build-all)
- [x] Create `.golangci.yml` with strict linting config
- [x] Add `.gitignore`
- [x] Run `go mod tidy`, verify `go build ./...` passes
- [x] Run `make lint` and `make test` — both must pass with zero output

### 1.2 Configuration
- [x] Create `internal/config/config.go` — load/save `~/.config/gcgo/properties.toml`
- [x] Support keys: `project`, `account`, `region`, `zone`
- [x] Environment variable overrides: `GCGO_PROJECT`, `GCGO_REGION`, `GCGO_ZONE`
- [x] Create `internal/config/commands.go` — `set`, `get`, `list`, `unset` subcommands
- [x] Wire config commands into root command
- [x] Unit tests for config load/save/override logic
- [x] E2E-style test: set a value, get it back, unset it, verify gone (in test/e2e/)

### 1.3 Authentication
- [x] Create `internal/auth/auth.go` — credential loading via ADC
- [x] `gcgo auth login` — OAuth browser flow with loopback redirect
- [x] `gcgo auth login --service-account-key=FILE` — service account JSON key auth
- [x] `gcgo auth list` — show active account/credential info
- [x] `gcgo auth revoke` — remove stored credentials
- [x] Wire auth commands into root command
- [x] Transparent token refresh — handled by cloud.google.com/go/auth/credentials
- [x] Unit tests for credential loading logic
- [x] Integration test: verify ADC detection works when credentials exist (in test/e2e/)

## Phase 2: Core Services

### 2.1 Projects
- [x] Create `internal/projects/client.go` — interface + GCP client wrapper
- [x] `gcgo projects list` — list accessible projects (ID, name, number)
- [x] `gcgo projects describe PROJECT` — full project details
- [x] Wire into root command
- [x] Unit tests with mock client
- [x] Table + JSON output for both commands

### 2.2 Compute Engine — Instances
- [x] Create `internal/compute/client.go` — interface for instance operations
- [x] `gcgo compute instances list` — list instances (name, zone, status, internal/external IP)
- [x] `gcgo compute instances describe INSTANCE` — full details as JSON
- [x] `gcgo compute instances create INSTANCE` — flags: `--machine-type`, `--image-family`, `--image-project`, `--disk-size`, `--zone`, `--network`, `--subnet`, `--tags`
- [x] `gcgo compute instances delete INSTANCE` — with confirmation prompt, `--quiet` skips
- [x] `gcgo compute instances start INSTANCE`
- [x] `gcgo compute instances stop INSTANCE`
- [x] `gcgo compute instances reset INSTANCE`
- [x] Wire into root command under `compute`
- [x] Unit tests for each operation with mock client
- [x] Table + JSON output

### 2.3 Compute Engine — SSH & SCP
- [x] `gcgo compute ssh INSTANCE` — resolve IP, exec system `ssh`
- [x] `gcgo compute scp SRC DST` — resolve IP, exec system `scp`
- [x] Auto-detect zone from instance name if only one match
- [x] Support `--zone` flag and config fallback
- [x] Unit test for IP resolution and command building

### 2.4 Compute Engine — Firewall
- [x] `gcgo compute firewall-rules list` — list rules
- [x] `gcgo compute firewall-rules create RULE` — flags: `--allow`, `--source-ranges`, `--target-tags`, `--network`
- [x] `gcgo compute firewall-rules delete RULE`
- [x] Unit tests with mock client

### 2.5 IAM
- [x] Create `internal/iam/client.go` — interface for IAM operations
- [x] `gcgo iam service-accounts list`
- [x] `gcgo iam service-accounts create NAME` — flags: `--display-name`
- [x] `gcgo iam service-accounts delete EMAIL`
- [x] `gcgo iam service-accounts keys list EMAIL`
- [x] `gcgo iam service-accounts keys create EMAIL --output-file=FILE`
- [x] `gcgo iam service-accounts keys delete KEY_ID --iam-account=EMAIL`
- [x] `gcgo iam policy get` — show project IAM policy
- [x] `gcgo iam policy add-binding` — flags: `--member`, `--role`
- [x] `gcgo iam policy remove-binding` — flags: `--member`, `--role`
- [x] Wire into root command
- [x] Unit tests with mock client

### 2.6 Cloud Storage
- [x] Create `internal/storage/client.go` — interface for storage operations
- [x] `gcgo storage ls [gs://BUCKET/PREFIX]` — list buckets or objects
- [x] `gcgo storage cp SRC DST` — local->GCS, GCS->local
- [x] `gcgo storage rm gs://BUCKET/OBJECT`
- [x] `gcgo storage mb gs://BUCKET` — create bucket, `--location` flag
- [x] `gcgo storage rb gs://BUCKET` — remove bucket
- [x] Parse `gs://` URIs correctly
- [x] Unit tests for URI parsing
- [ ] GCS-to-GCS copy support
- [ ] Large file parallel upload support (>5MB threshold)

## Phase 3: Platform Services

### 3.1 GKE (Container)
- [x] Create `internal/container/client.go` — interface for GKE operations
- [x] `gcgo container clusters list` — name, location, status, node count
- [x] `gcgo container clusters describe CLUSTER` — full details
- [x] `gcgo container clusters get-credentials CLUSTER` — write kubeconfig
- [x] Wire into root command
- [x] Unit tests with mock client

### 3.2 Cloud Run
- [x] Create `internal/run/client.go` — interface for Cloud Run operations
- [x] `gcgo run services list` — name, region, URL, last deployed
- [x] `gcgo run services describe SERVICE`
- [x] `gcgo run deploy SERVICE --image=IMAGE` — flags: `--region`, `--memory`, `--cpu`, `--port`, `--env`, `--allow-unauthenticated`
- [x] `gcgo run services delete SERVICE`
- [x] Wire into root command
- [x] Unit tests with mock client

### 3.3 Logging
- [x] Create `internal/logging/client.go` — interface for logging operations
- [x] `gcgo logging read FILTER` — read log entries, `--limit` flag
- [x] `gcgo logging tail FILTER` — stream logs via server-side streaming
- [x] Wire into root command
- [x] Unit tests for output formatting

## Phase 4: Polish & Testing

### 4.1 End-to-End Test Suite
- [x] Create `test/e2e/` framework — setup/teardown helpers, test project config
- [x] E2E: config set/get/list/unset round-trip
- [x] E2E: auth list + revoke idempotent
- [ ] E2E: create instance -> list -> describe -> stop -> start -> delete
- [ ] E2E: create bucket -> upload file -> list -> download -> delete file -> delete bucket
- [ ] E2E: create service account -> list -> add IAM binding -> remove binding -> delete
- [ ] E2E: GKE get-credentials (requires existing cluster)
- [ ] E2E: Cloud Run deploy -> describe -> delete (requires container image)
- [ ] E2E: logging read with known filter

### 4.2 Security Hardening
- [x] Audit: no credentials in error messages or logs
- [x] Audit: all user input validated before API calls
- [x] Audit: no shell injection in SSH/SCP command building (exec.Command, no shell)
- [x] Audit: service account key files read with restrictive permissions check
- [x] Fuzz testing on URI parsing (`gs://` paths)
- [x] Fuzz testing on config file parsing

### 4.3 Release Prep
- [x] Version injection via ldflags (`gcgo version`)
- [x] `goreleaser` config for cross-platform builds
- [x] Shell completions: bash, zsh, fish (`gcgo completion` — cobra provides this free)
- [x] README.md with installation and basic usage
