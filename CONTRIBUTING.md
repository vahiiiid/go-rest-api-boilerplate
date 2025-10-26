# Contributing to Go REST API Boilerplate

Thank you for your interest in contributing! This document provides guidelines and instructions for contributing to this project.

## Code of Conduct

- Be respectful and inclusive
- Follow Go best practices and idioms
- Write clear, maintainable code
- Add tests for new features

## Getting Started

1. Fork the repository
2. Clone your fork:
```bash
git clone https://github.com/YOUR_USERNAME/go-rest-api-boilerplate.git
cd go-rest-api-boilerplate
```

3. Create a new branch:
```bash
git checkout -b feature/your-feature-name
```

4. Install dependencies:
```bash
go mod download
```

## Development Guidelines

### Code Style

- Follow standard Go formatting: use `gofmt` or `goimports`
- Run the linter before committing: `make lint`
- Keep functions small and focused (< 50 lines ideally)
- Use descriptive variable and function names
- Add comments for exported functions and complex logic

### Testing

- Write tests for all new features and bug fixes
- Maintain or improve code coverage
- Tests should be independent and reproducible
- Use table-driven tests where appropriate

Run tests:
```bash
make test
```

### Commit Messages

Follow the [Conventional Commits](https://www.conventionalcommits.org/) specification:

- `feat: add new feature`
- `fix: bug fix`
- `docs: documentation changes`
- `test: add or update tests`
- `refactor: code refactoring`
- `chore: maintenance tasks`

Examples:
```
feat(auth): add password reset functionality
fix(user): handle duplicate email correctly
docs: update API documentation
test(handler): add tests for update endpoint
```

### Database Migrations

When adding or modifying database schema:

1. **Create Migration Files** (uses timestamp versioning):

   ```bash
   make migrate-create NAME=add_user_avatar
   ```

   This creates two files with timestamp format (YYYYMMDDHHMMSS):
   - `20251026103539_add_user_avatar.up.sql`
   - `20251026103539_add_user_avatar.down.sql`

2. **Write Up Migration**: Add SQL for schema changes in `*_up.sql`

   ```sql
   -- Migration: add_user_avatar
   -- Created: 2025-10-26T10:35:39Z
   -- Description: Add avatar column to users table

   BEGIN;

   ALTER TABLE users ADD COLUMN avatar VARCHAR(255);
   CREATE INDEX idx_users_avatar ON users(avatar);

   COMMIT;
   ```

3. **Write Down Migration**: Add rollback SQL in `*_down.sql`

   ```sql
   -- Migration: add_user_avatar (rollback)
   -- Created: 2025-10-26T10:35:39Z

   BEGIN;

   DROP INDEX IF EXISTS idx_users_avatar;
   ALTER TABLE users DROP COLUMN IF EXISTS avatar;

   COMMIT;
   ```

4. **Test Both Directions**:

   ```bash
   make migrate-up          # Apply migration
   make migrate-status      # Verify version
   make migrate-down        # Rollback
   make migrate-up          # Re-apply to confirm
   ```

5. **Migration Checklist**:
   - [ ] Both up and down migrations provided
   - [ ] Migrations are idempotent (use `IF EXISTS`, `IF NOT EXISTS`)
   - [ ] Wrapped in transactions (BEGIN/COMMIT)
   - [ ] Tested locally in development
   - [ ] Down migration tested and works correctly
   - [ ] No data loss in down migration (consider backup strategy)
   - [ ] Performance tested for large datasets
   - [ ] Compatible with zero-downtime deployments

6. **Advanced Migration Commands**:

   ```bash
   make migrate-down STEPS=3              # Rollback 3 migrations
   make migrate-goto VERSION=20251026     # Jump to specific version
   make migrate-force VERSION=20251026    # Force version (recovery)
   ```

**Migration Best Practices:**

- Use timestamp versioning (automatic, prevents conflicts)
- One logical change per migration
- Never modify existing migration files
- Test on production-like data volume
- Consider adding/removing columns in separate migrations
- Use `ALTER TABLE` instead of `DROP/CREATE` when possible

### Pull Request Process

1. Update documentation if needed
2. Add tests for new features
3. Ensure all tests pass: `make test`
4. Run linter: `make lint`
5. Update README.md if adding new features
6. Submit PR with clear description

### Project Structure

When adding new features, follow the existing structure:

```
internal/
├── <domain>/           # New domain/feature
│   ├── model.go        # Data models
│   ├── repository.go   # Data access
│   ├── service.go      # Business logic
│   ├── handler.go      # HTTP handlers
│   └── dto.go          # Request/response DTOs
```

### Adding New Endpoints

1. Define the model in `model.go`
2. Add repository methods in `repository.go`
3. Implement business logic in `service.go`
4. Create HTTP handlers in `handler.go`
5. Add Swagger annotations
6. Register routes in `internal/server/router.go`
7. Write tests in `tests/`

### Database Migrations (Legacy Reference)

This project uses **[golang-migrate](https://github.com/golang-migrate/migrate)** for production-grade database migrations with timestamp versioning.

For detailed migration documentation, see the [Migrations Guide](https://vahiiiid.github.io/go-rest-api-docs/MIGRATIONS_GUIDE/).

### Documentation

Documentation is maintained in a separate repository: [go-rest-api-docs](https://github.com/vahiiiid/go-rest-api-docs)

#### Contributing to Documentation

To contribute to the project documentation:

1. Visit the [documentation repository](https://github.com/vahiiiid/go-rest-api-docs)
2. Follow the contributing guidelines in that repository
3. Submit pull requests for documentation improvements

The documentation site is available at: https://vahiiiid.github.io/go-rest-api-docs/

#### Updating README

The main `README.md` in this repository should be kept concise and focused on:
- Quick start instructions
- Basic project overview
- Links to the full documentation site

When making significant changes to the codebase:
- Update the README.md if it affects quick start or basic usage
- Consider updating the full documentation in the docs repository

## Feature Requests

- Check existing issues first
- Open an issue to discuss major changes
- Provide use cases and examples

## Bug Reports

Include:
- Go version
- Steps to reproduce
- Expected vs actual behavior
- Error messages/logs
- Environment details

## Questions?

- Open a GitHub issue
- Check existing discussions
- Review the README.md

Thank you for contributing! 🎉

