# Contributing to Yougopher

Thank you for your interest in contributing to Yougopher!

## Development Setup

1. Fork and clone the repository
2. Ensure you have Go 1.23+ installed
3. Run `go mod download` to install dependencies
4. Run `go test ./...` to verify everything works

## Branching Strategy

- `test` - Development branch, PRs target here
- `main` - Production branch, promoted from test via release workflow

## Pull Request Process

1. Create a feature branch from `test`
2. Make your changes
3. Ensure tests pass: `go test -race ./...`
4. Ensure linting passes: `golangci-lint run`
5. Ensure 90% test coverage for new code
6. Submit PR to `test` branch

## Code Standards

### Testing
- Table-driven tests preferred
- Use `httptest.Server` for HTTP mocking
- Target 90% coverage

### Mutation Testing
- CI runs mutation testing with [mutagoph](https://github.com/Its-donkey/mutagoph)
- Uses diff-based testing: only mutates changed Go files
- Mutation reports are merged incrementally across PRs
- Local testing: `go install github.com/its-donkey/mutagoph/cmd/mutagoph@latest`
- Run locally: `mutagoph run -mutations standard -target ./...`

### Documentation
- All exported types, functions, and methods must have GoDoc comments
- Package-level documentation in `doc.go`

### Go Best Practices
- `context.Context` is always the first parameter
- Use functional options for configuration
- Errors should be wrapped with context

## Commit Messages

Use clear, descriptive commit messages:

```
feat: add SuperChat event handling
fix: correct quota calculation for search.list
docs: update README with moderation examples
test: add coverage for token refresh
```

## Questions?

Open an issue for discussion before starting large changes.
