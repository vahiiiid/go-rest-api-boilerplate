# GRAB - AI-Friendly Development Guide

**Version**: v2.0.0  
**Last Updated**: 2025-12-10  
**Purpose**: Universal AI assistant guidelines for GRAB (Go REST API Boilerplate)

> This file follows the OpenAI AGENTS.md standard and is compatible with all major AI coding assistants including GitHub Copilot, Cursor, Windsurf, JetBrains AI, and others.

---

## 📋 Project Overview

**GRAB (Go REST API Boilerplate)** is a production-ready Go REST API starter with Clean Architecture, comprehensive testing (89.81% coverage), and Docker-first development workflow.

### Technology Stack

- **Go**: Check version with `go version`
- **Gin**: HTTP router framework
- **GORM**: PostgreSQL ORM
- **PostgreSQL**: Check version with `make exec-db` then `psql --version`
- **Docker**: Check version with `docker --version`
- **golang-migrate**: Database migration tool
- **JWT**: Authentication with refresh token rotation
- **Air**: Hot-reload development
- **golangci-lint**: Code quality enforcement
- **Swagger**: OpenAPI documentation

### Documentation

- **Main Docs**: https://vahiiiid.github.io/go-rest-api-docs/
- **Repository**: https://github.com/vahiiiid/go-rest-api-boilerplate
- **Issues**: https://github.com/vahiiiid/go-rest-api-boilerplate/issues

---

## 🏗️ Architecture

### Clean Architecture Pattern

GRAB strictly follows Clean Architecture with clear layer separation:

```
Handler (HTTP) → Service (Business Logic) → Repository (Database)
```

**Domain Structure**:
```
internal/<domain>/
├── model.go       # GORM models with database tags
├── dto.go         # Data Transfer Objects (API contracts)
├── repository.go  # Database access interface + implementation
├── service.go     # Business logic interface + implementation
├── handler.go     # HTTP handlers with Gin + Swagger annotations
└── *_test.go      # Unit and integration tests
```

**Reference Implementation**: See `internal/user/` for complete domain example.

**Key Rules**:
- Handlers only handle HTTP concerns (bind, validate, respond)
- Services contain all business logic
- Repositories only interact with database
- Never skip layers or cross boundaries

---

## 🚀 Development Workflow

### Docker-First Approach

**Important**: Developers run `make` commands on the host machine. The Makefile automatically detects if Docker containers are running and executes commands in the appropriate context.

**No need to manually enter containers** - the Makefile handles this transparently.

```bash
# Start all containers
make up

# All commands below auto-detect Docker and execute accordingly
make test              # Run tests
make lint              # Run linting
make lint-fix          # Auto-fix linting issues
make swag              # Generate Swagger documentation
make migrate-up        # Apply database migrations
make logs              # View container logs
```

### Pre-Commit Checklist

Always run before committing:
```bash
make lint-fix    # Auto-fix linting issues
make lint        # Verify no remaining issues
make test        # Run all tests
make swag        # Update Swagger if API changed
```

---

## 📝 Common Tasks

### Adding a New Domain/Entity

**Step-by-step process**:

1. **Create directory structure**:
   ```bash
   mkdir -p internal/<domain>
   ```

2. **Create model** (`internal/<domain>/model.go`):
   ```go
   package <domain>
   
   import (
       "time"
       "gorm.io/gorm"
   )
   
   type <Entity> struct {
       ID        uint           `gorm:"primarykey" json:"id"`
       Name      string         `gorm:"not null" json:"name"`
       UserID    uint           `gorm:"not null" json:"user_id"`
       CreatedAt time.Time      `json:"created_at"`
       UpdatedAt time.Time      `json:"updated_at"`
       DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
   }
   ```

3. **Create DTOs** (`internal/<domain>/dto.go`):
   ```go
   package <domain>
   
   type Create<Entity>Request struct {
       Name string `json:"name" binding:"required,min=3,max=200"`
   }
   
   type Update<Entity>Request struct {
       Name string `json:"name" binding:"omitempty,min=3,max=200"`
   }
   
   type <Entity>Response struct {
       ID        uint      `json:"id"`
       Name      string    `json:"name"`
       UserID    uint      `json:"user_id"`
       CreatedAt time.Time `json:"created_at"`
   }
   ```

4. **Create repository** (`internal/<domain>/repository.go`):
   ```go
   package <domain>
   
   import (
       "context"
       "gorm.io/gorm"
   )
   
   type Repository interface {
       Create(ctx context.Context, entity *<Entity>) error
       FindByID(ctx context.Context, id uint) (*<Entity>, error)
       Update(ctx context.Context, entity *<Entity>) error
       Delete(ctx context.Context, id uint) error
   }
   
   type repository struct {
       db *gorm.DB
   }
   
   func NewRepository(db *gorm.DB) Repository {
       return &repository{db: db}
   }
   ```

