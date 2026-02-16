.PHONY: help build test lint run clean install

BINARY_NAME=cheapass
MAIN_PATH=cmd/cheapass/main.go

help:
	@echo "cheapass - AWS spend tracker"
	@echo ""
	@echo "Available targets:"
	@echo "  make build    - Build the binary"
	@echo "  make install  - Install to \$$GOPATH/bin"
	@echo "  make run      - Run locally"
	@echo "  make test     - Run tests"
	@echo "  make lint     - Run linter"
	@echo "  make clean    - Remove build artifacts"

build:
	@echo "Building $(BINARY_NAME)..."
	go build -o $(BINARY_NAME) $(MAIN_PATH)
	@echo "✓ Built: ./$(BINARY_NAME)"

install:
	@echo "Installing $(BINARY_NAME)..."
	go install $(MAIN_PATH)
	@echo "✓ Installed to \$$GOPATH/bin/$(BINARY_NAME)"

run: build
	./$(BINARY_NAME) cost --help

test:
	go test -v ./...

lint:
	go vet ./...
	go fmt ./...

clean:
	rm -f $(BINARY_NAME)
	go clean

deps:
	go mod download
	go mod tidy
