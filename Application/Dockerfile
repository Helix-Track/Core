# Build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies for CGO and SQLite
RUN apk add --no-cache gcc musl-dev sqlite-dev

WORKDIR /app

# Copy go mod files
COPY go.mod ./

# Copy source to generate go.sum
COPY . .

# Generate go.sum and download dependencies
RUN go mod tidy && go mod download

# Build the application
RUN CGO_ENABLED=1 go build -o /app/htCore main.go

# Runtime stage
FROM golang:1.24-alpine

WORKDIR /app

# Install sqlite + curl (curl is required by the compose healthcheck
# `curl -f http://localhost:8080/health`; without it the container reports
# (unhealthy) despite the app being healthy)
RUN apk add --no-cache sqlite-libs curl

# Copy binary from builder
COPY --from=builder /app/htCore .

# Copy configurations
COPY Configurations/ ./Configurations/

# Create Database directory
RUN mkdir -p Database

# Expose port
EXPOSE 8080

# Copy Database DDL directory for first-start auto-init
COPY Database/ ./Database/

# Run the application
# --space-root overrides the default config for per-space data isolation
CMD ["./htCore", "-config", "Configurations/default.json"]
