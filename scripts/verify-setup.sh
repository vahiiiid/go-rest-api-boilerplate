#!/bin/bash

echo "🔍 Go REST API Boilerplate - Setup Verification"
echo "================================================"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Success/Failure counters
SUCCESS=0
FAILURE=0

# Check function
check() {
    if [ $? -eq 0 ]; then
        echo -e "${GREEN}✅ $1${NC}"
        ((SUCCESS++))
    else
        echo -e "${RED}❌ $1${NC}"
        ((FAILURE++))
    fi
}

echo "Checking prerequisites..."
echo ""

# Check Go installation
go version > /dev/null 2>&1
check "Go is installed: $(go version 2>/dev/null | awk '{print $3}')"

# Check Docker
docker --version > /dev/null 2>&1
check "Docker is installed: $(docker --version 2>/dev/null | cut -d' ' -f3 | tr -d ',')"

# Check Docker Compose
docker-compose --version > /dev/null 2>&1
if [ $? -eq 0 ]; then
    check "Docker Compose is installed: $(docker-compose --version 2>/dev/null | cut -d' ' -f4 | tr -d ',')"
else
    # Try docker compose (v2 syntax)
    docker compose version > /dev/null 2>&1
    check "Docker Compose is installed: $(docker compose version 2>/dev/null | cut -d' ' -f4 | tr -d ',')"
fi

echo ""
echo "Checking Go development tools..."
echo ""

# Get GOPATH and GOBIN
GOBIN=$(go env GOPATH)/bin

# Check for swag
if command -v swag &> /dev/null || [ -f "$GOBIN/swag" ]; then
    if command -v swag &> /dev/null; then
        echo -e "${GREEN}✅ swag is installed: $(swag --version 2>/dev/null | head -n1)${NC}"
    else
        echo -e "${GREEN}✅ swag is installed at: $GOBIN/swag${NC}"
    fi
    ((SUCCESS++))
else
    echo -e "${RED}❌ swag is not installed${NC}"
    ((FAILURE++))
fi

# Check for golangci-lint
if command -v golangci-lint &> /dev/null || [ -f "$GOBIN/golangci-lint" ]; then
    if command -v golangci-lint &> /dev/null; then
        echo -e "${GREEN}✅ golangci-lint is installed: $(golangci-lint --version 2>/dev/null | head -n1 | awk '{print $4}')${NC}"
    else
        echo -e "${GREEN}✅ golangci-lint is installed at: $GOBIN/golangci-lint${NC}"
    fi
    ((SUCCESS++))
else
    echo -e "${RED}❌ golangci-lint is not installed${NC}"
    ((FAILURE++))
fi

# Check for migrate
if command -v migrate &> /dev/null || [ -f "$GOBIN/migrate" ]; then
    if command -v migrate &> /dev/null; then
        echo -e "${GREEN}✅ migrate is installed: $(migrate -version 2>/dev/null)${NC}"
    else
        echo -e "${GREEN}✅ migrate is installed at: $GOBIN/migrate${NC}"
    fi
    ((SUCCESS++))
else
    echo -e "${RED}❌ migrate is not installed${NC}"
    ((FAILURE++))
fi

# Check for air
if command -v air &> /dev/null || [ -f "$GOBIN/air" ]; then
    if command -v air &> /dev/null; then
        echo -e "${GREEN}✅ air is installed: $(air -v 2>/dev/null | head -n1)${NC}"
    else
        echo -e "${GREEN}✅ air is installed at: $GOBIN/air${NC}"
    fi
    ((SUCCESS++))
else
    echo -e "${RED}❌ air is not installed${NC}"
    ((FAILURE++))
fi

echo ""
echo "Checking project files..."
echo ""

# Check required files
FILES=(
    "go.mod"
    "go.sum"
    ".env.example"
    "Dockerfile"
    "docker-compose.yml"
    "Makefile"
    "README.md"
    "cmd/server/main.go"
    "internal/user/handler.go"
    "internal/auth/service.go"
    "tests/handler_test.go"
)

for file in "${FILES[@]}"; do
    if [ -f "$file" ]; then
        echo -e "${GREEN}✅ $file exists${NC}"
        ((SUCCESS++))
    else
        echo -e "${RED}❌ $file missing${NC}"
        ((FAILURE++))
    fi
done

echo ""
echo "Checking Go code..."
echo ""

# Check if go.mod is valid
go list ./... > /dev/null 2>&1
check "Go modules are valid"

# Run go vet
go vet ./... > /dev/null 2>&1
check "go vet passes"

# Check if code compiles
go build ./cmd/server > /dev/null 2>&1
check "Code compiles successfully"

# Clean up binary
rm -f server 2>/dev/null

echo ""
echo "Running tests..."
echo ""

# Run tests
go test ./... > /dev/null 2>&1
check "All tests pass"

echo ""
echo "================================================"
echo "Summary:"
echo -e "${GREEN}✅ Passed: $SUCCESS${NC}"
if [ $FAILURE -gt 0 ]; then
    echo -e "${RED}❌ Failed: $FAILURE${NC}"
fi
echo ""

if [ $FAILURE -eq 0 ]; then
    echo -e "${GREEN}🎉 Everything looks good! You're ready to go!${NC}"
    echo ""
    echo "Next steps:"
    echo "  1. Run: docker-compose up --build"
    echo "  2. Visit: http://localhost:8080/health"
    echo "  3. Check: http://localhost:8080/swagger/index.html"
    exit 0
else
    echo -e "${YELLOW}⚠️  Some checks failed. Please review the errors above.${NC}"
    echo ""
    if [[ ":$PATH:" != *":$GOBIN:"* ]]; then
        echo -e "${YELLOW}💡 TIP: Add Go bin to your PATH:${NC}"
        echo "  export PATH=\"\$PATH:$GOBIN\""
        echo ""
    fi
    exit 1
fi
