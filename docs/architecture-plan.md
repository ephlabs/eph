# Eph - Open Source Ephemeral Environment Controller Architecture Plan
*Ephemeral environments that make you say "What the eph?"*

## Executive Summary

Eph is an open-source ephemeral environment controller designed to automatically create, manage, and destroy temporary development and testing environments. The system responds to pull request events from Git providers, provisions isolated environments with their own compute resources and databases, and provides developers with unique URLs to preview their changes before merging.

The architecture emphasizes extensibility through a provider-agnostic plugin system, allowing deployment to diverse infrastructure backends including Kubernetes, Docker Compose, AWS ECS, and more. The design prioritizes developer experience with a simple CLI, automatic resource optimization, and flexible routing strategies.

Eph focuses exclusively on cloud-based ephemeral environments tied to pull requests, working with existing infrastructure rather than provisioning it. It orchestrates the deployment of pre-built container images - it is not a CI/CD system and does not build images. Local development workflows remain the domain of existing tools like Docker Compose.

## Non-Goals and Scope Boundaries

To maintain focus and avoid duplicating existing tools, Eph explicitly does not aim to:

**Replace local development tools**: Eph is not a substitute for Docker Compose, Air, or other local development workflows. Developers should continue using their preferred local tools.

**Build container images**: Eph is an orchestrator, not a CI/CD pipeline. It expects container images to already exist in a registry, tagged according to conventions or with explicit references provided by your CI system. Image building should be handled by your existing CI system (GitHub Actions, GitLab CI, Jenkins, etc.). Eph provides flexible image discovery to work with any CI system's tagging approach.

**Manage production environments**: Eph is specifically for ephemeral preview/test environments. Production deployments should use proper CD tools like ArgoCD or Flux.

**Provide a PaaS**: Eph manages temporary environments, not long-running applications. It's a developer productivity tool, not an application platform.

**Abstract away all platform differences**: While Eph provides a common interface, it embraces platform-specific features rather than lowest-common-denominator abstractions.

## Core Concepts and Terminology

An **ephemeral environment** is a temporary, isolated instance of an application created automatically for a specific purpose, typically to preview changes in a pull request. These environments include all necessary components: compute resources, databases, networking, and dependencies.

A **provider** in Eph terminology is a plugin that knows how to create and manage environments on a specific infrastructure platform. The Kubernetes provider, for example, creates namespaces and deployments in a Kubernetes cluster, while a Docker Compose provider might spin up containers on a single host or Docker Swarm cluster.

The **environment lifecycle** encompasses creation (triggered by a PR event), active use (developers testing features), idle periods (automatic scaling down), and eventual destruction (when the PR is merged or after a timeout).

## Reconciliation Philosophy

Eph embraces a **reconciliation-first architecture** inspired by Kubernetes controllers and proven in production systems. Rather than relying solely on event-driven updates, Eph continuously reconciles the desired state (pull requests with preview labels) against the actual state (running environments).

**Level-Based Primary**: The core reconciliation loop observes external sources of truth every 30 seconds, ensuring eventual consistency even when webhooks are missed, systems are partitioned, or components crash.

**Edge-Based Optimization**: Webhooks provide immediate responsiveness but are treated as hints to trigger reconciliation sooner, not as the primary source of truth.

**No Internal Source of Truth**: Eph treats external systems as authoritative:
- Git providers (GitHub/GitLab) define which environments should exist
- Infrastructure providers (Kubernetes/Docker) report which environments actually exist
- PostgreSQL serves only as an event log and cache, never as primary state

This philosophy ensures Eph can recover from any failure scenario by simply restarting and reconciling current state.

### The Power of Reconciliation

This reconciliation-first architecture provides remarkable simplicity and reliability:

- **No Split Brain**: Can't have inconsistent state when you don't store state
- **No Lost Events**: Can't lose events when you don't rely on events
- **No Race Conditions**: Can't have races when every operation is idempotent
- **No Complex Recovery**: Startup is recovery, recovery is normal operation
- **No Cascading Failures**: Each reconciliation is independent

By embracing eventual consistency and external state, Eph achieves the reliability of distributed systems like Kubernetes while maintaining the simplicity of a single reconciliation loop.

## Example Use Cases

Eph excels at scenarios requiring isolated, temporary environments:

**Frontend Preview Deployments**: Every PR gets a unique URL for designers and product managers to review changes without setting up development environments.

**API Testing**: QA engineers test API changes without affecting shared staging environments, eliminating "works on my machine" issues.

**Database Migration Testing**: Safely test schema changes with production-like data without risk to shared databases.

**Multi-Service Integration**: Test microservice changes with the full application stack, ensuring compatibility before merge.

**Customer Demos**: Spin up isolated environments for sales demos with specific feature flags or customizations.

**Training and Workshops**: Create identical environments for each workshop participant without manual setup.

**Integration Testing with Service Dependencies**: When a PR receives the "preview" label, Eph creates an environment for integration testing. The reconciliation loop waits for CI to build and push the required images, then deploys them. This enables testing service interactions in isolation - for example, deploying a new payment service version alongside stable versions of other services to verify API compatibility before merge.

Eph is not designed for:

**Local development workflows**: Use Docker Compose or similar tools for rapid local iteration.

**Production deployments**: Use proper CD tools like ArgoCD, Flux, or Spinnaker.

**Long-running test environments**: Use dedicated staging clusters for persistent testing needs.

**CI/CD build environments**: Use GitHub Actions, Jenkins, or GitLab CI for build and test automation.

## System Architecture Overview

The Eph system follows an event-driven architecture with clear separation of concerns between components. At its heart, the system consists of an API Gateway that receives webhook events from Git providers, an Event Processor that validates and queues these events, and an Environment Orchestrator that coordinates the actual provisioning of resources.

The following diagram illustrates the architecture of Eph, highlighting the separation between the 'Untrusted Environment' and the 'Trusted Environment - ephd Daemon'. The 'Untrusted Environment' includes components like the CLI and Web UI, which interact with the system via HTTPS but do not have direct access to infrastructure or sensitive resources. The 'Trusted Environment - ephd Daemon' encompasses the core engine, provider interface, and supporting services, which handle all privileged operations securely within controlled boundaries.

```mermaid
graph TB
    subgraph "Untrusted Environment"
        CLI[eph CLI]
        WebUI[Web Browser UI]
    end

    subgraph "Event Sources"
        GH[GitHub Webhooks]
        GL[GitLab Webhooks]
        API[Direct API Calls]
    end

    subgraph "Trusted Environment - ephd Daemon"
        RestAPI[REST API Server]

        subgraph "Core Engine"
            Gateway[API Gateway]
            Events[Event Processor]
            Orchestrator[Environment Orchestrator]
            State[State Store<br/>PostgreSQL]

            Gateway --> Events
            Events --> Orchestrator
            Orchestrator <--> State
        end

        subgraph "Provider Interface"
            ProviderAPI[Provider gRPC API]

            subgraph "Provider Processes"
                K8sProvider[Kubernetes Provider<br/>Separate Process]
                DockerProvider[Docker Provider<br/>Separate Process]
                CloudProviders[Cloud Providers<br/>ECS, Cloud Run, etc.]
            end

            Orchestrator <--> ProviderAPI
            ProviderAPI <--> K8sProvider
            ProviderAPI <--> DockerProvider
            ProviderAPI <--> CloudProviders
        end

        subgraph "Supporting Services"
            DNS[DNS Service]
            Auth[Auth Service]
            Metrics[Metrics Collector]
            Secrets[Secrets Manager]
        end

        RestAPI --> Gateway
        RestAPI --> Orchestrator
    end

    CLI -- "HTTPS" --> RestAPI
    WebUI -- "HTTPS" --> RestAPI
    GH --> Gateway
    GL --> Gateway
    API --> Gateway

    Orchestrator --> DNS
    Orchestrator --> Auth
    Orchestrator --> Metrics
    Orchestrator --> Secrets
```

**Important**: The eph CLI is a pure API client with zero direct access to infrastructure, databases, or providers. All operations flow through the ephd REST API.

## Go Project Structure and Code Organization

Eph follows idiomatic Go project structure patterns established by the Go community and used by popular projects like Kubernetes, Hugo, and Docker CLI.

### Directory Layout

