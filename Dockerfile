# Development stage with hot-reload
FROM golang:1.23-alpine AS development

# Install development tools
RUN apk add --no-cache git curl

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Install air for hot-reload (pinned to compatible version with Go 1.23)
RUN go install github.com/air-verse/air@v1.52.3

# Copy source code (in docker-compose, we'll mount a volume over this)
COPY . .

# Expose port
EXPOSE 8080

# Run with air for hot-reload
CMD ["air", "-c", ".air.toml"]

# Production builder stage
FROM golang:1.23-alpine AS builder

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
