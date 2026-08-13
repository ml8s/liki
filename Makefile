.PHONY: all build build-linux lint vet check test test-all clean run

BINARY=liki
VERSION=$(shell cat cmd/liki/VERSION 2>/dev/null || echo dev)
BUILD_TIME=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags="-X main.BuildTime=$(VERSION)@$(BUILD_TIME)"

all: build

build:
	go build $(LDFLAGS) -o $(BINARY) ./cmd/liki/

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/liki-engine ./cmd/liki/

lint:
	golangci-lint run ./...

vet:
	go vet ./...

check: lint vet build test

test:
	go test ./internal/... -count=1

update: ## 拉取最新引擎代码（部署/测试前必做）
	git pull --ff-only origin master

test-all: update
	scripts/ci-engine.sh

clean:
	rm -f $(BINARY) bin/liki-engine

run:
	go run ./cmd/liki/

VERSION_FILE := cmd/liki/VERSION

version-patch: ## Bump PATCH (1.9.0 → 1.9.1)
	@V=$$(cat $(VERSION_FILE)); \
	MAJOR=$${V%%.*}; \
	REST=$${V#*.}; \
	MINOR=$${REST%.*}; \
	PATCH=$${REST#*.}; \
	echo "$$MAJOR.$$MINOR.$$((PATCH + 1))" > $(VERSION_FILE); \
	echo "✅ $$V → $$(cat $(VERSION_FILE))"

version-minor: ## Bump MINOR (1.9.0 → 1.10.0)
	@V=$$(cat $(VERSION_FILE)); \
	MAJOR=$${V%%.*}; \
	REST=$${V#*.}; \
	MINOR=$${REST%.*}; \
	echo "$$MAJOR.$$((MINOR + 1)).0" > $(VERSION_FILE); \
	echo "✅ $$V → $$(cat $(VERSION_FILE))"

version-major: ## Bump MAJOR (1.9.0 → 2.0.0)
	@V=$$(cat $(VERSION_FILE)); \
	MAJOR=$${V%%.*}; \
	echo "$$((MAJOR + 1)).0.0" > $(VERSION_FILE); \
	echo "✅ $$V → $$(cat $(VERSION_FILE))"
