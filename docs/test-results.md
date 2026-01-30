---
layout: default
title: Test Results
description: Code coverage and mutation testing results for Yougopher.
---

## Overview

This page tracks test quality metrics for Yougopher.

## Code Coverage

| Package | Coverage | Status |
|---------|----------|--------|
| `youtube/core` | - | - |
| `youtube/auth` | - | - |
| `youtube/streaming` | - | - |
| `youtube/data` | - | - |
| `youtube/analytics` | - | - |

**Last updated:** -

### Running Coverage Locally

```bash
# Generate coverage report
go test -coverprofile=coverage.out ./...

# View coverage summary
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

## Mutation Testing

Mutation testing measures test effectiveness by introducing bugs (mutants) and checking if tests catch them.

| Metric | Value |
|--------|-------|
| Mutation Score | - |
| Total Mutants | - |
| Killed | - |
| Survived | - |
| Timeout | - |

**Last updated:** -

### Mutation Levels

| Level | Mutators | Use Case |
|-------|----------|----------|
| `lite` | Basic operators | Quick CI checks |
| `standard` | Common patterns | Default for PRs |
| `thorough` | Extended set | Pre-release validation |
| `mutilated` | All mutators | Comprehensive analysis |

### Running Mutation Tests Locally

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

### Viewing Reports

- **JSON reports:** Download from GitHub Actions artifacts
- **HTML viewer:** Open [mutation-report.html](../mutation-report.html) and load the JSON

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
| Coverage | Always | Generate coverage summary |
| Mutation (diff) | Push/PR | Test changed files only |
| Mutation (full) | Manual | Full codebase analysis |

### Triggering Full Mutation Testing

1. Go to [Actions](https://github.com/Its-donkey/yougopher/actions)
2. Select "Tests" workflow
3. Click "Run workflow"
4. Check "Run full mutation testing"
5. Click "Run workflow"

The full mutation test creates a PR with the HTML report when complete.
