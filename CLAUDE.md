# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

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
```

## Architecture Overview

Eph is an ephemeral environment orchestrator designed for pull request workflows. It follows an event-driven architecture with PostgreSQL as the single source of truth.

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

### Key Design Patterns

1. **Provider Plugin System**: Infrastructure providers (Kubernetes, AWS, etc.) are isolated plugins that:
   - Run only within the trusted daemon
   - Will use gRPC for communication (future)
   - Cannot be accessed directly by clients

2. **Event-Driven State Management**:
   - All state changes go through PostgreSQL transactions
   - LISTEN/NOTIFY for real-time updates
   - Outbox pattern for reliable event delivery

3. **Security Model**:
   - Zero-trust client architecture
   - Token-based authentication
   - Non-guessable environment URLs
   - All operations mediated through ephd

### Directory Structure Intent

- `internal/cli/`: Command implementations using cobra/viper
- `internal/server/`: HTTP API server (to be implemented)
- `internal/controller/`: Environment lifecycle orchestration
- `internal/providers/`: Infrastructure provider implementations
- `internal/state/`: Database models and migrations
- `internal/worker/`: Background job processing
- `internal/webhook/`: Git provider webhook handlers

### Current Implementation Status

The project is in pre-MVP phase with:
- Basic CLI structure complete (version, completion, wtf commands)
- Comprehensive CI/CD pipeline ready
- Architecture documented but core features not yet implemented
- Commands stubbed: up, down, list, logs, auth
