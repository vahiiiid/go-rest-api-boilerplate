#!/bin/bash
set -e

echo "Initializing Swagger documentation..."

# Determine the GOPATH
GOPATH="${GOPATH:-$HOME/go}"
GOBIN="${GOBIN:-$GOPATH/bin}"

# Check if swag is installed
if ! command -v swag &> /dev/null && [ ! -f "$GOBIN/swag" ]; then
    echo "Installing swag CLI..."
    go install github.com/swaggo/swag/cmd/swag@latest
    echo "Swag installed to: $GOBIN/swag"
fi

# Generate swagger docs (try both PATH and GOBIN)
echo "Generating Swagger docs..."
if command -v swag &> /dev/null; then
    swag init -g ./cmd/server/main.go -o ./api/docs
elif [ -f "$GOBIN/swag" ]; then
    "$GOBIN/swag" init -g ./cmd/server/main.go -o ./api/docs
else
    echo "❌ Error: swag not found. Please run: go install github.com/swaggo/swag/cmd/swag@latest"
    echo "Then add $GOBIN to your PATH"
    exit 1
fi

echo "✅ Swagger documentation generated successfully!"
echo "Access it at: http://localhost:8080/swagger/index.html"

