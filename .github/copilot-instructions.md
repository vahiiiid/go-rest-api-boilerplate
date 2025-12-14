# GitHub Copilot Instructions for GRAB (Go REST API Boilerplate)

**Version**: v2.0.0  
**Last Updated**: 2025-12-10  
**Purpose**: Developer-focused guidelines for building APIs with GRAB

---

## 📋 What is GRAB?

GRAB (Go REST API Boilerplate) is a production-ready Go REST API starter with:
- **Clean Architecture** (Handler → Service → Repository)
- **JWT Authentication** with refresh token rotation
- **Role-Based Access Control (RBAC)**
- **Database Migrations** (golang-migrate)
- **Docker-First Development**
- **89.81% Test Coverage**
- **Comprehensive Documentation**: https://vahiiiid.github.io/go-rest-api-docs/

---

## 🎯 Core Development Principles

### 1. **Environment Detection - Don't Hardcode Versions**
Instead of stating "Go 1.24" or "PostgreSQL 15", show how to check:

```bash
# Check Go version
go version

# Check Docker version
docker --version

# Check PostgreSQL version (inside container)
make exec-db
psql --version
```

### 2. **Docker-First Development**
- Developers run `make` commands on host
- **Makefile automatically detects** if Docker container is running
- Commands execute in container if available, host otherwise
- **No need to manually enter container** - the Makefile handles execution context

```bash
# Start containers first
make up

# Run tests (automatically in container if running)
make test

# Run linting (automatically in container if running)
make lint

# Apply lint fixes (automatically in container if running)
make lint-fix

# Generate Swagger docs (automatically in container if running)
make swag
```

### 3. **Clean Architecture Pattern**
Every domain follows this structure:

```
internal/
└── <domain>/
    ├── model.go       # Domain models (GORM)
    ├── dto.go         # Data Transfer Objects (API contracts)
    ├── repository.go  # Database access layer
    ├── service.go     # Business logic layer
    ├── handler.go     # HTTP handlers (Gin)
    └── *_test.go      # Tests for each layer
```

**Key Rules**:
- Handler → Service → Repository (never skip layers)
- No business logic in handlers
- No HTTP concerns in services
- Repository only talks to database

---

## 🚀 Common Development Tasks

### Adding a New Domain/Entity

**Example**: Adding a "Todo" entity

1. **Create directory structure**:
```bash
mkdir -p internal/todo
```

2. **Create model** (`internal/todo/model.go`):
```go
package todo

import (
    "time"
    "gorm.io/gorm"
)

type Todo struct {
    ID          uint           `gorm:"primarykey" json:"id"`
    Title       string         `gorm:"not null" json:"title"`
    Description string         `json:"description"`
    Completed   bool           `gorm:"default:false" json:"completed"`
    UserID      uint           `gorm:"not null" json:"user_id"`
    CreatedAt   time.Time      `json:"created_at"`
    UpdatedAt   time.Time      `json:"updated_at"`
    DeletedAt   gorm.DeletedAt `gorm:"index" json:"-"`
}
```

3. **Create DTO** (`internal/todo/dto.go`):
```go
package todo

type CreateTodoRequest struct {
    Title       string `json:"title" binding:"required,min=3,max=200"`
    Description string `json:"description" binding:"max=1000"`
}

type UpdateTodoRequest struct {
    Title       string `json:"title" binding:"omitempty,min=3,max=200"`
    Description string `json:"description" binding:"omitempty,max=1000"`
    Completed   *bool  `json:"completed"`
}

type TodoResponse struct {
    ID          uint      `json:"id"`
    Title       string    `json:"title"`
    Description string    `json:"description"`
    Completed   bool      `json:"completed"`
    UserID      uint      `json:"user_id"`
    CreatedAt   time.Time `json:"created_at"`
    UpdatedAt   time.Time `json:"updated_at"`
}
```

