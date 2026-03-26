BINARY_NAME=edge-checker
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/House-lovers7/edge-checker/internal/version.Version=$(VERSION) -X github.com/House-lovers7/edge-checker/internal/version.Commit=$(COMMIT) -X github.com/House-lovers7/edge-checker/internal/version.Date=$(DATE)"

.PHONY: build test clean

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/edge-checker

test:
	go test -race ./...

clean:
	rm -f $(BINARY_NAME)
