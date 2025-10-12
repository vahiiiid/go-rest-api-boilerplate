# Development stage with hot-reload
# Using Debian-based image for better CGO/SQLite compatibility
FROM golang:1.24-bookworm AS development

# Install system dependencies
RUN apt-get update && apt-get install -y \
    git \
    curl \
    bash \
    gcc \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Enable CGO for development (required for SQLite in tests)
ENV CGO_ENABLED=1

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install all development tools
RUN go install github.com/air-verse/air@v1.52.3 && \
    go install github.com/swaggo/swag/cmd/swag@latest && \
    go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest && \
    go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Copy source code (in docker compose, we'll mount a volume over this)
COPY . .

# Expose port
EXPOSE 8080

# Set executable permission for app_entrypoint.sh
RUN chmod +x scripts/app_entrypoint.sh

# Use entrypoint
ENTRYPOINT ["./scripts/app_entrypoint.sh"]

# Production build stage
FROM golang:1.24-alpine AS builder

# Install build dependencies
RUN apk add --no-cache git

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags="-w -s" -o main ./cmd/server

# Production final stage
FROM alpine:latest AS production

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/main .

# Expose port
EXPOSE 8080

# Run the application
CMD ["./main"]