5. **Create service** (`internal/<domain>/service.go`):
   ```go
   package <domain>
   
   import "context"
   
   type Service interface {
       Create<Entity>(ctx context.Context, userID uint, req *Create<Entity>Request) (*<Entity>Response, error)
       Get<Entity>(ctx context.Context, userID, id uint) (*<Entity>Response, error)
       Update<Entity>(ctx context.Context, userID, id uint, req *Update<Entity>Request) (*<Entity>Response, error)
       Delete<Entity>(ctx context.Context, userID, id uint) error
   }
   
   type service struct {
       repo Repository
   }
   
   func NewService(repo Repository) Service {
       return &service{repo: repo}
   }
   ```

6. **Create handler** (`internal/<domain>/handler.go`):
   ```go
   package <domain>
   
   import (
       "net/http"
       "strconv"
       
       "github.com/gin-gonic/gin"
       "go-rest-api-boilerplate/internal/contextutil"
       "go-rest-api-boilerplate/internal/errors"
   )
   
   type Handler struct {
       service Service
   }
   
   func NewHandler(service Service) *Handler {
       return &Handler{service: service}
   }
   
   // @Summary Create entity
   // @Tags <domain>
   // @Accept json
   // @Produce json
   // @Security BearerAuth
   // @Param request body Create<Entity>Request true "Entity data"
   // @Success 201 {object} <Entity>Response
   // @Failure 400 {object} errors.ErrorResponse
   // @Router /api/v1/<domain> [post]
   func (h *Handler) Create<Entity>(c *gin.Context) {
       userID := contextutil.GetUserID(c)
       
       var req Create<Entity>Request
       if err := c.ShouldBindJSON(&req); err != nil {
           errors.HandleValidationError(c, err)
           return
       }
       
       result, err := h.service.Create<Entity>(c.Request.Context(), userID, &req)
       if err != nil {
           errors.HandleError(c, err)
           return
       }
       
       c.JSON(http.StatusCreated, result)
   }
   ```

7. **Create database migration**:
   ```bash
   make migrate-create NAME=create_<table>_table
   ```
   
   Edit the generated `.up.sql` file:
   ```sql
   BEGIN;
   
   CREATE TABLE IF NOT EXISTS <table> (
       id SERIAL PRIMARY KEY,
       name VARCHAR(200) NOT NULL,
       user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
       created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
       deleted_at TIMESTAMP
   );
   
   CREATE INDEX idx_<table>_user_id ON <table>(user_id);
   CREATE INDEX idx_<table>_deleted_at ON <table>(deleted_at);
   
   COMMIT;
   ```
   
   Edit the `.down.sql` file:
   ```sql
   BEGIN;
   
   DROP TABLE IF EXISTS <table>;
   
   COMMIT;
   ```

8. **Register routes** in `internal/server/router.go`:
   ```go
   // Initialize components
   <domain>Repo := <domain>.NewRepository(db)
   <domain>Service := <domain>.NewService(<domain>Repo)
   <domain>Handler := <domain>.NewHandler(<domain>Service)
   
   // Register routes
   v1.Use(authMiddleware.RequireAuth()).Group("/<domain>").
       POST("", <domain>Handler.Create<Entity>).
       GET("/:id", <domain>Handler.Get<Entity>).
       PUT("/:id", <domain>Handler.Update<Entity>).
       DELETE("/:id", <domain>Handler.Delete<Entity>)
   ```

9. **Write tests** for all layers

10. **Apply changes**:
    ```bash
    make migrate-up      # Apply migration
    make test            # Run tests
    make lint            # Check code quality
    make swag            # Update Swagger docs
    ```

### Database Migrations

**Naming Convention**: `YYYYMMDDHHMMSS_verb_noun_table`

**Examples**:
- `20251025225126_create_users_table`
- `20251028000000_create_refresh_tokens_table`
- `20251210120000_add_avatar_to_users_table`
- `20251215143000_add_index_to_users_email`

**Commands**:
```bash
make migrate-create NAME=create_todos_table    # Create new migration
make migrate-up                                 # Apply all pending
make migrate-down                               # Rollback one
make migrate-status                             # Check status
make migrate-force VERSION=<version>           # Force version
```

**Best Practices**:
- Wrap in `BEGIN;` / `COMMIT;` transactions
- Use `IF NOT EXISTS` for safety
- Create indexes for foreign keys
- Create indexes for frequently queried columns
- Always write corresponding `.down.sql`
- Test rollback before committing

---

## 🔐 Authentication & Authorization

### Getting Current User

