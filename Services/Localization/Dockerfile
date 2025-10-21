# HelixTrack Localization Service Dockerfile
# Multi-stage build for optimized image size

# Stage 1: Build
FROM golang:1.22-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /build

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download
RUN go mod verify

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.serviceVersion=1.0.0 -X main.buildDate=$(date -u +%Y-%m-%d)" \
    -o htLocalization \
    cmd/main.go

# Stage 2: Runtime
FROM alpine:latest

# Install runtime dependencies
RUN apk add --no-cache \
    ca-certificates \
    tzdata \
    postgresql-client \
    jq \
    curl \
    bash

# Create non-root user
RUN addgroup -g 1000 helixtrack && \
    adduser -D -u 1000 -G helixtrack helixtrack

# Set working directory
WORKDIR /app

# Copy binary from builder
COPY --from=builder /build/htLocalization /app/htLocalization

# Copy configuration files
COPY --chown=helixtrack:helixtrack configs/ /app/configs/

# Copy seed data
COPY --chown=helixtrack:helixtrack seed-data/ /app/seed-data/

# Copy scripts
COPY --chown=helixtrack:helixtrack scripts/ /app/scripts/
RUN chmod +x /app/scripts/*.sh

# Create directories for data and logs
RUN mkdir -p /app/backups /app/logs /app/certs && \
    chown -R helixtrack:helixtrack /app

# Switch to non-root user
USER helixtrack

# Expose ports
# 8085: HTTP/3 QUIC (default, can be rotated 8085-8095)
EXPOSE 8085-8095

# Health check
HEALTHCHECK --interval=30s --timeout=10s --start-period=10s --retries=3 \
    CMD curl --insecure --fail https://localhost:8085/health || exit 1

# Environment variables
ENV GIN_MODE=release \
    SEED_DATA_PATH=seed-data \
    LOG_LEVEL=info

# Volume mounts for persistence
VOLUME ["/app/backups", "/app/logs"]

# Entry point
ENTRYPOINT ["/app/htLocalization"]

# Default command
CMD ["--config=configs/default.json"]