```
eph/
├── go.mod                 # Module definition: github.com/ephlabs/eph
├── go.sum                 # Dependency checksums
├── Makefile              # Build automation
├── LICENSE               # Apache 2.0 license
├── README.md             # Project documentation
├── cmd/                  # Binary entry points (thin wrappers only)
│   ├── eph/
│   │   └── main.go      # CLI binary: import internal/cli; cli.Execute()
│   └── ephd/
│       └── main.go      # Daemon binary: import internal/server; server.Run()
├── internal/            # Private application code (compiler enforced)
│   ├── server/          # HTTP server implementation
│   │   ├── server.go    # Core server logic
│   │   └── README.md    # Package documentation
│   ├── cli/             # CLI command implementation
│   │   ├── root.go      # Root Cobra command
│   │   ├── auth.go      # Authentication commands
│   │   ├── completion.go # Shell completion support
│   │   ├── down.go      # Environment destruction
│   │   ├── list.go      # List environments
│   │   ├── logs.go      # Stream logs
│   │   ├── up.go        # Create environment
│   │   ├── version.go   # Version command
│   │   ├── wtf.go       # Diagnostic command
│   │   └── *_test.go    # CLI tests
│   ├── server/          # HTTP server implementation
│   ├── cli/             # CLI command implementation
│   ├── api/             # API client/server shared code
│   ├── config/          # Configuration parsing and validation
│   ├── controller/      # Environment business logic and orchestration
│   ├── reconciler/      # Core reconciliation loop engine
│   ├── informers/       # GitHub and Kubernetes informers (cache external state)
│   ├── log/             # Logging utilities
│   ├── providers/       # Provider implementations
│   │   ├── interface.go # Provider interface
│   │   └── kubernetes/  # Kubernetes provider
│   ├── state/           # Event logging (not authoritative state)
│   ├── webhook/         # Git webhook handlers
│   └── worker/          # Background reconciliation loops
├── pkg/                 # Exportable packages (use sparingly)
│   └── version/         # Version information
│       ├── version.go   # Version constants and variables
│       └── README.md    # Package documentation
├── bin/                 # Build output directory (gitignored)
├── configs/             # Example configurations
├── docs/                # Project documentation
│   └── architecture-plan.md
├── migrations/          # Database migration files
├── scripts/             # Build and CI scripts
│   └── setup-ci.sh
└── web/                 # Web dashboard (React)
```

### Package Responsibilities

**`cmd/` packages**:
- Contain only minimal main functions
- Import and invoke code from `internal/` packages
- Handle command-line argument parsing specific to each binary
- Should not contain business logic or be directly testable

**`internal/` packages**:
- Contains all application business logic
- Cannot be imported by external projects (compiler enforced)
- Fully testable with unit and integration tests
- Shared between multiple binaries in the same project

**Key `internal/` package distinctions**:
- **`controller/`**: Environment business logic and orchestration (what should happen)
  - Contains domain-specific logic for managing environments
  - Makes decisions about desired state vs actual state
  - Orchestrates operations across multiple providers/resources
  - Stateless - no persistent state, only business rules

- **`reconciler/`**: Core reconciliation loop engine (how reconciliation works)
  - Implements the generic reconciliation control loop (every 30s)
  - Handles reconciliation mechanics, timing, and error handling
  - Coordinates between informers and controllers
  - Reusable across different resource types

- **`informers/`**: External state caching and watching
  - Cache current state from GitHub (PRs, labels) and Kubernetes (deployments, services)
  - Provide efficient access to external state without constant API calls
  - Detect changes and trigger reconciliation events

**`pkg/` packages**:
- Exportable packages safe for external use
- Use sparingly - most code should be in `internal/`
- Must maintain backward compatibility
- Should be genuinely reusable across projects

### Implementation Examples

**Minimal cmd/ephd/main.go**:
```go
package main

import (
    "log"
    "github.com/ephlabs/eph/internal/server"
)

func main() {
    if err := server.Run(); err != nil {
        log.Fatal(err)
    }
}
```

**Real logic in internal/server/server.go**:
```go
package server

import (
    "context"
    "net/http"
    "github.com/ephlabs/eph/internal/config"
    "github.com/ephlabs/eph/internal/controller"
)

type Server struct {
    config     *config.Config
    controller *controller.Controller
    httpServer *http.Server
}

func Run() error {
    cfg, err := config.Load()
    if err != nil {
        return err
    }

    s := &Server{
        config:     cfg,
        controller: controller.New(cfg),
    }

    return s.Start()
}

func (s *Server) Start() error {
    mux := s.setupRoutes()
    s.httpServer = &http.Server{
        Addr:    s.config.ServerAddr,
        Handler: mux,
    }

    return s.httpServer.ListenAndServe()
}
```

### Testing Strategy

The structure enables comprehensive testing:

```bash
# Test all business logic
go test ./internal/...

# Test specific packages
go test ./internal/server/
go test ./internal/cli/

# Build binaries
go build ./cmd/eph
go build ./cmd/ephd

# Integration tests
go test ./internal/server/ -tags=integration
```

### Benefits of This Structure

1. **Separation of Concerns**: Clear boundaries between binary entry points and business logic
2. **Testability**: All business logic is in testable packages
3. **Reusability**: `internal/` packages can be shared between `eph` and `ephd`
4. **Import Safety**: `internal/` prevents external dependencies on private code
5. **Standard Tooling**: Works seamlessly with Go build tools and IDEs
6. **Community Familiarity**: Follows patterns used by major Go projects

### Anti-Patterns to Avoid

- **Business logic in `cmd/`**: Keep main functions minimal
- **Overusing `pkg/`**: Most code should be internal unless truly reusable
- **Deep package nesting**: Prefer flat structure within `internal/`
- **Circular dependencies**: Use interfaces to break cycles between packages

This structure provides a solid foundation for implementing the Eph system while maintaining Go best practices and enabling effective testing and maintenance.

### Reconciliation-Based Processing Flow

Eph follows a continuous reconciliation pattern that ensures reliability through simplicity:

```mermaid
graph TB
    subgraph "External Sources of Truth"
        GitHub[GitHub API<br/>PR List + Labels]
        K8s[Kubernetes API<br/>Namespaces + Deployments]
    end

    subgraph "Eph Core - Informer Pattern"
        GHInformer[GitHub Informer<br/>Polls every 30s]
        K8sInformer[K8s Informer<br/>Watches continuously]
        Cache[In-Memory Cache<br/>Current state view]
    end

    subgraph "Reconciliation Engine"
        Reconciler[Reconciler<br/>Compares desired vs actual]
        Diff[Diff Calculator]
        Actions[Action Executor<br/>Create/Update/Delete]
    end

    subgraph "Optional Components"
        Webhooks[Webhook Handler<br/>Optional: Pokes reconciler<br/>No state, just a signal]
        EventLog[PostgreSQL<br/>Event log only]
    end

    GitHub --> GHInformer
    K8s --> K8sInformer
    GHInformer --> Cache
    K8sInformer --> Cache
    Cache --> Reconciler
    Reconciler --> Diff
    Diff --> Actions
    Actions --> K8s
    Actions -.-> EventLog
    Webhooks -.-> Reconciler
```

The reconciliation loop runs continuously:

1. **Observe Desired State**: Query GitHub for PRs with the "preview" label
2. **Observe Actual State**: Query Kubernetes for existing environments
3. **Compute Differences**: Determine what needs to be created, updated, or deleted
4. **Apply Changes**: Execute changes idempotently - safe to retry at any point
5. **Log Events**: Record actions for audit/analytics (non-critical path)

This approach handles all edge cases automatically:
- Rapid label changes (add/remove/add)
- Missed webhooks or network partitions
- Component crashes at any point
- Multiple ephd instances running

## REST API Design

### API-First Architecture

Eph follows an API-first design where the ephd daemon exposes a comprehensive REST API that serves as the **single source of truth** for all operations.

#### Core API Endpoints

**Environment Management**:
- `POST /api/v1/environments` - Create environment
- `GET /api/v1/environments` - List environments
- `GET /api/v1/environments/{id}` - Get environment details
- `DELETE /api/v1/environments/{id}` - Destroy environment
- `PUT /api/v1/environments/{id}/scale` - Scale environment

**Monitoring and Logs**:
- `GET /api/v1/environments/{id}/status` - Real-time status
- `GET /api/v1/environments/{id}/logs` - Stream logs
- `GET /api/v1/environments/{id}/metrics` - Resource metrics

**Configuration and Auth**:
- `POST /api/v1/auth/login` - Authentication
- `GET /api/v1/config/validate` - Validate eph.yaml
- `GET /api/v1/providers/capabilities` - Available providers

#### API Security
- All endpoints require authentication
- Rate limiting per user/token
- Request/response logging for audit
- Input validation and sanitization
- CORS policies for web dashboard access

### Provider Plugin Architecture

