#!/bin/bash
set -e

echo "Building dnsbalancer..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "Error: Go is not installed. Please install Go 1.21 or later."
    exit 1
fi

# Get version info
VERSION="${VERSION:-1.0.0}"
GIT_COMMIT=$(git rev-parse HEAD 2>/dev/null || echo "unknown")
BUILD_DATE=$(date -u +%Y-%m-%dT%H:%M:%SZ)

# Build flags
LDFLAGS="-X github.com/yourusername/dnsbalancer/cmd.Version=${VERSION} \
         -X github.com/yourusername/dnsbalancer/cmd.GitCommit=${GIT_COMMIT} \
         -X github.com/yourusername/dnsbalancer/cmd.BuildDate=${BUILD_DATE}"

# Tidy dependencies
echo "Downloading dependencies..."
go mod tidy
go mod download

# Build
echo "Building binary..."
go build -ldflags "${LDFLAGS}" -o dnsbalancer .

echo ""
echo "✅ Build complete: ./dnsbalancer"
echo ""
echo "Next steps:"
echo "  1. Generate config: ./dnsbalancer genconfig"
echo "  2. Edit config.yaml with your DNS backends"
echo "  3. Validate: ./dnsbalancer validate"
echo "  4. Test connectivity: ./dnsbalancer healthcheck"
echo "  5. Run: sudo ./dnsbalancer serve --debug"
echo ""
