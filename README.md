<div align="center">

[![GRAB Logo](https://grabapi.dev/docs/images/logo.png)](https://grabapi.dev/)

# Go REST API Boilerplate

Production-ready in 90 seconds. No headaches, just clean code.

*GRAB is a Go boilerplate that doesn't waste your time — highly tested, Docker-ready, fully documented, **AI-assistant optimized**, with everything you need. Built for developers who want to code with AI, not fight it.*

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![Release](https://img.shields.io/github/v/release/vahiiiid/go-rest-api-boilerplate?color=blue)](https://github.com/vahiiiid/go-rest-api-boilerplate/releases/tag/v2.0.0)
[![codecov](https://codecov.io/gh/vahiiiid/go-rest-api-boilerplate/graph/badge.svg?branch=main)](https://codecov.io/gh/vahiiiid/go-rest-api-boilerplate)
[![CI](https://github.com/vahiiiid/go-rest-api-boilerplate/workflows/CI/badge.svg)](https://github.com/vahiiiid/go-rest-api-boilerplate/actions)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![Go Report Card](https://goreportcard.com/badge/github.com/vahiiiid/go-rest-api-boilerplate)](https://goreportcard.com/report/github.com/vahiiiid/go-rest-api-boilerplate)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-read%20the%20docs-brightgreen?logo=readthedocs&logoColor=white)](https://grabapi.dev/docs/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![GitHub Stars](https://img.shields.io/github/stars/vahiiiid/go-rest-api-boilerplate?style=social)](https://github.com/vahiiiid/go-rest-api-boilerplate/stargazers)

**[🌐 Official Website](https://grabapi.dev/)** • **[📖 Full Documentation](https://grabapi.dev/docs/)** • **[🚀 Quick Start](#-quick-start)** • **[✨ Live Demo](#-see-it-in-action)**

</div>

---

## 🤖 AI-Friendly Development

GRAB is designed to work seamlessly with your favorite AI coding assistants:

[![GitHub Copilot](https://img.shields.io/badge/GitHub%20Copilot-optimized-7F52FF?logo=github&logoColor=white)](https://github.com/features/copilot)
[![Cursor](https://img.shields.io/badge/Cursor-ready-7C3AED?logo=cursor&logoColor=white)](https://cursor.sh/)
[![Windsurf](https://img.shields.io/badge/Windsurf-supported-00C7B7?logoColor=white)](https://codeium.com/windsurf)
[![GoLand](https://img.shields.io/badge/GoLand-dual%20AI-087CFA?logo=goland&logoColor=white)](https://www.jetbrains.com/go/)
[![AGENTS.md](https://img.shields.io/badge/AGENTS.md-compliant-orange?logo=openai&logoColor=white)](AGENTS.md)

**Out-of-the-box AI integration** with comprehensive guidelines for:
- **GitHub Copilot** (VS Code, GoLand, Visual Studio)
- **Cursor IDE** (with dedicated `.cursor/rules/`)
- **Windsurf IDE** (with dedicated `.windsurf/rules/`)
- **JetBrains AI** (via AGENTS.md standard)
- Any AI assistant supporting AGENTS.md standard

> **Note**: GoLand users get dual AI support through both GitHub Copilot (via `.github/copilot-instructions.md`) and JetBrains AI (via `AGENTS.md`). No IDE-specific configuration needed.

AI assistants understand GRAB's Clean Architecture, Docker-first workflow, migration patterns, and testing conventions. Get intelligent code completions, accurate refactoring suggestions, and context-aware help.

**📚 Learn More**: [AI-Friendly Development Guide](https://grabapi.dev/docs/AI_FRIENDLY/)

---

## 🕒 Why Waste Days on Setup?

You know the pain: Starting a new Go project means days of configuring Docker, wiring up authentication, setting up migrations, writing boilerplate code, and praying your hot-reload actually works.

**GRAB changes that.**

```bash
make quick-start  # ← One command. 90 seconds. You're building features.
```

**This is the real deal.** The production-grade boilerplate you wish you had from day one:

✅ **Clean Architecture** — Handler → Service → Repository (GO industry standard)  
✅ **AI-Optimized Guidelines** — Built-in rules for GitHub Copilot, Cursor, Windsurf & AGENTS.md  
✅ **Security & JWT Auth** — OAuth 2.0 BCP compliant with refresh token rotation, rate limiting, CORS  
✅ **Role-Based Access Control** — Many-to-many RBAC with JWT integration and secure admin CLI  
✅ **Database Migrations** — PostgreSQL with version control & rollback  
✅ **Comprehensive Tests** — Unit + integration with CI/CD pipeline  
✅ **Interactive Docs** — Auto-generated Swagger + Postman collection  
✅ **Structured Logging** — JSON logs with request IDs and tracing  
✅ **Standardized API Responses** — Consistent envelope format for all endpoints  
✅ **Structured Error Handling** — Machine-readable error codes with details  
✅ **Production Docker** — Multi-stage builds, health checks, optimized images  
✅ **Environment-Aware** — Dev/staging/prod configs + Make automation & more  
✅ **Graceful Shutdown** — Zero-downtime deployments with configurable timeouts  
✅ **Hot-Reload (2 seconds!)** — Powered by Air, not magic  

**And that's just scratching the surface.** Check the [full documentation](https://grabapi.dev/docs/) to see everything GRAB offers.

### 🏆 Built Following Go Standards

Not some random structure — follows **[official Go project layout](https://go.dev/doc/modules/layout)** + battle-tested community patterns from **[golang-standards/project-layout](https://github.com/golang-standards/project-layout)**. The same architecture used by Gin, GORM, and production Go services.

### 🎯 Perfect For

- 🚀 **Shipping Fast** — Launch MVPs and production APIs in days, not weeks  
- 👥 **Team Projects** — Consistent standards everyone understands  
- 🏗️ **Scaling Up** — Architecture that grows with your business
- 📖 **Learning Go** — See how pros structure real-world applications

---

## 🚀 Quick Start

Get your API running in **under 2 minutes**:

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and [Docker Compose](https://docs.docker.com/compose/install/)
- [Git](https://git-scm.com/downloads)

> **💡 Want to run without Docker?** See the [Manual Setup Guide](https://grabapi.dev/docs/SETUP/) in the documentation.

### One-Command Setup ⚡

```bash
git clone https://github.com/vahiiiid/go-rest-api-boilerplate.git
cd go-rest-api-boilerplate
make quick-start
```

<div align="center">
  <img src="https://grabapi.dev/docs/images/quick-start-light.gif" alt="Quick Start Demo" width="800">
</div>

**🎉 Done!** Your API is now running at:

- **API Base URL:** <http://localhost:8080/api/v1>
- **Swagger UI:** <http://localhost:8080/swagger/index.html>
- **Health Checks:** <http://localhost:8080/health> • [/health/live](http://localhost:8080/health/live) • [/health/ready](http://localhost:8080/health/ready)

**Create Admin User:**

```bash
make create-admin              # Interactive: prompts for email, name, password
make promote-admin ID=1        # Promote existing user to admin by ID
```

---

## ✨ See It In Action

### Interactive Swagger Documentation

<div align="center">
  <img src="https://grabapi.dev/docs/images/swagger-ui.png" alt="Swagger UI" width="700">
</div>

Open [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) to explore and test all endpoints interactively.

### Or Use Postman

<div align="center">
  <img src="https://grabapi.dev/docs/images/postman-collection.png" alt="Postman Collection" width="700">
</div>

Import the pre-configured collection from `api/postman_collection.json` with example requests and tests.

**🚀 Ready to Build?**

- 📖 **[Development Guide](https://grabapi.dev/docs/DEVELOPMENT_GUIDE/)** — Learn how to add models, routes, and handlers
- 💡 **[TODO List Tutorial](https://grabapi.dev/docs/TODO_EXAMPLE/)** — Complete step-by-step feature implementation from scratch

---

## 💎 What Makes GRAB Different?

### Not Just Features — A Complete Development Experience

Most boilerplates give you code. **GRAB gives you a professional development workflow.**

#### 🔐 Authentication That Actually Works

- **OAuth 2.0 BCP compliant** — JWT-based auth (HS256) with refresh token rotation and automatic reuse detection
- **Enhanced security** — Refresh tokens with family tracking, secure token invalidation, and breach detection
- **Context helpers** — Type-safe user extraction (no more casting nightmares)
- **Password security** — Bcrypt hashing with best-practice cost factor
- **Rate limiting** — Token-bucket protection against abuse built-in

👉 [Authentication Guide](https://grabapi.dev/docs/AUTHENTICATION/) | [Context Helpers](https://grabapi.dev/docs/CONTEXT_HELPERS/)

#### 🔑 Role-Based Access Control (RBAC)

- **Many-to-many architecture** — Flexible roles system with extensible permissions
- **Secure admin CLI** — Interactive admin creation with strong password enforcement (no defaults in code)
- **JWT-integrated authorization** — Roles embedded in tokens for server-side validation
- **Protected endpoints** — Middleware-based access control (RequireRole, RequireAdmin)
- **Three-endpoint pattern** — `/auth/me` (current user), `/users/:id` (specific), `/users` (admin list)
- **Paginated user management** — Admin-only user listing with filtering and search

👉 [RBAC Guide](https://grabapi.dev/docs/RBAC/)

#### 🗄️ Database Setup That Doesn't Fight You

- **PostgreSQL + GORM** — Production-grade ORM with relationship support
- **golang-migrate** — Industry-standard migrations with timestamp versioning
- **Complete migration CLI** — Create, apply, rollback with ease

  ```bash
  make migrate-create NAME=add_posts_table  # Create with timestamp
  make migrate-up                            # Apply all pending
  make migrate-down                          # Rollback last (safe)
  make migrate-down STEPS=3                  # Rollback multiple
  make migrate-status                        # Check current version
  make migrate-goto VERSION=<timestamp>      # Jump to specific version
  ```

- **Safety features** — Confirmation prompts, dirty state detection
- **Transaction support** — BEGIN/COMMIT wrappers for data integrity
- **Connection pooling** — Configured for performance out of the box

👉 [Migrations Guide](https://grabapi.dev/docs/MIGRATIONS_GUIDE/)

#### 🐳 Docker That Saves Your Sanity

- **2-second hot-reload** — Powered by Air, actually works in Docker
- **One command to rule them all** — `make quick-start` handles everything
- **Development & production** — Separate optimized configs
- **Multi-stage builds** — Tiny production images (~20MB)

👉 [Docker Guide](https://grabapi.dev/docs/DOCKER/)

#### 🏥 Production-Grade Health Checks

- **Kubernetes-ready probes** — Liveness (`/health/live`) and readiness (`/health/ready`) endpoints
- **Database health monitoring** — Response time tracking with pass/warn/fail thresholds
- **RFC-compliant responses** — Following IETF draft standards for health check format
- **Zero-downtime deployments** — Smart readiness checks for load balancer integration
- **Extensible architecture** — Easy to add custom health checkers (Redis, external APIs, etc.)

👉 [Health Checks Guide](https://grabapi.dev/docs/HEALTH_CHECKS/)

#### 📚 Documentation That Exists (And Helps!)

- **Auto-generated Swagger** — Interactive API explorer at `/swagger/index.html`
- **Full documentation site** — Not just README, real guides at [grabapi.dev/docs](https://grabapi.dev/docs/)
- **Step-by-step tutorials** — Build a TODO app from scratch
- **Postman collection** — Import and test immediately

👉 [Full Documentation](https://grabapi.dev/docs/)

#### 🧪 Tests That Give You Confidence

- **Comprehensive coverage** — Handlers, services, and repositories all tested
- **In-memory SQLite** — No external dependencies for tests
- **Table-driven tests** — Go idiomatic testing patterns
- **CI/CD ready** — GitHub Actions configured and working

👉 [Testing Guide](https://grabapi.dev/docs/TESTING/)

#### 📦 Standardized API Responses

- **Consistent envelope format** — All responses wrapped in `{success, data, error, meta}` structure
- **JSend-inspired design** — Industry best practice for API response formatting
- **Type-safe responses** — Predictable structure for frontend integration
- **Metadata support** — Pagination, timestamps, request IDs built-in

👉 [API Response Format Guide](https://grabapi.dev/docs/API_RESPONSE_FORMAT/)

#### ⚠️ Error Handling That Makes Sense

- **Structured API errors** — Machine-readable codes (NOT_FOUND, VALIDATION_ERROR, etc.)
- **Detailed error info** — Code, message, details, timestamp, path, request ID
- **Validation details** — Clear field-level error messages for bad requests
- **Centralized middleware** — Single error handler for consistent responses
- **Rate limit errors** — Includes `retry_after` for proper backoff logic

👉 [Error Handling Guide](https://grabapi.dev/docs/ERROR_HANDLING/)

#### 🏗️ Architecture That Scales

- **Clean layers** — Handler → Service → Repository (no shortcuts)
- **Dependency injection** — Proper DI, easy to mock and test
- **Domain-driven** — Organize by feature, not by layer
- **Official Go layout** — Follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

👉 [Development Guide](https://grabapi.dev/docs/DEVELOPMENT_GUIDE/)

---

## 🛠️ Development

### With Docker (Recommended)

The easiest way to develop with hot-reload and zero setup:

```bash
make up        # Start containers with hot-reload
make logs      # View logs
make test      # Run all tests
make lint      # Check code quality
make lint-fix  # Auto-fix linting issues
make down      # Stop containers
```

**What you get:**

- 🔥 **Hot-reload** — Code changes reflect in ~2 seconds (powered by Air)
- 📦 **Volume mounts** — Edit code in your IDE, runs in container
- 🗄️ **PostgreSQL** — Database on internal Docker network
- 📚 **All tools pre-installed** — No Go installation needed on host

### Database Migrations

Production-grade migrations using golang-migrate:

```bash
make migrate-create NAME=add_todos_table  # Create new migration
make migrate-up                            # Apply all pending
make migrate-down                          # Rollback last migration
make migrate-status                        # Check current version
```

For long-running migrations:

```bash
go run cmd/migrate/main.go up --timeout=30m --lock-timeout=1m
```

All environments use SQL migrations for consistency and safety.

👉 **[Complete Migration Guide](https://grabapi.dev/docs/MIGRATIONS_GUIDE/)**

### Without Docker

Want to run natively? You'll need Go 1.24+ installed.

```bash
make build-binary    # Build binary to bin/server
make run-binary      # Build and run (requires PostgreSQL on localhost)
```

👉 **[Full Setup Guide](https://grabapi.dev/docs/SETUP/)** for native development

---

## 🚢 Deployment

### Production-Ready From Day One

GRAB includes optimized production builds:

```bash
make docker-up-prod  # Start production containers
```

**What's included:**

- ✅ Multi-stage Docker builds (minimal image size)
- ✅ Production-grade health checks (liveness & readiness probes)
- ✅ Environment-based configuration
- ✅ No development dependencies
- ✅ Production logging

### Deploy Anywhere

Ready for:

- **AWS ECS/Fargate** — Container orchestration
- **Google Cloud Run** — Serverless containers
- **DigitalOcean App Platform** — Platform-as-a-service
- **Kubernetes** — Self-managed orchestration
- **Any VPS** — Using Docker Compose

👉 **[Deployment Guide](https://grabapi.dev/docs/SETUP/)** for step-by-step instructions

---

## 📖 Documentation

### 🌐 Full Documentation Site

**[📚 Read the Docs →](https://grabapi.dev/docs/)**

Complete guides covering everything:

- 🚀 [Getting Started](https://grabapi.dev/docs/SETUP/) — Installation and configuration
- 💻 [Development Guide](https://grabapi.dev/docs/DEVELOPMENT_GUIDE/) — Building features
- 💡 [TODO Tutorial](https://grabapi.dev/docs/TODO_EXAMPLE/) — Step-by-step implementation
- 🐳 [Docker Guide](https://grabapi.dev/docs/DOCKER/) — Container workflows
- 🗄️ [Migrations](https://grabapi.dev/docs/MIGRATIONS_GUIDE/) — Database schema management
- 🏥 [Health Checks](https://grabapi.dev/docs/HEALTH_CHECKS/) — Kubernetes probes and monitoring
- 🧪 [Testing](https://grabapi.dev/docs/TESTING/) — Writing and running tests
- 📦 [API Response Format](https://grabapi.dev/docs/API_RESPONSE_FORMAT/) — Standardized response envelope
- ⚠️ [Error Handling](https://grabapi.dev/docs/ERROR_HANDLING/) — Structured API errors
- 📚 [Swagger](https://grabapi.dev/docs/SWAGGER/) — API documentation
- ⚙️ [Configuration](https://grabapi.dev/docs/CONFIGURATION/) — Environment setup

### 🤝 Contributing to Documentation

Documentation lives in a [separate repository](https://github.com/vahiiiid/go-rest-api-docs). To contribute:

1. Visit [github.com/vahiiiid/go-rest-api-docs](https://github.com/vahiiiid/go-rest-api-docs)
2. Follow the contributing guidelines
3. Submit pull requests for improvements

For code contributions, see [CONTRIBUTING.md](CONTRIBUTING.md)

---

## 🤝 Contributing

We ❤️ contributions! See [CONTRIBUTING.md](CONTRIBUTING.md) for:

- Code style guidelines
- Pull request process
- Testing requirements
- Commit conventions

### Quick Start

1. Fork the repository
2. Create a feature branch (`git checkout -b feat/amazing-feature`)
3. Make your changes
4. Run tests and linter (`make lint-fix && make lint && make test`)
5. Commit your changes (`git commit -m 'feat: add amazing feature'`)
6. Push to the branch (`git push origin feat/amazing-feature`)
7. Open a Pull Request

---

## 🙏 Built With Amazing Tools

- **[Gin](https://github.com/gin-gonic/gin)** — Fast HTTP web framework
- **[GORM](https://github.com/go-gorm/gorm)** — Developer-friendly ORM
- **[golang-migrate](https://github.com/golang-migrate/migrate)** — Database migration toolkit
- **[Viper](https://github.com/spf13/viper)** — Configuration management
- **[golang-jwt](https://github.com/golang-jwt/jwt)** — JWT implementation
- **[swaggo](https://github.com/swaggo/swag)** — Swagger documentation generator
- **[Air](https://github.com/air-verse/air)** — Hot-reload for development

---

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

---

## 💬 Support & Community

- 🌐 [Official Website](https://grabapi.dev/)
- 📖 [Read the Documentation](https://grabapi.dev/docs/)
- 🐛 [Report Bugs](https://github.com/vahiiiid/go-rest-api-boilerplate/issues)
- 💬 [Ask Questions](https://github.com/vahiiiid/go-rest-api-boilerplate/discussions)
- ⭐ [Star this repo](https://github.com/vahiiiid/go-rest-api-boilerplate) if you find it helpful!

---

<div align="center">

**Made with ❤️ for the Go community**

[🌐 Website](https://grabapi.dev/) • [⭐ Star](https://github.com/vahiiiid/go-rest-api-boilerplate) • [📖 Docs](https://grabapi.dev/docs/) • [🐛 Issues](https://github.com/vahiiiid/go-rest-api-boilerplate/issues) • [💬 Discussions](https://github.com/vahiiiid/go-rest-api-boilerplate/discussions)

</div>
