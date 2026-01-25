.PHONY: all build test clean install install-dev lint fmt vet

# Binary name
BINARY_NAME=craizy
BINARY_PATH=./bin/$(BINARY_NAME)

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOTEST=$(GOCMD) test
GOCLEAN=$(GOCMD) clean
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod
GOFMT=$(GOCMD) fmt
GOVET=$(GOCMD) vet

# Main package path
MAIN_PATH=./cmd/craizy

# Build flags
LDFLAGS=-ldflags "-s -w"

all: test build

build:
	@echo "Building $(BINARY_NAME)..."
	@mkdir -p bin
	$(GOBUILD) $(LDFLAGS) -o $(BINARY_PATH) $(MAIN_PATH)
	@echo "Built: $(BINARY_PATH)"

test:
	@echo "Running tests..."
	$(GOTEST) -v -race -coverprofile=coverage.txt -covermode=atomic ./...

test-short:
	@echo "Running short tests..."
	$(GOTEST) -v -short ./...

coverage:
	@echo "Generating coverage report..."
	$(GOTEST) -coverprofile=coverage.txt -covermode=atomic ./...
	$(GOCMD) tool cover -html=coverage.txt -o coverage.html
	@echo "Coverage report: coverage.html"

lint:
	@echo "Running linters..."
	@which golangci-lint > /dev/null || (echo "golangci-lint not found. Install from https://golangci-lint.run/usage/install/" && exit 1)
	golangci-lint run ./...

fmt:
	@echo "Formatting code..."
	$(GOFMT) ./...

vet:
	@echo "Running go vet..."
	$(GOVET) ./...

clean:
	@echo "Cleaning..."
	$(GOCLEAN)
	rm -rf bin/
	rm -f coverage.txt coverage.html

deps:
	@echo "Downloading dependencies..."
	$(GOMOD) download
	$(GOMOD) tidy

install: build
	@echo "Installing $(BINARY_NAME)..."
	cp $(BINARY_PATH) $(GOPATH)/bin/$(BINARY_NAME)
	@echo "Installed to: $(GOPATH)/bin/$(BINARY_NAME)"

install-dev:
	@echo "Installing $(BINARY_NAME)-dev..."
	$(GOBUILD) -o $(shell go env GOPATH)/bin/$(BINARY_NAME)-dev $(MAIN_PATH)
	@echo "Installed $(BINARY_NAME)-dev to $(shell go env GOPATH)/bin/"

run: build
	$(BINARY_PATH)

help:
	@echo "Available targets:"
	@echo "  all          - Run tests and build"
	@echo "  build        - Build the binary"
	@echo "  test         - Run all tests with race detection"
	@echo "  test-short   - Run short tests"
	@echo "  coverage     - Generate coverage report"
	@echo "  lint         - Run linters"
	@echo "  fmt          - Format code"
	@echo "  vet          - Run go vet"
	@echo "  clean        - Clean build artifacts"
	@echo "  deps         - Download dependencies"
	@echo "  install      - Build and install binary"
	@echo "  install-dev  - Build and install as craizy-dev"
	@echo "  run          - Build and run"
