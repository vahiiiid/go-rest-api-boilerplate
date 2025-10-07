# Changelog

All notable changes to the Go REST API Boilerplate (GRAB) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Planned
- Rate limiting middleware
- Email verification for user registration
- Password reset functionality
- Refresh token support

---

## [1.1.0] - 2025-01-15

### Added
- ✨ **Request Logging Middleware** - Structured JSON logging for all HTTP requests
  - Logs request method, path, status code, duration, client IP, and request ID
  - Uses Go's standard `log/slog` for structured logging
  - Configurable to skip specific paths (e.g., health checks)
  - Automatic request ID generation and propagation
  - Log level adjustment based on response status (INFO/WARN/ERROR)
  - Comprehensive unit tests with 100% coverage
  - Location: `internal/middleware/logger.go`

### Changed
- Updated project structure documentation to reflect new directories
- Enhanced README.md to clarify configurable request logging feature

### Fixed
- Updated go.mod dependencies after adding yaml.v3 import
- Fixed GitHub Actions go mod tidy error

---

## [1.0.0] - 2025-01-05

### 🎉 Initial Release

The first production-ready release of GRAB - Go REST API Boilerplate.

### Added

#### Core Features
- **JWT Authentication** - Secure token-based authentication with HS256
- **User Management** - Complete CRUD operations for users
- **PostgreSQL Integration** - GORM ORM with PostgreSQL support
- **RESTful API** - Clean REST endpoints following best practices
- **Swagger Documentation** - Interactive API documentation with swaggo
- **Database Migrations** - Version-controlled schema with golang-migrate

#### Development Experience
- **Docker Development** - Hot-reload with Air (~2 sec feedback)
- **Docker Production** - Optimized multi-stage builds
- **Make Commands** - Simplified workflow automation
- **Helper Scripts** - Quick setup and verification tools
  - `quick-start.sh` - Automated setup
  - `verify-setup.sh` - Project verification
  - `init-swagger.sh` - Swagger generation
  - `install-tools.sh` - Development tools installation

#### Testing & Quality
- **Unit Tests** - Handler and service layer tests
- **Integration Tests** - Full request/response cycle tests
- **GitHub Actions CI** - Automated linting and testing
- **Code Linting** - golangci-lint configuration
- **Test Coverage** - Coverage reports

#### Security
- **Password Hashing** - Bcrypt with configurable cost
- **JWT Token Validation** - Secure token verification
- **Input Validation** - Request validation with Gin binding
- **SQL Injection Protection** - GORM parameterized queries
- **CORS Support** - Configurable cross-origin requests

#### Documentation
- **Comprehensive README** - Quick start and feature overview
- **Setup Guide** - Detailed installation instructions
- **Docker Guide** - Container deployment guide
- **Swagger Guide** - API documentation guide
- **Quick Reference** - Command cheat sheet
- **Development Guide** - Architecture and patterns
- **TODO Example** - Complete implementation tutorial
- **Postman Collection** - Pre-configured API tests

#### API Endpoints
- `POST /api/v1/auth/register` - User registration
- `POST /api/v1/auth/login` - User login
- `GET /api/v1/users/:id` - Get user by ID (protected)
- `PUT /api/v1/users/:id` - Update user (protected)
- `DELETE /api/v1/users/:id` - Delete user (protected)
- `GET /health` - Health check endpoint

#### Infrastructure
- Multi-stage Dockerfile for development and production
- Docker Compose for local development
- Environment variable configuration
- Database connection pooling
- Graceful shutdown handling

### Technical Stack
- **Language**: Go 1.23+
- **Framework**: Gin Web Framework
- **ORM**: GORM
- **Database**: PostgreSQL 15
- **Authentication**: JWT (golang-jwt)
- **Documentation**: Swagger (swaggo)
- **Testing**: Go testing + httptest
- **Hot Reload**: Air
- **Linting**: golangci-lint
- **Migrations**: golang-migrate

### Project Structure
```
go-rest-api-boilerplate/
├── cmd/server/          # Application entry point
├── internal/            # Private application code
│   ├── auth/           # Authentication logic
│   ├── user/           # User domain
│   ├── db/             # Database connection
│   └── server/         # Router setup
├── api/                # API documentation
├── migrations/         # Database migrations
├── tests/              # Test files
├── scripts/            # Helper scripts
└── configs/            # Configuration files
```

### Dependencies
- github.com/gin-gonic/gin v1.10.0
- gorm.io/gorm v1.25.12
- gorm.io/driver/postgres v1.5.9
- github.com/golang-jwt/jwt/v5 v5.2.1
- github.com/swaggo/swag v1.16.4
- golang.org/x/crypto v0.31.0

### Notes
- This is the first stable release suitable for production use
- All core features are tested and documented
- Breaking changes will follow semantic versioning
- See [Documentation](https://vahiiiid.github.io/go-rest-api-docs/) for detailed guides

---

## Version History

- **1.0.0** (2025-01-05) - Initial stable release

---

## Links

- [Documentation](https://vahiiiid.github.io/go-rest-api-docs/)
- [GitHub Repository](https://github.com/vahiiiid/go-rest-api-boilerplate)
- [Report Issues](https://github.com/vahiiiid/go-rest-api-boilerplate/issues)
- [Contributing Guidelines](CONTRIBUTING.md)

---

**Legend:**
- 🎉 Major release
- ✨ New feature
- 🐛 Bug fix
- 📝 Documentation
- 🔒 Security
- ⚡ Performance
- 🔧 Configuration
- 🗑️ Deprecated
- ❌ Removed
