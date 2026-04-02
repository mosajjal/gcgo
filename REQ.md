# gcgo Requirements

## What This Is

A Go-native CLI for Google Cloud Platform. Single binary, instant startup, no Python runtime.
Not a full gcloud replacement — covers the commands people actually use daily.

## Functional Requirements

### FR-01: Authentication
- Support Application Default Credentials (ADC) — the standard `gcloud auth application-default login` flow
- `gcgo auth login` — opens browser OAuth flow, stores credentials
- `gcgo auth login --service-account-key=FILE` — authenticate with a service account JSON key
- `gcgo auth list` — show active account
- `gcgo auth revoke` — remove stored credentials
- Token refresh handled transparently — user never sees token expiry errors

### FR-02: Configuration
- `gcgo config set KEY VALUE` — set a property (project, region, zone, account)
- `gcgo config get KEY` — read a property
- `gcgo config list` — show all properties
- `gcgo config unset KEY` — remove a property
- Config stored in `~/.config/gcgo/properties.toml`
- Environment variable overrides: `GCGO_PROJECT`, `GCGO_REGION`, `GCGO_ZONE`
- `--project` flag overrides config on any command

### FR-03: Compute Engine
- `gcgo compute instances list` — list instances (name, zone, status, IP)
- `gcgo compute instances describe INSTANCE` — full instance details
- `gcgo compute instances create INSTANCE` — create VM with common flags (machine-type, image, disk-size, network)
- `gcgo compute instances delete INSTANCE` — delete VM (with confirmation prompt)
- `gcgo compute instances start/stop/reset INSTANCE` — lifecycle operations
- `gcgo compute ssh INSTANCE` — SSH into instance (calls system ssh, handles key setup)
- `gcgo compute scp SRC DST` — SCP files to/from instances
- `gcgo compute firewall-rules list/create/delete` — basic firewall management

### FR-04: IAM
- `gcgo iam service-accounts list` — list service accounts
- `gcgo iam service-accounts create/delete NAME` — manage service accounts
- `gcgo iam service-accounts keys create/list/delete` — manage SA keys
- `gcgo iam policy get` — show project IAM policy
- `gcgo iam policy add-binding --member=X --role=Y` — add IAM binding
- `gcgo iam policy remove-binding --member=X --role=Y` — remove IAM binding

### FR-05: Cloud Storage
- `gcgo storage ls [BUCKET/PREFIX]` — list buckets or objects
- `gcgo storage cp SRC DST` — copy files (local<->GCS, GCS<->GCS)
- `gcgo storage rm URI` — delete objects
- `gcgo storage mb BUCKET` — create bucket
- `gcgo storage rb BUCKET` — remove bucket
- Support `gs://` URI scheme
- Parallel uploads for large files

### FR-06: GKE (Container)
- `gcgo container clusters list` — list GKE clusters
- `gcgo container clusters get-credentials CLUSTER` — write kubeconfig entry
- `gcgo container clusters describe CLUSTER` — cluster details

### FR-07: Cloud Run
- `gcgo run services list` — list services
- `gcgo run services describe SERVICE` — service details
- `gcgo run deploy SERVICE --image=IMAGE` — deploy a service
- `gcgo run services delete SERVICE` — delete service

### FR-08: Logging
- `gcgo logging read FILTER` — read log entries with filter
- `gcgo logging tail FILTER` — stream logs in real-time (server-side streaming)

### FR-09: Projects
- `gcgo projects list` — list accessible projects
- `gcgo projects describe PROJECT` — project details

## Non-Functional Requirements

### NFR-01: Performance
- Binary startup to first output: <100ms (cold)
- `gcgo config get project` completes in <10ms
- No lazy-loading tricks that push latency to first API call — initialize what you need, skip what you don't

### NFR-02: Binary Size
- Target: <30MB static binary
- No CGO unless absolutely required (prefer pure Go TLS, DNS)
- `CGO_ENABLED=0` build by default

### NFR-03: Cross-Platform
- Linux amd64/arm64 (primary)
- macOS amd64/arm64
- Windows amd64 (best effort, no special-casing)

### NFR-04: Security
- Never store tokens in plaintext config — use OS keyring or ADC file
- No credentials in CLI history — support `--service-account-key` from file, not inline
- Sanitize error messages — never leak tokens or keys in stderr output
- Validate all user input before passing to API calls

### NFR-05: Compatibility
- Match gcloud flag names where sensible — lower migration friction
- `--format json` everywhere
- Exit codes: 0 success, 1 error
- Machine-parseable JSON output must be stable — no breaking changes

### NFR-06: Error Handling
- Every error message includes what went wrong and what to do about it
- API errors show HTTP status + GCP error message, not raw gRPC codes
- Auth errors suggest `gcgo auth login`
- Missing project errors suggest `gcgo config set project PROJECT_ID`

### NFR-07: Testability
- Every command testable without real GCP credentials (interface-based mocking)
- E2E test suite runs against a real GCP project for release validation
- Race detector clean (`go test -race ./...`)
