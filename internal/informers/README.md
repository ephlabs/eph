# Internal Informers Package

This package handles external state caching and watching for the Eph application.

## Purpose

The informers package implements caches for external state from:
- GitHub API (PRs, labels, branches)
- Kubernetes API (namespaces, deployments, services)

## Responsibilities

- Cache current state from external systems
- Provide efficient access to external state without constant API calls
- Detect changes and trigger reconciliation events
- Follow Kubernetes informer pattern for consistency

## Architecture

This follows the Kubernetes informer pattern where:
- Informers watch external APIs for changes
- They maintain local caches of current state
- Controllers can query the cache instead of hitting APIs directly
- Change events trigger reconciliation loops

## Implementation Status

🚧 **Placeholder** - Not yet implemented
