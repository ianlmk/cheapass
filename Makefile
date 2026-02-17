.PHONY: help build build-all install install-all test lint run clean deps

CHEAPASS_MAIN=cmd/cheapass/main.go
STATE_CHECK_MAIN=cmd/state_check/main.go

help:
	@echo "cheapass - AWS spend & resource audit suite"
	@echo ""
	@echo "Available targets:"
	@echo "  make build        - Build cheapass binary only"
	@echo "  make build-all    - Build all binaries (cheapass, state-check)"
	@echo "  make install      - Install all tools to \$$GOPATH/bin"
	@echo "  make install-old  - Install cheapass only (legacy)"
	@echo "  make run          - Run cheapass cost command"
	@echo "  make audit        - Run state_check audit"
	@echo "  make test         - Run tests"
	@echo "  make lint         - Run linter"
	@echo "  make clean        - Remove build artifacts"
	@echo "  make deps         - Update dependencies"

build:
	@echo "Building cheapass..."
	go build -o cheapass $(CHEAPASS_MAIN)
	@echo "✓ Built: ./cheapass"

build-all:
	@echo "Building cheapass..."
	go build -o cheapass $(CHEAPASS_MAIN)
	@echo "✓ Built: ./cheapass"
	@echo "Building state-check..."
	go build -o state-check $(STATE_CHECK_MAIN)
	@echo "✓ Built: ./state-check"

install:
	@echo "Installing all tools to \$$GOPATH/bin..."
	go install ./cmd/cheapass
	go install ./cmd/state_check
	@echo "✓ Installed cheapass and state-check"

install-old:
	@echo "Installing cheapass only (legacy)..."
	go install $(CHEAPASS_MAIN)
	@echo "✓ Installed to \$$GOPATH/bin/cheapass"

run: build
	./cheapass cost --help

audit:
	@echo "Running state_check audit..."
	go run $(STATE_CHECK_MAIN) --region us-east-2

test:
	go test -v ./...

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -f cheapass state-check
	go clean

deps:
	go mod download
	go mod tidy
