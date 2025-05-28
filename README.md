# Eph 🌊

[![CI Status](https://github.com/ephlabs/eph/actions/workflows/ci.yml/badge.svg)](https://github.com/ephlabs/eph/actions/workflows/ci.yml)
[![Go Version](https://img.shields.io/badge/Go-1.24.3-00ADD8?style=flat&logo=go)](https://go.dev)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Pre-commit](https://img.shields.io/badge/pre--commit-enabled-brightgreen?logo=pre-commit&logoColor=white)](https://github.com/pre-commit/pre-commit)
[![gRPC](https://img.shields.io/badge/gRPC-Provider%20Plugins-00ADD8?style=flat&logo=grpc)](https://grpc.io)
[![PostgreSQL](https://img.shields.io/badge/PostgreSQL-16%2B-316192?style=flat&logo=postgresql&logoColor=white)](https://www.postgresql.org)

> *Ephemeral environments that make you say "What the eph?"*

Eph is an open-source ephemeral environment controller that automatically creates, manages, and destroys temporary preview environments for pull requests. Get a full, isolated environment with every PR - complete with its own URL, database, and resources.

## Vision

```
Developer: *opens PR*
Developer: *adds 'preview' label*
Eph: "Here's your environment: https://myapp-gentle-stream-42.preview.company.com"
Developer: "What the eph? That was fast!"
```

No more fighting over staging servers. No more "works on my machine." Just push code, get a preview.

## How It Works

```mermaid
graph LR
    A[Open PR] --> B[CI Builds Images]
    B --> C[Add 'preview' label]
    C --> D[Eph creates environment]
    D --> E[Access via unique URL]
    E --> F[Auto-cleanup on merge]
```

**Key principles:**
- Eph orchestrates environments, it doesn't build images (that's CI's job)
- Works with your existing infrastructure (Kubernetes, Docker, cloud providers)
- Extensible via gRPC-based provider plugins
- Secure by default with non-guessable URLs

## Architecture

```mermaid
graph TB
    subgraph "Event Sources"
        GH[GitHub Webhooks]
        GL[GitLab Webhooks]
        API[Direct API Calls]
    end

    subgraph "Core Engine"
        Gateway[API Gateway]
        Events[Event Processor]
        Orchestrator[Environment Orchestrator]
        State[State Store<br/>PostgreSQL]
    end

    subgraph "Provider Plugins"
        K8s[Kubernetes Provider]
        Docker[Docker Provider]
        Cloud[Cloud Providers<br/>ECS, Cloud Run, etc.]
    end

    GH --> Gateway
    GL --> Gateway
    API --> Gateway
    Gateway --> Events
    Events --> Orchestrator
    Orchestrator <--> State
    Orchestrator <--> K8s
    Orchestrator <--> Docker
    Orchestrator <--> Cloud
```

## Configuration (Target State)

Projects will configure Eph via `eph.yaml`:

```yaml
version: "1.0"
name: my-app

triggers:
  - type: pr_label
    labels: ["preview"]
    wait_for_checks: ["build"]

environment:
  name_template: "{project}-{words}-{number}"  # my-app-gentle-stream-42
  ttl: 72h
  idle_timeout: 4h

kubernetes:
  manifests:
    - ./k8s/base
    - ./k8s/overlays/preview
  images:
    - name: api
      newTag: "pr-{pr_number}"

database:
  enabled: true
  instances:
    - name: postgres
      version: "15"
      template:
        strategy: seed
        seed:
          scripts: ["./db/schema.sql"]
```

## Current Status

**🚧 Pre-MVP** - Architecture planning complete, implementation starting

## Roadmap

### MVP (Current Focus)
- [ ] Core event processing engine
- [ ] Kubernetes provider (built-in)
- [ ] GitHub webhook integration
- [ ] PostgreSQL state management
- [ ] Basic CLI (`eph list`, `eph logs`, `eph down`)
- [ ] Minimal web dashboard

### Phase 1: Production Ready
- [ ] gRPC provider plugin system
- [ ] GitLab support
- [ ] OAuth/OIDC authentication
- [ ] Helm chart support
- [ ] Prometheus metrics

### Phase 2: Multi-Provider
- [ ] Docker Compose provider
- [ ] AWS ECS provider
- [ ] Google Cloud Run provider
- [ ] Provider plugin SDK

### Future Vision
- [ ] Header-based routing (Signadot-style)
- [ ] Intelligent resource optimization
- [ ] Multi-cluster support
- [ ] IDE integrations

## Technical Decisions

- **PostgreSQL** for state (not etcd/Redis) - ACID guarantees, JSON support, single dependency
- **gRPC** for provider plugins - Language agnostic, streaming support, process isolation
- **Event-driven** architecture - Scalable, resilient, auditable
- **Kubernetes-first** - Most complex target, proves the provider abstraction

## Project Structure

```
eph/
├── cmd/
│   ├── eph/          # CLI commands
│   └── ephd/         # Server daemon
├── pkg/
│   ├── api/          # HTTP API handlers
│   ├── controller/   # Environment orchestration
│   ├── providers/    # Provider implementations
│   ├── state/        # PostgreSQL state management
│   └── webhook/      # Git webhook handlers
├── web/              # React dashboard
└── docs/             # Documentation
```

## Development & CI

### Prerequisites
- Go 1.24.3+
- PostgreSQL 16+
- golangci-lint (for linting)
- pre-commit (for local development hooks)

### Setting Up Local Development
1. Clone the repository
2. Run the CI setup script:
   ```bash
   ./scripts/setup-ci.sh
   ```
   This installs required tools and sets up pre-commit hooks.

### CI/CD Pipeline
Our continuous integration runs on every push to `main` and all pull requests:

- **Linting**: golangci-lint with comprehensive rules
- **Testing**: Unit tests with PostgreSQL integration
- **Building**: Cross-platform binaries for all major platforms
- **Security Scanning**: Trivy vulnerability scanner and gosec
- **Integration Tests**: End-to-end testing with ephd daemon

#### Branch Protection
The `main` branch is protected and requires:
- All status checks to pass
- One approved review
- Branches to be up to date before merging

#### Pre-commit Hooks
Local pre-commit hooks run automatically before each commit:
- `go fmt` and `go imports`
- `go mod tidy`
- `go vet`
- `golangci-lint` (fast mode)
- Basic YAML/text validation

To run manually: `pre-commit run --all-files`

### Commands
```bash
# Run tests
go test ./...

# Run linting
golangci-lint run

# Build binaries
go build -o bin/eph ./cmd/eph
go build -o bin/ephd ./cmd/ephd
```

## Getting Started

See [docs/architecture-plan.md](docs/architecture-plan.md) for the complete architectural vision.

---

*When in doubt, just eph it!* 🚀