```go
import "go-rest-api-boilerplate/internal/contextutil"

func (h *Handler) SomeHandler(c *gin.Context) {
    userID := contextutil.GetUserID(c)
    userEmail := contextutil.GetUserEmail(c)
    userRole := contextutil.GetUserRole(c)
    
    // Use user information...
}
```

### Protecting Routes

```go
// Require authentication
v1.Use(authMiddleware.RequireAuth()).Group("/todos")

// Require specific role (admin)
v1.Use(authMiddleware.RequireAuth()).
   Use(rbacMiddleware.RequireRole("admin")).
   POST("/admin/users", userHandler.CreateUser)

// Multiple roles
v1.Use(authMiddleware.RequireAuth()).
   Use(rbacMiddleware.RequireRole("admin", "moderator")).
   GET("/admin/reports", reportHandler.GetReports)
```

---

## ❌ Error Handling

GRAB uses centralized error handling:

```go
import "go-rest-api-boilerplate/internal/errors"

// Validation errors (automatic field extraction)
if err := c.ShouldBindJSON(&req); err != nil {
    errors.HandleValidationError(c, err)
    return
}

// Standard errors
errors.HandleError(c, errors.ErrNotFound)
errors.HandleError(c, errors.ErrUnauthorized)
errors.HandleError(c, errors.ErrForbidden)
errors.HandleError(c, errors.ErrBadRequest)
errors.HandleError(c, errors.ErrInternalServer)

// Service/repository errors
result, err := h.service.CreateUser(ctx, req)
if err != nil {
    errors.HandleError(c, err)
    return
}
```

---

## 🧪 Testing

### Test Structure

Use table-driven tests:

```go
func TestService_CreateEntity(t *testing.T) {
    tests := []struct {
        name        string
        userID      uint
        request     *CreateEntityRequest
        setupMocks  func(*MockRepository)
        expectError bool
        errorType   error
    }{
        {
            name:   "success",
            userID: 1,
            request: &CreateEntityRequest{Name: "Test"},
            setupMocks: func(m *MockRepository) {
                m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
            },
            expectError: false,
        },
        {
            name:        "validation_error",
            userID:      1,
            request:     &CreateEntityRequest{Name: ""},
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            ctrl := gomock.NewController(t)
            defer ctrl.Finish()
            
            mockRepo := NewMockRepository(ctrl)
            if tt.setupMocks != nil {
                tt.setupMocks(mockRepo)
            }
            
            service := NewService(mockRepo)
            result, err := service.CreateEntity(context.Background(), tt.userID, tt.request)
            
            if tt.expectError {
                assert.Error(t, err)
                if tt.errorType != nil {
                    assert.Equal(t, tt.errorType, err)
                }
            } else {
                assert.NoError(t, err)
                assert.NotNil(t, result)
            }
        })
    }
}
```

### Test Commands

```bash
make test              # Run all tests
make test-coverage     # Generate coverage report (opens in browser)
make test-verbose      # Run with verbose output
```

---

## 📚 Swagger/OpenAPI Documentation

### Annotations

```go
// @Summary Create user
// @Description Create a new user account
// @Tags users
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateUserRequest true "User creation data"
// @Success 201 {object} UserResponse
// @Failure 400 {object} errors.ErrorResponse
// @Failure 401 {object} errors.ErrorResponse
// @Router /api/v1/users [post]
func (h *Handler) CreateUser(c *gin.Context) {
    // Handler implementation
}
```

### Update Documentation

```bash
make swag    # Regenerate Swagger docs

# View at: http://localhost:8080/swagger/index.html
```

---

## ⚙️ Configuration

### Configuration Files

- `configs/config.yaml` - Base configuration
- `configs/config.development.yaml` - Development overrides
- `configs/config.staging.yaml` - Staging overrides
- `configs/config.production.yaml` - Production overrides

### Environment Variables

Override any config value with environment variables:

```bash
DATABASE_PASSWORD=secret      # Overrides database.password
JWT_SECRET=secret            # Overrides jwt.secret
APP_ENVIRONMENT=production   # Overrides app.environment
RATE_LIMIT_ENABLED=true      # Overrides ratelimit.enabled
```

**Full Configuration Guide**: https://vahiiiid.github.io/go-rest-api-docs/CONFIGURATION/

---

## 🎯 Out-of-the-Box Features

GRAB includes these production-ready features:

