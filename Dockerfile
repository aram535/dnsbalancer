# Multi-stage build
FROM golang:1.21-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git make

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY . .

# Build the application
ARG VERSION=dev
ARG GIT_COMMIT=unknown
ARG BUILD_DATE=unknown

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags "-X github.com/aram535/dnsbalancer/cmd.Version=${VERSION} \
              -X github.com/aram535/dnsbalancer/cmd.GitCommit=${GIT_COMMIT} \
              -X github.com/aram535/dnsbalancer/cmd.BuildDate=${BUILD_DATE}" \
    -o dnsbalancer .

# Final stage
FROM alpine:latest

# Install ca-certificates for HTTPS health checks (if needed in future)
RUN apk --no-cache add ca-certificates

# Create non-root user (though we'll need root for port 53)
RUN addgroup -g 1000 dnsbalancer && \
    adduser -D -u 1000 -G dnsbalancer dnsbalancer

# Create directories
RUN mkdir -p /etc/dnsbalancer /var/log/dnsbalancer && \
    chown -R dnsbalancer:dnsbalancer /var/log/dnsbalancer

# Copy binary from builder
COPY --from=builder /build/dnsbalancer /usr/local/bin/dnsbalancer

# Copy example config
COPY config.example.yaml /etc/dnsbalancer/config.yaml

# Expose DNS port
EXPOSE 53/udp

# Use root for port 53 binding
USER root

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD ["/usr/local/bin/dnsbalancer", "healthcheck"]

# Run the application
ENTRYPOINT ["/usr/local/bin/dnsbalancer"]
CMD ["serve", "--config", "/etc/dnsbalancer/config.yaml"]
