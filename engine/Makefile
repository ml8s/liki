# 语言原语层：单命令入口。全量测试/推送检查统一走仓库根 Makefile。
.PHONY: all build build-linux lint vet test clean run

VERSION=$(shell cat cmd/engine-mcp/VERSION 2>/dev/null || echo dev)
LDFLAGS=-ldflags="-X main.BuildTime=$(VERSION)"

all: build

build:
	go build $(LDFLAGS) -o engine-mcp ./cmd/engine-mcp/

build-linux:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build $(LDFLAGS) -o bin/engine-mcp ./cmd/engine-mcp/

lint:
	golangci-lint run ./...

vet:
	go vet ./...

test:
	go test ./internal/... -count=1

clean:
	rm -f engine-mcp bin/engine-mcp

run:
	go run ./cmd/engine-mcp/