1. **JWT Authentication** - Access + refresh tokens with rotation ([Docs](https://vahiiiid.github.io/go-rest-api-docs/AUTHENTICATION/))
2. **RBAC** - Role-based access control ([Docs](https://vahiiiid.github.io/go-rest-api-docs/RBAC/))
3. **Database Migrations** - Versioned SQL migrations ([Docs](https://vahiiiid.github.io/go-rest-api-docs/MIGRATIONS_GUIDE/))
4. **Health Checks** - `/health`, `/live`, `/ready` endpoints ([Docs](https://vahiiiid.github.io/go-rest-api-docs/HEALTH_CHECKS/))
5. **Rate Limiting** - Token bucket algorithm ([Docs](https://vahiiiid.github.io/go-rest-api-docs/RATE_LIMITING/))
6. **Structured Logging** - JSON logs with context ([Docs](https://vahiiiid.github.io/go-rest-api-docs/LOGGING/))
7. **API Response Format** - Standardized responses ([Docs](https://vahiiiid.github.io/go-rest-api-docs/API_RESPONSE_FORMAT/))
8. **Error Handling** - Centralized error management ([Docs](https://vahiiiid.github.io/go-rest-api-docs/ERROR_HANDLING/))
9. **Graceful Shutdown** - Clean termination ([Docs](https://vahiiiid.github.io/go-rest-api-docs/GRACEFUL_SHUTDOWN/))
10. **Swagger/OpenAPI** - Auto-generated docs ([Docs](https://vahiiiid.github.io/go-rest-api-docs/SWAGGER/))
11. **Context Helpers** - Request utilities ([Docs](https://vahiiiid.github.io/go-rest-api-docs/CONTEXT_HELPERS/))

---

## 🔧 Quick Reference

### Essential Commands

| Task | Command |
|------|---------|
| Start development | `make up` |
| Stop containers | `make down` |
| Run tests | `make test` |
| Lint code | `make lint` |
| Fix linting | `make lint-fix` |
| Create migration | `make migrate-create NAME=<name>` |
| Apply migrations | `make migrate-up` |
| Rollback migration | `make migrate-down` |
| Migration status | `make migrate-status` |
| Update Swagger | `make swag` |
| View logs | `make logs` |
| Enter app container | `make exec` |
| Enter DB container | `make exec-db` |
| Clean restart | `make down && make up` |
| Health check | `curl localhost:8080/health` |
| View all commands | `make help` |

### Project Structure

```
go-rest-api-boilerplate/
├── .github/              # GitHub workflows, templates
├── .cursor/              # Cursor AI rules
├── .windsurf/            # Windsurf AI rules
├── api/                  # API documentation
├── cmd/                  # Application entry points
├── configs/              # Configuration files
├── internal/             # Application code
│   ├── auth/             # Authentication
│   ├── config/           # Config management
│   ├── contextutil/      # Context helpers
│   ├── db/               # Database setup
│   ├── errors/           # Error handling
│   ├── health/           # Health checks
│   ├── middleware/       # HTTP middleware
│   ├── migrate/          # Migration logic
│   ├── server/           # Router setup
│   └── user/             # User domain (reference)
├── migrations/           # SQL migration files
├── scripts/              # Helper scripts
├── tests/                # Integration tests
├── Dockerfile            # Docker image
├── docker-compose.yml    # Development compose
├── Makefile              # Development commands
├── AGENTS.md             # This file
└── README.md             # Project overview
```

---

## 💡 Best Practices for AI Assistants

1. **Reference Existing Code**: Always check `internal/user/` for patterns before creating new domains
2. **Follow Clean Architecture**: Never skip Handler → Service → Repository layers
3. **Use Context Helpers**: Import `contextutil` for user information from JWT
4. **Minimal Comments**: Write self-documenting code, comment WHY not WHAT
5. **Test Coverage**: Maintain 85%+ test coverage for all new code
6. **Check Makefile**: All development commands available in `make help`
7. **Read Documentation**: Comprehensive guides at https://vahiiiid.github.io/go-rest-api-docs/
8. **Version Checking**: Show commands to check versions, don't hardcode
9. **Docker-First**: Assume Docker containers running, use `make` commands
10. **Migration Naming**: Follow `YYYYMMDDHHMMSS_verb_noun_table` pattern

---

## 🔗 Additional Resources

- **Full Documentation**: https://vahiiiid.github.io/go-rest-api-docs/
- **GitHub Repository**: https://github.com/vahiiiid/go-rest-api-boilerplate
- **Issue Tracker**: https://github.com/vahiiiid/go-rest-api-boilerplate/issues
- **Discussions**: https://github.com/vahiiiid/go-rest-api-boilerplate/discussions
- **Development Guide**: https://vahiiiid.github.io/go-rest-api-docs/DEVELOPMENT_GUIDE/
- **Quick Reference**: https://vahiiiid.github.io/go-rest-api-docs/QUICK_REFERENCE/

---

**Version**: v2.0.0  
**Last Updated**: 2025-12-10  
**Maintained By**: GRAB Contributors
