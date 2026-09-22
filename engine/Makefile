# 语言原语层：单命令入口。编排（全量测试/推送检查）统一走仓库根 Makefile
# （根 test-engine → scripts/ci-engine.sh）——单一编排入口，不在此重复。
.PHONY: all build build-linux lint vet test clean run

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

test:
	go test ./internal/... -count=1

clean:
	rm -f $(BINARY) bin/liki-engine

run:
	go run ./cmd/liki/