The provider plugin system is the core of Eph's extensibility. Each provider runs as a separate process and communicates with the core system via gRPC. This design offers several advantages over traditional in-process plugins: providers can be written in any language that supports gRPC, they can't crash the core system, and they can be developed and versioned independently.

**Security Note**: Provider plugins run exclusively within the ephd daemon process or as trusted gRPC services. The eph CLI never directly communicates with providers - all provider operations are mediated by the ephd API layer.

**Stateless Provider Design**: Providers in Eph are stateless - they report current state and execute operations but never store state. This aligns perfectly with the reconciliation model where the reconciler asks "what exists?" and "please create/update/delete this" without the provider needing to track anything.

The gRPC interface defines a standard set of operations that all providers must implement:

```protobuf
syntax = "proto3";
package eph.provider.v1;

service Provider {
  // Lifecycle Management
  rpc CreateEnvironment(CreateEnvironmentRequest) returns (stream OperationUpdate);
  rpc DestroyEnvironment(DestroyEnvironmentRequest) returns (stream OperationUpdate);
  rpc ScaleEnvironment(ScaleEnvironmentRequest) returns (stream OperationUpdate);

  // Status & Monitoring
  rpc GetEnvironmentStatus(GetEnvironmentStatusRequest) returns (EnvironmentStatus);
  rpc StreamLogs(StreamLogsRequest) returns (stream LogEntry);
  rpc GetMetrics(GetMetricsRequest) returns (EnvironmentMetrics);

  // Provider Capabilities
  rpc GetCapabilities(Empty) returns (ProviderCapabilities);
  rpc ValidateConfiguration(ValidateConfigurationRequest) returns (ValidationResult);
}

message ProviderCapabilities {
  bool supports_scale_to_zero = 1;
  bool supports_custom_domains = 2;
  bool supports_persistent_storage = 3;
  bool supports_database_provisioning = 4;
  repeated string supported_databases = 5;
  map<string, ConfigSchema> configuration_schema = 6;
  ResourceLimits resource_limits = 7;
}

message OperationUpdate {
  string operation_id = 1;
  OperationStatus status = 2;
  string message = 3;
  int32 progress_percent = 4;
  map<string, string> outputs = 5;
  repeated ResourceInfo created_resources = 6;
}
```

Each provider declares its capabilities when queried, allowing the core system to understand what features are available. For example, a Kubernetes provider might support scale-to-zero and custom domains, while a simpler Docker Compose provider might only support basic environment creation.

### Image Resolution Strategy

Since Eph orchestrates environments but doesn't build images, it needs a flexible strategy to discover which container images to deploy for any given Git ref (PR, branch, or tag). The system supports multiple resolution strategies that can be combined for maximum compatibility with different CI/CD systems.

#### Image Discovery Methods

**1. Convention-Based Tags**: CI systems push images with predictable tag patterns
```yaml
# eph.yaml
environment:
  images:
    - name: api
      repository: ghcr.io/myorg/myapp
      tag_template: "{ref_type}-{ref_name}-{commit_sha:0:7}"
      # Results in: pr-123-abc1234, branch-main-def5678, tag-v1.0.0-ghi9012
```

**2. Git Annotations**: CI systems communicate image locations via Git
```bash
# CI writes image location to git notes
git notes add -m "eph-images: api=ghcr.io/myorg/api:pr-123-abc1234"
git push origin refs/notes/*
```

**3. Status Check Integration**: Parse image refs from CI check outputs
```yaml
triggers:
  - type: pr_label
    labels: ["preview"]
    wait_for_checks: ["docker-build"]  # Must complete first
```

**4. Registry Scanning**: Dynamically discover images by pattern matching
```yaml
images:
  - name: api
    repository: ghcr.io/myorg/api
    tag_pattern: "pr-{pr_number}-*"  # Find newest matching
    max_age: 7d                      # Don't use stale images
```

#### Resolution Priority

Eph attempts each strategy in order until a valid image is found:

1. **Explicit Git annotations** (most specific)
2. **CI check outputs** (when integrated with CI)
3. **Convention-based templates** (predictable patterns)
4. **Registry scanning** (pattern matching)
5. **Fallback tags** (last resort)

This multi-strategy approach ensures Eph works with any CI system without requiring specific integrations.

**Important**: Image resolution happens during reconciliation, not as a trigger. The reconciler:
1. Determines an environment should exist (based on Git)
2. Attempts to find the appropriate image
3. Creates/updates the environment if image is found
4. Marks environment as "WaitingForImage" if not found
5. Retries on the next reconciliation cycle

Images are dependencies, not triggers. Their existence never causes environment creation.

### Informer Pattern and State Management

Eph uses the Informer pattern popularized by Kubernetes to maintain an eventually-consistent view of external state without a central database.

```go
// Conceptual implementation
type GitHubInformer struct {
    client    *github.Client
    cache     cache.Store        // Thread-safe in-memory store
    resync    time.Duration      // How often to poll
}

type KubernetesInformer struct {
    informer  cache.SharedIndexInformer  // K8s native informer
    lister    listers.NamespaceLister    // Cached reads
}

type EnvironmentReconciler struct {
    github     *GitHubInformer
    kubernetes *KubernetesInformer
    providers  map[string]provider.Provider
}
```

Benefits of this approach:
- **No Split Brain**: External systems are always authoritative
- **Fast Reads**: All queries served from in-memory cache
- **Crash Recovery**: Simply restart and re-sync from external state
- **Natural Sharding**: Multiple ephd instances can reconcile different repos

The Informer pattern is critical to Eph's reliability:

**GitHubInformer**: Polls GitHub every 30 seconds for PRs/branches/tags
- Caches the list of all PRs with 'preview' label
- Provides consistent view even during GitHub outages
- Never misses changes due to continuous polling

**KubernetesInformer**: Uses native Kubernetes watch API
- Maintains real-time cache of all Eph-managed resources
- Receives immediate updates when resources change
- Resync periodically to catch any missed events

**No Persistent State**: The informers maintain only in-memory caches
- On restart, caches are rebuilt from external sources
- No risk of stale or inconsistent state
- Natural crash recovery through re-synchronization

### State Management Philosophy

Eph follows a "stateless controller" pattern where PostgreSQL serves only non-critical roles:

**Event Log** (Nice to Have):
- Audit trail of all actions taken
- Analytics on environment usage
- Debugging failed operations
- Can be lost without affecting correctness

**Coordination** (For Multi-Instance):
- Leader election ensures only one reconciler runs at a time
- NOT for work distribution - there's no work queue
- Prevents multiple instances from taking conflicting actions
- If PostgreSQL is down, fall back to single-instance mode
- System remains functional, just without HA

**Never Stored in Database**:
- Which environments should exist (always from Git)
- Which environments do exist (always from providers)
- Environment configuration (always from eph.yaml in Git)
- Current environment status (always from providers)

```sql
-- Minimal schema focused on events and coordination
CREATE TABLE events (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    timestamp   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    type        TEXT NOT NULL,  -- 'env_created', 'env_deleted', 'image_found', etc
    environment TEXT NOT NULL,
    repository  TEXT NOT NULL,
    pr_number   INTEGER,
    details     JSONB,
    -- This is for analytics and debugging only
    -- Deleting this entire table would not affect Eph's operation
);

CREATE TABLE coordination (
    key         TEXT PRIMARY KEY,
    holder      TEXT NOT NULL,
    expires_at  TIMESTAMPTZ NOT NULL,
    -- Used only for leader election in multi-instance deployments
);

-- No environments table!
-- No jobs table!
-- No state machines!
-- Git + Kubernetes = Source of Truth
```

### Security Model

Eph implements security at multiple levels while maintaining developer-friendly defaults.

**API Authentication**: The system uses token-based authentication for all API calls. Developers generate personal access tokens through the web UI, which are then stored securely in their local configuration. Every CLI and API request must include a valid bearer token. The MVP implements simple token validation with basic scoping (read, write, admin), with plans to add OAuth2/OIDC support in later phases.

#### Client Authentication and Authorization

**Token-Based Authentication**:
- Developers authenticate via personal access tokens or OAuth flows
- Tokens are issued and managed by ephd daemon
- CLI stores tokens locally (encrypted at rest)
- All API requests include Bearer token authentication

**Zero Trust Architecture**:
- CLI clients are considered untrusted and have no privileged access
- All authorization decisions made server-side by ephd
- No sensitive operations can be performed without server validation
- Audit logging captures all API calls with user identity

**Token Management**:
- `eph auth login` - Interactive OAuth or token setup
- `eph auth logout` - Clear local token storage
- `eph auth status` - Show current authentication state
- Automatic token refresh for long-lived sessions

