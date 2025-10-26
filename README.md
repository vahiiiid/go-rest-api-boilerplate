<div align="center">

![GRAB Logo](https://vahiiiid.github.io/go-rest-api-docs/images/logo.png)

# Go REST API Boilerplate

Production-ready in 90 seconds. No headaches, just clean code.

*GRAB is a Go boilerplate that doesn't waste your time — highly tested, Docker-ready, fully documented, with everything you need.*

[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8?logo=go&logoColor=white)](https://go.dev/)
[![codecov](https://codecov.io/gh/vahiiiid/go-rest-api-boilerplate/graph/badge.svg?branch=main)](https://codecov.io/gh/vahiiiid/go-rest-api-boilerplate)
[![CI](https://github.com/vahiiiid/go-rest-api-boilerplate/workflows/CI/badge.svg)](https://github.com/vahiiiid/go-rest-api-boilerplate/actions)
[![Docker](https://img.shields.io/badge/Docker-ready-2496ED?logo=docker&logoColor=white)](https://www.docker.com/)
[![Go Report Card](https://goreportcard.com/badge/github.com/vahiiiid/go-rest-api-boilerplate)](https://goreportcard.com/report/github.com/vahiiiid/go-rest-api-boilerplate)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Documentation](https://img.shields.io/badge/docs-read%20the%20docs-brightgreen?logo=readthedocs&logoColor=white)](https://vahiiiid.github.io/go-rest-api-docs/)
[![PRs Welcome](https://img.shields.io/badge/PRs-welcome-brightgreen.svg)](CONTRIBUTING.md)
[![GitHub Stars](https://img.shields.io/github/stars/vahiiiid/go-rest-api-boilerplate?style=social)](https://github.com/vahiiiid/go-rest-api-boilerplate/stargazers)

**[📖 Full Documentation](https://vahiiiid.github.io/go-rest-api-docs/)** • **[🚀 Quick Start](#-quick-start)** • **[✨ Live Demo](#-see-it-in-action)**

</div>

---

## 🕒 Why Waste Days on Setup?

You know the pain: Starting a new Go project means days of configuring Docker, wiring up authentication, setting up migrations, writing boilerplate code, and praying your hot-reload actually works.

**GRAB changes that.**

```bash
make quick-start  # ← One command. 90 seconds. You're building features.
```

**This is the real deal.** The production-grade boilerplate you wish you had from day one:

✅ **Clean Architecture** — Handler → Service → Repository (GO industry standard)  
✅ **Security & JWT Auth** — Rate limiting, CORS, input validation built-in  
✅ **Database Migrations** — PostgreSQL with version control & rollback  
✅ **Comprehensive Tests** — Unit + integration with CI/CD pipeline  
✅ **Interactive Docs** — Auto-generated Swagger + Postman collection  
✅ **Structured Logging** — JSON logs with request IDs and tracing  
✅ **Production Docker** — Multi-stage builds, health checks, optimized images  
✅ **Environment-Aware** — Dev/staging/prod configs + Make automation & more  
✅ **Graceful Shutdown** — Zero-downtime deployments with configurable timeouts  
✅ **Hot-Reload (2 seconds!)** — Powered by Air, not magic  

**And that's just scratching the surface.** Check the [full documentation](https://vahiiiid.github.io/go-rest-api-docs/) to see everything GRAB offers.

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

> **💡 Want to run without Docker?** See the [Manual Setup Guide](https://vahiiiid.github.io/go-rest-api-docs/SETUP/) in the documentation.

### One-Command Setup ⚡

```bash
git clone https://github.com/vahiiiid/go-rest-api-boilerplate.git
cd go-rest-api-boilerplate
make quick-start
```

<div align="center">
  <img src="https://vahiiiid.github.io/go-rest-api-docs/images/quick-start-light.gif" alt="Quick Start Demo" width="800">
</div>

**🎉 Done!** Your API is now running at:

- **API Base URL:** <http://localhost:8080/api/v1>
- **Swagger UI:** <http://localhost:8080/swagger/index.html>
- **Health Check:** <http://localhost:8080/health>

---

## ✨ See It In Action

### Interactive Swagger Documentation

<div align="center">
  <img src="https://vahiiiid.github.io/go-rest-api-docs/images/swagger-ui.png" alt="Swagger UI" width="700">
</div>

Open [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html) to explore and test all endpoints interactively.

### Or Use Postman

<div align="center">
  <img src="https://vahiiiid.github.io/go-rest-api-docs/images/postman-collection.png" alt="Postman Collection" width="700">
</div>

Import the pre-configured collection from `api/postman_collection.json` with example requests and tests.

**🚀 Ready to Build?**

- 📖 **[Development Guide](https://vahiiiid.github.io/go-rest-api-docs/DEVELOPMENT_GUIDE/)** — Learn how to add models, routes, and handlers
- 💡 **[TODO List Tutorial](https://vahiiiid.github.io/go-rest-api-docs/TODO_EXAMPLE/)** — Complete step-by-step feature implementation from scratch

---

## 💎 What Makes GRAB Different?

### Not Just Features — A Complete Development Experience

Most boilerplates give you code. **GRAB gives you a professional development workflow.**

#### 🔐 Authentication That Actually Works

- **JWT-based auth** (HS256) — Secure, stateless, ready for production
- **Context helpers** — Type-safe user extraction (no more casting nightmares)
- **Password security** — Bcrypt hashing with best-practice cost factor
- **Rate limiting** — Token-bucket protection against abuse built-in

👉 [Context Helpers Guide](https://vahiiiid.github.io/go-rest-api-docs/CONTEXT_HELPERS/)

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

👉 [Migrations Guide](https://vahiiiid.github.io/go-rest-api-docs/MIGRATIONS_GUIDE/)

#### 🐳 Docker That Saves Your Sanity

- **2-second hot-reload** — Powered by Air, actually works in Docker
- **One command to rule them all** — `make quick-start` handles everything
- **Development & production** — Separate optimized configs
- **Multi-stage builds** — Tiny production images (~20MB)

👉 [Docker Guide](https://vahiiiid.github.io/go-rest-api-docs/DOCKER/)

#### 📚 Documentation That Exists (And Helps!)

- **Auto-generated Swagger** — Interactive API explorer at `/swagger/index.html`
- **Full documentation site** — Not just README, real guides at [vahiiiid.github.io/go-rest-api-docs](https://vahiiiid.github.io/go-rest-api-docs/)
- **Step-by-step tutorials** — Build a TODO app from scratch
- **Postman collection** — Import and test immediately

👉 [Full Documentation](https://vahiiiid.github.io/go-rest-api-docs/)

#### 🧪 Tests That Give You Confidence

- **Comprehensive coverage** — Handlers, services, and repositories all tested
- **In-memory SQLite** — No external dependencies for tests
- **Table-driven tests** — Go idiomatic testing patterns
- **CI/CD ready** — GitHub Actions configured and working

👉 [Testing Guide](https://vahiiiid.github.io/go-rest-api-docs/TESTING/)

#### 🏗️ Architecture That Scales

- **Clean layers** — Handler → Service → Repository (no shortcuts)
- **Dependency injection** — Proper DI, easy to mock and test
- **Domain-driven** — Organize by feature, not by layer
- **Official Go layout** — Follows [golang-standards/project-layout](https://github.com/golang-standards/project-layout)

👉 [Development Guide](https://vahiiiid.github.io/go-rest-api-docs/DEVELOPMENT_GUIDE/)

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

👉 **[Complete Migration Guide](https://vahiiiid.github.io/go-rest-api-docs/MIGRATIONS_GUIDE/)**

### Without Docker

Want to run natively? You'll need Go 1.24+ installed.

```bash
make build-binary    # Build binary to bin/server
make run-binary      # Build and run (requires PostgreSQL on localhost)
```

👉 **[Full Setup Guide](https://vahiiiid.github.io/go-rest-api-docs/SETUP/)** for native development

---

## 🚢 Deployment

### Production-Ready From Day One

GRAB includes optimized production builds:

```bash
make docker-up-prod  # Start production containers
```

**What's included:**

- ✅ Multi-stage Docker builds (minimal image size)
- ✅ Health check endpoints
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

👉 **[Deployment Guide](https://vahiiiid.github.io/go-rest-api-docs/SETUP/)** for step-by-step instructions

---

## 🎃 Hacktoberfest 2025

<div align="center">

![Hacktoberfest](https://img.shields.io/badge/Hacktoberfest-2025-orange?style=for-the-badge&logo=digitalocean&logoColor=white)

**We're participating in Hacktoberfest 2025! 🚀**

</div>

We welcome contributions from developers of all skill levels! Pick up any [open issues](https://github.com/vahiiiid/go-rest-api-boilerplate/issues) labeled `hacktoberfest` or `good first issue`, fork the repository, make your changes, and submit a pull request. Whether it's bug fixes, new features, documentation improvements, or test enhancements - every contribution counts! 🎉

---

## 📖 Documentation

### 🌐 Full Documentation Site

**[📚 Read the Docs →](https://vahiiiid.github.io/go-rest-api-docs/)**

Complete guides covering everything:

- 🚀 [Getting Started](https://vahiiiid.github.io/go-rest-api-docs/SETUP/) — Installation and configuration
- 💻 [Development Guide](https://vahiiiid.github.io/go-rest-api-docs/DEVELOPMENT_GUIDE/) — Building features
- 💡 [TODO Tutorial](https://vahiiiid.github.io/go-rest-api-docs/TODO_EXAMPLE/) — Step-by-step implementation
- 🐳 [Docker Guide](https://vahiiiid.github.io/go-rest-api-docs/DOCKER/) — Container workflows
- 🗄️ [Migrations](https://vahiiiid.github.io/go-rest-api-docs/MIGRATIONS_GUIDE/) — Database schema management
- 🧪 [Testing](https://vahiiiid.github.io/go-rest-api-docs/TESTING/) — Writing and running tests
- 📚 [Swagger](https://vahiiiid.github.io/go-rest-api-docs/SWAGGER/) — API documentation
- ⚙️ [Configuration](https://vahiiiid.github.io/go-rest-api-docs/CONFIGURATION/) — Environment setup

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
- **[GORM](https://gorm.io/)** — Developer-friendly ORM
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

- 📖 [Read the Documentation](https://vahiiiid.github.io/go-rest-api-docs/)
- 🐛 [Report Bugs](https://github.com/vahiiiid/go-rest-api-boilerplate/issues)
- 💬 [Ask Questions](https://github.com/vahiiiid/go-rest-api-boilerplate/discussions)
- ⭐ [Star this repo](https://github.com/vahiiiid/go-rest-api-boilerplate) if you find it helpful!

---

<div align="center">

**Made with ❤️ for the Go community**

[⭐ Star](https://github.com/vahiiiid/go-rest-api-boilerplate) • [📖 Docs](https://vahiiiid.github.io/go-rest-api-docs/) • [🐛 Issues](https://github.com/vahiiiid/go-rest-api-boilerplate/issues) • [💬 Discussions](https://github.com/vahiiiid/go-rest-api-boilerplate/discussions)

</div>