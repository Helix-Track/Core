# Build stage
FROM golang:1.22-alpine AS builder

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
FROM golang:1.22-alpine

WORKDIR /app

# Install sqlite
RUN apk add --no-cache sqlite-libs

# Copy binary from builder
COPY --from=builder /app/htCore .

# Copy configurations
COPY Configurations/ ./Configurations/

# Create Database directory
RUN mkdir -p Database

# Expose port
EXPOSE 8080

# Run the application
CMD ["./htCore", "-config", "Configurations/default.json"]