4. **Create repository** (`internal/todo/repository.go`):
```go
package todo

import (
    "context"
    "gorm.io/gorm"
)

type Repository interface {
    Create(ctx context.Context, todo *Todo) error
    FindByID(ctx context.Context, id uint) (*Todo, error)
    FindByUserID(ctx context.Context, userID uint) ([]Todo, error)
    Update(ctx context.Context, todo *Todo) error
    Delete(ctx context.Context, id uint) error
}

type repository struct {
    db *gorm.DB
}

func NewRepository(db *gorm.DB) Repository {
    return &repository{db: db}
}

// Implementation methods...
```

5. **Create service** (`internal/todo/service.go`):
```go
package todo

import (
    "context"
    "go-rest-api-boilerplate/internal/errors"
)

type Service interface {
    CreateTodo(ctx context.Context, userID uint, req *CreateTodoRequest) (*TodoResponse, error)
    GetTodo(ctx context.Context, userID, todoID uint) (*TodoResponse, error)
    GetUserTodos(ctx context.Context, userID uint) ([]TodoResponse, error)
    UpdateTodo(ctx context.Context, userID, todoID uint, req *UpdateTodoRequest) (*TodoResponse, error)
    DeleteTodo(ctx context.Context, userID, todoID uint) error
}

type service struct {
    repo Repository
}

func NewService(repo Repository) Service {
    return &service{repo: repo}
}

// Implementation methods...
```

6. **Create handler** (`internal/todo/handler.go`):
```go
package todo

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

// @Summary Create todo
// @Tags todos
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param request body CreateTodoRequest true "Todo creation request"
// @Success 201 {object} TodoResponse
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo}
// @Router /api/v1/todos [post]
func (h *Handler) CreateTodo(c *gin.Context) {
    userID := contextutil.GetUserID(c)
    
    var req CreateTodoRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        _ = c.Error(apiErrors.FromGinValidation(err))
        return
    }
    
    todo, err := h.service.CreateTodo(c.Request.Context(), userID, &req)
    if err != nil {
        _ = c.Error(apiErrors.InternalServerError(err))
        return
    }
    
    c.JSON(http.StatusCreated, todo)
}

// Additional handler methods...
```

7. **Create migration**:
```bash
make migrate-create NAME=create_todos_table
```

Edit the generated migration file:
```sql
-- migrations/YYYYMMDDHHMMSS_create_todos_table.up.sql
BEGIN;

CREATE TABLE IF NOT EXISTS todos (
    id SERIAL PRIMARY KEY,
    title VARCHAR(200) NOT NULL,
    description TEXT,
    completed BOOLEAN DEFAULT FALSE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP
);

CREATE INDEX idx_todos_user_id ON todos(user_id);
CREATE INDEX idx_todos_deleted_at ON todos(deleted_at);

COMMIT;
```

```sql
-- migrations/YYYYMMDDHHMMSS_create_todos_table.down.sql
BEGIN;

DROP TABLE IF EXISTS todos;

COMMIT;
```

8. **Register routes** in `internal/server/router.go`:
```go
// Initialize todo components
todoRepo := todo.NewRepository(db)
todoService := todo.NewService(todoRepo)
todoHandler := todo.NewHandler(todoService)

// Register todo routes
v1.Use(authMiddleware.RequireAuth()).Group("/todos").
    POST("", todoHandler.CreateTodo).
    GET("", todoHandler.GetUserTodos).
    GET("/:id", todoHandler.GetTodo).
    PUT("/:id", todoHandler.UpdateTodo).
    DELETE("/:id", todoHandler.DeleteTodo)
```

9. **Write tests** (`internal/todo/*_test.go`)

10. **Run migration and tests**:
```bash
make migrate-up
make test
make lint
make swag
```

---

### Creating Database Migrations

**Pattern**: `YYYYMMDDHHMMSS_verb_noun_table`

**Examples**:
- `20251025225126_create_users_table`
- `20251028000000_create_refresh_tokens_table`
- `20251122153000_create_roles_table`
- `20251210120000_add_avatar_to_users_table`

