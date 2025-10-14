#!/bin/bash

set -e

# Color codes
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;36m'
NC='\033[0m' # No Color

echo "🚀 Go REST API Boilerplate - Quick Start"
echo "========================================"
echo ""

# Check Docker
if ! command -v docker &> /dev/null; then
    echo -e "${RED}❌ Docker is not installed${NC}"
    echo ""
    echo "Please install Docker first:"
    echo "  https://docs.docker.com/get-docker/"
    echo ""
    echo "Or see manual setup instructions:"
    echo "  https://vahiiiid.github.io/go-rest-api-docs/SETUP/"
    exit 1
fi

# Check Docker Compose
if ! docker compose version &> /dev/null 2>&1; then
    echo -e "${RED}❌ Docker Compose is not installed${NC}"
    echo ""
    echo "Please install Docker Compose:"
    echo "  https://docs.docker.com/compose/install/"
    exit 1
fi

echo -e "${GREEN}✅ Docker and Docker Compose are installed${NC}"
echo ""

# Create .env file if it doesn't exist
if [ ! -f .env ]; then
    echo "Creating .env file from .env.example..."
    cp .env.example .env
    echo -e "${GREEN}✅ .env file created${NC}"
    echo ""
    echo -e "${YELLOW}⚠️  Please review .env and update JWT_SECRET for production${NC}"
else
    echo -e "${GREEN}✅ .env file exists${NC}"
fi

echo ""
echo "Reading .env file..."
echo ""
# Load environment variables from .env file
if [ -f .env ]; then
    # parsing: exclude comments, empty lines, and invalid variable names
    export $(cat .env | grep -E '^[A-Za-z_][A-Za-z0-9_]*=' | grep -v '#' | xargs)
fi

echo -e "${GREEN}✅ .env file read${NC}"

# Fallback for env variable(s)
PORT=${PORT:-8080}

echo ""
echo "🐳 Starting Docker containers..."
echo ""

# Stop existing containers if running
if docker compose ps | grep -q "Up"; then
    echo "Stopping existing containers..."
    docker compose down
fi

# Start containers
if docker compose up -d --build --wait; then
    echo ""
    echo -e "${GREEN}✅ Containers started successfully${NC}"
else
    echo ""
    echo -e "${RED}❌ Failed to start containers${NC}"
    echo ""
    echo "Check logs with: docker compose logs"
    exit 1
fi

echo ""
echo "================================================"
echo -e "${GREEN}🎉 Success! Your API is ready!${NC}"
echo "================================================"
echo ""
echo "📍 Your API is running at:"
echo "   • API Base:    http://localhost:${PORT}/api/v1"
echo "   • Swagger UI:  http://localhost:${PORT}/swagger/index.html"
echo "   • Health:      http://localhost:${PORT}/health"
echo ""
echo "🐳 Docker Commands:"
echo "   • View logs:   docker compose logs -f app"
echo "   • Stop:        docker compose down"
echo "   • Restart:     docker compose restart"
echo ""
echo "🛠️  Development Commands:"
echo "   • Run tests:   make test"
echo "   • Run linter:  make lint"
echo "   • Update docs: make swag"
echo ""
echo "📚 Documentation:"
echo "   https://vahiiiid.github.io/go-rest-api-docs/"
echo ""
