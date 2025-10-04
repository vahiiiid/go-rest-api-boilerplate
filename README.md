<div align="center">

![GRAB Logo](https://vahiiiid.github.io/go-rest-api-docs/images/logo.png)

# 🧩 GRAB
**G**o **R**EST **A**PI **B**oilerplate

*Grab it and Go — a clean, lightweight, production-ready REST API starter kit in Go with JWT, PostgreSQL, Docker, and Swagger.*

**🚀 Start building in under 2 minutes** • **📚 Fully documented** • **🧪 100% tested** • **🐳 Docker ready**

[![Version](https://img.shields.io/badge/version-1.0.0-blue.svg)](https://github.com/vahiiiid/go-rest-api-boilerplate/releases/tag/v1.0.0)
[![CI](https://github.com/vahiiiid/go-rest-api-boilerplate/workflows/CI/badge.svg)](https://github.com/vahiiiid/go-rest-api-boilerplate/actions)
[![Go Report Card](https://goreportcard.com/badge/github.com/vahiiiid/go-rest-api-boilerplate)](https://goreportcard.com/report/github.com/vahiiiid/go-rest-api-boilerplate)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-latest-brightgreen.svg)](https://vahiiiid.github.io/go-rest-api-docs/)

</div>

## ✨ Features

- ✅ **JWT Authentication** - Secure token-based auth (HS256)
- ✅ **User Management** - Complete CRUD with validation
- ✅ **PostgreSQL + GORM** - Robust database with ORM
- ✅ **Docker Development** - Hot-reload with Air (~2 sec feedback)
- ✅ **Docker Production** - Optimized multi-stage builds
- ✅ **Swagger/OpenAPI** - Interactive API documentation
- ✅ **Database Migrations** - Version-controlled schema changes
- ✅ **Automated Testing** - Unit & integration tests
- ✅ **GitHub Actions CI** - Automated linting and testing
- ✅ **Make Commands** - Simplified workflow automation
- ✅ **Helper Scripts** - Quick setup and verification tools
- ✅ **Postman Collection** - Pre-configured API tests
- ✅ **Clean Architecture** - Layered, maintainable structure
- ✅ **Security Best Practices** - Bcrypt hashing, input validation
- ✅ **CORS Support** - Configurable cross-origin requests

## 🎯 Why GRAB?

**Stop wasting time on boilerplate. Start building features.**

- **⚡ 2-Minute Setup** - One command gets you a fully working API
- **📚 Learn by Example** - Complete TODO list tutorial in the docs
- **🔥 Hot-Reload** - See changes in ~2 seconds, no restart needed
- **✅ Production Ready** - Used in real projects, not just a demo
- **📖 Actually Documented** - Every feature explained with examples
- **🧪 Fully Tested** - All endpoints have working tests you can learn from

**Perfect for:**
- 🚀 Starting new Go projects quickly
- 📖 Learning Go web development best practices
- 🏗️ Building production-ready APIs
- 👥 Team projects with consistent structure

## 📑 Table of Contents

- [Quick Start](#-quick-start)
- [Development](#-development)
- [Production](#-production)
- [API Documentation](#-api-documentation)
- [Testing](#-testing)
- [Documentation](#-documentation)
- [Project Structure](#️-project-structure)
- [Contributing](#-contributing)

---

## 🚀 Quick Start

Get up and running in **under 2 minutes**:

### Prerequisites
- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/downloads)

### One-Command Setup ⚡

```bash
git clone https://github.com/vahiiiid/go-rest-api-boilerplate.git
cd go-rest-api-boilerplate
make quick-start
```

**🎉 Done!** Your API is now running at:
- **API Base URL:** http://localhost:8080/api/v1
- **Swagger UI:** http://localhost:8080/swagger/index.html
- **Health Check:** http://localhost:8080/health

### What Just Happened?

The `quick-start` command automatically:
1. ✅ Installed development tools (swag, golangci-lint, migrate, air)
2. ✅ Verified all prerequisites and dependencies
3. ✅ Created `.env` file from template
4. ✅ Generated Swagger documentation
5. ✅ Built and started Docker containers
6. ✅ Ran database migrations (via AutoMigrate)

### 🚀 Ready to Build Your Own Features?

Now that your API is running, learn how to add your own endpoints!

**📖 [Development Guide](https://vahiiiid.github.io/go-rest-api-docs/DEVELOPMENT_GUIDE/)** - Complete guide showing you how to:
- Understand the codebase structure
- Add new models, routes, and handlers
- Implement CRUD operations
- Follow best practices

**Or if you learn by example**, check out the **[TODO List Example](https://vahiiiid.github.io/go-rest-api-docs/TODO_EXAMPLE/)** - a complete step-by-step tutorial implementing a TODO feature from scratch!

Visit the [documentation site](https://vahiiiid.github.io/go-rest-api-docs/) for the full guide!

### Try It Out 🧪

```bash
# Check health
curl http://localhost:8080/health

# Register a user
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Alice Smith",
    "email": "alice@example.com",
    "password": "secret123"
  }'

# Visit Swagger UI for interactive docs
open http://localhost:8080/swagger/index.html
```

---

## 💻 Development

### Automated Setup (Recommended)

```bash
make quick-start
```

This starts the development environment with:
- 🔥 **Hot-reload** - Code changes reflect in ~2 seconds (powered by Air)
- 📦 **Volume mounts** - Edit code in your IDE, runs in container
- 🗄️ **PostgreSQL** - Database on internal Docker network
- 📚 **Swagger** - Auto-generated API documentation

### Manual Docker Setup

If you prefer to see what's happening step-by-step:

```bash
# 1. Install development tools
make install-tools
# or
./scripts/install-tools.sh

# 2. Create environment file
cp .env.example .env

# 3. Generate Swagger docs
make swag
# or
./scripts/init-swagger.sh

# 4. Start containers with hot-reload
docker-compose up --build
```

**That's it!** The API will automatically reload when you edit code.

> 💡 **Next Step:** Check out the **[Development Guide](https://vahiiiid.github.io/go-rest-api-docs/development-guide/)** to learn how to add your own endpoints! It includes a complete TODO list example with all the code you need.

### Development Workflow

```bash
# Start containers
make docker-up

# Edit code in your IDE
# Changes auto-reload in ~2 seconds! ✨

# Check code quality
make lint

# Run tests
make test

# Generate/update Swagger docs
make swag

# Stop containers
make docker-down

# View logs
docker-compose logs -f app
```

### Available Make Commands

```bash
make help              # Show all available commands
make quick-start       # Complete automated setup
make docker-up         # Start development containers
make docker-down       # Stop containers
make lint              # Run linter (golangci-lint)
make lint-fix          # Auto-fix linting issues
make test              # Run all tests
make test-coverage     # Run tests with coverage report
make swag              # Generate Swagger documentation
make migrate-create    # Create new migration file
make migrate-docker-up # Run migrations in container
make verify            # Verify project setup
make install-tools     # Install development tools
```

### Without Docker (Manual Development)

For local development without Docker, see the [Setup Guide](https://vahiiiid.github.io/go-rest-api-docs/SETUP/) for detailed instructions on:
- Installing Go 1.23+
- Setting up PostgreSQL locally
- Running the app directly on your machine
- Manual migration commands

---

## 🏭 Production

### Overview

GRAB provides optimized production builds with:
- ✅ Multi-stage Docker builds (minimal image size)
- ✅ No development dependencies
- ✅ No mounted volumes
- ✅ Production-ready configuration
- ✅ Health checks

### Simple Deployment (VPS/Server)

```bash
# Clone repository on your server
git clone https://github.com/vahiiiid/go-rest-api-boilerplate.git
cd go-rest-api-boilerplate

# Install development tools (needed for Swagger generation)
make install-tools

# Create production environment file
cp .env.example .env
nano .env  # Edit with production values (database, JWT secret, etc.)

# Generate API documentation
make swag

# Start production containers
make docker-up-prod
```

### Production Configuration

Update `.env` with production values:

```env
ENV=production
PORT=8080
DB_HOST=your-db-host
DB_PORT=5432
DB_USER=your-db-user
DB_PASSWORD=your-strong-password
DB_NAME=your-db-name
JWT_SECRET=your-very-strong-random-secret
JWT_TTL_HOURS=24
```

### Docker Production Build

```bash
# Build production image
docker build -t grab-api:latest .

# Run production container
docker run -p 8080:8080 --env-file .env grab-api:latest
```

### Cloud Deployment

GRAB is ready to deploy to:
- **AWS ECS/Fargate** - Container orchestration
- **Google Cloud Run** - Serverless containers
- **Azure Container Instances** - Managed containers
- **DigitalOcean App Platform** - Platform-as-a-service
- **Kubernetes** - Self-managed orchestration
- **Any VPS** - Using Docker Compose

For detailed deployment guides, database migrations, and Docker production setup, see the [Setup Guide](https://vahiiiid.github.io/go-rest-api-docs/SETUP/) and [Docker Guide](https://vahiiiid.github.io/go-rest-api-docs/DOCKER/).

---

## 📚 API Documentation

### Swagger UI

Interactive API documentation is available at:

```
http://localhost:8080/swagger/index.html
```

Try endpoints directly from your browser with the built-in "Try it out" feature.

### Postman Collection

Import the pre-configured Postman collection with example requests and tests:

```
api/postman_collection.json
```

**Includes:**
- Pre-configured requests for all endpoints
- Environment variables
- Automated tests
- Example payloads

### API Endpoints

#### Authentication

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/api/v1/auth/register` | Register new user |
| POST | `/api/v1/auth/login` | Login user |

#### Users (Protected)

| Method | Endpoint | Description | Auth Required |
|--------|----------|-------------|---------------|
| GET | `/api/v1/users` | List users (paginated) | ✅ |
| GET | `/api/v1/users/:id` | Get user by ID | ✅ |
| PUT | `/api/v1/users/:id` | Update user | ✅ |
| DELETE | `/api/v1/users/:id` | Delete user | ✅ |

#### Health

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/health` | Health check |

### Example Requests

**Register User:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

**Login:**
```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "securepass123"
  }'
```

**List Users (with auth):**
```bash
curl -X GET http://localhost:8080/api/v1/users \
  -H "Authorization: Bearer YOUR_JWT_TOKEN"
```

For more examples, see the [Quick Reference Guide](https://vahiiiid.github.io/go-rest-api-docs/QUICK_REFERENCE/).

---

## 🧪 Testing

### Run Tests

```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run tests in verbose mode
go test ./... -v
```

### Test Coverage

The project includes:
- ✅ **Unit tests** - Handler, service, repository layers
- ✅ **Integration tests** - Full request/response cycle
- ✅ **In-memory SQLite** - No external dependencies for tests
- ✅ **Test fixtures** - Reusable test data
- ✅ **HTTP mocking** - Using `httptest` package

**Test Suites:**
- `TestRegisterHandler` - User registration flows
- `TestLoginHandler` - Authentication flows
- `TestProtectedEndpoints` - Authorization checks
- `TestHealthEndpoint` - Health check validation

### Continuous Integration

GitHub Actions automatically runs on every push:
- ✅ Run all tests
- ✅ Check code with `go vet`
- ✅ Run `golangci-lint`
- ✅ Generate coverage reports

See `.github/workflows/ci.yml` for CI configuration.

---

## 📖 Documentation

Full API documentation, usage guides, and tutorials are maintained in a separate repository:

### 📘 Documentation Site

**🌐 [View Full Documentation](https://vahiiiid.github.io/go-rest-api-docs/)**

Complete, searchable documentation site featuring:
- 🚀 Getting Started guides
- 💻 Development tutorials with examples
- 🏗️ Architecture overview
- 🐳 Docker deployment guides
- 📚 API reference with Swagger
- 🗄️ Database migration guides

### 📦 Documentation Repository

**👉 [go-rest-api-docs](https://github.com/vahiiiid/go-rest-api-docs)**

The documentation repository includes:
- Complete setup and deployment guides
- Step-by-step development tutorials
- TODO list implementation example
- Best practices and patterns
- Troubleshooting guides
- Contributing guidelines

### 🤝 Contributing to Documentation

To contribute to the documentation:
1. Visit the [docs repository](https://github.com/vahiiiid/go-rest-api-docs)
2. Follow the contributing guidelines
3. Submit pull requests for improvements

For contributing to the codebase, see [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 🏗️ Project Structure

```
go-rest-api-boilerplate/
├── api/
│   ├── docs/                      # Swagger documentation (auto-generated)
│   │   ├── docs.go
│   │   ├── swagger.json
│   │   └── swagger.yaml
│   └── postman_collection.json    # Postman API tests with examples
├── cmd/
│   └── server/
│       └── main.go                # Application entry point
├── configs/
│   └── config.yaml                # Configuration file example
├── internal/                      # Private application code
│   ├── auth/                      # Authentication & authorization
│   │   ├── dto.go                 # JWT claims & auth DTOs
│   │   ├── middleware.go          # JWT middleware
│   │   └── service.go             # Token generation & validation
│   ├── db/
│   │   └── db.go                  # Database connection setup
│   ├── server/
│   │   └── router.go              # Route definitions & middleware
│   └── user/                      # User domain (example feature)
│       ├── dto.go                 # Request/Response DTOs
│       ├── handler.go             # HTTP handlers with Swagger docs
│       ├── model.go               # GORM database model
│       ├── repository.go          # Data access layer (CRUD)
│       └── service.go             # Business logic layer
├── migrations/                    # Database migration files
│   ├── 000001_create_users_table.up.sql
│   └── 000001_create_users_table.down.sql
├── scripts/                       # Helper automation scripts
│   ├── install-tools.sh           # Install dev tools (swag, air, etc.)
│   ├── quick-start.sh             # One-command setup
│   ├── init-swagger.sh            # Generate Swagger docs
│   └── verify-setup.sh            # Verify project setup
├── tests/
│   ├── handler_test.go            # Integration tests (httptest)
│   └── README.md                  # Testing guide
├── .air.toml                      # Hot-reload configuration
├── .env.example                   # Environment variables template
├── .gitignore
├── .golangci.yml                  # Linter configuration
├── .github/
│   └── workflows/
│       └── ci.yml                 # GitHub Actions CI/CD
├── CHANGELOG.md                   # Version history
├── CONTRIBUTING.md                # Contribution guidelines
├── docker-compose.yml             # Development with hot-reload
├── docker-compose.prod.yml        # Production optimized
├── Dockerfile                     # Multi-stage build (dev + prod)
├── go.mod                         # Go module dependencies
├── go.sum                         # Dependency checksums
├── LICENSE                        # MIT License
├── Makefile                       # Build automation & shortcuts
└── README.md                      # This file
```

**📚 Documentation**: All guides, tutorials, and API docs are in a [separate repository](https://github.com/vahiiiid/go-rest-api-docs) and published at [vahiiiid.github.io/go-rest-api-docs](https://vahiiiid.github.io/go-rest-api-docs/)

---

## 🏛️ Architecture

GRAB follows **clean architecture** principles with clear separation of concerns:

### Layers

```
┌─────────────────────────────────────┐
│         Handler Layer               │  ← HTTP handlers, request/response
│   (internal/user/handler.go)        │     validation, error handling
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         Service Layer               │  ← Business logic, orchestration
│   (internal/user/service.go)        │     transactions, domain rules
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│       Repository Layer              │  ← Data access, CRUD operations
│  (internal/user/repository.go)      │     database queries
└─────────────────────────────────────┘
              ↓
┌─────────────────────────────────────┐
│         Database (PostgreSQL)       │  ← Data persistence
└─────────────────────────────────────┘
```

### Key Principles

- ✅ **Separation of Concerns** - Each layer has a single responsibility
- ✅ **Dependency Injection** - Loose coupling between layers
- ✅ **Testability** - Easy to mock and test each layer
- ✅ **Maintainability** - Clear structure, easy to navigate
- ✅ **Scalability** - Easy to extend with new features

### Want to Build Your Own Features?

The **[Development Guide](https://vahiiiid.github.io/go-rest-api-docs/development-guide/)** provides a complete walkthrough of this architecture with a real **TODO list implementation** showing you exactly how to:
- Add new models and database tables
- Create repositories, services, and handlers
- Register routes and add Swagger documentation
- Follow the same patterns used in the user management system

**[📖 View Full Documentation](https://vahiiiid.github.io/go-rest-api-docs/)**

---

## 🔐 Security Features

- **Password Hashing** - Bcrypt with configurable cost (default: 10)
- **JWT Tokens** - Secure token generation and validation (HS256)
- **Input Validation** - Request validation using Gin binding tags
- **SQL Injection Protection** - GORM parameterized queries
- **CORS** - Configurable cross-origin resource sharing
- **Rate Limiting** - (Add via middleware - see docs)
- **Environment Variables** - Sensitive data never hardcoded

⚠️ **Production Checklist:**
- [ ] Change `JWT_SECRET` to a strong, random value
- [ ] Use strong database passwords
- [ ] Enable HTTPS/TLS
- [ ] Configure proper CORS origins
- [ ] Set up rate limiting
- [ ] Enable database connection encryption
- [ ] Regular dependency updates

---

## 🤝 Contributing

Contributions are welcome! See [CONTRIBUTING.md](CONTRIBUTING.md) for:
- Code style guidelines
- Pull request process
- Testing requirements
- Commit conventions

### Quick Guidelines

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Make your changes
4. Run tests and linter (`make lint && make test`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feature/amazing-feature`)
7. Open a Pull Request

---

## 📋 Changelog

See [CHANGELOG.md](CHANGELOG.md) for a detailed history of changes.

**Current Version**: [v1.0.0](https://github.com/vahiiiid/go-rest-api-boilerplate/releases/tag/v1.0.0) - Initial stable release

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 🙏 Acknowledgments

Built with these amazing tools:

- [Gin](https://github.com/gin-gonic/gin) - HTTP web framework
- [GORM](https://gorm.io/) - ORM library
- [golang-jwt](https://github.com/golang-jwt/jwt) - JWT implementation
- [swaggo](https://github.com/swaggo/swag) - Swagger documentation
- [Air](https://github.com/air-verse/air) - Hot-reload for development
- [golang-migrate](https://github.com/golang-migrate/migrate) - Database migrations

---

## 📧 Support

- 📖 Check the [documentation](https://vahiiiid.github.io/go-rest-api-docs/)
- 🐛 [Report bugs](https://github.com/vahiiiid/go-rest-api-boilerplate/issues)
- 💬 [Ask questions](https://github.com/vahiiiid/go-rest-api-boilerplate/discussions)
- ⭐ [Star this repo](https://github.com/vahiiiid/go-rest-api-boilerplate) if you find it helpful!

---

<div align="center">

**Made with ❤️ for the Go community**

**[⭐ Star this repo](https://github.com/vahiiiid/go-rest-api-boilerplate)** if you find it useful!

</div>
