.PHONY: help quick-start up down restart logs build test test-coverage lint lint-fix swag migrate-create migrate-up migrate-down migrate-version build-binary run-binary clean

# Container name (from docker-compose.yml)
CONTAINER_NAME := go_api_app

# Check if container is running
CONTAINER_RUNNING := $(shell docker ps --format '{{.Names}}' 2>/dev/null | grep -E '^$(CONTAINER_NAME)$$')

# Determine execution command
ifdef CONTAINER_RUNNING
	EXEC_CMD = docker exec $(CONTAINER_RUNNING)
	ENV_MSG = 🐳 Running in Docker container
else
	EXEC_CMD = 
	ENV_MSG = 💻 Running on host (Docker not available)
endif

## help: Show this help message
help:
	@echo "Go REST API Boilerplate - Available Commands"
	@echo "=============================================="
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make quick-start    - Complete setup and start (Docker required)"
	@echo ""
	@echo "🐳 Docker Commands:"
	@echo "  make up             - Start containers"
	@echo "  make down           - Stop containers"
	@echo "  make restart        - Restart containers"
	@echo "  make logs           - View container logs"
	@echo "  make build          - Rebuild containers"
	@echo ""
	@echo "🧪 Development Commands:"
	@echo "  make test           - Run tests"
	@echo "  make test-coverage  - Run tests with coverage"
	@echo "  make lint           - Run linter"
	@echo "  make lint-fix       - Run linter and fix issues"
	@echo "  make swag           - Generate Swagger docs"
	@echo ""
	@echo "🗄️  Database Commands:"
	@echo "  make migrate-create NAME=<name>  - Create new migration"
	@echo "  make migrate-up                  - Run migrations"
	@echo "  make migrate-down                - Drop all migration tables"
	@echo "  make migrate-version             - Show migration status"
	@echo ""
	@echo "⚙️  Native Build (requires Go on host):"
	@echo "  make build-binary   - Build Go binary directly (no Docker)"
	@echo "  make run-binary     - Build and run binary directly (no Docker)"
	@echo ""
	@echo "🧹 Utility:"
	@echo "  make clean          - Clean build artifacts"
	@echo ""
	@echo "💡 Most commands auto-detect Docker/host environment"
	@echo "💡 Native build commands require Go installed on your machine"

## quick-start: Complete setup and start the project
quick-start:
	@chmod +x scripts/quick-start.sh
	@./scripts/quick-start.sh

## up: Start Docker containers
up:
	@echo "🐳 Starting Docker containers..."
	@docker compose up -d --build --wait
	@echo "✅ Containers started and healthy"
	@echo "📍 API: http://localhost:8080"

## down: Stop Docker containers
down:
	@echo "🛑 Stopping Docker containers..."
	@docker compose down
	@echo "✅ Containers stopped"

## restart: Restart Docker containers
restart:
	@echo "🔄 Restarting Docker containers..."
	@docker compose restart
	@echo "✅ Containers restarted"

## logs: View container logs
logs:
	@docker compose logs -f app

## build: Rebuild Docker containers
build:
	@echo "🔨 Building Docker containers..."
	@docker compose build
	@echo "✅ Build complete"

## test: Run tests
test:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go test ./... -v
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go test ./... -v; \
	else \
		echo "❌ Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

## test-coverage: Run tests with coverage
test-coverage:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go test ./... -v -coverprofile=coverage.out
	@$(EXEC_CMD) go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report: coverage.html"
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go test ./... -v -coverprofile=coverage.out; \
		go tool cover -html=coverage.out -o coverage.html; \
		echo "✅ Coverage report: coverage.html"; \
	else \
		echo "❌ Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

## lint: Run linter
lint:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@echo "🔍 Running golangci-lint..."
	@$(EXEC_CMD) golangci-lint run --timeout=5m && echo "✅ No linting issues found!" || exit 1
else
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		echo "🔍 Running golangci-lint..."; \
		golangci-lint run --timeout=5m && echo "✅ No linting issues found!" || exit 1; \
	else \
		echo "❌ Error: Docker container not running and golangci-lint not installed"; \
		echo "Please run: make up"; \
		echo "Or install: go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		exit 1; \
	fi
