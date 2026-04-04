# gcgo

A fast, single-binary Google Cloud CLI written in Go. Covers the commands people actually use daily — no Python runtime, instant startup.

> **For AI agents and LLMs:** a full command reference is in [`llms.txt`](./llms.txt). A Claude Code skill is at [`.claude/skills/gcgo/SKILL.md`](./.claude/skills/gcgo/SKILL.md) — install it to get gcgo-aware assistance in any project.

## Install

### From source

```sh
go install github.com/mosajjal/gcgo/cmd/gcgo@latest
gcgo version
```

That installs the latest tagged or default branch version into `$(go env GOPATH)/bin` or `$(go env GOBIN)`. `gcgo version` prints the binary version, commit ID, and build date. You can also install a specific release or revision:

```sh
go install github.com/mosajjal/gcgo/cmd/gcgo@v0.1.0
go install github.com/mosajjal/gcgo/cmd/gcgo@1cd6947
```

### From releases

Download from [GitHub Releases](https://github.com/mosajjal/gcgo/releases) and put the binary in your PATH.

### Build locally

```sh
git clone https://github.com/mosajjal/gcgo.git
cd gcgo
make build
# binary is at ./bin/gcgo
./bin/gcgo version
```

## Auth

gcgo delegates authentication to Application Default Credentials (ADC). If you already use `gcloud`, your existing credentials work automatically.

```sh
# Option 1: use existing gcloud ADC
gcloud auth application-default login

# Option 2: gcgo browser flow (writes ADC-compatible credentials)
gcgo auth login

# Option 3: service account key file
gcgo auth login --service-account-key=key.json

# Option 4: GOOGLE_APPLICATION_CREDENTIALS env var (honoured automatically)
export GOOGLE_APPLICATION_CREDENTIALS=/path/to/key.json

# Check what's active
gcgo auth list
```

## CI / GitHub Actions

gcgo works with [Workload Identity Federation](https://cloud.google.com/iam/docs/workload-identity-federation) — no service account key file needed. The `google-github-actions/auth` action writes an ADC-compatible external credential config and sets `GOOGLE_APPLICATION_CREDENTIALS`; gcgo picks it up automatically on the next call.

```yaml
jobs:
  deploy:
    runs-on: ubuntu-latest
    permissions:
      id-token: write   # required for OIDC token request
      contents: read

    steps:
      - uses: actions/checkout@v4

      - uses: actions/setup-go@v5
        with:
          go-version: stable

      - name: Install gcgo
        run: go install github.com/mosajjal/gcgo/cmd/gcgo@latest

      - uses: google-github-actions/auth@v2
        with:
          workload_identity_provider: projects/PROJECT_NUMBER/locations/global/workloadIdentityPools/POOL/providers/PROVIDER
          service_account: deployer@PROJECT_ID.iam.gserviceaccount.com

      - name: Deploy
        run: |
          gcgo run deploy my-service \
            --image gcr.io/$PROJECT_ID/my-image:${{ github.sha }} \
            --region us-central1
```

For Docker pushes in the same job:

```yaml
      - name: Authenticate Docker and push
        run: |
          gcgo auth configure-docker
          docker build -t gcr.io/$PROJECT_ID/my-image:${{ github.sha }} .
          docker push gcr.io/$PROJECT_ID/my-image:${{ github.sha }}
```

For Terraform:

```yaml
      - name: Terraform apply
        run: GOOGLE_OAUTH_ACCESS_TOKEN=$(gcgo token) terraform apply -auto-approve
```

`GCGO_CONFIG_DIR` is useful when you want to isolate credentials from the default `~/.config/gcgo/` path:

```yaml
      - run: GCGO_CONFIG_DIR=/tmp/gcgo-ci gcgo run services list --region us-central1
```

## Config

```sh
gcgo config set project my-project-id
gcgo config set region us-central1
gcgo config set zone us-central1-a
gcgo config list
```

Config lives at `~/.config/gcgo/properties.toml`. Environment variables and `--project` / `--region` / `--zone` flags override config values.

## Environment Variables

| Variable | Purpose |
|---|---|
| `GCGO_CONFIG_DIR` | Override config and credential directory (default: `~/.config/gcgo/`) |
| `GCGO_CREDENTIALS_HELPER` | External credential helper binary — see below |
| `GCGO_PROJECT` | Default project (same as `gcgo config set project`) |
| `GCGO_REGION` | Default region |
| `GCGO_ZONE` | Default zone |
| `GOOGLE_APPLICATION_CREDENTIALS` | Standard ADC env var, honoured by gcgo |

### GCGO_CONFIG_DIR

Points gcgo at a different directory for both `properties.toml` and `credentials.json`. Useful in CI/CD where secrets are mounted into a known path:

```sh
# CI: credentials are mounted at /run/secrets/gcgo/credentials.json
GCGO_CONFIG_DIR=/run/secrets/gcgo gcgo run services list --region us-central1
```

### GCGO_CREDENTIALS_HELPER

Delegates credential storage and retrieval to an external binary. gcgo calls the helper with a single verb argument:

| Invocation | Direction | Purpose |
|---|---|---|
| `helper get` | helper → stdout (JSON) | Retrieve stored credentials |
| `helper store` | stdin (JSON) → helper | Store new credentials |
| `helper erase` | — | Remove stored credentials |

The JSON format for `get` and `store` is the same as Google's ADC file:

```json
{
  "client_id": "...",
  "client_secret": "...",
  "refresh_token": "...",
  "type": "authorized_user"
}
```

Or for service accounts:

```json
{
  "type": "service_account",
  "project_id": "...",
  "private_key_id": "...",
  "private_key": "...",
  "client_email": "..."
}
```

The helper exits 0 on success. Any non-zero exit code is treated as an error — stderr is shown to the user.

#### Example helpers

**age encryption** (matches the typical dotfiles security model):

```sh
#!/usr/bin/env bash
# ~/.local/bin/gcgo-age-helper
KEY="$HOME/.ssh/id_ed25519"
STORE="$HOME/.config/gcgo/credentials.age"
case "$1" in
  get)   age --decrypt -i "$KEY" "$STORE" ;;
  store) age --encrypt -i "$KEY" -o "$STORE" ;;
  erase) rm -f "$STORE" ;;
esac
```

```sh
export GCGO_CREDENTIALS_HELPER=gcgo-age-helper
```

**OS keyring** (Linux `secret-tool`, macOS `security`):

```sh
#!/usr/bin/env bash
# ~/.local/bin/gcgo-keyring-helper
case "$1" in
  get)   secret-tool lookup service gcgo ;;
  store) secret-tool store --label="gcgo credentials" service gcgo ;;
  erase) secret-tool clear service gcgo ;;
esac
```

**HashiCorp Vault:**

```sh
#!/usr/bin/env bash
# ~/.local/bin/gcgo-vault-helper
case "$1" in
  get)   vault kv get -field=credentials secret/gcgo ;;
  store) vault kv put secret/gcgo credentials=@- ;;
  erase) vault kv delete secret/gcgo ;;
esac
```

**1Password:**

```sh
#!/usr/bin/env bash
# ~/.local/bin/gcgo-op-helper
case "$1" in
  get)   op read "op://Personal/gcgo/credentials" ;;
  store) op item edit gcgo credentials=@- ;;
  erase) op item delete gcgo ;;
esac
```

The helper path can include arguments — `GCGO_CREDENTIALS_HELPER="my-helper --profile prod"` works.

## Shortcuts

Convenience commands at the top level — each is a shortcut for a longer canonical command.

```sh
# Show active identity, project, region, and zone
gcgo whoami
gcgo whoami --format json

# Set project (and optionally region/zone) in one step
gcgo use my-project-id
gcgo use my-project-id --region us-central1 --zone us-central1-a

# Open the GCP console in your browser
gcgo open                  # project dashboard
gcgo open run              # Cloud Run console
gcgo open logs             # Cloud Logging
gcgo open sql              # Cloud SQL
gcgo open gke              # GKE clusters
gcgo open storage          # Cloud Storage browser
gcgo open iam              # IAM & Admin
gcgo open secrets          # Secret Manager
gcgo open builds           # Cloud Build
gcgo open monitoring       # Cloud Monitoring
# …and more: compute, functions, spanner, firestore, bigtable, kms, pubsub, redis, artifacts

# SSH into a VM
gcgo ssh my-vm --zone us-central1-a
gcgo ssh my-vm --zone us-central1-a -- -L 8080:localhost:8080

# Read or stream logs
gcgo logs 'severity>=ERROR'
gcgo logs 'resource.type="cloud_run_revision"' --limit 100
gcgo logs --tail

# Print the current access token (useful for scripting)
gcgo token
gcgo token | docker login -u oauth2accesstoken --password-stdin gcr.io
GOOGLE_OAUTH_ACCESS_TOKEN=$(gcgo token) terraform apply
```

## Commands

```
gcgo projects list
gcgo projects describe PROJECT_ID

gcgo compute instances list --zone us-central1-a
gcgo compute instances create my-vm --machine-type e2-medium --zone us-central1-a
gcgo compute instances delete my-vm --zone us-central1-a
gcgo compute instances start|stop|reset my-vm --zone us-central1-a
gcgo compute instances add-tags my-vm --zone us-central1-a --tags web,prod
gcgo compute instances set-machine-type my-vm --zone us-central1-a --machine-type n2-standard-4
gcgo compute instances attach-disk my-vm --zone us-central1-a --disk my-disk
gcgo compute ssh my-vm --zone us-central1-a
gcgo compute scp ./local.txt my-vm:/tmp/remote.txt --zone us-central1-a
gcgo compute firewall-rules list
gcgo compute firewall-rules create allow-http --allow tcp:80 --source-ranges 0.0.0.0/0
gcgo compute ssl-certificates list
gcgo compute ssl-certificates create my-cert --domains example.com,www.example.com
gcgo compute security-policies list
gcgo compute security-policies create my-policy
gcgo compute security-policies add-rule my-policy --priority 1000 --action allow --src-ip-ranges 1.2.3.4/32
gcgo compute images list
gcgo compute vpn-tunnels list --region us-central1

gcgo dns managed-zones list
gcgo dns managed-zones create my-zone --dns-name example.com. --visibility public
gcgo dns record-sets list --zone my-zone
gcgo dns record-sets create www.example.com. --zone my-zone --type A --ttl 300 --rrdatas 1.2.3.4

gcgo iam service-accounts list
gcgo iam service-accounts create my-sa --display-name "My SA"
gcgo iam service-accounts keys create my-sa@proj.iam.gserviceaccount.com --output-file=key.json
gcgo iam policy get
gcgo iam policy add-binding --member user:alice@example.com --role roles/viewer
gcgo iam deny-policies list ATTACHMENT_POINT
gcgo iam org-policies list projects/my-project

gcgo storage ls
gcgo storage ls gs://my-bucket/prefix/
gcgo storage cp ./file.txt gs://my-bucket/file.txt
gcgo storage cp gs://my-bucket/file.txt ./file.txt
gcgo storage mv gs://my-bucket/old.txt gs://my-bucket/new.txt
gcgo storage rsync ./dist gs://my-bucket/static
gcgo storage rm gs://my-bucket/file.txt
gcgo storage mb gs://new-bucket --location us-central1
gcgo storage rb gs://old-bucket
gcgo storage iam get-policy gs://my-bucket
gcgo storage lifecycle describe gs://my-bucket
gcgo storage retention describe gs://my-bucket

gcgo container clusters list
gcgo container clusters get-credentials my-cluster --location us-central1
gcgo container clusters resize my-cluster --node-pool default-pool --num-nodes 5 --location us-central1
gcgo container node-pools list --cluster my-cluster --location us-central1

gcgo run services list --region us-central1
gcgo run deploy my-service --image gcr.io/proj/img:latest --region us-central1
gcgo run services delete my-service --region us-central1
gcgo run revisions list --region us-central1
gcgo run revisions delete my-revision --region us-central1
gcgo run domain-mappings list --region us-central1

gcgo functions list --region us-central1
gcgo functions deploy my-fn --region us-central1 --runtime go122 --entry-point Handler
gcgo functions call my-fn --region us-central1 --data '{"key":"value"}'

gcgo logging read 'severity=ERROR' --limit 100
gcgo logging tail 'resource.type="gce_instance"'
gcgo logging sinks list
gcgo logging sinks update my-sink --destination bigquery.googleapis.com/projects/proj/datasets/ds
gcgo logging metrics list
gcgo logging exclusions list
gcgo logging buckets list --location global

gcgo pubsub topics list
gcgo pubsub subscriptions list
gcgo pubsub subscriptions pull my-sub --max-messages 10
gcgo pubsub subscriptions ack my-sub --ack-ids ID1,ID2
gcgo pubsub subscriptions seek my-sub --time 2026-01-01T00:00:00Z

gcgo sql instances list
gcgo sql databases list --instance my-db
gcgo sql users list --instance my-db
gcgo sql backups list --instance my-db

gcgo spanner instances list
gcgo spanner databases list --instance my-instance
gcgo spanner databases execute-sql my-db --instance my-instance --sql "SELECT 1"
gcgo spanner backups list --instance my-instance
gcgo spanner operations list --instance my-instance

gcgo scheduler jobs list --location us-central1
gcgo artifacts repositories list --location us-central1
gcgo builds list
gcgo ai models list --region us-central1
gcgo alloydb clusters list --location us-central1
gcgo asset search-all-resources --scope projects/my-project
gcgo billing accounts list
gcgo bigtable instances list
gcgo firestore list
gcgo kms key-rings list --location us-central1
gcgo secrets list
gcgo secrets versions access my-secret latest
gcgo monitoring dashboards list
gcgo redis instances list --location us-central1
gcgo services enable storage.googleapis.com
gcgo tasks queues list --location us-central1
gcgo workflows list --location us-central1
gcgo scc findings list organizations/123/sources/-
gcgo folders list --parent organizations/123456789
gcgo organizations list
gcgo composer environments list --location us-central1
gcgo dataflow jobs list
gcgo dataproc clusters list --region us-central1
gcgo dataplex lakes list --location us-central1
gcgo deploy delivery-pipelines list --location us-central1
gcgo eventarc triggers list --location us-central1
```

Every list/describe command supports `--format json`. Errors go to stderr, data to stdout. Exit code 0 on success, 1 on error.

## Shell Completion

```sh
# bash
gcgo completion bash > /etc/bash_completion.d/gcgo

# zsh
gcgo completion zsh > "${fpath[1]}/_gcgo"

# fish
gcgo completion fish > ~/.config/fish/completions/gcgo.fish
```

## Development

```sh
make build      # build binary
make test       # run unit tests with race detector
make lint       # run golangci-lint
make test-e2e   # run E2E tests (needs GCGO_TEST_PROJECT + auth)
make build-all  # cross-compile for all platforms
make clean      # remove build artifacts
```

## License

[MIT](./LICENSE) — Copyright (c) 2026 Ali Mosajjal
