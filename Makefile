.PHONY: build test lint fmt clean build-all vet

BINARY  = gcgo
BUILD   = bin
GO      = go
VERSION = $(shell git describe --tags --always --dirty 2>/dev/null || echo dev)
COMMIT  = $(shell git rev-parse --short HEAD 2>/dev/null || echo none)
LDFLAGS = -s -w -X github.com/mosajjal/gcgo/internal/version.Version=$(VERSION) \
          -X github.com/mosajjal/gcgo/internal/version.GitCommit=$(COMMIT) \
          -X github.com/mosajjal/gcgo/internal/version.BuildTime=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)

build:
	CGO_ENABLED=0 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD)/$(BINARY) ./cmd/gcgo

test:
	$(GO) test -race -count=1 ./...

test-cover:
	$(GO) test -race -coverprofile=coverage.out ./...
	$(GO) tool cover -func=coverage.out

test-e2e:
	$(GO) test -race -tags=integration -count=1 ./test/e2e/...

lint:
	golangci-lint run ./...

vet:
	$(GO) vet ./...

fmt:
	gofmt -w .
	goimports -w .

clean:
	rm -rf $(BUILD) coverage.out

build-all:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD)/$(BINARY)-linux-amd64 ./cmd/gcgo
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD)/$(BINARY)-linux-arm64 ./cmd/gcgo
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD)/$(BINARY)-darwin-amd64 ./cmd/gcgo
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD)/$(BINARY)-darwin-arm64 ./cmd/gcgo
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build -ldflags "$(LDFLAGS)" -o $(BUILD)/$(BINARY)-windows-amd64.exe ./cmd/gcgo
