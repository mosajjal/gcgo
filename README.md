# gcgo

A fast, single-binary Google Cloud CLI written in Go. Covers the commands people actually use daily — no Python runtime, instant startup.

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

# Option 2: service account key
gcgo auth login --service-account-key=key.json

# Check what's active
gcgo auth list
```

## Config

```sh
gcgo config set project my-project-id
gcgo config set region us-central1
gcgo config set zone us-central1-a
gcgo config list
```

Config lives at `~/.config/gcgo/properties.toml`. Environment variables (`GCGO_PROJECT`, `GCGO_REGION`, `GCGO_ZONE`) and `--project` flags override config values.

## Commands

```
gcgo projects list
gcgo projects describe PROJECT_ID

gcgo compute instances list --zone us-central1-a
gcgo compute instances create my-vm --machine-type e2-medium --zone us-central1-a
gcgo compute instances delete my-vm --zone us-central1-a
gcgo compute instances start|stop|reset my-vm --zone us-central1-a
gcgo compute ssh my-vm --zone us-central1-a
gcgo compute scp ./local.txt my-vm:/tmp/remote.txt --zone us-central1-a
gcgo compute firewall-rules list
gcgo compute firewall-rules create allow-http --allow tcp:80 --source-ranges 0.0.0.0/0

gcgo iam service-accounts list
gcgo iam service-accounts create my-sa --display-name "My SA"
gcgo iam service-accounts keys create my-sa@proj.iam.gserviceaccount.com --output-file=key.json
gcgo iam policy get
gcgo iam policy add-binding --member user:alice@example.com --role roles/viewer

gcgo storage ls
gcgo storage ls gs://my-bucket/prefix/
gcgo storage cp ./file.txt gs://my-bucket/file.txt
gcgo storage cp gs://my-bucket/file.txt ./file.txt
gcgo storage rm gs://my-bucket/file.txt
gcgo storage mb gs://new-bucket --location us-central1
gcgo storage rb gs://old-bucket

gcgo container clusters list
gcgo container clusters get-credentials my-cluster --location us-central1

gcgo run services list --region us-central1
gcgo run deploy my-service --image gcr.io/proj/img:latest --region us-central1
gcgo run services delete my-service --region us-central1

gcgo logging read 'severity=ERROR' --limit 100
gcgo logging tail 'resource.type="gce_instance"'

gcgo pubsub topics list
gcgo scheduler jobs list --location us-central1
gcgo artifacts repositories list --location us-central1
gcgo builds list
gcgo functions list --region us-central1
gcgo ai models list --region us-central1
gcgo alloydb clusters list --location us-central1
gcgo asset search-all-resources --scope projects/my-project
gcgo billing accounts list
gcgo bigtable instances list
gcgo folders list --parent organizations/123456789
gcgo organizations list
gcgo composer environments list --region us-central1
gcgo datacatalog entry-groups list --region us-central1
gcgo dataflow jobs list --region us-central1
gcgo dataplex lakes list --region us-central1
gcgo dataproc clusters list --region us-central1
gcgo deploy delivery-pipelines list --location us-central1
gcgo eventarc triggers list --location us-central1
gcgo firestore list
gcgo spanner instances list
gcgo sql instances list
gcgo kms keyrings list --location us-central1
gcgo secrets list
gcgo monitoring channels list
gcgo redis instances list --location us-central1
gcgo scc sources 123456789
gcgo services list
gcgo tasks queues list --location us-central1
gcgo workflows list --location us-central1
```

Every list/describe command supports `--format json`.

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
make test-e2e   # run E2E tests (needs GCGO_TEST_PROJECT + auth, plus service-specific env vars for some flows)
make build-all  # cross-compile for all platforms
make clean      # remove build artifacts
```

## License

MIT