**Steps**:
```bash
# 1. Generate migration files
make migrate-create NAME=create_todos_table

# 2. Edit the .up.sql file
vim migrations/YYYYMMDDHHMMSS_create_todos_table.up.sql

# 3. Edit the .down.sql file (for rollback)
vim migrations/YYYYMMDDHHMMSS_create_todos_table.down.sql

# 4. Apply migration
make migrate-up

# 5. Verify
make migrate-status

# 6. If needed, rollback
make migrate-down
```

**Migration Best Practices**:
- Always wrap in `BEGIN;` / `COMMIT;` transactions
- Include indexes for foreign keys and frequently queried columns
- Use `IF NOT EXISTS` for safety
- Write corresponding `.down.sql` for every `.up.sql`
- Test rollback before committing

---

### Working with Authentication

**Protected Routes** (require valid JWT):
```go
v1.Use(authMiddleware.RequireAuth()).Group("/todos")
```

**Role-Based Access**:
```go
// Admin-only route
v1.Use(authMiddleware.RequireAuth()).
   Use(rbacMiddleware.RequireRole("admin")).
   POST("/admin/users", userHandler.CreateUser)
```

**Getting Current User**:
```go
import "go-rest-api-boilerplate/internal/contextutil"

func (h *Handler) MyHandler(c *gin.Context) {
    userID := contextutil.GetUserID(c)
    userEmail := contextutil.GetEmail(c)
    userRoles := contextutil.GetRoles(c)
    
    // Use user information...
}
```

---

### Error Handling

Use the centralized error handling:

```go
import apiErrors "go-rest-api-boilerplate/internal/errors"

// Validation errors
if err := c.ShouldBindJSON(&req); err != nil {
    _ = c.Error(apiErrors.FromGinValidation(err))
    return
}

// Custom errors
if todo == nil {
    _ = c.Error(apiErrors.NotFound("Todo not found"))
    return
}

// Service errors
result, err := h.service.CreateTodo(ctx, userID, req)
if err != nil {
    _ = c.Error(apiErrors.InternalServerError(err))
    return
}
```

**Available Error Types**:
- `apiErrors.NotFound(message)` - 404 Not Found
- `apiErrors.Unauthorized(message)` - 401 Unauthorized
- `apiErrors.Forbidden(message)` - 403 Forbidden
- `apiErrors.BadRequest(message)` - 400 Bad Request
- `apiErrors.InternalServerError(err)` - 500 Internal Server Error

---

### Testing

**Test Structure**:
```go
func TestService_CreateTodo(t *testing.T) {
    tests := []struct {
        name        string
        userID      uint
        request     *CreateTodoRequest
        setupMocks  func(*MockRepository)
        expectError bool
        errorType   error
    }{
        {
            name:   "success",
            userID: 1,
            request: &CreateTodoRequest{
                Title: "Test Todo",
            },
            setupMocks: func(m *MockRepository) {
                m.EXPECT().Create(gomock.Any(), gomock.Any()).Return(nil)
            },
            expectError: false,
        },
        // More test cases...
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
            result, err := service.CreateTodo(context.Background(), tt.userID, tt.request)
            
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

**Run tests**:
```bash
make test              # Run all tests
make test-coverage     # Generate coverage report
make test-verbose      # Run with verbose output
```

---

### Updating Swagger Documentation

After adding/modifying endpoints:

```bash
# Regenerate Swagger docs
make swag

