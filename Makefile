.PHONY: build clean test install run dev fmt vet lint help

# Variables
BINARY_NAME=dnsbalancer
VERSION?=1.0.0
GIT_COMMIT=$(shell git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(shell date -u +%Y-%m-%dT%H:%M:%SZ)
LDFLAGS=-ldflags "-X github.com/yourusername/dnsbalancer/cmd.Version=$(VERSION) \
                   -X github.com/yourusername/dnsbalancer/cmd.GitCommit=$(GIT_COMMIT) \
                   -X github.com/yourusername/dnsbalancer/cmd.BuildDate=$(BUILD_DATE)"

## help: Display this help message
help:
	@echo "Available targets:"
	@echo "  build      - Build the binary"
	@echo "  clean      - Remove build artifacts"
	@echo "  test       - Run tests"
	@echo "  install    - Install binary to /usr/local/bin"
	@echo "  run        - Run the application"
	@echo "  dev        - Run in debug mode"
	@echo "  fmt        - Format code"
	@echo "  vet        - Run go vet"
	@echo "  lint       - Run golangci-lint (if installed)"
	@echo "  deps       - Download dependencies"
	@echo "  genconfig  - Generate example config"

## build: Build the binary with version info
build:
	@echo "Building $(BINARY_NAME) v$(VERSION)..."
	go build $(LDFLAGS) -o $(BINARY_NAME) .
	@echo "Build complete: ./$(BINARY_NAME)"

## clean: Remove build artifacts
clean:
	@echo "Cleaning..."
	rm -f $(BINARY_NAME)
	rm -f config.yaml
	go clean

## test: Run tests
test:
	go test -v -race -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html

## install: Install to /usr/local/bin
install: build
	@echo "Installing to /usr/local/bin..."
	sudo install -m 755 $(BINARY_NAME) /usr/local/bin/
	sudo mkdir -p /etc/dnsbalancer
	@echo "Installed successfully!"
	@echo "Generate config: dnsbalancer genconfig -o /etc/dnsbalancer/config.yaml"

## run: Run the application
run: build
	sudo ./$(BINARY_NAME) serve

## dev: Run in debug mode
dev: build
	sudo ./$(BINARY_NAME) serve --debug

## fmt: Format code
fmt:
	go fmt ./...

## vet: Run go vet
vet:
	go vet ./...

## lint: Run golangci-lint
lint:
	@which golangci-lint > /dev/null || (echo "golangci-lint not installed"; exit 1)
	golangci-lint run

## deps: Download dependencies
deps:
	go mod download
	go mod tidy

## genconfig: Generate example config file
genconfig: build
	./$(BINARY_NAME) genconfig

## validate: Validate config file
validate: build
	./$(BINARY_NAME) validate

## healthcheck: Test backend connectivity
healthcheck: build
	./$(BINARY_NAME) healthcheck

## docker-build: Build Docker image
docker-build:
	docker build -t $(BINARY_NAME):$(VERSION) .

.DEFAULT_GOAL := help