**URL Enumeration Prevention**: Rather than using predictable URLs like `app-pr-123`, Eph generates human-readable but non-guessable identifiers using a combination of adjectives, nouns, and numbers (e.g., `app-serene-ocean-42`). This approach, inspired by Heroku and similar platforms, prevents unauthorized discovery of environments while remaining memorable and shareable. The system maintains internal mappings between PR numbers and generated names, with optional redirects from PR-based aliases for convenience.

**Environment Access Control**: By default, environments are publicly accessible to support common use cases like sharing preview links with designers, product managers, or external stakeholders. For sensitive environments, developers can enable protection through their `eph.yaml` configuration:
- Basic authentication for simple password protection
- OAuth-based access control (future phase) for organization-based restrictions
- IP allowlisting for additional network-level security

**Audit Logging**: All API actions and environment lifecycle events are logged with user identity, timestamp, action details, and outcome. This provides accountability, aids in debugging, and supports compliance requirements. Logs are structured for easy querying and can be exported to external logging systems.

**Resource Isolation**: Each environment runs in its own Kubernetes namespace with appropriate RBAC policies, network policies, and resource quotas. This ensures environments cannot interfere with each other or exceed their allocated resources.

### Crash-Only Security Design

Security in Eph doesn't depend on graceful cleanup:

**Automatic Secret Rotation**: Secrets have built-in TTLs and are never stored in ephd's memory beyond immediate use. Provider credentials are re-fetched from secret stores on each reconciliation loop.

**Stateless Authentication**: All tokens are validated against external auth providers on every request. No session state means no session cleanup required.

**Resource Quotas**: Kubernetes namespaces have hard resource limits that survive ephd crashes. Even if cleanup fails, resources are bounded.

**Fail-Safe Cleanup**: The garbage collector runs continuously, not just on environment deletion. Orphaned resources are automatically detected and cleaned up.

### Trust Boundary Architecture

Eph operates on a **zero-trust client** model similar to kubectl/Kubernetes or docker/dockerd:

**Trusted Components (ephd daemon)**:
- Runs on trusted infrastructure (cluster, server, cloud)
- Direct database access (PostgreSQL)
- Provider operations (Kubernetes, Docker, etc.)
- Environment orchestration and business logic
- Webhook handling from Git providers
- Secrets and credentials management
- Configuration authority for eph.yaml files

**Untrusted Components (eph CLI)**:
- Runs on developer laptops and workstations
- **Zero database access**
- **Zero direct infrastructure access**
- **Zero business logic**
- Pure API client that calls ephd REST endpoints
- Local token storage and user preferences only

**Communication**: All client-server communication via authenticated REST API over HTTPS.

## Configuration Schema

Projects define their ephemeral environment requirements in an `eph.yaml` file at the repository root. This file describes triggers, resource requirements, and provider-specific configuration.

#### Configuration Authority

**Server Authority**: The ephd daemon is authoritative for all eph.yaml configurations:
- Validates configuration syntax and permissions
- Applies security policies and resource limits
- Resolves environment variables and secrets
- Enforces organizational constraints

**Client Role**: The eph CLI may cache configuration for user experience:
- Local preferences and defaults
- Recently used configurations for faster commands
- **Never** authoritative - always defers to server validation

```yaml
version: "1.0"
name: my-application

# Provider configuration with fallback support
providers:
  primary: kubernetes
  fallback: docker-compose

# Trigger configuration
triggers:
  # PR label triggers
  - type: pr_label
    labels: ["preview", "eph:deploy"]
    # Optional: wait for CI to complete
    wait_for_checks: ["build", "test"]

  # PR comment triggers
  - type: pr_comment
    patterns: ["/deploy", "/preview"]

  # Automatic triggers for certain branches
  - type: auto
    branches: ["feature/*", "fix/*"]
    ignore_draft: true

# Environment configuration
environment:
  # Naming and networking
  name_template: "{project}-{words}-{number}"  # e.g., "myapp-serene-ocean-42"
  subdomain_template: "{name}.{base_domain}"
  base_domain: "${EPH_BASE_DOMAIN:-preview.example.com}"

  # Human-friendly aliases that redirect to the generated name
  alias_template: "{project}-pr-{pr_number}"  # Optional PR-based redirect

  # Lifecycle management
  ttl: 72h  # Total time to live
  idle_timeout: 4h  # Scale down after inactivity
  wake_on_access: true  # Auto-wake scaled environments

  # Resource constraints
  resources:
    cpu_request: "100m"
    cpu_limit: "2"
    memory_request: "128Mi"
    memory_limit: "4Gi"

  # Environment variables available to all services
  env:
    APP_ENV: "preview"
    FEATURE_FLAGS: "preview-mode"

  # Image resolution configuration
  images:
    - name: api
      repository: ghcr.io/myorg/api

      # Primary: Look for explicit image reference in git notes
      tag_source: git_note
      git_note_ref: "eph-images"

      # Secondary: Use templated convention
      tag_template: "{ref_type}-{ref_name}-{commit_sha:0:7}"
      # Template variables:
      # - {ref_type}: "pr", "branch", or "tag"
      # - {ref_name}: "123", "main", "v1.0.0"
      # - {pr_number}: "123" (for PRs)
      # - {commit_sha}: full SHA
      # - {commit_sha:0:7}: substring syntax
      # - {branch_name}: sanitized branch name

      # Tertiary: Scan registry for matching tags
      tag_pattern: "pr-{pr_number}-*"

      # Constraints
      max_age: 7d              # Reject images older than 7 days
      required_labels:         # Image must have these labels
        - "eph.io/commit={commit_sha}"
        - "eph.io/pr={pr_number}"

      # Fallback
      fallback_tag: "latest"   # Last resort
      fallback_behavior: "fail" # or "use-fallback", "wait"

    - name: frontend
      repository: ghcr.io/myorg/frontend
      # Simpler configuration for predictable CI
      tag_template: "pr-{pr_number}"

    - name: database-migrator
      repository: ghcr.io/myorg/migrator
      # Can use specific versions for tools
      tag: "v2.1.0"  # Fixed version

# Provider-specific configurations
kubernetes:
  # Target cluster configuration
  context: "${K8S_CONTEXT}"
  namespace_template: "{project}-pr-{pr_number}"

  # Manifest sources (applied in order)
  manifests:
    - path: ./k8s/base
    - kustomization: ./k8s/overlays/preview
      patches:
        - target:
            kind: Deployment
            name: api-server
          patch: |
            - op: replace
              path: /spec/replicas
              value: 1

  # Image overrides
  images:
    - name: api-server
      newName: "{registry}/{project}/api"
      newTag: "pr-{pr_number}-{commit_sha:0:7}"

  # Image pull configuration
  imagePullSecrets:
    - name: registry-credentials

  # Common image tag patterns:
  # PR-based: "pr-{pr_number}"
  # Commit-based: "{commit_sha}" or "{commit_sha:0:7}"
  # Combined: "pr-{pr_number}-{commit_sha:0:7}"
  # Branch-based: "{branch_name}-{commit_sha:0:7}"
  #
  # Your CI must push images with these tags BEFORE Eph deploys

  # Ingress configuration
  ingress:
    class: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
      nginx.ingress.kubernetes.io/proxy-body-size: "10m"

docker-compose:
  # Compose file selection
  compose_files:
    - docker-compose.yml
    - docker-compose.preview.yml

  # Environment file
  env_file: .env.preview

  # Service scaling overrides
  scale:
    web: 1
    worker: 1

  # Note: Eph will use the 'image:' directives from your compose files
  # It will NOT execute 'build:' directives - images must exist
  # Your CI should build and push images before triggering Eph

# Database configuration
database:
  enabled: true

  # Database instances needed
  instances:
    - name: main
      type: postgres
      version: "15"

      # Template strategy
      template:
        # Options: "empty", "seed", "snapshot", "branch"
        strategy: seed

        # For seed strategy
        seed:
          # SQL scripts to run after creation
          scripts:
            - ./db/schema.sql
            - ./db/migrations/*.sql
            - ./db/seed-preview.sql

        # For snapshot strategy (future)
        # snapshot:
        #   source: "${DB_SNAPSHOT_ID}"
        #   max_age: 7d

      # Connection configuration
      connection:
        # Environment variable to inject
        env_var: DATABASE_URL
        # Database name (templated)
        database: "app_pr_{pr_number}"

# Service dependencies
services:
  # Internal services (Eph manages these)
  - name: redis
    type: internal
    image: redis:7-alpine
    persistent: false

  # External services (references to existing systems)
  - name: auth-service
    type: external
    endpoint: "${AUTH_SERVICE_URL:-https://auth.staging.example.com}"

# Secrets management
secrets:
  # Provider selection
  provider: kubernetes  # or "vault", "aws-secrets-manager"

  # Kubernetes secrets
  kubernetes:
    # Copy secrets from source namespace
    copy_from_namespace: default
    secrets:
      - app-secrets
      - database-credentials

  # # Vault configuration (alternative)
  # vault:
  #   path: "secret/data/preview/{environment_name}"
  #   role: "preview-environments"

# Security configuration
security:
  # Environment access control
  environment_access:
    # Default access level for all environments
    default: public  # or "protected"

    # Protection for specific environments
    protection:
      # Basic auth (simple password protection)
      type: none  # or "basic", "oauth" (future)

      # # For basic auth
      # basic_auth:
      #   username: preview
      #   password: "${PREVIEW_PASSWORD}"

      # # For OAuth (future)
      # oauth:
      #   provider: github
      #   allowed_orgs: ["mycompany"]
      #   allowed_teams: ["developers"]

  # URL generation strategy
  naming:
    # Use readable random names to prevent enumeration
    strategy: readable  # e.g., "serene-ocean-42"
    include_project: true  # Results in "myapp-serene-ocean-42"

# Networking configuration
networking:
  # Routing strategy
  routing:
    # Options: "subdomain", "path", "header"
    strategy: subdomain

    # For path-based routing
    # path_prefix: "/preview/{name}"

    # For header-based routing (advanced)
    # header: "X-Eph-Environment"

  # TLS configuration
  tls:
    enabled: true
    provider: cert-manager  # or "letsencrypt", "self-signed"

# Hooks for custom logic
hooks:
  # Pre-creation hooks (run before environment creation)
  pre_create:
    - name: validate-dependencies
      command: ["./scripts/check-deps.sh"]
      timeout: 30s

  # Post-creation hooks (run after environment is ready)
  post_create:
    - name: warm-cache
      command: ["./scripts/warm-cache.sh", "${environment_url}"]
      timeout: 5m

    - name: run-smoke-tests
      command: ["./scripts/smoke-test.sh", "${environment_url}"]
      timeout: 10m
      continueOnError: true

  # Pre-destroy hooks (run before environment destruction)
  pre_destroy:
    - name: backup-data
      command: ["./scripts/backup-preview-data.sh", "{environment_name}"]
      timeout: 5m

# Observability configuration
observability:
  # Metrics collection
  metrics:
    enabled: true
    # Scrape annotations for Prometheus
    annotations:
      prometheus.io/scrape: "true"
      prometheus.io/port: "9090"

  # Distributed tracing
  tracing:
    enabled: true
    # Automatic trace context propagation
    propagate_context: true

  # Log aggregation
  logs:
    # Log shipping configuration
    ship_to: "${LOG_DESTINATION:-stdout}"
    include_pod_logs: true

# Advanced features (optional)
advanced:
  # Multi-region deployment (future)
  # regions:
  #   - us-west-2
  #   - eu-west-1

  # Canary deployment support (future)
  # canary:
  #   enabled: true
  #   analysis:
  #     metrics:
  #       - name: error-rate
  #         threshold: 5
```