# View docs
open http://localhost:8080/swagger/index.html
```

**Swagger Annotations**:
```go
// @Summary Short description
// @Description Detailed description
// @Tags tag-name
// @Accept json
// @Produce json
// @Security BearerAuth
// @Param id path int true "User ID"
// @Param request body CreateUserRequest true "User creation request"
// @Success 200 {object} UserResponse
// @Failure 400 {object} errors.Response{success=bool,error=errors.ErrorInfo} "Validation error"
// @Router /api/v1/users [post]
```

---

## 📚 Out-of-the-Box Features

GRAB comes with these production-ready features:

1. **Authentication**: JWT with refresh tokens → [Docs](https://vahiiiid.github.io/go-rest-api-docs/AUTHENTICATION/)
2. **RBAC**: Role-based access control → [Docs](https://vahiiiid.github.io/go-rest-api-docs/RBAC/)
3. **Migrations**: Versioned database migrations → [Docs](https://vahiiiid.github.io/go-rest-api-docs/MIGRATIONS_GUIDE/)
4. **Health Checks**: `/health`, `/live`, `/ready` → [Docs](https://vahiiiid.github.io/go-rest-api-docs/HEALTH_CHECKS/)
5. **Rate Limiting**: Token bucket algorithm → [Docs](https://vahiiiid.github.io/go-rest-api-docs/RATE_LIMITING/)
6. **Structured Logging**: JSON logs with context → [Docs](https://vahiiiid.github.io/go-rest-api-docs/LOGGING/)
7. **API Response Format**: Standardized responses → [Docs](https://vahiiiid.github.io/go-rest-api-docs/API_RESPONSE_FORMAT/)
8. **Error Handling**: Centralized error management → [Docs](https://vahiiiid.github.io/go-rest-api-docs/ERROR_HANDLING/)
9. **Graceful Shutdown**: Clean server termination → [Docs](https://vahiiiid.github.io/go-rest-api-docs/GRACEFUL_SHUTDOWN/)
10. **Swagger/OpenAPI**: Auto-generated API docs → [Docs](https://vahiiiid.github.io/go-rest-api-docs/SWAGGER/)
11. **Context Helpers**: Request context utilities → [Docs](https://vahiiiid.github.io/go-rest-api-docs/CONTEXT_HELPERS/)

---

## 🔧 Configuration

Configuration uses YAML files + environment variables:

```yaml
# configs/config.yaml (base)
app:
  name: "GRAB API"
  environment: "development"

# Override with environment-specific files:
# - configs/config.development.yaml
# - configs/config.staging.yaml
# - configs/config.production.yaml

# Override individual values with env vars:
# DATABASE_PASSWORD=secret
# JWT_SECRET=secret
```

**Environment Variable Mapping**:
- `DATABASE_PASSWORD` → `database.password`
- `JWT_SECRET` → `jwt.secret`
- `APP_ENVIRONMENT` → `app.environment`

Full config guide: https://vahiiiid.github.io/go-rest-api-docs/CONFIGURATION/

---

## 🐳 Docker Commands

```bash
# Start all services
make up

# Stop all services
make down

# View logs
make logs

# Rebuild containers
make build

# Execute command in app container
make exec

# Execute command in db container
make exec-db

# Clean restart
make down && make up
```

---

## 🧪 Pre-Commit Checklist

Before committing code:

```bash
# 1. Fix linting issues automatically
make lint-fix

# 2. Check for remaining issues
make lint

# 3. Run all tests
make test

# 4. Update Swagger docs (if API changed)
make swag

# 5. Verify everything works
make up
curl http://localhost:8080/health
```

---

## 📖 Additional Resources

- **Documentation Site**: https://vahiiiid.github.io/go-rest-api-docs/
- **Main Repository**: https://github.com/vahiiiid/go-rest-api-boilerplate
- **Issues**: https://github.com/vahiiiid/go-rest-api-boilerplate/issues
- **Discussions**: https://github.com/vahiiiid/go-rest-api-boilerplate/discussions

---

## 💡 Tips for AI Assistants

- **Always reference existing patterns**: Look at `internal/user/` for domain structure examples
- **Follow Clean Architecture**: Never skip layers (Handler → Service → Repository)
- **Use context helpers**: `contextutil.GetUserID(c)` for authenticated user info
- **Minimal comments**: Write self-documenting code, comment WHY not WHAT
- **Test thoroughly**: Maintain 85%+ test coverage
- **Check Makefile**: All development commands are in `make help`
- **Read the docs**: Comprehensive guides at https://vahiiiid.github.io/go-rest-api-docs/

---

**Last Updated**: 2025-12-10  
**GRAB Version**: v2.0.0
