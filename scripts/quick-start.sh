#!/bin/bash

echo "🚀 Go REST API Boilerplate - Quick Start"
echo "========================================"
echo ""

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

# Step 1: Install development tools
echo -e "${BLUE}Step 1: Installing development tools...${NC}"
echo ""

./scripts/install-tools.sh
INSTALL_STATUS=$?

if [ $INSTALL_STATUS -ne 0 ]; then
    echo ""
    echo -e "${RED}❌ Failed to install development tools.${NC}"
    echo ""
    echo "Please check the errors above and try again."
    echo "You can also install tools manually with: make install-tools"
    exit 1
fi

echo ""

# Step 2: Verify setup
echo -e "${BLUE}Step 2: Verifying setup...${NC}"
echo ""

./scripts/verify-setup.sh
VERIFY_STATUS=$?

if [ $VERIFY_STATUS -ne 0 ]; then
    echo ""
    echo -e "${YELLOW}⚠️  Initial verification failed. Attempting to fix...${NC}"
    echo ""
    
    # Try to install tools again
    echo "Retrying tool installation..."
    ./scripts/install-tools.sh
    
    # Verify again
    echo ""
    echo "Verifying setup again..."
    ./scripts/verify-setup.sh
    VERIFY_STATUS=$?
    
    if [ $VERIFY_STATUS -ne 0 ]; then
        echo ""
        echo -e "${RED}❌ Setup verification failed after retry.${NC}"
        echo ""
        echo "Please review the errors above and fix them manually."
        echo ""
        echo "Common issues:"
        echo "  • Go not installed: https://go.dev/dl/"
        echo "  • Docker not installed: https://docs.docker.com/get-docker/"
        echo "  • GOBIN not in PATH: export PATH=\"\$PATH:\$(go env GOPATH)/bin\""
        echo ""
        exit 1
    fi
fi

echo ""

# Step 3: Check environment file
echo -e "${BLUE}Step 3: Setting up environment...${NC}"
echo ""

if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo -e "${GREEN}✅ .env file created${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  Please review .env and update if needed (especially JWT_SECRET for production)${NC}"
else
    echo -e "${GREEN}✅ .env file already exists${NC}"
fi

echo ""

# Step 4: Generate Swagger documentation
echo -e "${BLUE}Step 4: Generating Swagger documentation...${NC}"
echo ""

# Get GOBIN
GOBIN=$(go env GOPATH)/bin

# Run swag init
if command -v swag &> /dev/null; then
    swag init -g ./cmd/server/main.go -o ./api/docs
elif [ -f "$GOBIN/swag" ]; then
    "$GOBIN/swag" init -g ./cmd/server/main.go -o ./api/docs
else
    echo -e "${RED}❌ swag not found (should have been installed in Step 1)${NC}"
    exit 1
fi

if [ $? -eq 0 ]; then
    echo -e "${GREEN}✅ Swagger docs generated${NC}"
else
    echo -e "${RED}❌ Failed to generate Swagger docs${NC}"
    exit 1
fi

echo ""

# Step 5: Start Docker containers
echo -e "${BLUE}Step 5: Starting Docker containers...${NC}"
echo ""

echo "Building and starting containers (this may take a few minutes on first run)..."
docker-compose up --build -d

if [ $? -eq 0 ]; then
    echo ""
    echo -e "${GREEN}✅ Containers started successfully!${NC}"
else
    echo ""
    echo -e "${RED}❌ Failed to start containers${NC}"
    echo ""
    echo "Please check Docker is running and try again."
    exit 1
fi

echo ""
echo "Waiting for services to be ready..."
sleep 3

# Check if containers are running
if docker ps | grep -q "go_api_app"; then
    echo -e "${GREEN}✅ Application container is running${NC}"
else
    echo -e "${RED}❌ Application container failed to start${NC}"
    echo ""
    echo "Check logs with: docker-compose logs app"
    exit 1
fi

if docker ps | grep -q "go_api_db"; then
    echo -e "${GREEN}✅ Database container is running${NC}"
else
    echo -e "${RED}❌ Database container failed to start${NC}"
    echo ""
    echo "Check logs with: docker-compose logs db"
    exit 1
fi

echo ""
echo "========================================"
echo -e "${GREEN}🎉 Setup Complete!${NC}"
echo "========================================"
echo ""
echo "Your API is now running!"
echo ""
echo "📍 Access Points:"
echo "  • API Base URL:  http://localhost:8080/api/v1"
echo "  • Swagger UI:    http://localhost:8080/swagger/index.html"
echo "  • Health Check:  http://localhost:8080/health"
echo ""
echo "🧪 Try it out:"
echo "  curl http://localhost:8080/health"
echo ""
echo "📚 Next Steps:"
echo "  • Check Swagger UI for interactive API docs"
echo "  • Import Postman collection from: api/postman_collection.json"
echo "  • View logs: docker-compose logs -f app"
echo "  • Stop services: docker-compose down"
echo ""
echo "📖 Documentation:"
echo "  • Quick Reference: docs/QUICK_REFERENCE.md"
echo "  • Setup Guide: docs/SETUP.md"
echo "  • Docker Guide: docs/DOCKER.md"
echo ""
echo "💡 Development Tips:"
echo "  • Code changes auto-reload in ~2 seconds (hot-reload enabled)"
echo "  • Run tests: make test"
echo "  • Check code: make lint"
echo "  • Create migration: make migrate-create NAME=your_migration"
echo ""
