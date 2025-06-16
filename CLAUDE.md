# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Structure

```
eph/
├── cmd/                    # Main applications
│   ├── eph/               # CLI binary
│   │   └── main.go
│   └── ephd/              # Daemon binary
│       └── main.go
├── internal/              # Private application code
│   ├── api/               # API definitions
│   ├── cli/               # CLI command implementations
│   ├── config/            # Configuration management
│   ├── controller/        # Environment orchestration logic
│   ├── log/               # Logging utilities
│   ├── migrate/           # Database migrations
│   ├── providers/         # Infrastructure provider implementations
│   │   └── kubernetes/    # Kubernetes provider
│   ├── server/            # HTTP API server
│   ├── state/             # Event logging (not authoritative state)
│   ├── webhook/           # Git provider webhook handlers
│   └── worker/            # Background reconciliation loops
├── pkg/                   # Public library code
│   └── version/           # Version information
├── configs/               # Configuration files
├── migrations/            # Database migration files
├── scripts/               # Build and deployment scripts
├── web/                   # Web UI assets (future)
└── docs/                  # Documentation
```

## Commands

### Build and Development
```bash
# Build both eph CLI and ephd daemon binaries
make build

# Run tests with coverage
make test

# Run the daemon server
make run-daemon

# Run the CLI
make run-cli

# Clean build artifacts
make clean

# Run linting
make lint
```

### Testing
```bash
# Run all tests
go test ./...

# Run tests for a specific package
go test ./internal/cli/...

# Run tests with verbose output
go test -v ./...

# Run a specific test
go test -run TestFunctionName ./internal/cli/

# Run tests with coverage (excludes cmd/ files)
make test-ci
```

### Code Coverage Rules

The project excludes certain files from unit test coverage metrics:

- **cmd/ directory**: Main application entry points (cmd/eph/main.go, cmd/ephd/main.go) are excluded because:
  - They contain minimal logic (mostly just calling into internal packages)
  - They are integration points that are better tested through end-to-end tests
  - Testing main functions requires complex setup and provides little value
  - Coverage tools should focus on business logic in internal/ and pkg/ directories

Coverage exclusions help maintain meaningful coverage metrics by focusing on testable business logic rather than application wiring code.

### Pre-commit Hooks
```bash
# Run pre-commit hooks manually
pre-commit run --all-files

# The project uses pre-commit hooks for:
# - Go formatting and imports
# - Module tidying
# - Comprehensive linting
# - Running short tests
# - YAML validation
# - File formatting checks
```

## CI/CD Integration

Eph orchestrates environments but doesn't build images. Your CI must:
1. Build and push container images before Eph deploys
2. Use predictable tagging conventions (e.g., `pr-123-abc1234`)
3. Eph will discover images through flexible resolution strategies

Example GitHub Actions integration:
```yaml
- name: Build and Push
  run: |
    IMAGE="ghcr.io/${{ github.repository }}:pr-${{ github.event.pull_request.number }}-${GITHUB_SHA:0:7}"
    docker build -t $IMAGE .
    docker push $IMAGE
```

## Architecture Overview

Eph is an ephemeral environment orchestrator designed for pull request workflows. It follows a reconciliation-first architecture where external systems are the source of truth.

### Core Components

**ephd (Daemon)**: The central orchestration service that:
- Manages environment lifecycle through REST API
- Handles Git webhooks for PR events
- Coordinates with infrastructure providers
- Processes background jobs via worker pools

**eph (CLI)**: A stateless API client that:
- Authenticates users via tokens
- Sends all operations to ephd
- Never directly touches infrastructure

### Reconciliation-First Architecture

Eph follows a **reconciliation-first architecture** inspired by Kubernetes controllers:
- **Level-Based Primary**: Core reconciliation loop observes external sources every 30 seconds
- **Edge-Based Optimization**: Webhooks provide immediate responsiveness but are treated as hints
- **No Internal Source of Truth**: External systems (GitHub, Kubernetes) are authoritative
- **Crash-Only Design**: Recovery from any failure is simply restarting and reconciling

### Key Design Patterns

1. **Provider Plugin System**: Infrastructure providers (Kubernetes, AWS, etc.) are isolated plugins that:
   - Run only within the trusted daemon
   - Will use gRPC for communication (MVP uses built-in provider)
   - Cannot be accessed directly by clients

2. **Reconciliation-Based State Management**:
   - Continuous reconciliation loop every 30 seconds
   - External systems (GitHub, Kubernetes) are the only sources of truth
   - PostgreSQL stores event logs only, never authoritative state
   - PostgreSQL is for event logging only, not state storage
   - System continues functioning if PostgreSQL is down
   - No job queues - reconciliation handles all work distribution

3. **Security Model**:
   - Zero-trust client architecture
   - Token-based authentication
   - Non-guessable environment URLs
   - All operations mediated through ephd

### Reconciliation Philosophy

  The core principle of Eph is that reconciliation is primary, webhooks are optimization:
  - Environments are created within 30s even without webhooks
  - Missed webhooks have no impact on correctness
  - Multiple webhook deliveries are handled idempotently
  - Recovery from any failure is automatic through reconciliation

  Git (PR labels, branches, tags) is the ONLY trigger for environments - never CI webhooks or image pushes.

### Directory Structure Intent

- `internal/cli/`: Command implementations using cobra/viper
- `internal/server/`: HTTP API server (to be implemented)
- `internal/controller/`: Environment orchestration logic (stateless)
- `internal/reconciler/`: Core reconciliation loop implementation
- `internal/informers/`: GitHub and Kubernetes informers (cache external state)
- `internal/providers/`: Infrastructure provider implementations
- `internal/state/`: Event logging (not authoritative state)
- `internal/webhook/`: Git provider webhook handlers
- `internal/worker/`: Background reconciliation loops

### Current Implementation Status

The project is in pre-MVP phase with:
- Basic CLI structure complete (version, completion, wtf commands)
- HTTP server with middleware and placeholder API endpoints
- Comprehensive CI/CD pipeline ready
- Architecture documented but core features not yet implemented
- Commands stubbed: up, down, list, logs, auth

## Development Guidelines

### Library Usage and Documentation

When using Go standard library or third-party packages:
1. **Check Current Documentation**: Always verify you're using the most up-to-date APIs and patterns
   - For Go standard library: check https://pkg.go.dev/std
   - For third-party packages: check their latest documentation
   - Pay special attention to version-specific features (e.g., Go 1.22+ routing enhancements)

2. **Version Awareness**: This project uses Go 1.24.3, which includes:
   - Enhanced HTTP routing with method-based patterns and wildcards (since Go 1.22)
   - `http.Request.PathValue()` for extracting path parameters (since Go 1.22)
   - Range over integers (since Go 1.22)
   - All the latest Go features and performance improvements
   - Always check the Go release notes for version-specific features

3. **Avoid Outdated Patterns**: Don't rely on memory of older patterns; verify current best practices

### Code Style Guidelines

- Minimize unnecessary comments - code should be self-documenting
- Use standard library constants instead of string literals (e.g., `http.MethodGet` not `"GET"`)
- Handle all errors appropriately - never ignore error returns
- Keep functions focused and testable
- Comprehensive CI/CD pipeline with pre-commit hooks
- Architecture documented but core features not yet implemented
- Commands stubbed: up, down, list, logs, auth
- Security scanning and linting configured
- Branch protection rules enforced