## The Minimal Viable Product (MVP)

The MVP focuses on delivering a production-ready Kubernetes ephemeral environment controller. This deliberate constraint allows us to build a solid foundation while proving the core architecture.

### MVP Prerequisites

The MVP assumes users have:

**Access to a Kubernetes cluster**: Whether minikube for testing, or a production cluster (EKS, GKE, AKS, etc.)

**A wildcard DNS domain they control**: For example, `*.preview.company.com` pointing to their ingress controller

**A Git repository with Kubernetes manifests**: The application should already be deployable to Kubernetes

**Container images built by CI**: Eph expects images to exist in a registry. Your CI pipeline should build and push images tagged with the PR number or commit SHA

**A container registry**: Docker Hub, GitHub Container Registry, ECR, etc. that your cluster can pull from

Eph does not provision infrastructure or build images - it orchestrates environments using existing resources.

### Typical Workflow

The expected workflow combines CI/CD for building with Eph for deployment:

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant Git as GitHub
    participant CI as CI/CD System
    participant Reg as Container Registry
    participant Eph as Eph
    participant K8s as Kubernetes

    Dev->>Git: Open PR
    Git->>CI: Trigger build
    CI->>CI: Build & test code
    CI->>Reg: Push image (api:pr-123-abc1234)
    CI->>Git: ✅ Build successful

    Dev->>Git: Add 'preview' label
    Git->>Eph: Webhook: PR labeled
    Eph->>Reg: Verify image exists
    Eph->>K8s: Deploy with image api:pr-123-abc1234
    K8s->>Reg: Pull image
    Eph->>Git: Comment: Environment ready!
