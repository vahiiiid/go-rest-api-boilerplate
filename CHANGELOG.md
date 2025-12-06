# Changelog

All notable changes to the Go REST API Boilerplate (GRAB) project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

---

## [2.0.0] - 2025-12-06

### 🎉 Major Release

This release represents a significant evolution of GRAB with enterprise-grade features, dramatically improved test coverage, and production-hardened security.

### Added

#### 🔐 Security & Authentication
- ✨ **Refresh Token System** (PR #85) - OAuth 2.0 BCP compliant implementation
  - Token rotation with family tracking for reuse detection
  - Automatic revocation on suspicious activity
  - SHA-256 token hashing for secure storage
  - Configurable TTLs (Access: 15m, Refresh: 7 days)
  - New endpoints: `/api/v1/auth/refresh`, `/api/v1/auth/logout`
  - Complete integration and unit test coverage
  - Migration: `20251028000000_create_refresh_tokens_table`
  
- ✨ **Role-Based Access Control (RBAC)** - Many-to-many role architecture
  - Built-in roles: `user` (default), `admin`
  - Secure admin CLI (`cmd/createadmin/`) with interactive password creation
  - JWT-integrated authorization (roles in token claims)
  - Middleware: `RequireRole()`, `RequireAdmin()`
  - Protected admin endpoints with paginated user management
  - Context helpers: `HasRole()`, `IsAdmin()`, `GetRoles()`
  - Database migrations for `roles` and `user_roles` tables
  - Comprehensive RBAC documentation

#### 🏥 Health & Reliability
- ✨ **Production-Grade Health Checks** (PR #96)
  - `/health` - Basic health endpoint with uptime
  - `/health/live` - Kubernetes liveness probe
  - `/health/ready` - Readiness probe with database health checks
  - RFC-compliant response format (IETF draft standard)
  - Graceful degradation for database failures
  - Response time tracking with configurable thresholds
  - Extensible health checker architecture

- ✨ **Graceful Shutdown** (PR #79)
  - Signal handling (SIGINT, SIGTERM)
  - HTTP server timeouts (read/write/idle)
  - Connection draining with configurable timeout
  - Database connection cleanup
  - Zero-downtime deployment support

#### 🔧 Configuration & Error Handling
- ✨ **Centralized Configuration** (PR #50)
  - Viper-based configuration management
  - Environment-aware configs (development/staging/production)
  - YAML + environment variable override support
  - Docker-friendly configuration
  - Startup validation for required secrets

- ✨ **Structured Error Handling** (PR #84)
  - Centralized error package with standard error codes
  - Consistent error response format across all endpoints
  - Error middleware with structured logging
  - Request ID tracking for debugging
  - Machine-readable error details

- ✨ **API Response Format Standardization**
  - Envelope format: `{success, data, error, meta}`
  - Consistent structure across all endpoints
  - Pagination metadata support
  - Error codes and structured details

#### 🗄️ Database & Migrations
- ✨ **Migration System** (PR #83, #75)
  - Removed GORM AutoMigrate anti-pattern
  - Implemented golang-migrate with versioned SQL files
  - Added migration commands: `make migrate-up`, `make migrate-down`, `make migrate-status`
  - Migration testing script
  - Four migrations: users, refresh_tokens, roles, user_roles

#### 📝 Documentation & Community
- ✨ **Community Governance**
  - Code of Conduct (PR #45)
  - Security Policy (PR #46)
  - Contributing Guidelines (PR #47)
  - Issue Templates (PR #48)

- ✨ **Comprehensive Documentation**
  - Authentication guide with refresh token flows
  - RBAC implementation guide
  - Health checks guide
  - Graceful shutdown guide
  - Error handling documentation
  - Configuration guide
  - Context helpers reference

### Changed

#### 🔒 Security Improvements
- 🔒 **JWT Secrets** (PR #95) - Removed from config files, now required via environment variables
- 🔒 **Production Validation** - Startup checks enforce secure configuration
- 🔒 **Password Hashing** - Bcrypt with configurable cost (default: 10)
- 🔒 **Rate Limiting** - Now enabled by default in production configurations

#### 📦 Package & Architecture
- 📦 **Package Rename** (PR #89) - `contexthelpers` → `contextutil` (more specific, avoids conflicts)
- 📦 **Clean Architecture** - Maintained strict Handler → Service → Repository separation

#### ⚡ Developer Experience  
- ⚡ **Environment Variables** (PR #82) - Full override support for all configuration values
- ⚡ **Comment Reduction** (PR #91) - Reduced from 6% to ~3% (industry standard)
- ⚡ **Code Quality** - Consistent Go file structure and ordering standards

### Fixed

- 🐛 **Migration Reliability** - Removed AutoMigrate, preventing schema drift
- 🐛 **Error Responses** - Standardized format prevents inconsistent API contracts
- 🐛 **Health Check Accuracy** - Database checks now properly reflect system readiness

### Improved

#### 📊 Test Coverage
- ✅ **Increased from ~17% to 89.81%** (+72.81% improvement)
- ✅ Comprehensive unit tests for all layers
- ✅ Integration tests for critical paths
- ✅ Handler, service, and repository coverage
- ✅ RBAC middleware test coverage
- ✅ Refresh token rotation and reuse detection tests
- ✅ Health check system tests

#### 🔍 CI/CD
- ✅ **Codecov Integration** (PR #64, #76) - Automated coverage reporting
- ✅ GitHub Actions CI with comprehensive checks
- ✅ Automated linting and testing
- ✅ Coverage thresholds enforcement

### Security

- 🔒 Refresh token rotation prevents token theft
- 🔒 Token reuse detection with automatic family revocation
- 🔒 JWT secrets enforced via environment variables
- 🔒 RBAC with many-to-many role architecture
- 🔒 Admin CLI with secure interactive password creation
- 🔒 Rate limiting enabled by default

### Breaking Changes

#### API Response Format
All API responses now use standardized envelope format:

**Before (v1.1.0):**
```json
{
  "id": 1,
  "name": "John Doe"
}
```

**After (v2.0.0):**
```json
{
  "success": true,
  "data": {
    "id": 1,
    "name": "John Doe"
  }
}
```

#### Error Response Format
Error responses now include structured error codes:

**Before (v1.1.0):**
```json
{
  "error": "invalid credentials"
}
```

**After (v2.0.0):**
```json
{
  "success": false,
  "error": {
    "code": "UNAUTHORIZED",
    "message": "Invalid credentials",
    "details": {}
  }
}
```

#### Authentication Flow
- Login now returns both access and refresh tokens
- Refresh endpoint (`/api/v1/auth/refresh`) added for token rotation
- Logout endpoint (`/api/v1/auth/logout`) added for token revocation

#### Package Naming
- Package `contexthelpers` renamed to `contextutil`

### Migration Notes

#### Database Migrations
Run migrations to add new tables:
```bash
make migrate-up
```

This adds:
- `refresh_tokens` table (OAuth 2.0 token rotation)
- `roles` table (RBAC system)
- `user_roles` junction table (many-to-many relationships)

#### Environment Variables
Add required environment variables:
```bash
# Required in production
JWT_SECRET=your-secret-here
JWT_ACCESS_TOKEN_TTL=15m
JWT_REFRESH_TOKEN_TTL=168h
```

#### User Roles
All new registrations automatically receive the `user` role. To create admins:
```bash
make create-admin
```

### Technical Stack Updates

- **Go**: 1.24.0
- **Gin**: v1.9.1
- **GORM**: v1.25.10
- **PostgreSQL**: 15
- **JWT**: golang-jwt/jwt/v5 v5.2.0
- **Viper**: v1.21.0
- **golang-migrate**: v4.19.0

### API Endpoints (v2.0.0)

#### Authentication
- `POST /api/v1/auth/register` - User registration (returns token pair)
- `POST /api/v1/auth/login` - User login (returns token pair)
- `POST /api/v1/auth/refresh` - Refresh access token (NEW)
- `POST /api/v1/auth/logout` - Revoke refresh token (NEW)
- `GET /api/v1/auth/me` - Get current user profile (NEW)

#### Users
- `GET /api/v1/users` - List all users (admin only, NEW)
- `GET /api/v1/users/:id` - Get user by ID (own or admin)
- `PUT /api/v1/users/:id` - Update user (own or admin)
- `DELETE /api/v1/users/:id` - Delete user (own or admin)

#### Health
- `GET /health` - Basic health check (NEW)
- `GET /health/live` - Liveness probe (NEW)
- `GET /health/ready` - Readiness probe with DB checks (NEW)

### Documentation

- **Full Documentation**: https://vahiiiid.github.io/go-rest-api-docs/
- **RBAC Guide**: https://vahiiiid.github.io/go-rest-api-docs/RBAC/
- **Authentication Guide**: https://vahiiiid.github.io/go-rest-api-docs/AUTHENTICATION/
- **Health Checks**: https://vahiiiid.github.io/go-rest-api-docs/HEALTH_CHECKS/
- **Migration Guide**: https://vahiiiid.github.io/go-rest-api-docs/MIGRATIONS_GUIDE/

### Credits

Special thanks to all contributors who made this release possible!

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
- **Language**: Go 1.25
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
