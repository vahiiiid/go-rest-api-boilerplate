.PHONY: run build docker-build docker-up docker-down test swag clean help

# Variables
BINARY_NAME=server
DOCKER_IMAGE=go-rest-api-boilerplate

help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

run: ## Run the application locally binary
	go run ./cmd/server

build: ## Build the application binary
	go build -o bin/$(BINARY_NAME) ./cmd/server

quick-start: ## Quick start with Docker (runs the script)
	@./scripts/quick-start.sh

docker-build: ## Build Docker image
	docker build -t $(DOCKER_IMAGE) .

docker-up: ## Start Docker Compose services (development with hot-reload)
	docker-compose up --build

docker-up-prod: ## Start Docker Compose services (production build)
	docker-compose -f docker-compose.prod.yml up --build

docker-down: ## Stop Docker Compose services
	docker-compose down

docker-down-prod: ## Stop production Docker Compose services
	docker-compose -f docker-compose.prod.yml down

test: ## Run tests
	go test ./... -v -cover

test-coverage: ## Run tests with coverage report
	go test ./... -coverprofile=coverage.out
	go tool cover -html=coverage.out

swag: ## Generate Swagger documentation
	@./scripts/init-swagger.sh

lint: ## Run linter
	@GOBIN=$$(go env GOPATH)/bin; \
	if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run; \
	elif [ -f "$$GOBIN/golangci-lint" ]; then \
		$$GOBIN/golangci-lint run; \
	else \
		echo "❌ golangci-lint not installed."; \
		echo ""; \
		echo "Install with:"; \
		echo "  make install-tools"; \
		echo ""; \
		echo "Or manually:"; \
		echo "  go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest"; \
		echo ""; \
		echo "Or using brew (macOS):"; \
		echo "  brew install golangci-lint"; \
		echo ""; \
		echo "Note: After installation, add Go bin to PATH:"; \
		echo "  export PATH=\"\$$PATH:\$$(go env GOPATH)/bin\""; \
		echo ""; \
		exit 1; \
	fi

lint-fix: ## Run linter and auto-fix issues
	@GOBIN=$$(go env GOPATH)/bin; \
	if command -v golangci-lint >/dev/null 2>&1; then \
		golangci-lint run --fix; \
	elif [ -f "$$GOBIN/golangci-lint" ]; then \
		$$GOBIN/golangci-lint run --fix; \
	else \
		echo "❌ golangci-lint not installed. Run: make install-tools"; \
		exit 1; \
	fi

verify: ## Verify project setup
	@./scripts/verify-setup.sh

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out

deps: ## Download dependencies
	go mod download
	go mod tidy

install-tools: ## Install development tools (swag, golangci-lint, migrate, air)
	@./scripts/install-tools.sh

migrate-up: ## Run database migrations (requires golang-migrate)
	@GOBIN=$$(go env GOPATH)/bin; \
	if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_api?sslmode=disable" up; \
	elif [ -f "$$GOBIN/migrate" ]; then \
		$$GOBIN/migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_api?sslmode=disable" up; \
	else \
		echo "❌ golang-migrate not installed. Run: make install-tools"; \
		echo "Or see: migrations/MIGRATIONS.md"; \
	fi

migrate-down: ## Rollback last migration
	@GOBIN=$$(go env GOPATH)/bin; \
	if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_api?sslmode=disable" down 1; \
	elif [ -f "$$GOBIN/migrate" ]; then \
		$$GOBIN/migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_api?sslmode=disable" down 1; \
	else \
		echo "❌ golang-migrate not installed. Run: make install-tools"; \
	fi

migrate-create: ## Create new migration (usage: make migrate-create NAME=add_user_field)
	@if [ -z "$(NAME)" ]; then \
		echo "❌ Please provide NAME. Usage: make migrate-create NAME=add_user_field"; \
		exit 1; \
	fi
	@GOBIN=$$(go env GOPATH)/bin; \
	if command -v migrate >/dev/null 2>&1; then \
		migrate create -ext sql -dir migrations -seq $(NAME); \
		echo "✅ Created migration files for: $(NAME)"; \
	elif [ -f "$$GOBIN/migrate" ]; then \
		$$GOBIN/migrate create -ext sql -dir migrations -seq $(NAME); \
		echo "✅ Created migration files for: $(NAME)"; \
	else \
		echo "❌ golang-migrate not installed. Run: make install-tools"; \
	fi

migrate-version: ## Show current migration version
	@GOBIN=$$(go env GOPATH)/bin; \
	if command -v migrate >/dev/null 2>&1; then \
		migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_api?sslmode=disable" version; \
	elif [ -f "$$GOBIN/migrate" ]; then \
		$$GOBIN/migrate -path migrations -database "postgres://postgres:postgres@localhost:5432/go_api?sslmode=disable" version; \
	else \
		echo "❌ golang-migrate not installed. Run: make install-tools"; \
	fi

migrate-docker-up: ## Run migrations inside Docker container
	@if docker ps | grep -q go_api_app; then \
		docker exec go_api_app sh -c 'go run -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path /app/migrations -database "postgres://postgres:postgres@db:5432/go_api?sslmode=disable" up'; \
	else \
		echo "❌ Container not running. Start with: make docker-up"; \
	fi

migrate-docker-down: ## Rollback last migration inside Docker container
	@if docker ps | grep -q go_api_app; then \
		docker exec go_api_app sh -c 'go run -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path /app/migrations -database "postgres://postgres:postgres@db:5432/go_api?sslmode=disable" down 1'; \
	else \
		echo "❌ Container not running. Start with: make docker-up"; \
	fi

migrate-docker-version: ## Show migration version inside Docker container
	@if docker ps | grep -q go_api_app; then \
		docker exec go_api_app sh -c 'go run -tags "postgres" github.com/golang-migrate/migrate/v4/cmd/migrate@latest -path /app/migrations -database "postgres://postgres:postgres@db:5432/go_api?sslmode=disable" version'; \
	else \
		echo "❌ Container not running. Start with: make docker-up"; \
	fi

.DEFAULT_GOAL := help