```

This separation of concerns keeps each tool focused on what it does best: CI builds and tests, Eph deploys and manages environments.

### Reconciliation is Primary, Webhooks are Optimization

While the diagram above shows the webhook flow for clarity, it's important to understand that **reconciliation is the primary mechanism**:

1. **Without webhooks**: Eph would still create the environment within 30 seconds when the reconciliation loop runs
2. **With webhooks**: The environment might be created within seconds as webhooks trigger immediate reconciliation
3. **If webhook fails**: No problem - the next reconciliation loop catches it
4. **If multiple webhooks fire**: No problem - reconciliation is idempotent

The reconciliation loop ensures eventual consistency regardless of webhook delivery.

### CI/CD Integration Requirements

For Eph to successfully deploy environments, CI systems must:

1. **Build and push images before triggering Eph**
2. **Use predictable tagging conventions OR communicate image locations**
3. **Include metadata in images for validation**

#### Recommended CI Patterns

**GitHub Actions Example**:
```yaml
name: Build for Eph
on:
  pull_request:
    types: [opened, synchronize, labeled]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build and Push
        run: |
          IMAGE="ghcr.io/${{ github.repository }}:pr-${{ github.event.pull_request.number }}-${GITHUB_SHA:0:7}"

          docker build \
            --label "eph.io/pr=${{ github.event.pull_request.number }}" \
            --label "eph.io/commit=${{ github.sha }}" \
            --label "eph.io/built-at=$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
            -t $IMAGE .

          docker push $IMAGE

      - name: Notify Eph
        run: |
          # Option 1: Git notes
          git notes add -m "eph-image: $IMAGE"
          git push origin refs/notes/*

          # Option 2: PR comment (if using PR comments for triggering)
          gh pr comment ${{ github.event.pull_request.number }} \
            --body "/eph deploy --image=$IMAGE"
```

**GitLab CI Example**:
```yaml
build-for-eph:
  stage: build
  only:
    - merge_requests
  script:
    # Build with metadata
    - |
      docker build \
        --label "eph.io/mr=$CI_MERGE_REQUEST_IID" \
        --label "eph.io/commit=$CI_COMMIT_SHA" \
        -t $CI_REGISTRY_IMAGE:mr-$CI_MERGE_REQUEST_IID-$CI_COMMIT_SHORT_SHA .

    - docker push $CI_REGISTRY_IMAGE:mr-$CI_MERGE_REQUEST_IID-$CI_COMMIT_SHORT_SHA

    # Also tag with predictable name for Eph
    - docker tag $CI_REGISTRY_IMAGE:mr-$CI_MERGE_REQUEST_IID-$CI_COMMIT_SHORT_SHA \
                 $CI_REGISTRY_IMAGE:mr-$CI_MERGE_REQUEST_IID
    - docker push $CI_REGISTRY_IMAGE:mr-$CI_MERGE_REQUEST_IID
```

### Git as the Single Source of Truth for Triggers

Eph uses Git refs (PR labels, branches, tags) as the **only** mechanism for triggering environment creation:

```yaml
# Examples of Git-based triggers
triggers:
  # PR with label
  - type: pr_label
    labels: ["preview", "demo"]

  # Git branches matching pattern
  - type: git_branch
    pattern: "preview/*"

  # Git tags matching pattern
  - type: git_tag
    pattern: "demo/*"
```

**What Git changes trigger:**
- Add "preview" label to PR → Environment should be created
- Remove "preview" label from PR → Environment should be destroyed
- Push tag "demo/v1" → Environment should be created
- Delete tag "demo/v1" → Environment should be destroyed

**What does NOT trigger environment creation:**
- CI building an image
- Webhook from CI/CD system
- Image appearing in registry
- Manual API calls (unless they modify Git)

This ensures a single, auditable source of truth for all environments.

### MVP Scope and Features

The MVP implements these core capabilities:

**GitHub Pull Request Integration**: The system responds to webhooks from GitHub when pull requests are created or updated. Developers add a "preview" label to their PR, triggering environment creation. The system posts status updates back to the PR as comments, providing the environment URL when ready.

**Kubernetes Environment Provisioning**: The MVP includes a built-in Kubernetes provider that creates isolated namespaces for each environment. It applies Kubernetes manifests from the repository, supports Kustomize overlays for environment-specific modifications, and handles image tag substitution to deploy the PR-specific build.

**Image Resolution Logic**: The MVP implements a hybrid image resolution system that attempts multiple strategies to find the correct image for an environment. It validates image existence, age, and metadata before deployment. If no valid image is found, the environment enters a "WaitingForImage" state and retries on the next reconciliation loop.

**DNS and Routing**: Environments are accessible via wildcard DNS subdomains (e.g., `myapp-pr-123.preview.company.com`). The system creates appropriate Ingress resources in Kubernetes and manages DNS records via Route53 or Cloudflare APIs.

**Command-Line Interface**: A simple but powerful CLI provides essential commands:
- `eph init` - Initialize a new project with a template `eph.yaml`
- `eph auth login` - Authenticate with personal access token
- `eph up` - Manually create an environment (useful for testing)
- `eph down` - Destroy an environment
- `eph list` - Show all active environments
- `eph logs` - Stream logs from an environment
- `eph exec` - Execute commands in an environment
- `eph wtf` - Diagnostic command that shows detailed status and common issues

**CLI Command Implementation**: All eph commands are API clients:
- `eph up` → `POST /api/v1/environments`
- `eph down {env}` → `DELETE /api/v1/environments/{env}`
- `eph list` → `GET /api/v1/environments`
- `eph logs {env}` → `GET /api/v1/environments/{env}/logs`
- `eph status {env}` → `GET /api/v1/environments/{env}/status`

The CLI provides user-friendly interfaces and output formatting but contains zero business logic.

**Web Dashboard**: A basic web interface shows all active environments, their status, resource usage, and recent events. Developers can view logs, destroy environments, see configuration details, and generate personal access tokens for CLI authentication.

### MVP Architecture Decisions

The MVP makes several pragmatic choices to reduce complexity:

**Monolithic Core**: Instead of microservices, the MVP uses a single Go binary containing the API server, event processor, and orchestrator. This simplifies deployment and debugging while maintaining clean internal module boundaries.

**Built-in Kubernetes Provider**: Rather than implementing the full gRPC provider interface, the Kubernetes provider is compiled into the core binary. This eliminates inter-process communication overhead and simplifies the initial implementation.

**PostgreSQL for Logging Only**: The MVP uses PostgreSQL purely for event logging and audit trails. There's no job queue (reconciliation handles this), no environment state table (Git and Kubernetes are authoritative), and no critical dependency on the database. If PostgreSQL is down, Eph continues to function normally, just without audit logs.

**Simplified Database Handling**: Environments either connect to existing databases using isolated schemas or run ephemeral PostgreSQL containers within the environment. No snapshotting or advanced templating in the MVP.

**Basic Security Model**: API authentication uses simple bearer tokens generated through the web UI. Environment URLs use readable but non-guessable names (e.g., `gentle-stream-42`) to prevent enumeration while remaining user-friendly. Environment access is public by default to support sharing with stakeholders.

**Informer-Based Architecture**: The MVP implements simplified informers that poll GitHub every 30 seconds and watch Kubernetes continuously. This provides the same reliability benefits as full Kubernetes-style informers while keeping the implementation straightforward.

**No Job Queue**: Unlike traditional architectures, there's no job queue or state machine. The reconciliation loop naturally handles all state transitions by continuously converging toward the desired state.

**Crash-Only by Default**: The MVP assumes it can be killed at any time. All operations are idempotent, all state is external, and startup always begins with a full reconciliation.

### MVP Reconciliation Flow

The MVP implements a pure reconciliation model without job queues or complex state machines:

```mermaid
sequenceDiagram
    participant Dev as Developer
    participant GH as GitHub
    participant Rec as Reconciler (every 30s)
    participant K8s as Kubernetes
    participant Reg as Registry

    Dev->>GH: Add 'preview' label to PR
    GH->>Rec: Webhook (optional optimization)

    Note over Rec: Next reconciliation cycle

    loop Every 30 seconds
        Rec->>GH: List PRs with 'preview' label
        GH-->>Rec: PR #123, #124, #126

        Rec->>K8s: List managed namespaces
        K8s-->>Rec: ns-123, ns-124 (missing ns-126)

        Rec->>Rec: Compute diff
        Note over Rec: Need to create env for PR #126

        Rec->>GH: Fetch eph.yaml for PR #126
        Rec->>Reg: Check if image exists

        alt Image exists
            Rec->>K8s: Create namespace, deploy
            Rec->>GH: Comment with URL
        else Image not ready
            Rec->>Rec: Mark as "WaitingForImage"
            Note over Rec: Will retry next cycle
        end
    end
```

Key points:
- No job queue - just a simple reconciliation loop
- No database state tracking - Git and Kubernetes are the sources of truth
- Webhooks only trigger faster reconciliation, they don't carry state
- Every operation is idempotent and safe to retry

The implementation centers on the reconciliation loop:
- Informers in `internal/informers` maintain cached views of GitHub and Kubernetes
- Reconciler in `internal/reconciler` computes and applies differences
- Providers in `internal/providers` execute idempotent infrastructure operations
- Webhooks in `internal/webhook` simply trigger reconciliation sooner

### Reconciliation-First Implementation

The MVP implements a simplified but production-ready reconciliation loop:

```go
// Simplified MVP reconciliation
func (r *Reconciler) Reconcile(ctx context.Context) error {
    // 1. List all PRs with 'preview' label
    prs, err := r.github.ListPullRequests(ctx, "preview")
    if err != nil {
        return fmt.Errorf("list PRs: %w", err)
    }

    // 2. List all Kubernetes namespaces we manage
    namespaces, err := r.k8s.ListNamespaces(ctx, "eph.io/managed=true")
    if err != nil {
        return fmt.Errorf("list namespaces: %w", err)
    }

    // 3. Build lookup maps
    desired := make(map[string]*PullRequest)
    for _, pr := range prs {
        name := GenerateEnvironmentName(pr)
        desired[name] = pr
    }

    actual := make(map[string]*Namespace)
    for _, ns := range namespaces {
        actual[ns.Name] = ns
    }

    // 4. Create missing environments
    for name, pr := range desired {
        if _, exists := actual[name]; !exists {
            r.createEnvironment(ctx, pr)
        }
    }

    // 5. Delete orphaned environments
    for name := range actual {
        if _, exists := desired[name]; !exists {
            r.deleteEnvironment(ctx, name)
        }
    }

    return nil
}
```

This simple loop handles all edge cases through continuous reconciliation.

### MVP Code Structure

The MVP follows the idiomatic Go project structure described in the "Go Project Structure and Code Organization" section above, with a clean separation between layers:

```
eph/
├── cmd/
│   ├── eph/              # CLI binary entry point
│   └── ephd/             # Server daemon entry point
├── internal/             # Private application code
│   ├── server/           # HTTP server implementation
│   ├── cli/              # CLI command implementation
│   ├── api/              # HTTP API client/server shared code
│   ├── config/           # Configuration parsing and validation
│   ├── controller/       # Environment business logic and orchestration
│   ├── reconciler/       # Core reconciliation loop engine
│   ├── informers/        # GitHub and Kubernetes informers (cache external state)
│   ├── log/              # Logging utilities
│   ├── providers/        # Provider implementations
│   │   ├── interface.go  # Provider interface
│   │   └── kubernetes/   # Kubernetes provider
│   ├── state/            # Event logging (not authoritative state)
│   ├── webhook/          # Git webhook handlers
│   └── worker/           # Background reconciliation loops
├── pkg/
│   └── version/          # Exportable version information
├── web/                  # Web dashboard (React)
├── configs/              # Example configurations
├── migrations/           # Database migrations
└── docs/                 # Documentation
```

This structure provides:

1. Clean separation between API handlers, business logic, and infrastructure code
2. Proper testability of all core components
3. Clear boundaries between the CLI client and server daemon
4. Adherence to Go community standards for package organization

The web dashboard (React application) and documentation exist alongside the Go code but follow their own best practices for organization.

### Dogfooding Strategy

Once the MVP is functional, the Eph project itself will use Eph for testing pull requests. This provides real-world validation and rapid feedback:

```yaml
# eph.yaml in the Eph repository
version: "1.0"
name: eph-controller

provider: kubernetes
triggers:
  - type: pr_label
    labels: ["preview", "test-deployment"]

environment:
  name: "eph-{words}-{number}"  # e.g., "eph-mystic-canyon-7"
  domain: "{name}.dogfood.eph.dev"

# Ensure our CI builds images first
triggers:
  - type: pr_label
    labels: ["preview", "test-deployment"]
    wait_for_checks: ["docker-build"]

kubernetes:
  manifests:
    - ./deploy/eph-controller.yaml
    - ./deploy/eph-webhook-handler.yaml

database:
  instances:
    - name: state
      type: postgres
      version: "15"
      template:
        strategy: seed
        seed:
          scripts:
            - ./deploy/test-schema.sql
```

This approach ensures that every Eph PR is tested in a real Eph-managed environment, providing confidence in changes before merging.

The Eph project will particularly benefit from testing crash recovery:

```yaml
# Additional dogfooding tests
testing:
  chaos:
    # Randomly kill ephd during reconciliation
    - name: chaos-monkey
      schedule: "*/10 * * * *"  # Every 10 minutes
      action: "kubectl delete pod -l app=ephd --random"

    # Verify environments remain consistent
    - name: consistency-check
      schedule: "*/5 * * * *"   # Every 5 minutes
      action: "./scripts/verify-environments.sh"
