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

### Database Migrations

This project uses GORM AutoMigrate. For production, consider using a migration tool like:
- [golang-migrate](https://github.com/golang-migrate/migrate)
- [goose](https://github.com/pressly/goose)

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

