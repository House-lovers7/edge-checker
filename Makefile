BINARY_NAME=edge-checker
VERSION=$(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT=$(shell git rev-parse --short HEAD 2>/dev/null || echo "none")
DATE=$(shell date -u +"%Y-%m-%dT%H:%M:%SZ")
LDFLAGS=-ldflags "-X github.com/House-lovers7/edge-checker/internal/version.Version=$(VERSION) -X github.com/House-lovers7/edge-checker/internal/version.Commit=$(COMMIT) -X github.com/House-lovers7/edge-checker/internal/version.Date=$(DATE)"

.PHONY: build test clean vet ci web-install web-dev web-build docker-build release-local

## Go

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/edge-checker

test:
	go test -race ./internal/...

vet:
	go vet ./...

clean:
	rm -f $(BINARY_NAME) edge-checker-*

## CI (runs all checks locally)

ci: vet test build web-build
	@echo "All checks passed."

## Web

web-install:
	cd web && npm install

web-dev:
	cd web && npm run dev

web-build:
	cd web && npm run build

## Docker

docker-build:
	docker build --build-arg VERSION=$(VERSION) --build-arg COMMIT=$(COMMIT) --build-arg DATE=$(DATE) -t edge-checker:$(VERSION) .

## Release (local cross-compile)

PLATFORMS=darwin/amd64 darwin/arm64 linux/amd64 linux/arm64

release-local:
	@for platform in $(PLATFORMS); do \
		GOOS=$${platform%/*} GOARCH=$${platform#*/} \
		go build $(LDFLAGS) -o $(BINARY_NAME)-$${platform%/*}-$${platform#*/} ./cmd/edge-checker; \
		echo "Built: $(BINARY_NAME)-$${platform%/*}-$${platform#*/}"; \
	done