```

This ensures our crash-only design actually works under real conditions.

### Internal Testing Configuration

The Eph team will use a special configuration for testing Eph itself:

```yaml
# eph-internal.yaml - used only by Eph team
version: "1.0"
name: eph-controller-test

provider: kubernetes
triggers:
  - type: pr_label
    labels: ["test-eph"]

environment:
  name: "eph-{words}-{number}"  # e.g., "eph-bold-summit-23"
  # Preview instances get their own test subdomain
  domain: "{name}.test.eph.dev"

kubernetes:
  namespace: "eph-test-{words}-{number}"
  manifests:
    - ./deploy/eph-controller.yaml

  # Preview instances get limited credentials
  secrets:
    - name: eph-test-credentials
      data:
        # Can only create *.test.eph.dev domains
        DNS_DOMAIN_PATTERN: "*.test.eph.dev"
        # Can only use eph-test-cluster
        K8S_CLUSTER: "eph-test-cluster"
        # Can only create in eph-preview-* namespaces
        K8S_NAMESPACE_PREFIX: "eph-preview-"

# Run integration tests after deployment
hooks:
  post_create:
    - name: integration-tests
      command: ["./scripts/test-preview-instance.sh"]
      timeout: 15m
```

This keeps the complexity isolated to the Eph team's workflow without affecting the general architecture. The preview instances have limited permissions and can only create resources in test subdomains and namespaces, ensuring security while enabling true end-to-end testing of the PR's code.

## Feature Development Roadmap

### Phase 1: Production Hardening

After the MVP proves the core concept, the focus shifts to reliability and security features necessary for production use.

**Enhanced Security Model**: Implement proper authentication using OAuth2/OIDC providers, allowing teams to use their existing identity systems. Add role-based access control (RBAC) to control who can create, view, and destroy environments. Implement audit logging for all actions, storing who did what and when. Add support for encrypting sensitive configuration values.

**CI/CD Integration**: Add support for waiting on CI status checks before deploying, ensuring images are built and tests pass. Implement webhooks to notify CI systems when environments are ready for integration testing. Add image existence validation to fail fast if images aren't available. Support for multiple container registries with different authentication methods.

**Operational Excellence**: Integrate with Prometheus for detailed metrics about environment creation times, resource usage, and system health. Structure all logs with consistent fields for easy searching and debugging. Implement circuit breakers for external service calls to prevent cascade failures. Add comprehensive health check endpoints for monitoring systems.

**Advanced Kubernetes Features**: Support for Helm charts in addition to raw manifests, allowing teams to use their existing Helm-based deployments. Implement proper resource quotas and limit ranges to prevent runaway resource consumption. Add support for persistent volume claims for stateful workloads. Enable horizontal pod autoscaling for environments under load.

**Database Seeding Framework**: Develop a robust system for initializing databases with test data. Support multiple seeding strategies including SQL scripts, programmatic seeders, and partial production data copies. Implement automatic migration running for applications that manage their own schemas. Add database connection pooling and management.

### Phase 2: Provider Ecosystem

With a solid foundation, expand beyond Kubernetes to support other deployment targets.

**gRPC Provider Interface**: Implement the full gRPC-based provider plugin system. Create comprehensive documentation and examples for provider developers. Build a test harness for validating provider implementations. Develop a provider SDK in Go to simplify common operations.

**Docker Compose Provider**: Build a provider that can deploy Docker Compose applications to Docker hosts or Swarm clusters. This provider will execute `docker-compose up` using the image references in your compose files, but will NOT execute build directives. Handle volume management and inter-container networking. Support Compose-specific features like healthchecks and depends_on. Implement log aggregation from multiple containers. Note that users must ensure their compose files reference images that exist in accessible registries.

**Cloud Provider Integrations**: Develop an AWS ECS/Fargate provider for teams using Amazon's container services. Create a Google Cloud Run provider for serverless container deployments. Build an Azure Container Instances provider. Each should leverage cloud-native features like IAM roles and managed databases.

**Provider Registry**: Create a central registry where community members can publish and discover providers. Implement provider versioning and compatibility checking. Add automated testing for published providers. Build a CLI plugin manager for easy provider installation.

### Phase 3: Advanced Routing and Networking

Sophisticated routing capabilities enable more complex testing scenarios.

**Service Mesh Integration**: Build deep integration with Istio for advanced traffic management. Support canary deployments within ephemeral environments. Enable distributed tracing across services. Implement mutual TLS between services.

**Header-Based Routing**: Develop a routing system similar to Signadot's sandboxes, where specific services can be overridden while sharing others. Support routing rules based on HTTP headers, allowing multiple versions to coexist. Build browser extensions and CLI tools to simplify header injection. Create SDKs for propagating routing context.

**Multi-Cluster Support**: Enable environments that span multiple Kubernetes clusters. Implement cross-cluster service discovery. Handle network connectivity between clusters. Support geo-distributed environments for testing latency-sensitive applications.

### Phase 4: Intelligent Resource Management

Move beyond simple rules to intelligent optimization of resources.

**Predictive Scaling**: Analyze historical usage patterns to predict when environments will be active. Pre-warm environments before developers typically start work. Scale down proactively during predicted idle periods. Learn from individual developer and team patterns.

**Cost Analytics**: Track detailed resource consumption per environment, team, and project. Provide cost breakdowns by compute, storage, and network. Generate recommendations for optimizing resource requests. Alert on unusual spending patterns.

**Smart Cleanup**: Implement intelligent TTL adjustment based on PR activity and developer behavior. Detect abandoned environments through Git activity analysis. Archive environment state before destruction for debugging. Provide easy environment resurrection from archives.

### Phase 5: Developer Experience Enhancements

Polish the developer experience to make Eph a joy to use.

**IDE Integrations**: Build extensions for VS Code, IntelliJ, and other popular IDEs. Show environment status directly in the IDE. Enable one-click environment creation and access. Support remote development in ephemeral environments.

**Local Development Bridge**: Create tools that connect local development to cloud ephemeral environments. Port forwarding to access remote services locally. Bidirectional file synchronization for hot reload. Environment variable and secret injection. Similar to Telepresence or Okteto's synchronization, this bridges the gap between local development and cloud environments without trying to replace local tools.

**GitOps Integration**: Create Kubernetes operators that reconcile Eph environments as custom resources. Support ArgoCD ApplicationSets for environment definitions. Enable Flux integration for GitOps workflows. Build Terraform providers for infrastructure-as-code.

**Collaborative Features**: Add real-time collaboration features to environments. Implement shared debugging sessions. Build comment and annotation systems. Create environment snapshots for sharing specific states.

**Advanced CLI Features**: Add shell completions and interactive prompts. Implement environment templates and blueprints. Build a TUI (terminal UI) for managing multiple environments. Create workflow automation commands.

## Reliability Patterns

Eph incorporates several reliability patterns proven in production systems:

**1. Idempotent Everything**
```go
// Safe to call multiple times
func CreateEnvironment(env Environment) error {
    if exists(env) {
        return UpdateEnvironment(env)
    }
    return actuallyCreate(env)
}
```

**2. Crash-Only Components**
- No graceful shutdown required for correctness
- All state is external or reconstructible
- Recovery is just normal startup

**3. Level-Triggered Reconciliation**
- Don't track what changed, observe current state
- Don't queue events, compute needed actions
- Don't save progress, operations are atomic

**4. Fast Failure Recovery**
```go
func (s *Server) Run() error {
    // Fail fast on startup
    if err := s.validateProviderAccess(); err != nil {
        return fmt.Errorf("provider access check: %w", err)
    }

    // But once running, never give up
    for {
        if err := s.reconciler.Reconcile(ctx); err != nil {
            log.Error("Reconciliation failed, will retry", "error", err)
            time.Sleep(time.Second) // Brief backoff
            continue
        }
        time.Sleep(30 * time.Second) // Normal interval
    }
}
```

**5. Continuous Garbage Collection**
- Don't wait for delete events
- Scan for orphaned resources periodically
- Clean up expired environments automatically

```go
// Garbage collection is part of normal reconciliation
func (gc *GarbageCollector) CollectGarbage(ctx context.Context) error {
    // 1. Find environments past their TTL
    environments := gc.k8s.ListEnvironments(ctx)
    for _, env := range environments {
        if env.Age() > env.TTL {
            gc.markForDeletion(env)
        }
    }

    // 2. Find environments with no corresponding Git ref
    gitRefs := gc.github.ListAllRefs(ctx)
    for _, env := range environments {
        if !gitRefs.Contains(env.SourceRef) {
            // Git ref was deleted
            gc.markForDeletion(env)
        }
    }

    // 3. Clean up orphaned resources
    // This handles cases where deletion partially failed
    gc.cleanupOrphanedResources(ctx)

    return nil
}
```

**Key principle**: Garbage collection doesn't rely on delete events. It continuously ensures the actual state matches the desired state, including cleaning up environments that shouldn't exist.

**6. Observability First**
- Reconciliation metrics (duration, errors, actions taken)
- Cache hit rates and sync latency
- Provider API call rates and errors
- Clear logs showing reconciliation decisions

**7. Image Availability Resilience**
```go
// Images might not be ready immediately after PR creation
func (r *Reconciler) handleImageNotReady(env Environment) error {
    age := time.Since(env.CreatedAt)

    if age < 10*time.Minute {
        // Young environments wait for CI
        env.Status = "WaitingForImage"
        return nil  // Retry next reconciliation
    }

    if age < 30*time.Minute && env.Config.FallbackTag != "" {
        // Fall back after reasonable wait
        env.Status = "UsingFallbackImage"
        env.UseImage(env.Config.FallbackTag)
        return nil
    }

    // Old environments with no image are marked failed
    env.Status = "ImageNotFound"
    env.Error = "CI may have failed to build image"
    return nil
}
```

### Reconciliation Timing and Efficiency

The reconciliation loop timing balances responsiveness with efficiency:

**Default Timing**:
- GitHub polling: Every 30 seconds
- Kubernetes watch: Continuous (event-driven)
- Full reconciliation: Every 30 seconds
- Webhook-triggered: Immediate (sub-second)

**Optimization Strategies**:
```go
// Intelligent polling based on activity
func (r *Reconciler) determineNextSync() time.Duration {
    if r.recentActivity() {
        return 10 * time.Second  // More frequent during active development
    }
    if r.isWeekend() {
        return 2 * time.Minute   // Less frequent on weekends
    }
    return 30 * time.Second      // Default interval
}
```

**Preventing Thundering Herd**:
- Multiple ephd instances use jittered start times
- Reconciliation for different repos can be sharded
- Rate limiting on external API calls
- Caching reduces redundant API requests

### Common Misconceptions About Eph's Architecture

**"CI triggers Eph to create environments"**
- **Wrong**: CI never triggers Eph. Git (via PR labels/tags/branches) is the only trigger.
- **Right**: CI builds images. Eph watches Git, waits for images, then deploys.

**"Eph stores environment state in PostgreSQL"**
- **Wrong**: PostgreSQL never stores which environments exist or their status.
- **Right**: Git defines what should exist, Kubernetes reports what does exist. PostgreSQL only stores event logs.

**"Webhooks are required for Eph to work"**
- **Wrong**: Webhooks are purely an optimization for responsiveness.
- **Right**: The reconciliation loop runs every 30 seconds regardless. Webhooks just make it run sooner.

**"Failed webhook delivery will cause missed environments"**
- **Wrong**: Missed webhooks have no impact on correctness.
- **Right**: The next reconciliation loop (within 30s) will detect and correct any differences.

**"Eph needs a job queue to handle concurrent requests"**
- **Wrong**: Traditional job queues add complexity and failure modes.
- **Right**: The reconciliation pattern naturally handles all concurrency through idempotent operations.

## Technical Design Decisions

### Why gRPC for Provider Plugins?

The choice of gRPC over alternatives like REST APIs or native Go plugins comes from several technical requirements. Providers need to stream logs and progress updates in real-time, which gRPC handles elegantly with its streaming RPC support. The strongly-typed protocol buffer definitions ensure compatibility between core and plugins. Language agnosticism allows provider developers to use their preferred languages - a Rust developer can build a provider without learning Go. Process isolation means a buggy provider can't crash the core system. The gRPC ecosystem provides excellent tooling for testing, debugging, and monitoring.

Additionally, the gRPC interface aligns perfectly with reconciliation-based design. Providers don't need to maintain state - they simply report current state and execute idempotent operations. The streaming interface allows providers to send progress updates during long-running operations without blocking the reconciliation loop.

### PostgreSQL as Event Log and Coordinator

PostgreSQL in Eph serves a fundamentally different role than in traditional architectures. Rather than storing authoritative state, it provides:

**Event Streaming**: An append-only log of all actions taken, enabling audit trails, analytics, and debugging without being required for correctness.

**Work Coordination**: When running multiple ephd instances, PostgreSQL provides simple distributed locking to prevent duplicate work without complex consensus protocols.

**Metrics Storage**: Time-series data about environment usage, creation times, and resource consumption for capacity planning.

The key insight: if PostgreSQL is unavailable, Eph continues to function correctly, just without audit logs or multi-instance coordination. The reconciliation loop ensures environments remain synchronized with Git regardless.

### Kubernetes-First Strategy

Starting with Kubernetes as the primary provider isn't just about market share. Kubernetes represents the most complex deployment target, exercising all aspects of the provider interface. If the abstraction works for Kubernetes, simpler providers like Docker Compose become straightforward. The Kubernetes ecosystem provides solutions for many challenges: Ingress controllers for routing, cert-manager for TLS, persistent volumes for storage. The declarative model aligns well with ephemeral environments. Most importantly, solving Kubernetes well captures the majority of the cloud-native market.

### Event-Driven Architecture

The event-driven design provides several benefits for ephemeral environment management. Webhook events can be acknowledged quickly, providing good user experience even when provisioning is slow. The system can scale by adding more workers without architectural changes. Failed operations can be retried automatically with exponential backoff. The event log provides a complete audit trail of all operations. Components remain loosely coupled, making testing and evolution easier.

### Image Resolution Without Building

Eph's design principle of not building images requires a sophisticated resolution system. Rather than prescribing a single CI integration pattern, Eph supports multiple discovery methods that work with any CI system. This flexibility is achieved through:

1. **Template-based conventions** that map Git refs to image tags
2. **Git-native communication** via notes and tags
3. **Registry introspection** to find matching images
4. **Validation logic** to ensure images are recent and correct

The reconciliation loop naturally handles the async nature of CI - if an image isn't ready, the environment simply waits until the next reconciliation finds it.

## Open Source Philosophy

Eph embraces open source principles from the ground up. The project uses the Apache 2.0 license to encourage both community and commercial adoption. All development happens in the open on GitHub, with public roadmaps and design discussions. The core team commits to reviewing community pull requests within 48 hours and maintaining comprehensive documentation for contributors.

The provider plugin system is specifically designed to foster community contributions. By using gRPC instead of language-specific plugins, we lower the barrier for contributors who might not know Go. The provider interface is versioned and stable, ensuring community providers continue working as Eph evolves.

Community engagement happens through multiple channels. GitHub Discussions for design proposals and questions, Discord for real-time chat and support, monthly community calls for updates and demos, and a public roadmap for transparency about future development.

The project maintains clear contribution guidelines covering code standards, testing requirements, and the review process. New contributors are welcomed with good first issues and mentorship. Regular contributors can join the core team based on sustained engagement.

## Conclusion

Eph represents a pragmatic approach to ephemeral environment management. By starting with a focused MVP that solves real problems for Kubernetes users, we can validate the architecture while delivering immediate value. The extensible provider system ensures Eph can grow to support any infrastructure, while the emphasis on developer experience makes it a tool developers actually want to use.

The open source model ensures Eph remains community-driven, with development priorities set by real user needs rather than commercial pressures. As the ecosystem grows, Eph has the potential to become the standard interface for ephemeral environments across the industry.

*When in doubt, just eph it!*
