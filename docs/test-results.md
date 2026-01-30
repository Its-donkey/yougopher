---
layout: default
title: Test Results
description: Code coverage and mutation testing results for Yougopher.
---

## Interactive Results

View the full interactive test results: **[Test Results Viewer](test-results.html)**

The viewer displays:
- Code coverage by package
- Mutation testing score and details
- Per-file mutation results with filtering

## Running Tests Locally

### Coverage

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Mutation Testing

```bash
# Install mutagoph
go install github.com/its-donkey/mutagoph/cmd/mutagoph@latest

# Run diff-based (changed files only)
mutagoph run -target ./... --diff-base HEAD~1

# Run full mutation testing
mutagoph run -target ./... --dynamic-level standard

# Generate HTML report
mutagoph run -target ./... --output html --output-file mutation-report.html
```

## Mutation Levels

| Level | Mutators | Use Case |
|-------|----------|----------|
| `lite` | Basic operators | Quick CI checks |
| `standard` | Common patterns | Default for PRs |
| `thorough` | Extended set | Pre-release validation |
| `mutilated` | All mutators | Comprehensive analysis |

## CI Integration

Tests run automatically on:
- Push to `main` or `test` branches
- Pull requests to `main` or `test`

### Workflow Jobs

| Job | Trigger | Description |
|-----|---------|-------------|
| Test | Always | Run tests with race detector |
| Lint | Always | golangci-lint checks |
| Build | Always | Verify compilation |
| Coverage | Always | Generate coverage report |
| Mutation (diff) | Push/PR | Test changed files only |
| Mutation (full) | Manual | Full codebase analysis |
| Update Results | Push to test | Merge and commit results |

### Triggering Full Mutation Testing

1. Go to [Actions](https://github.com/Its-donkey/yougopher/actions)
2. Select "Tests" workflow
3. Click "Run workflow"
4. Check "Run full mutation testing"
5. Click "Run workflow"