endif

## lint-fix: Run linter and fix issues
lint-fix:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@echo "🔧 Running golangci-lint with auto-fix..."
	@$(EXEC_CMD) golangci-lint run --fix --timeout=5m && echo "✅ Linting complete! Issues auto-fixed where possible." || exit 1
else
	@if command -v golangci-lint >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		echo "🔧 Running golangci-lint with auto-fix..."; \
		golangci-lint run --fix --timeout=5m && echo "✅ Linting complete! Issues auto-fixed where possible." || exit 1; \
	else \
		echo "❌ Error: Docker container not running and golangci-lint not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

## swag: Generate Swagger documentation
swag:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) swag init -g ./cmd/server/main.go -o ./api/docs
	@echo "✅ Swagger docs generated"
else
	@if command -v swag >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		swag init -g ./cmd/server/main.go -o ./api/docs; \
		echo "✅ Swagger docs generated"; \
	else \
		echo "❌ Error: Docker container not running and swag not installed"; \
		echo "Please run: make up"; \
		echo "Or install: go install github.com/swaggo/swag/cmd/swag@latest"; \
		exit 1; \
	fi
endif

## migrate-create: Create a new migration file
migrate-create:
ifndef NAME
	@echo "❌ Error: NAME is required"
	@echo "Usage: make migrate-create NAME=create_users_table"
	@exit 1
endif
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) migrate create -ext sql -dir migrations -seq $(NAME)
	@echo "✅ Migration created: migrations/*_$(NAME).sql"
else
	@if command -v migrate >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		migrate create -ext sql -dir migrations -seq $(NAME); \
		echo "✅ Migration created: migrations/*_$(NAME).sql"; \
	else \
		echo "❌ Error: Docker container not running and migrate not installed"; \
		echo "Please run: make up"; \
		echo "Or install: go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest"; \
		exit 1; \
	fi
endif

## migrate-up: Run database migrations
migrate-up:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go run cmd/migrate/main.go -action=up
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go run cmd/migrate/main.go -action=up; \
	else \
		echo "❌ Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

## migrate-down: Drop all migration tables
migrate-down:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go run cmd/migrate/main.go -action=down
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go run cmd/migrate/main.go -action=down; \
	else \
		echo "❌ Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

## migrate-version: Show current migration version
migrate-version:
ifdef CONTAINER_RUNNING
	@echo "$(ENV_MSG)"
	@$(EXEC_CMD) go run cmd/migrate/main.go -action=status
else
	@if command -v go >/dev/null 2>&1; then \
		echo "$(ENV_MSG)"; \
		go run cmd/migrate/main.go -action=status; \
	else \
		echo "❌ Error: Docker container not running and Go not installed"; \
		echo "Please run: make up"; \
		exit 1; \
	fi
endif

## build-binary: Build Go binary directly on host (requires Go)
build-binary:
	@if ! command -v go >/dev/null 2>&1; then \
		echo "❌ Error: Go is not installed on your machine"; \
		echo ""; \
		echo "Please install Go first:"; \
		echo "  https://golang.org/doc/install"; \
		echo ""; \
		echo "Or use Docker instead:"; \
		echo "  make up"; \
		exit 1; \
	fi
	@echo "🔨 Building Go binary..."
	@mkdir -p bin
	@go build -o bin/server ./cmd/server
	@echo "✅ Binary built successfully: bin/server"
	@echo ""
	@echo "To run the binary:"
	@echo "  make run-binary"
	@echo "  OR"
	@echo "  ./bin/server"

## run-binary: Build and run Go binary directly on host (requires Go)
run-binary: build-binary
	@echo ""
	@echo "🚀 Starting server..."
	@echo ""
	@echo "⚠️  Note: Ensure PostgreSQL is running on localhost:5432"
	@echo "⚠️  Note: Set environment variables or use .env file"
	@echo ""
	@./bin/server

## clean: Clean build artifacts
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f coverage.out coverage.html
	@rm -f bin/*
	@docker compose down -v 2>/dev/null || true
	@echo "✅ Clean complete"
