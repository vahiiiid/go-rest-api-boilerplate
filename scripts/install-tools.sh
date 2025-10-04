#!/bin/bash

echo "🔧 Installing Development Tools"
echo "================================"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

# Success/Failure counters
SUCCESS=0
FAILURE=0

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo -e "${RED}❌ Go is not installed!${NC}"
    echo ""
    echo "Please install Go first:"
    echo "  - Visit: https://go.dev/dl/"
    echo "  - Or use: brew install go (macOS)"
    echo ""
    exit 1
fi

echo -e "${GREEN}✅ Go is installed: $(go version | awk '{print $3}')${NC}"
echo ""

# Get GOPATH and GOBIN
GOPATH=$(go env GOPATH)
GOBIN="${GOPATH}/bin"

echo "📦 Installing Go tools to: $GOBIN"
echo ""

# Function to install a Go tool
install_tool() {
    local tool_name=$1
    local tool_path=$2
    local install_cmd=$3
    
    echo -e "${BLUE}Installing $tool_name...${NC}"
    
    if eval "$install_cmd" > /dev/null 2>&1; then
        echo -e "${GREEN}✅ $tool_name installed successfully${NC}"
        ((SUCCESS++))
    else
        echo -e "${RED}❌ Failed to install $tool_name${NC}"
        ((FAILURE++))
    fi
}

# Install Swagger CLI (swag)
install_tool "Swagger CLI (swag)" \
    "github.com/swaggo/swag/cmd/swag" \
    "go install github.com/swaggo/swag/cmd/swag@latest"

# Install golangci-lint
install_tool "golangci-lint" \
    "github.com/golangci/golangci-lint/cmd/golangci-lint" \
    "go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"

# Install golang-migrate
install_tool "golang-migrate" \
    "github.com/golang-migrate/migrate/v4/cmd/migrate" \
    "go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"

# Install Air (hot-reload for development)
install_tool "Air (hot-reload)" \
    "github.com/air-verse/air" \
    "go install github.com/air-verse/air@v1.52.3"

echo ""
echo "================================"
echo "Summary:"
echo -e "${GREEN}✅ Installed: $SUCCESS${NC}"
if [ $FAILURE -gt 0 ]; then
    echo -e "${RED}❌ Failed: $FAILURE${NC}"
fi
echo ""

# Check if GOBIN is in PATH
if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
    echo -e "${YELLOW}⚠️  WARNING: $GOBIN is not in your PATH${NC}"
    echo ""
    echo "To use these tools, add this to your shell profile (~/.bashrc, ~/.zshrc, etc.):"
    echo ""
    echo "  export PATH=\"\$PATH:$GOBIN\""
    echo ""
    echo "Then run: source ~/.bashrc (or restart your terminal)"
    echo ""
fi

if [ $FAILURE -eq 0 ]; then
    echo -e "${GREEN}🎉 All tools installed successfully!${NC}"
    echo ""
    echo "Installed tools:"
    echo "  • swag         - Swagger documentation generator"
    echo "  • golangci-lint - Go linter"
    echo "  • migrate      - Database migration tool"
    echo "  • air          - Hot-reload for development"
    echo ""
    exit 0
else
    echo -e "${RED}⚠️  Some tools failed to install. Please check the errors above.${NC}"
    exit 1
fi

